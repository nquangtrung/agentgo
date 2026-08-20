package openai

import (
	"context"

	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/responses"
)

//go:generate mockgen -destination=./mocks/mock_open_ai_response.go -package=mocks trontria.com/agentgo/providers/openai openAIResponsesService
type openAIResponsesService interface {
	New(ctx context.Context, body responses.ResponseNewParams, opts ...option.RequestOption) (*responses.Response, error)
	NewStreaming(ctx context.Context, body responses.ResponseNewParams, opts ...option.RequestOption) *ssestream.Stream[responses.ResponseStreamEventUnion]
}
