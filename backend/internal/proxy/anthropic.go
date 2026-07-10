package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []AnthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Stream    bool               `json:"stream,omitempty"`
}

type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type AnthropicResponse struct {
	ID      string             `json:"id"`
	Model   string             `json:"model"`
	Content []AnthropicContent `json:"content"`
	Usage   AnthropicUsage     `json:"usage"`
}

// Stream structures
type AnthropicStreamMessageStart struct {
	Type    string `json:"type"`
	Message struct {
		ID    string         `json:"id"`
		Usage AnthropicUsage `json:"usage"`
	} `json:"message"`
}

type AnthropicStreamContentBlockDelta struct {
	Type  string `json:"type"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

type AnthropicStreamMessageDelta struct {
	Type  string `json:"type"`
	Usage struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type AnthropicTranslator struct{}

func NewAnthropicTranslator() *AnthropicTranslator {
	return &AnthropicTranslator{}
}

func (a *AnthropicTranslator) Name() string {
	return "anthropic"
}

func (a *AnthropicTranslator) TranslateRequest(ctx context.Context, openAIReq *ChatCompletionRequest) ([]byte, string, map[string]string, error) {
	// Extract system messages from the messages array
	var systemPrompt string
	var anthropicMessages []AnthropicMessage

	for _, msg := range openAIReq.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			// Map role names if they differ
			role := msg.Role
			if role == "user" || role == "assistant" {
				anthropicMessages = append(anthropicMessages, AnthropicMessage{
					Role:    role,
					Content: msg.Content,
				})
			} else {
				// Fallback to user role
				anthropicMessages = append(anthropicMessages, AnthropicMessage{
					Role:    "user",
					Content: msg.Content,
				})
			}
		}
	}

	anthropicReq := AnthropicRequest{
		Model:     openAIReq.Model,
		Messages:  anthropicMessages,
		MaxTokens: 4096, // required field by Anthropic, default to safe 4k
		System:    systemPrompt,
		Stream:    openAIReq.Stream,
	}

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, "", nil, err
	}

	headers := map[string]string{
		"Content-Type":      "application/json",
		"anthropic-version": "2023-06-01",
	}

	return body, "https://api.anthropic.com/v1/messages", headers, nil
}

func (a *AnthropicTranslator) TranslateResponse(ctx context.Context, statusCode int, respBody []byte, requestedModel string) (*ChatCompletionResponse, error) {
	if statusCode >= 400 {
		return nil, fmt.Errorf("anthropic error response: %s", string(respBody))
	}

	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil {
		return nil, err
	}

	var textContent string
	for _, content := range anthropicResp.Content {
		if content.Type == "text" {
			textContent += content.Text
		}
	}

	openAIResp := &ChatCompletionResponse{
		ID:      anthropicResp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   requestedModel,
		Choices: []Choice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: textContent,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}

	return openAIResp, nil
}

func (a *AnthropicTranslator) TranslateStreamChunk(ctx context.Context, chunk []byte, requestedModel string) ([]byte, int, int, bool, error) {
	line := string(chunk)
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, 0, 0, false, nil
	}

	// Anthropic uses event: and data: lines. We only care about data: JSON bodies
	if !strings.HasPrefix(line, "data:") {
		return nil, 0, 0, false, nil
	}

	data := strings.TrimPrefix(line, "data:")
	data = strings.TrimSpace(data)

	// Quick check of the type field in the JSON
	var typeCheck struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(data), &typeCheck); err != nil {
		return nil, 0, 0, false, nil
	}

	switch typeCheck.Type {
	case "message_start":
		var ms AnthropicStreamMessageStart
		_ = json.Unmarshal([]byte(data), &ms)
		promptTokens := ms.Message.Usage.InputTokens

		// Build a dummy start block for OpenAI (optional, but good practice)
		openAIChunk := ChatCompletionStreamResponse{
			ID:      ms.Message.ID,
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   requestedModel,
			Choices: []StreamChoice{
				{
					Index: 0,
					Delta: ChatMessage{
						Role: "assistant",
					},
				},
			},
		}
		outBytes, _ := json.Marshal(openAIChunk)
		formatted := []byte(fmt.Sprintf("data: %s\n\n", string(outBytes)))
		return formatted, promptTokens, 0, false, nil

	case "content_block_delta":
		var cbd AnthropicStreamContentBlockDelta
		_ = json.Unmarshal([]byte(data), &cbd)

		openAIChunk := ChatCompletionStreamResponse{
			ID:      "chatcmpl-stream",
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   requestedModel,
			Choices: []StreamChoice{
				{
					Index: 0,
					Delta: ChatMessage{
						Content: cbd.Delta.Text,
					},
				},
			},
		}
		outBytes, _ := json.Marshal(openAIChunk)
		formatted := []byte(fmt.Sprintf("data: %s\n\n", string(outBytes)))
		return formatted, 0, 0, false, nil

	case "message_delta":
		var md AnthropicStreamMessageDelta
		_ = json.Unmarshal([]byte(data), &md)
		completionTokens := md.Usage.OutputTokens

		openAIChunk := ChatCompletionStreamResponse{
			ID:      "chatcmpl-stream",
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   requestedModel,
			Choices: []StreamChoice{
				{
					Index:        0,
					Delta:        ChatMessage{},
					FinishReason: "stop",
				},
			},
		}
		outBytes, _ := json.Marshal(openAIChunk)
		formatted := []byte(fmt.Sprintf("data: %s\n\n", string(outBytes)))
		return formatted, 0, completionTokens, false, nil

	case "message_stop":
		return []byte("data: [DONE]\n\n"), 0, 0, true, nil

	default:
		// Unknown event type, just skip and return empty
		return nil, 0, 0, false, nil
	}
}
