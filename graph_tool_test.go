package agentgo

import (
	"context"
	"testing"

	"github.com/nquangtrung/agentgo/mocks"
	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// The graph must record the model's own tool requests before their results, so
// providers that pair a result with the call that produced it see both.
func TestResolveToolRecordsToolCallBeforeResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	mockProvider.EXPECT().Context().AnyTimes().Return(models.LanguageModelContext{ModelName: "test-model"})

	calls := []models.ToolCall{
		{ToolName: "mock_tool", ID: "call_1", Params: map[string]any{"p": "1"}},
		{ToolName: "mock_tool", ID: "call_2", Params: map[string]any{"p": "2"}},
	}

	mockProvider.EXPECT().ResolveToolCall(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		models.LanguageModelToolCallResolveOutput{
			ToolCalls: calls,
			Text:      "Checking both.",
		}, nil,
	)

	ctx := context.WithValue(context.Background(), models.ProviderContextKey, providers.AgentProvider(mockProvider))
	ctx = context.WithValue(ctx, models.ToolsContextKey, []models.BaseTool{
		models.NewTool(models.NewToolParams{
			Name: "mock_tool",
			Fn:   func(models.ToolExecuteParams) models.ToolExecuteOutput { return models.ToolExecuteOutput{} },
		}),
	})

	delta, err := resolveTool(ctx, agentState{
		messages: []models.Message{models.NewHumanStringMessage("go")},
	})
	require.NoError(t, err)

	require.Len(t, delta.appendMessages, 1)
	message := delta.appendMessages[0]

	assert.Equal(t, models.MessageRoleAssistant, message.Type())
	assert.Equal(t, "Checking both.", message.Content().Text())

	recorded := message.Content().ToolCalls()
	require.Len(t, recorded, 2, "both parallel calls should be recorded on one turn")
	assert.Equal(t, "call_1", recorded[0].ID())
	assert.Equal(t, "call_2", recorded[1].ID())
}

// A step that resolved no calls must not append an empty assistant turn.
func TestResolveToolAppendsNothingWithoutCalls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProvider := mocks.NewMockAgentProvider(ctrl)
	mockProvider.EXPECT().ResolveToolCall(gomock.Any(), gomock.Any(), gomock.Any()).Return(
		models.LanguageModelToolCallResolveOutput{}, nil,
	)

	ctx := context.WithValue(context.Background(), models.ProviderContextKey, providers.AgentProvider(mockProvider))
	ctx = context.WithValue(ctx, models.ToolsContextKey, []models.BaseTool{})

	delta, err := resolveTool(ctx, agentState{
		messages: []models.Message{models.NewHumanStringMessage("go")},
	})
	require.NoError(t, err)
	assert.Empty(t, delta.appendMessages)
}
