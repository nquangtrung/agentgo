package deepseek

import (
	"context"

	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/responses"
)

// deepSeekResponsesService is the slice of the OpenAI Responses API that the
// DeepSeek provider depends on. DeepSeek serves the Responses API at its own
// base URL, so the same generated client works once the base URL is overridden.
//
//go:generate mockgen -destination=./mocks/mock_deepseek_responses_service.go -package=mocks github.com/nquangtrung/agentgo/providers/deepseek deepSeekResponsesService
type deepSeekResponsesService interface {
	New(ctx context.Context, body responses.ResponseNewParams, opts ...option.RequestOption) (*responses.Response, error)
	NewStreaming(ctx context.Context, body responses.ResponseNewParams, opts ...option.RequestOption) *ssestream.Stream[responses.ResponseStreamEventUnion]
}
