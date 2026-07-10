package proxy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gateway/internal/alert"
	"gateway/internal/cache"
	"gateway/internal/model"
	"gateway/internal/storage"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const (
	gatewayKeyContextKey contextKey = "gatewayKey"
	teamIDContextKey     contextKey = "teamID"
)

type ProxyServer struct {
	storage    storage.Storage
	cache      cache.Cache
	registry   *TranslatorRegistry
	httpClient *http.Client
}

func NewProxyServer(s storage.Storage, c cache.Cache) *ProxyServer {
	return &ProxyServer{
		storage:    s,
		cache:      c,
		registry:   NewTranslatorRegistry(),
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (ps *ProxyServer) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error": "Invalid Authorization header format"}`, http.StatusUnauthorized)
			return
		}

		apiKey := parts[1]
		keyHash := HashKey(apiKey)

		ctx := r.Context()

		// Try cache first
		gKey, found, err := ps.cache.GetGatewayKey(ctx, keyHash)
		if err != nil {
			// Log and fallback to storage
			gKey = nil
			found = false
		}

		if !found {
			// Try storage
			gKey, err = ps.storage.GetGatewayKey(ctx, keyHash)
			if err != nil {
				http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
				return
			}
			if gKey == nil {
				http.Error(w, `{"error": "Unauthorized: invalid gateway key"}`, http.StatusUnauthorized)
				return
			}
			// Write back to cache
			_ = ps.cache.SetGatewayKey(ctx, keyHash, gKey)
		}

		if gKey.Status != "active" {
			http.Error(w, `{"error": "Unauthorized: gateway key is suspended or revoked"}`, http.StatusUnauthorized)
			return
		}

		// Inject to context
		ctx = context.WithValue(ctx, gatewayKeyContextKey, gKey)
		ctx = context.WithValue(ctx, teamIDContextKey, gKey.TeamID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (ps *ProxyServer) BudgetMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		teamID, ok := ctx.Value(teamIDContextKey).(string)
		if !ok {
			http.Error(w, `{"error": "Internal server error: team context missing"}`, http.StatusInternalServerError)
			return
		}

		// Try cache first
		limit, used, found, err := ps.cache.GetTeamBudget(ctx, teamID)
		if err != nil {
			found = false
		}

		if !found {
			// Query storage
			team, err := ps.storage.GetTeam(ctx, teamID)
			if err != nil {
				http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
				return
			}
			if team == nil {
				http.Error(w, `{"error": "Team not found"}`, http.StatusNotFound)
				return
			}
			limit = team.BudgetLimit
			used = team.BudgetUsed
			// Set cache
			_ = ps.cache.SetTeamBudget(ctx, teamID, limit, used)
		}

		if used >= limit {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired) // 402
			_, _ = w.Write([]byte(`{"error": {"message": "Budget limit exceeded for team. Please top up.", "code": "budget_limit_exceeded"}}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (ps *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.URL.Path == "/v1/chat/completions" {
		ps.HandleChatCompletions(w, r)
		return
	}
	http.Error(w, `{"error": "Not Found"}`, http.StatusNotFound)
}

func (ps *ProxyServer) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamID := ctx.Value(teamIDContextKey).(string)

	shadowModel := r.Header.Get("X-Shadow-Model")
	if shadowModel == "" {
		shadowModel = r.Header.Get("x-shadow-model")
	}

	// Read and parse original request
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}

	var openAIReq ChatCompletionRequest
	if err := json.Unmarshal(bodyBytes, &openAIReq); err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Invalid request JSON: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	// Local PII & Secrets Redaction Guard
	piiRedactedCount := 0
	mergedPlaceholderMap := make(map[string]string)
	for i, msg := range openAIReq.Messages {
		scrubRes := ScrubText(msg.Content)
		if scrubRes.RedactedCount > 0 {
			openAIReq.Messages[i].Content = scrubRes.Text
			piiRedactedCount += scrubRes.RedactedCount
			for placeholder, val := range scrubRes.PlaceholderMap {
				mergedPlaceholderMap[placeholder] = val
			}
		}
	}

	// Resolve provider
	translator, providerName := ps.registry.ResolveByModel(openAIReq.Model)

	// Fetch provider configuration (API Key)
	providerConf, err := ps.storage.GetProviderConfig(ctx, providerName)
	if err != nil {
		http.Error(w, `{"error": "Internal server error reading provider config"}`, http.StatusInternalServerError)
		return
	}
	if providerConf == nil || providerConf.APIKey == "" {
		http.Error(w, fmt.Sprintf(`{"error": "Provider '%s' not configured"}`, providerName), http.StatusInternalServerError)
		return
	}

	// Translate request (will use the redacted openAIReq!)
	translatedBody, targetURL, headers, err := translator.TranslateRequest(ctx, &openAIReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Translation error: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Create request to target provider
	outReq, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(translatedBody))
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to create downstream request: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	// Apply standard headers and credentials
	for k, v := range headers {
		outReq.Header.Set(k, v)
	}

	if providerName == "openai" {
		outReq.Header.Set("Authorization", "Bearer "+providerConf.APIKey)
	} else if providerName == "anthropic" {
		outReq.Header.Set("x-api-key", providerConf.APIKey)
	}

	primaryLogID := uuid.New().String()
	startTime := time.Now()
	resp, err := ps.httpClient.Do(outReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Downstream request failed: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if openAIReq.Stream {
		ps.handleStreamResponse(w, resp, translator, openAIReq.Model, teamID, startTime, primaryLogID, shadowModel, &openAIReq, piiRedactedCount, mergedPlaceholderMap)
	} else {
		ps.handleStandardResponse(w, resp, translator, openAIReq.Model, teamID, startTime, primaryLogID, shadowModel, &openAIReq, piiRedactedCount, mergedPlaceholderMap)
	}
}

func (ps *ProxyServer) handleStandardResponse(w http.ResponseWriter, resp *http.Response, translator ProviderTranslator, modelName string, teamID string, startTime time.Time, primaryLogID string, shadowModel string, openAIReq *ChatCompletionRequest, piiRedactedCount int, mergedPlaceholderMap map[string]string) {
	ctx := context.Background() // Use background to complete logging even if request cancels at final millisecond

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read downstream response"}`, http.StatusBadGateway)
		return
	}

	// Translate back to OpenAI format
	openAIResp, err := translator.TranslateResponse(ctx, resp.StatusCode, respBytes, modelName)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(respBytes) // Return downstream response as fallback if translation fails
		return
	}

	// Re-inject original sensitive data back into response content (un-redaction)
	for i, choice := range openAIResp.Choices {
		if choice.Message.Content != "" {
			openAIResp.Choices[i].Message.Content = UnscrubText(choice.Message.Content, mergedPlaceholderMap)
		}
	}

	latency := time.Since(startTime).Milliseconds()
	cost := CalculateCost(modelName, openAIResp.Usage.PromptTokens, openAIResp.Usage.CompletionTokens)

	// Perform budget deduction & usage logging asynchronously or synchronously
	// Synchronous here to prevent race conditions on budget limits
	newUsed, _ := ps.cache.IncrementTeamBudgetUsed(ctx, teamID, cost)
	_ = ps.storage.UpdateTeamBudget(ctx, teamID, cost)
	ps.checkBudgetAlerts(ctx, teamID, newUsed)

	logEntry := &model.UsageLog{
		ID:               primaryLogID,
		TeamID:           teamID,
		Model:            modelName,
		PromptTokens:     openAIResp.Usage.PromptTokens,
		CompletionTokens: openAIResp.Usage.CompletionTokens,
		Cost:             cost,
		LatencyMS:        latency,
		CreatedAt:        time.Now(),
		IsShadow:         false,
		PrimaryLogID:     "",
		PIIRedactedCount: piiRedactedCount,
	}
	_ = ps.storage.CreateUsageLog(ctx, logEntry)

	// Execute shadow routing in a non-blocking background goroutine if requested
	if shadowModel != "" {
		go ps.executeShadowCall(context.Background(), teamID, openAIReq, shadowModel, primaryLogID)
	}

	// Write back to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(openAIResp)
}

