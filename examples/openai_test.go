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

func TestGenerateTextOpenAIWithInput(t *testing.T) {
	utils.LoadEnv("../.env")

	modelName := "gpt-5-mini"

	var messages = []models.Message{
		models.NewSystemStringMessage("You are a helpful assistant."),
		models.NewHumanStringMessage("My name is John. Can you tell me a joke?"),
	}

	output1, err := agentgo.GenerateText(agentgo.GenerateTextParams{
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
	log.Printf("Cached Tokens: %d\n", output1.Usage.CachedTokens)
	log.Printf("Reasoning Tokens: %d\n", output1.Usage.ReasoningTokens)

	messages = append(
		messages,
		models.NewAssistantStringMessage(output1.Text),
		models.NewHumanStringMessage("What is my name?"),
	)
	output2, err := agentgo.GenerateText(agentgo.GenerateTextParams{
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
	log.Printf("Cached Tokens: %d\n", output2.Usage.CachedTokens)
	log.Printf("Reasoning Tokens: %d\n", output2.Usage.ReasoningTokens)
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
			if stepStartPart, ok := part.AsStepStartPart(); ok {
				log.Printf("[step start] - %s\n", stepStartPart.GetStepName())
			}
		case models.PartTypeStepEnd:
			if stepEndPart, ok := part.AsStepEndPart(); ok {
				log.Printf("[step end] - %s\n", stepEndPart.GetStepName())
			}
		case models.PartTypeText:
			if textPart, ok := part.AsTextPart(); ok {
				log.Printf("[text] - %s\n", textPart.GetText())
			}
		default:
			log.Printf("[unknown] - %s\n", part.GetType())
		}
	}
}

func TestStreamTextOpenAIWithInput(t *testing.T) {
	utils.LoadEnv("../.env")

	modelName := "gpt-5-mini"

	var messages = []models.Message{
		models.NewSystemStringMessage("You are a helpful assistant."),
		models.NewHumanStringMessage("My name is John. Can you tell me a joke?"),
	}

	output1 := agentgo.StreamText(agentgo.StreamTextParams{
		ModelName: modelName,
		Messages:  messages,
	})
	text1 := ""
	text2 := ""
	if output1 == (models.LanguageModelStreamOutput{}) {
		panic("stream output is nil")
	}
	for part := range output1.Channel {
		switch part.GetType() {
		case models.PartTypeStepStart:
			if stepStartPart, ok := part.AsStepStartPart(); ok {
				log.Printf("[step start] - %s\n", stepStartPart.GetStepName())
			}
		case models.PartTypeStepEnd:
			if stepEndPart, ok := part.AsStepEndPart(); ok {
				log.Printf("[step end] - %s\n", stepEndPart.GetStepName())
			}
		case models.PartTypeText:
			if textPart, ok := part.AsTextPart(); ok {
				log.Printf("[text] - %s\n", textPart.GetText())
				text1 += textPart.GetText()
			}
		default:
			log.Printf("[unknown] - %s\n", part.GetType())
		}
	}

	messages = append(
		messages,
		models.NewAssistantStringMessage(text1),
		models.NewHumanStringMessage("What is my name?"),
	)

	output2 := agentgo.StreamText(agentgo.StreamTextParams{
		ModelName: modelName,
		Messages:  messages,
	})
	if output2 == (models.LanguageModelStreamOutput{}) {
		panic("stream output is nil")
	}
	for part := range output2.Channel {
		switch part.GetType() {
		case models.PartTypeStepStart:
			if stepStartPart, ok := part.AsStepStartPart(); ok {
				log.Printf("[step start] - %s\n", stepStartPart.GetStepName())
			}
		case models.PartTypeStepEnd:
			if stepEndPart, ok := part.AsStepEndPart(); ok {
				log.Printf("[step end] - %s\n", stepEndPart.GetStepName())
			}
		case models.PartTypeText:
			if textPart, ok := part.AsTextPart(); ok {
				log.Printf("[text] - %s\n", textPart.GetText())
				text2 += textPart.GetText()
			}
		default:
			log.Printf("[unknown] - %s\n", part.GetType())
		}
	}

	log.Printf("[output 1] - %s\n", output1.ModelName)
	log.Printf("Text: %s\n", text1)
	log.Printf("[output 2] - %s\n", output2.ModelName)
	log.Printf("Text: %s\n", text2)
}
