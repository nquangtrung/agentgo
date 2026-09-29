package agentgo

import (
	"context"
	"fmt"

	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/providers"
)

// defaultMaxObjectRetries is the number of object generation attempts before
// giving up. Matches the AI SDK's default maxRetries.
const defaultMaxObjectRetries = 2

func resolveMaxObjectRetries(ctx context.Context) int {
	if v, ok := ctx.Value(models.MaxObjectRetriesContextKey).(int); ok && v > 0 {
		return v
	}
	return defaultMaxObjectRetries
}

func prepareObject(ctx context.Context, state agentState) (agentStateDelta, error) {
	stream := ctx.Value(models.StreamContextKey).(bool)

	return agentStateDelta{
		from:   PREPARE_OBJECT,
		stream: stream,
	}, nil
}

// resolveObjectText decodes the model's raw JSON output and validates it
// against the schema. It returns the decoded value on success, or an
// *models.ObjectParseError / *models.ObjectValidationError on failure.
func resolveObjectText(raw string, schema models.ObjectSchema) (any, error) {
	value, err := schema.DecodeJSON([]byte(raw))
	if err != nil {
		return nil, &models.ObjectParseError{Text: raw, Err: err}
	}
	if err := schema.Validate(value); err != nil {
		return nil, err
	}
	return value, nil
}

// objectRepairMessages builds the messages fed back to the model after a
// failed attempt: the model's own bad output, then an instruction to fix it.
func objectRepairMessages(raw string, err error) []models.Message {
	return []models.Message{
		models.NewAssistantStringMessage(raw),
		models.NewHumanStringMessage(fmt.Sprintf(
			"The previous response failed validation: %v. "+
				"Please generate a corrected response that satisfies the schema.",
			err,
		)),
	}
}

func generateObject(ctx context.Context, state agentState) (agentStateDelta, error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	schema := ctx.Value(models.SchemaContextKey).(models.ObjectSchema)
	messages := state.messages

	if state.currentStep.prepareStepResult.Messages != nil {
		messages = *state.currentStep.prepareStepResult.Messages
	}

	output, err := provider.GenerateText(ctx, providers.AgentProviderPromptMessageParams{
		Messages:       messages,
		ResponseFormat: responseFormatFromSchema(schema),
	})

	if err != nil {
		return agentStateDelta{}, err
	}

	value, validationErr := resolveObjectText(output.Text, schema)
	if validationErr != nil {
		return agentStateDelta{
			from:                 GENERATE_OBJECT,
			incrementObjectAttempt: true,
			objectError:          validationErr,
			objectRaw:            output.Text,
			appendMessages:       objectRepairMessages(output.Text, validationErr),
			usage:                &output.Usage,
			addError:             validationErr,
		}, nil
	}

	outputAsToolExecuteOutput := resolveObjectOutputAsToolExecuteOutput(output, value)
	return agentStateDelta{
		from:              GENERATE_OBJECT,
		archiveToolResult: &outputAsToolExecuteOutput,
		objectGenerated:   true,
	}, nil
}

func streamObject(ctx context.Context, state agentState) (agentStateDelta, error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	schema := ctx.Value(models.SchemaContextKey).(models.ObjectSchema)
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)
	messages := state.messages

	if state.currentStep.prepareStepResult.Messages != nil {
		messages = *state.currentStep.prepareStepResult.Messages
	}

	output, err := provider.StreamText(
		ctx,
		providers.AgentProviderPromptMessageParams{
			Messages:       messages,
			ResponseFormat: responseFormatFromSchema(schema),
		},
		*emitter,
	)

	if err != nil {
		// TODO Handle error and retry
		return agentStateDelta{}, err
	}

	value, validationErr := resolveObjectText(output.Text, schema)
	if validationErr != nil {
		return agentStateDelta{
			from:                 STREAM_OBJECT,
			incrementObjectAttempt: true,
			objectError:          validationErr,
			objectRaw:            output.Text,
			appendMessages:       objectRepairMessages(output.Text, validationErr),
			usage:                &output.Usage,
			addError:             validationErr,
		}, nil
	}

	outputAsToolExecuteOutput := resolveObjectOutputAsToolExecuteOutput(output, value)
	return agentStateDelta{
		from:              STREAM_OBJECT,
		archiveToolResult: &outputAsToolExecuteOutput,
		objectGenerated:   true,
	}, nil
}

func endObject(ctx context.Context, state agentState) (agentStateDelta, error) {
	provider := ctx.Value(models.ProviderContextKey).(providers.AgentProvider)
	emitter := ctx.Value(models.PartEmitterContextKey).(*models.PartEmitter)
	maxRetries := resolveMaxObjectRetries(ctx)

	if state.objectError != nil {
		if state.objectAttempts >= maxRetries {
			emitter.Emit(models.NewObjectErrorPart(provider.Context(), state.objectError, state.objectRaw))
			return agentStateDelta{
				from:      END_OBJECT,
				shouldEnd: true,
			}, nil
		}
		// Retries remain: let the loop continue without emitting an error part,
		// so the repair is transparent to the consumer.
		return agentStateDelta{from: END_OBJECT}, nil
	}

	if lastRecord := state.toolExecutionsArchive.LastRecord(); lastRecord != nil && lastRecord.ToolResult != nil {
		emitter.Emit(models.NewObjectPart(
			provider.Context(),
			lastRecord.ToolResult.Output["object"],
			lastRecord.ToolResult.Output["raw"].(string),
		))
	}

	return agentStateDelta{from: END_OBJECT}, nil
}