func (ps *ProxyServer) handleStreamResponse(w http.ResponseWriter, resp *http.Response, translator ProviderTranslator, modelName string, teamID string, startTime time.Time, primaryLogID string, shadowModel string, openAIReq *ChatCompletionRequest, piiRedactedCount int, mergedPlaceholderMap map[string]string) {
	ctx := context.Background()

	// Ensure headers are correct for streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx proxy buffering

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error": "Streaming not supported by client connection"}`, http.StatusInternalServerError)
		return
	}

	reader := bufio.NewReader(resp.Body)
	var promptTokens, completionTokens int

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			// Log error and stop stream
			break
		}

		outBytes, pTokens, cTokens, done, err := translator.TranslateStreamChunk(ctx, line, modelName)
		if err != nil {
			// Log and skip or write error chunk
			continue
		}

		if pTokens > 0 {
			promptTokens = pTokens
		}
		if cTokens > 0 {
			completionTokens = cTokens
		}

		if len(outBytes) > 0 {
			_, _ = w.Write(outBytes)
			flusher.Flush()
		}

		if done {
			break
		}
	}

	latency := time.Since(startTime).Milliseconds()
	cost := CalculateCost(modelName, promptTokens, completionTokens)

	// Deduct and log usage
	newUsed, _ := ps.cache.IncrementTeamBudgetUsed(ctx, teamID, cost)
	_ = ps.storage.UpdateTeamBudget(ctx, teamID, cost)
	ps.checkBudgetAlerts(ctx, teamID, newUsed)

	logEntry := &model.UsageLog{
		ID:               primaryLogID,
		TeamID:           teamID,
		Model:            modelName,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		Cost:             cost,
		LatencyMS:        latency,
		CreatedAt:        time.Now(),
		IsShadow:         false,
		PrimaryLogID:     "",
		PIIRedactedCount: piiRedactedCount,
	}
	_ = ps.storage.CreateUsageLog(ctx, logEntry)

	// Execute shadow routing in a non-blocking background goroutine if requested
	if shadowModel != "" {
		go ps.executeShadowCall(context.Background(), teamID, openAIReq, shadowModel, primaryLogID)
	}
}

