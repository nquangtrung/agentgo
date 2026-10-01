package claude

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/ssestream"
)

// Claude is reached through Anthropic's OpenAI SDK compatibility layer, which
// speaks the Chat Completions shape rather than the Responses API that the
// OpenAI provider uses. Only the two methods this provider needs are part of
// the interface.
//
//go:generate mockgen -destination=./mocks/mock_claude_chat_service.go -package=mocks github.com/nquangtrung/agentgo/providers/claude claudeChatService
type claudeChatService interface {
	New(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (*openai.ChatCompletion, error)
	NewStreaming(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) *ssestream.Stream[openai.ChatCompletionChunk]
}
