package main

import (
	"log"

	"trontria.com/agentgo"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/utils"

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
	log.Printf("[output] - %s\n", output.ModelName)
	log.Printf("Text: %s\n", output.Text)
	log.Printf("Input Tokens: %d\n", output.Usage.InputTokens)
	log.Printf("Output Tokens: %d\n", output.Usage.OutputTokens)
	log.Printf("Cached Tokens: %d\n", output.Usage.CachedTokens)
	log.Printf("Reasoning Tokens: %d\n", output.Usage.ReasoningTokens)
}

func TestStreamTextOpenAI(t *testing.T) {
	utils.LoadEnv("../.env")
	output := agentgo.StreamText(agentgo.StreamTextParams{
		ModelName: "gpt-5-mini",
		Prompt:    "Say \"this is a test\" 10 times fast.",
	})
	if output == (models.LanguageModelStreamOutput{}) {
		panic("stream output is nil")
	}
	for part := range output.Channel {
		switch part.GetType() {
		case models.PartTypeStepStart:
			if stepStartPart, ok := models.ToStepStartPart(part); ok {
				log.Printf("[step start] - %s\n", stepStartPart.GetStepName())
			}
		case models.PartTypeStepEnd:
			if stepEndPart, ok := models.ToStepEndPart(part); ok {
				log.Printf("[step end] - %s\n", stepEndPart.GetStepName())
			}
		case models.PartTypeText:
			if textPart, ok := models.ToTextPart(part); ok {
				log.Printf("[text] - %s\n", textPart.GetText())
			}
		default:
			log.Printf("[unknown] - %s\n", part.GetType())
		}
	}
}
