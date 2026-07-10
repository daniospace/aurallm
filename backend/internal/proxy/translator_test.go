package proxy

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestOpenAITranslator(t *testing.T) {
	ctx := context.Background()
	trans := NewOpenAITranslator()

	req := &ChatCompletionRequest{
		Model: "gpt-4o",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello!"},
		},
		Stream: false,
	}

	// 1. Test Request Translation
	body, url, headers, err := trans.TranslateRequest(ctx, req)
	if err != nil {
		t.Fatalf("TranslateRequest failed: %v", err)
	}

	if url != "https://api.openai.com/v1/chat/completions" {
		t.Errorf("Unexpected target URL: %s", url)
	}

	if headers["Content-Type"] != "application/json" {
		t.Errorf("Unexpected header Content-Type: %s", headers["Content-Type"])
	}

	var parsedReq ChatCompletionRequest
	if err := json.Unmarshal(body, &parsedReq); err != nil {
		t.Fatalf("Failed to unmarshal translated request body: %v", err)
	}
	if parsedReq.Model != "gpt-4o" || parsedReq.Messages[0].Content != "Hello!" {
		t.Errorf("Translated body has mismatch: %+v", parsedReq)
	}

	// 2. Test Response Translation
	respBody := `{"id":"chatcmpl-123","object":"chat.completion","created":1670000000,"model":"gpt-4o","choices":[{"index":0,"message":{"role":"assistant","content":"Hi back!"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":15,"total_tokens":25}}`
	openAIResp, err := trans.TranslateResponse(ctx, 200, []byte(respBody), "gpt-4o")
	if err != nil {
		t.Fatalf("TranslateResponse failed: %v", err)
	}

	if openAIResp.ID != "chatcmpl-123" || openAIResp.Usage.TotalTokens != 25 || openAIResp.Choices[0].Message.Content != "Hi back!" {
		t.Errorf("Translated response mismatch: %+v", openAIResp)
	}

	// 3. Test Stream Chunk Translation
	chunk := []byte(`data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1670000000,"model":"gpt-4o","choices":[{"index":0,"delta":{"content":"Hi"},"finish_reason":null}],"usage":{"prompt_tokens":10,"completion_tokens":20}}`)
	outBytes, pTokens, cTokens, done, err := trans.TranslateStreamChunk(ctx, chunk, "gpt-4o")
	if err != nil {
		t.Fatalf("TranslateStreamChunk failed: %v", err)
	}

	if pTokens != 10 || cTokens != 20 || done {
		t.Errorf("Unexpected values: pTokens=%d, cTokens=%d, done=%v", pTokens, cTokens, done)
	}

	if string(outBytes) != string(chunk) {
		t.Errorf("Expected stream chunk to pass through unchanged, got %s", string(outBytes))
	}
}

