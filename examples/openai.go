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

	var provider providers.AgentProvider = providers.OpenAIProvider{APIKey: apiKey}
	output, err := agentgo.GenerateText(provider, "Say this is a test")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output.Text)
}
