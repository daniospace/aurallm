package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type OpenAITranslator struct{}

func NewOpenAITranslator() *OpenAITranslator {
	return &OpenAITranslator{}
}

func (o *OpenAITranslator) Name() string {
	return "openai"
}

func (o *OpenAITranslator) TranslateRequest(ctx context.Context, openAIReq *ChatCompletionRequest) ([]byte, string, map[string]string, error) {
	body, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, "", nil, err
	}
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	return body, "https://api.openai.com/v1/chat/completions", headers, nil
}

func (o *OpenAITranslator) TranslateResponse(ctx context.Context, statusCode int, respBody []byte, requestedModel string) (*ChatCompletionResponse, error) {
	if statusCode >= 400 {
		return nil, fmt.Errorf("openai error response: %s", string(respBody))
	}
	var resp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (o *OpenAITranslator) TranslateStreamChunk(ctx context.Context, chunk []byte, requestedModel string) ([]byte, int, int, bool, error) {
	line := string(chunk)
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, 0, 0, false, nil
	}

	if line == "data: [DONE]" {
		return chunk, 0, 0, true, nil
	}

	if !strings.HasPrefix(line, "data:") {
		return chunk, 0, 0, false, nil
	}

	data := strings.TrimPrefix(line, "data:")
	data = strings.TrimSpace(data)

	var streamResp ChatCompletionStreamResponse
	if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
		return chunk, 0, 0, false, nil
	}

	promptTokens := 0
	completionTokens := 0
	if streamResp.Usage != nil {
		promptTokens = streamResp.Usage.PromptTokens
		completionTokens = streamResp.Usage.CompletionTokens
	}

	return chunk, promptTokens, completionTokens, false, nil
}
