package proxy

import (
	"context"
)

type ProviderTranslator interface {
	Name() string
	TranslateRequest(ctx context.Context, openAIReq *ChatCompletionRequest) (translatedBody []byte, targetURL string, headers map[string]string, err error)
	TranslateResponse(ctx context.Context, statusCode int, respBody []byte, requestedModel string) (*ChatCompletionResponse, error)

	// TranslateStreamChunk processes a raw stream line from the provider (e.g. SSE chunk)
	// and returns:
	// - openAIChunkBytes: the formatted OpenAI SSE line to send to the client (including "data: " prefix and "\n\n")
	// - promptTokensUsed: any prompt token count identified in this chunk (0 if not present)
	// - completionTokensUsed: any completion token count identified in this chunk (0 if not present)
	// - done: true if the stream is complete
	// - err: any processing error
	TranslateStreamChunk(ctx context.Context, chunk []byte, requestedModel string) (openAIChunkBytes []byte, promptTokensUsed int, completionTokensUsed int, done bool, err error)
}

// Registry to resolve translators
type TranslatorRegistry struct {
	translators map[string]ProviderTranslator
}

func NewTranslatorRegistry() *TranslatorRegistry {
	r := &TranslatorRegistry{
		translators: make(map[string]ProviderTranslator),
	}
	r.Register(NewOpenAITranslator())
	r.Register(NewAnthropicTranslator())
	return r
}

func (r *TranslatorRegistry) Register(t ProviderTranslator) {
	r.translators[t.Name()] = t
}

func (r *TranslatorRegistry) Get(name string) (ProviderTranslator, bool) {
	t, ok := r.translators[name]
	return t, ok
}

func (r *TranslatorRegistry) ResolveByModel(model string) (ProviderTranslator, string) {
	// Simple resolution logic: if model starts with "claude", route to anthropic. Otherwise openai.
	if len(model) >= 6 && model[:6] == "claude" {
		return r.translators["anthropic"], "anthropic"
	}
	return r.translators["openai"], "openai"
}
