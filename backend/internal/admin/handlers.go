package admin

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"gateway/internal/cache"
	"gateway/internal/model"
	"gateway/internal/proxy"
	"gateway/internal/storage"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type AdminHandler struct {
	storage storage.Storage
	cache   cache.Cache
}

func NewAdminHandler(s storage.Storage, c cache.Cache) *AdminHandler {
	return &AdminHandler{storage: s, cache: c}
}

func (ah *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Enable CORS for frontend development
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch r.URL.Path {
	case "/api/teams":
		if r.Method == http.MethodPost {
			ah.handleCreateTeam(w, r)
		} else if r.Method == http.MethodGet {
			ah.handleListTeams(w, r)
		} else if r.Method == http.MethodDelete {
			ah.handleDeleteTeam(w, r)
		} else {
			http.Error(w, `{"error": "Method Not Allowed"}`, http.StatusMethodNotAllowed)
		}
	case "/api/keys":
		if r.Method == http.MethodPost {
			ah.handleCreateKey(w, r)
		} else if r.Method == http.MethodGet {
			ah.handleListKeys(w, r)
		} else if r.Method == http.MethodDelete {
			ah.handleDeleteKey(w, r)
		} else {
			http.Error(w, `{"error": "Method Not Allowed"}`, http.StatusMethodNotAllowed)
		}
	case "/api/provider-configs":
		if r.Method == http.MethodPost {
			ah.handleCreateProviderConfig(w, r)
		} else {
			http.Error(w, `{"error": "Method Not Allowed"}`, http.StatusMethodNotAllowed)
		}
	case "/api/usage":
		if r.Method == http.MethodGet {
			ah.handleGetUsage(w, r)
		} else {
			http.Error(w, `{"error": "Method Not Allowed"}`, http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, `{"error": "Not Found"}`, http.StatusNotFound)
	}
}

func (ah *AdminHandler) handleCreateTeam(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string  `json:"name"`
		BudgetLimit float64 `json:"budget_limit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error": "Team name is required"}`, http.StatusBadRequest)
		return
	}

	team := &model.Team{
		ID:          uuid.New().String(),
		Name:        req.Name,
		BudgetLimit: req.BudgetLimit,
		BudgetUsed:  0.0,
	}

	if err := ah.storage.CreateTeam(r.Context(), team); err != nil {
		http.Error(w, `{"error": "Failed to create team"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(team)
}

func (ah *AdminHandler) handleListTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := ah.storage.ListTeams(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Failed to list teams"}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(teams)
}

func (ah *AdminHandler) handleDeleteTeam(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error": "id parameter is required"}`, http.StatusBadRequest)
		return
	}

	if err := ah.storage.DeleteTeam(r.Context(), id); err != nil {
		log.Printf("ERROR deleting team with ID %s: %v", id, err)
		http.Error(w, `{"error": "Failed to delete team from database"}`, http.StatusInternalServerError)
		return
	}

	// Invalidate team budget from memory cache
	_ = ah.cache.InvalidateTeam(r.Context(), id)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "deleted"}`))
}

func (ah *AdminHandler) handleCreateKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TeamID string `json:"team_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if req.TeamID == "" {
		http.Error(w, `{"error": "Team ID is required"}`, http.StatusBadRequest)
		return
	}

	// Generate key cleartext
	rawKeyBytes := make([]byte, 16)
	_, _ = rand.Read(rawKeyBytes)
	cleartextKey := "gw-" + hex.EncodeToString(rawKeyBytes)

	keyHash := proxy.HashKey(cleartextKey)

	key := &model.GatewayKey{
		KeyHash:   keyHash,
		TeamID:    req.TeamID,
		CreatedAt: time.Now(),
		Status:    "active",
	}

	if err := ah.storage.CreateGatewayKey(r.Context(), key); err != nil {
		http.Error(w, `{"error": "Failed to store gateway key"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"key":      cleartextKey,
		"key_hash": keyHash,
		"team_id":  req.TeamID,
		"status":   "active",
	})
}

func (ah *AdminHandler) handleListKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := ah.storage.ListGatewayKeys(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Failed to list keys"}`, http.StatusInternalServerError)
		return
	}

	teams, err := ah.storage.ListTeams(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Failed to list teams"}`, http.StatusInternalServerError)
		return
	}

	teamMap := make(map[string]string)
	for _, t := range teams {
		teamMap[t.ID] = t.Name
	}

	type KeyResponse struct {
		KeyHash   string    `json:"key_hash"`
		TeamID    string    `json:"team_id"`
		TeamName  string    `json:"team_name"`
		CreatedAt time.Time `json:"created_at"`
		Status    string    `json:"status"`
	}

	var resp []KeyResponse
	for _, k := range keys {
		name := teamMap[k.TeamID]
		if name == "" {
			name = "Unknown Team"
		}
		resp = append(resp, KeyResponse{
			KeyHash:   k.KeyHash,
			TeamID:    k.TeamID,
			TeamName:  name,
			CreatedAt: k.CreatedAt,
			Status:    k.Status,
		})
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func (ah *AdminHandler) handleDeleteKey(w http.ResponseWriter, r *http.Request) {
	keyHash := r.URL.Query().Get("key_hash")
	if keyHash == "" {
		http.Error(w, `{"error": "key_hash parameter is required"}`, http.StatusBadRequest)
		return
	}

	if err := ah.storage.DeleteGatewayKey(r.Context(), keyHash); err != nil {
		http.Error(w, `{"error": "Failed to delete key from database"}`, http.StatusInternalServerError)
		return
	}

	// Invalidate key from memory cache
	_ = ah.cache.InvalidateGatewayKey(r.Context(), keyHash)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "deleted"}`))
}

func (ah *AdminHandler) handleCreateProviderConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProviderName string `json:"provider_name"`
		APIKey       string `json:"api_key"`
		RoutingRules string `json:"routing_rules"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if req.ProviderName == "" || req.APIKey == "" {
		http.Error(w, `{"error": "provider_name and api_key are required"}`, http.StatusBadRequest)
		return
	}

	config := &model.ProviderConfig{
		ProviderName: req.ProviderName,
		APIKey:       req.APIKey,
		RoutingRules: req.RoutingRules,
	}

	if err := ah.storage.CreateProviderConfig(r.Context(), config); err != nil {
		http.Error(w, `{"error": "Failed to create provider config"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "saved"}`))
}

func (ah *AdminHandler) handleGetUsage(w http.ResponseWriter, r *http.Request) {
	teamID := r.URL.Query().Get("team_id")
	if teamID == "" {
		http.Error(w, `{"error": "team_id parameter is required"}`, http.StatusBadRequest)
		return
	}

	logs, err := ah.storage.GetUsageLogs(r.Context(), teamID)
	if err != nil {
		http.Error(w, `{"error": "Failed to retrieve usage logs"}`, http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(logs)
}
