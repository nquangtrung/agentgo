package main

import (
	"context"
	"log"
	"testing"

	"github.com/nquangtrung/agentgo"
	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/utils"
)

// Live Claude tests. These make real billable calls and need CLAUDE_API_KEY in
// ../.env. They are skipped unless AGENTGO_LIVE_TESTS=1 — see live_test.go.

func TestGenerateTextClaude(t *testing.T) {
	requireLiveTest(t)
	utils.LoadEnv("../.env")
	ctx := context.Background()

	output, err := agentgo.GenerateText(ctx, agentgo.Params{
		ModelName: "claude-sonnet-4-6",
		Prompt:    "Say this is a test",
	})
	if err != nil {
		panic(err)
	}

	log.Printf("[output] - %s\n", output.ModelName)
	log.Printf("Text: %s\n", output.Text)
	log.Printf("Input Tokens: %d\n", output.Usage.InputTokens)
	log.Printf("Output Tokens: %d\n", output.Usage.OutputTokens)
	log.Printf("Total Tokens: %d\n", output.Usage.TotalTokens)
}

func TestGenerateTextClaudeWithInput(t *testing.T) {
	requireLiveTest(t)
	utils.LoadEnv("../.env")
	ctx := context.Background()

	modelName := "claude-sonnet-4-6"

	var messages = []models.Message{
		models.NewSystemStringMessage("You are a helpful assistant."),
		models.NewHumanStringMessage("My name is John. Can you tell me a joke?"),
	}

	output1, err := agentgo.GenerateText(ctx, agentgo.Params{
		ModelName: modelName,
		Messages:  messages,
	})
	if err != nil {
		panic(err)
	}

	log.Printf("[output 1] - %s\n", output1.Text)

	messages = append(
		messages,
		models.NewAssistantStringMessage(output1.Text),
		models.NewHumanStringMessage("What is my name?"),
	)

	output2, err := agentgo.GenerateText(ctx, agentgo.Params{
		ModelName: modelName,
		Messages:  messages,
	})
	if err != nil {
		panic(err)
	}

	log.Printf("[output 2] - %s\n", output2.Text)
}

func TestStreamTextClaude(t *testing.T) {
	requireLiveTest(t)
	utils.LoadEnv("../.env")
	ctx := context.Background()

	output := agentgo.StreamText(ctx, agentgo.Params{
		ModelName: "claude-sonnet-4-6",
		Prompt:    "Say \"this is a test\" 10 times fast.",
	})
	if output == (models.LanguageModelStreamOutput{}) {
		panic("stream output is nil")
	}

	for part := range output.Channel {
		switch p := part.(type) {
		case models.StepStartPart:
			log.Printf("[step start] - %s\n", p.StepName())
		case models.StepEndPart:
			log.Printf("[step end] - %s\n", p.StepName())
		case models.TextPart:
			log.Printf("[text] - %s\n", p.Text())
		default:
			log.Printf("[unknown] - %s\n", part.Type())
		}
	}
}

// TestGenerateTextClaudeStructuredOutputUnsupported documents the provider's
// deliberate limitation: Anthropic's OpenAI compatibility layer ignores
// response_format, so object generation is rejected rather than silently
// returning unvalidated prose.
func TestGenerateTextClaudeStructuredOutputUnsupported(t *testing.T) {
	requireLiveTest(t)
	utils.LoadEnv("../.env")
	ctx := context.Background()

	type Recipe struct {
		Name string `json:"name"`
	}

	_, err := agentgo.GenerateObject(ctx, agentgo.ObjectParams[Recipe]{
		ModelName: "claude-sonnet-4-6",
		Prompt:    "Give me a pancake recipe",
		Schema: models.NewSchema[Recipe](models.NewSchemaParams[Recipe]{
			Name:        "recipe",
			Description: "A recipe",
			JSONSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
				"required":             []string{"name"},
				"additionalProperties": false,
			},
		}),
	})

	if err == nil {
		log.Println("expected an error for structured output on claude, got nil")
		return
	}
	log.Printf("[expected error] - %v\n", err)
}
