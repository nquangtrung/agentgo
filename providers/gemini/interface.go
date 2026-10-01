package gemini

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/ssestream"
)

// Gemini is reached through Google's OpenAI compatibility endpoint, which
// speaks the Chat Completions shape rather than the Responses API that the
// OpenAI provider uses. Only the two methods this provider needs are part of
// the interface.
//
//go:generate mockgen -destination=./mocks/mock_gemini_chat_service.go -package=mocks github.com/nquangtrung/agentgo/providers/gemini geminiChatService
type geminiChatService interface {
	New(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error)
	NewStreaming(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) *ssestream.Stream[openai.ChatCompletionChunk]
}