func (ps *ProxyServer) executeShadowCall(ctx context.Context, teamID string, originalReq *ChatCompletionRequest, shadowModel string, primaryLogID string) {
	// 1. Clone original request and modify model
	shadowReq := *originalReq
	shadowReq.Model = shadowModel
	shadowReq.Stream = false // Shadow calls are always standard non-stream to simplify token parsing

	// 2. Resolve provider
	translator, providerName := ps.registry.ResolveByModel(shadowModel)

	// 3. Fetch provider configuration (API Key)
	bgCtx := context.Background()
	providerConf, err := ps.storage.GetProviderConfig(bgCtx, providerName)
	if err != nil || providerConf == nil || providerConf.APIKey == "" {
		// Log and skip shadow call silently
		return
	}

	// 4. Translate request
	translatedBody, targetURL, headers, err := translator.TranslateRequest(bgCtx, &shadowReq)
	if err != nil {
		return
	}

	// 5. Create request to target provider
	outReq, err := http.NewRequestWithContext(bgCtx, http.MethodPost, targetURL, bytes.NewReader(translatedBody))
	if err != nil {
		return
	}

	// Apply standard headers and credentials
	for k, v := range headers {
		outReq.Header.Set(k, v)
	}

	if providerName == "openai" {
		outReq.Header.Set("Authorization", "Bearer "+providerConf.APIKey)
	} else if providerName == "anthropic" {
		outReq.Header.Set("x-api-key", providerConf.APIKey)
	}

	startTime := time.Now()
	resp, err := ps.httpClient.Do(outReq)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Translate back to OpenAI format
	openAIResp, err := translator.TranslateResponse(bgCtx, resp.StatusCode, respBytes, shadowModel)
	if err != nil {
		return
	}

	latency := time.Since(startTime).Milliseconds()
	cost := CalculateCost(shadowModel, openAIResp.Usage.PromptTokens, openAIResp.Usage.CompletionTokens)

	// Save shadow usage log (Note: does NOT update team budget because it is a shadow call! Perfect!)
	logEntry := &model.UsageLog{
		ID:               uuid.New().String(),
		TeamID:           teamID,
		Model:            shadowModel,
		PromptTokens:     openAIResp.Usage.PromptTokens,
		CompletionTokens: openAIResp.Usage.CompletionTokens,
		Cost:             cost,
		LatencyMS:        latency,
		CreatedAt:        time.Now(),
		IsShadow:         true,
		PrimaryLogID:     primaryLogID,
	}
	_ = ps.storage.CreateUsageLog(bgCtx, logEntry)
}

func (ps *ProxyServer) checkBudgetAlerts(ctx context.Context, teamID string, newUsed float64) {
	limit, _, found, err := ps.cache.GetTeamBudget(ctx, teamID)
	var teamName string
	if err != nil || !found {
		team, err := ps.storage.GetTeam(ctx, teamID)
		if err != nil || team == nil {
			return
		}
		limit = team.BudgetLimit
		teamName = team.Name
		_ = ps.cache.SetTeamBudget(ctx, teamID, limit, team.BudgetUsed)
	} else {
		team, err := ps.storage.GetTeam(ctx, teamID)
		if err == nil && team != nil {
			teamName = team.Name
		} else {
			teamName = fmt.Sprintf("Team %s", teamID[:8])
		}
	}

	if limit <= 0 {
		return
	}

	utilization := (newUsed / limit) * 100.0
	if utilization >= 100.0 {
		alert.SendBudgetAlert(ctx, ps.cache, teamID, teamName, newUsed, limit, 100)
	} else if utilization >= 80.0 {
		alert.SendBudgetAlert(ctx, ps.cache, teamID, teamName, newUsed, limit, 80)
	}
}