func TestAnthropicTranslator(t *testing.T) {
	ctx := context.Background()
	trans := NewAnthropicTranslator()

	req := &ChatCompletionRequest{
		Model: "claude-3-5-sonnet",
		Messages: []ChatMessage{
			{Role: "system", Content: "Be concise."},
			{Role: "user", Content: "Hello!"},
		},
		Stream: false,
	}

	// 1. Test Request Translation
	body, url, headers, err := trans.TranslateRequest(ctx, req)
	if err != nil {
		t.Fatalf("TranslateRequest failed: %v", err)
	}

	if url != "https://api.anthropic.com/v1/messages" {
		t.Errorf("Unexpected target URL: %s", url)
	}

	if headers["anthropic-version"] != "2023-06-01" {
		t.Errorf("Unexpected anthropic-version header: %s", headers["anthropic-version"])
	}

	var parsedReq AnthropicRequest
	if err := json.Unmarshal(body, &parsedReq); err != nil {
		t.Fatalf("Failed to unmarshal AnthropicRequest: %v", err)
	}

	if parsedReq.Model != "claude-3-5-sonnet" {
		t.Errorf("Model mismatch: %s", parsedReq.Model)
	}
	if parsedReq.System != "Be concise." {
		t.Errorf("System prompt not extracted: %s", parsedReq.System)
	}
	if len(parsedReq.Messages) != 1 || parsedReq.Messages[0].Content != "Hello!" || parsedReq.Messages[0].Role != "user" {
		t.Errorf("Messages translation mismatch: %+v", parsedReq.Messages)
	}

	// 2. Test Response Translation
	anthropicRespBody := `{"id":"msg_123","type":"message","role":"assistant","content":[{"type":"text","text":"Hi from Claude!"}],"model":"claude-3-5-sonnet","stop_reason":"end_turn","usage":{"input_tokens":12,"output_tokens":18}}`
	openAIResp, err := trans.TranslateResponse(ctx, 200, []byte(anthropicRespBody), "claude-3-5-sonnet")
	if err != nil {
		t.Fatalf("TranslateResponse failed: %v", err)
	}

	if openAIResp.ID != "msg_123" || openAIResp.Usage.PromptTokens != 12 || openAIResp.Usage.CompletionTokens != 18 {
		t.Errorf("Usage translation mismatch: %+v", openAIResp.Usage)
	}
	if len(openAIResp.Choices) != 1 || openAIResp.Choices[0].Message.Content != "Hi from Claude!" {
		t.Errorf("Content translation mismatch: %+v", openAIResp.Choices)
	}

	// 3. Test Stream Chunk Translation
	// Start Event
	startChunk := []byte(`data: {"type":"message_start","message":{"id":"msg_456","usage":{"input_tokens":15}}}`)
	outBytes, pTokens, cTokens, done, err := trans.TranslateStreamChunk(ctx, startChunk, "claude-3-5-sonnet")
	if err != nil {
		t.Fatalf("TranslateStreamChunk message_start failed: %v", err)
	}
	if pTokens != 15 || cTokens != 0 || done {
		t.Errorf("message_start values mismatch: pTokens=%d, cTokens=%d, done=%v", pTokens, cTokens, done)
	}
	if !strings.Contains(string(outBytes), "msg_456") {
		t.Errorf("Translated OpenAI start chunk missing original message id, got %s", string(outBytes))
	}

	// Delta Event
	deltaChunk := []byte(`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"hello "}}`)
	outBytes, pTokens, cTokens, done, err = trans.TranslateStreamChunk(ctx, deltaChunk, "claude-3-5-sonnet")
	if err != nil {
		t.Fatalf("TranslateStreamChunk content_block_delta failed: %v", err)
	}
	if pTokens != 0 || cTokens != 0 || done {
		t.Errorf("content_block_delta values mismatch: pTokens=%d, cTokens=%d, done=%v", pTokens, cTokens, done)
	}
	if !strings.Contains(string(outBytes), "hello ") {
		t.Errorf("Translated OpenAI delta chunk missing text content, got %s", string(outBytes))
	}

	// Message Delta (with final usage)
	msgDeltaChunk := []byte(`data: {"type":"message_delta","usage":{"output_tokens":25}}`)
	outBytes, pTokens, cTokens, done, err = trans.TranslateStreamChunk(ctx, msgDeltaChunk, "claude-3-5-sonnet")
	if err != nil {
		t.Fatalf("TranslateStreamChunk message_delta failed: %v", err)
	}
	if pTokens != 0 || cTokens != 25 || done {
		t.Errorf("message_delta values mismatch: pTokens=%d, cTokens=%d, done=%v", pTokens, cTokens, done)
	}
	if !strings.Contains(string(outBytes), `"finish_reason":"stop"`) {
		t.Errorf("Translated OpenAI finish chunk missing stop reason, got %s", string(outBytes))
	}

	// Stop Event
	stopChunk := []byte(`data: {"type":"message_stop"}`)
	outBytes, pTokens, cTokens, done, err = trans.TranslateStreamChunk(ctx, stopChunk, "claude-3-5-sonnet")
	if err != nil {
		t.Fatalf("TranslateStreamChunk message_stop failed: %v", err)
	}
	if pTokens != 0 || cTokens != 0 || !done {
		t.Errorf("message_stop values mismatch: pTokens=%d, cTokens=%d, done=%v", pTokens, cTokens, done)
	}
	if string(outBytes) != "data: [DONE]\n\n" {
		t.Errorf("Expected data: [DONE]\n\n, got %q", string(outBytes))
	}
}
