package agentgo

import (
	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
	"github.com/nquangtrung/agentgo/utils"
)

func resolveMessages(params Params) []models.Message {
	// Copy before appending: appending to the caller's slice would write into
	// their backing array whenever capacity allows, so reusing one slice across
	// calls could silently clobber their messages.
	messages := make([]models.Message, 0, len(params.Messages)+1)
	messages = append(messages, params.Messages...)
	if params.Prompt != "" {
		messages = append(messages, models.NewHumanStringMessage(params.Prompt))
	}

	return messages
}

func resolveProviderFromParams(params Params) (providers.AgentProvider, error) {
	if params.ModelName != "" {
		var provider, err = CreateAgentProvider(AgentProviderFactoryParams{
			ModelName: params.ModelName,
		})
		if err != nil {
			return nil, err
		}
		params.Provider = provider
	}

	if params.Provider == nil {
		return nil, &models.UnsupportedModelError{ModelName: "nil provider"}
	}

	return params.Provider, nil
}

func mustResolveProviderFromParams(params Params) providers.AgentProvider {
	return utils.Must(resolveProviderFromParams(params))
}

func resolveToolFromToolCall(toolCall models.ToolCall, tools []models.BaseTool) (models.Tool, error) {
	for _, tool := range tools {
		if tool.Name() == toolCall.ToolName {
			return tool, nil
		}
	}
	return nil, &models.ToolNotFoundError{ToolName: toolCall.ToolName}
}

func resolveTextOutputAsToolExecuteOutput(textOutput models.LanguageModelOutput, err error) models.ToolExecuteOutput {
	if err != nil {
		return models.ToolExecuteOutput{
			Output: nil,
			Error:  err,
			Usage:  textOutput.Usage,
		}
	}
	return models.ToolExecuteOutput{
		Output: map[string]any{
			"text": textOutput.Text,
		},
		Error: nil,
		Usage: textOutput.Usage,
	}
}

func resolveObjectOutputAsToolExecuteOutput(output models.LanguageModelOutput, value any) models.ToolExecuteOutput {
	return models.ToolExecuteOutput{
		Output: map[string]any{
			"object": value,
			"raw":    output.Text,
		},
		Error: nil,
		Usage: output.Usage,
	}
}

func responseFormatFromSchema(schema models.ObjectSchema) *models.ResponseFormat {
	return &models.ResponseFormat{
		Name:        schema.Name(),
		Description: schema.Description(),
		JSONSchema:  schema.JSONSchema(),
	}
}

// accumulateArchiveUsage sums the token usage across every archived record.
func accumulateArchiveUsage(context *models.ToolExecutionsArchive) models.LanguageModelUsage {
	return utils.Reduce(
		context.Records(),
		func(acc models.LanguageModelUsage, step models.ToolExecutionRecord) models.LanguageModelUsage {
			return models.LanguageModelUsage{
				InputTokens: acc.InputTokens + step.ToolResult.Usage.InputTokens,
				InputTokensDetails: models.LanguageModelUsageInputTokensDetails{
					CachedTokens:     acc.InputTokensDetails.CachedTokens + step.ToolResult.Usage.InputTokensDetails.CachedTokens,
					CacheWriteTokens: acc.InputTokensDetails.CacheWriteTokens + step.ToolResult.Usage.InputTokensDetails.CacheWriteTokens,
				},
				OutputTokens: acc.OutputTokens + step.ToolResult.Usage.OutputTokens,
				OutputTokensDetails: models.LanguageModelUsageOutputTokensDetails{
					ReasoningTokens: acc.OutputTokensDetails.ReasoningTokens + step.ToolResult.Usage.OutputTokensDetails.ReasoningTokens,
				},
				TotalTokens: acc.TotalTokens + step.ToolResult.Usage.TotalTokens,
			}
		},
		models.LanguageModelUsage{},
	)
}

func resolveExecutionContextAsTextOutput(context *models.ToolExecutionsArchive) (models.LanguageModelOutput, error) {
	if context.LastRecord() == nil {
		return models.LanguageModelOutput{}, &models.ExecutionContextError{
			Message: "no steps in execution context",
		}
	}
	lastOutput, ok := context.LastRecord().ToolResult.Output["text"].(string)
	text := utils.Ternary(ok, lastOutput, "")

	return models.LanguageModelOutput{
		Text:      text,
		Usage:     accumulateArchiveUsage(context),
		ModelName: context.ModelName(),
		Context:   context,
	}, nil
}

// resolveExecutionContextAsObjectResult pulls the last archived object out of
// the execution context. The value is typed as any; the caller asserts it to
// the schema's concrete type.
func resolveExecutionContextAsObjectResult(context *models.ToolExecutionsArchive) (any, string, error) {
	if context.LastRecord() == nil {
		return nil, "", &models.ExecutionContextError{
			Message: "no steps in execution context",
		}
	}

	result := context.LastRecord().ToolResult
	object, _ := result.Output["object"].(any)
	raw, _ := result.Output["raw"].(string)

	return object, raw, nil
}
