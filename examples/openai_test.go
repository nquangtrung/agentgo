package main

import (
	"trontria.com/agentgo"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/utils"

	"fmt"
	"testing"
)

func TestGenerateTextOpenAI(t *testing.T) {
	utils.LoadEnv("../.env")
	output, err := agentgo.GenerateText(agentgo.GenerateTextParams{
		ModelName: "gpt-5-mini",
		Prompt:    "Say this is a test",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("[output] - %s\n", output.ModelName)
	fmt.Printf("Text: %s\n", output.Text)
	fmt.Printf("Input Tokens: %d\n", output.Usage.InputTokens)
	fmt.Printf("Output Tokens: %d\n", output.Usage.OutputTokens)
	fmt.Printf("Cached Tokens: %d\n", output.Usage.CachedTokens)
	fmt.Printf("Reasoning Tokens: %d\n", output.Usage.ReasoningTokens)
}

func TestStreamTextOpenAI(t *testing.T) {
	utils.LoadEnv("../.env")
	c := agentgo.StreamText(agentgo.StreamTextParams{
		ModelName: "gpt-5-mini",
		Prompt:    "Say this is a test 10 times fast.",
	})
	if c == nil {
		panic("channel is nil")
	}
	for part := range c {
		switch part.GetType() {
		case models.PartTypeStepStart:
			if stepStartPart, ok := models.ToStepStartPart(part); ok {
				fmt.Printf("[step start] - %s\n", stepStartPart.GetStepName())
			}
		case models.PartTypeStepEnd:
			if stepEndPart, ok := models.ToStepEndPart(part); ok {
				fmt.Printf("[step end] - %s\n", stepEndPart.GetStepName())
			}
		case models.PartTypeText:
			if textPart, ok := models.ToTextPart(part); ok {
				fmt.Printf("[text] - %s\n", textPart.GetText())
			}
		default:
			fmt.Printf("[unknown] - %s\n", part.GetType())
		}
	}
}
