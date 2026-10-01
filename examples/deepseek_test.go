package main

import (
	"context"
	"log"
	"testing"

	"github.com/nquangtrung/agentgo"
	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/utils"
)

// Live DeepSeek tests. These make real billable calls and need DEEPSEEK_API_KEY
// in ../.env. They are skipped unless AGENTGO_LIVE_TESTS=1 — see live_test.go.

func TestGenerateTextDeepSeek(t *testing.T) {
	requireLiveTest(t)
	utils.LoadEnv("../.env")
	ctx := context.Background()

	output, err := agentgo.GenerateText(ctx, agentgo.Params{
		ModelName: "deepseek-flash",
		Prompt:    "Say this is a test",
	})
	if err != nil {
		panic(err)
	}

	log.Printf("[output] - %s\n", output.ModelName)
	log.Printf("Text: %s\n", output.Text)
	log.Printf("Input Tokens: %d\n", output.Usage.InputTokens)
	log.Printf("Output Tokens: %d\n", output.Usage.OutputTokens)
	log.Printf("Cached Tokens: %d\n", output.Usage.InputTokensDetails.CachedTokens)
	log.Printf("Reasoning Tokens: %d\n", output.Usage.OutputTokensDetails.ReasoningTokens)
	log.Printf("Total Tokens: %d\n", output.Usage.TotalTokens)
}

func TestGenerateTextDeepSeekWithInput(t *testing.T) {
	requireLiveTest(t)
	utils.LoadEnv("../.env")
	ctx := context.Background()

	modelName := "deepseek-flash"

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

	log.Printf("[output 1] - %s\n", output1.ModelName)
	log.Printf("Text: %s\n", output1.Text)

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

	log.Printf("[output 2] - %s\n", output2.ModelName)
	log.Printf("Text: %s\n", output2.Text)
}

func TestStreamTextDeepSeek(t *testing.T) {
	requireLiveTest(t)
	utils.LoadEnv("../.env")
	ctx := context.Background()

	output := agentgo.StreamText(ctx, agentgo.Params{
		ModelName: "deepseek-flash",
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
