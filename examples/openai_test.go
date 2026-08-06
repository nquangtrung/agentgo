package main

import (
	"trontria.com/agentgo"
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
