package main

import (
	"context"
	"log"

	"trontria.com/agentgo"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/utils"

	"testing"
)

func TestGenerateTextOpenAI(t *testing.T) {
	utils.LoadEnv("../.env")
	ctx := context.Background()

	output, err := agentgo.GenerateText(ctx, agentgo.Params{
		ModelName: "gpt-5-mini",
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
	log.Printf("Cache Write Tokens: %d\n", output.Usage.InputTokensDetails.CacheWriteTokens)
	log.Printf("Reasoning Tokens: %d\n", output.Usage.OutputTokensDetails.ReasoningTokens)
	log.Printf("Total Tokens: %d\n", output.Usage.TotalTokens)
}

func TestGenerateTextOpenAIWithInput(t *testing.T) {
	utils.LoadEnv("../.env")
	ctx := context.Background()

	modelName := "gpt-5-mini"

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
	log.Printf("Input Tokens: %d\n", output1.Usage.InputTokens)
	log.Printf("Output Tokens: %d\n", output1.Usage.OutputTokens)
	log.Printf("Cached Tokens: %d\n", output1.Usage.InputTokensDetails.CachedTokens)
	log.Printf("Cache Write Tokens: %d\n", output1.Usage.InputTokensDetails.CacheWriteTokens)
	log.Printf("Reasoning Tokens: %d\n", output1.Usage.OutputTokensDetails.ReasoningTokens)
	log.Printf("Total Tokens: %d\n", output1.Usage.TotalTokens)

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
	log.Printf("Input Tokens: %d\n", output2.Usage.InputTokens)
	log.Printf("Output Tokens: %d\n", output2.Usage.OutputTokens)
	log.Printf("Cached Tokens: %d\n", output2.Usage.InputTokensDetails.CachedTokens)
	log.Printf("Cache Write Tokens: %d\n", output2.Usage.InputTokensDetails.CacheWriteTokens)
	log.Printf("Reasoning Tokens: %d\n", output2.Usage.OutputTokensDetails.ReasoningTokens)
	log.Printf("Total Tokens: %d\n", output2.Usage.TotalTokens)
}

func TestStreamTextOpenAI(t *testing.T) {
	utils.LoadEnv("../.env")
	ctx := context.Background()
	output := agentgo.StreamText(ctx, agentgo.Params{
		ModelName: "gpt-5-mini",
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

func TestStreamTextOpenAIWithInput(t *testing.T) {
	utils.LoadEnv("../.env")
	ctx := context.Background()
	modelName := "gpt-5-mini"

	var messages = []models.Message{
		models.NewSystemStringMessage("You are a helpful assistant."),
		models.NewHumanStringMessage("My name is John. Can you tell me a joke?"),
	}

	output1 := agentgo.StreamText(ctx, agentgo.Params{
		ModelName: modelName,
		Messages:  messages,
	})
	text1 := ""
	text2 := ""
	if output1 == (models.LanguageModelStreamOutput{}) {
		panic("stream output is nil")
	}
	for part := range output1.Channel {
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

	messages = append(
		messages,
		models.NewAssistantStringMessage(text1),
		models.NewHumanStringMessage("What is my name?"),
	)

	output2 := agentgo.StreamText(ctx, agentgo.Params{
		ModelName: modelName,
		Messages:  messages,
	})
	if output2 == (models.LanguageModelStreamOutput{}) {
		panic("stream output is nil")
	}
	for part := range output2.Channel {
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

	log.Printf("[output 1] - %s\n", output1.ModelName)
	log.Printf("Text: %s\n", text1)
	log.Printf("[output 2] - %s\n", output2.ModelName)
	log.Printf("Text: %s\n", text2)
}
