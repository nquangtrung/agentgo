package main

import (
	"trontria.com/agentgo"
	"trontria.com/agentgo/providers"

	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		panic(".env file not found")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")

	var provider providers.AgentProvider = providers.OpenAIProvider{
		APIKey: apiKey,
		Model:  "gpt-5-mini",
	}
	output, err := agentgo.GenerateText(agentgo.GenerateTextParams{
		Provider: provider,
		Prompt:   "Say this is a test",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("[output] - %s\n", output.Model)
	fmt.Printf("Text: %s\n", output.Text)
	fmt.Printf("Input Tokens: %d\n", output.Usage.InputTokens)
	fmt.Printf("Output Tokens: %d\n", output.Usage.OutputTokens)
	fmt.Printf("Cached Tokens: %d\n", output.Usage.CachedTokens)
	fmt.Printf("Reasoning Tokens: %d\n", output.Usage.ReasoningTokens)
}
