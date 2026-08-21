package main

import (
	"log"

	"trontria.com/agentgo"
	"trontria.com/agentgo/models"
	"trontria.com/agentgo/utils"

	"testing"
)

func TestGenerateTextWithToolOpenAI(t *testing.T) {
	utils.LoadEnv("../.env")
	output, err := agentgo.GenerateText(agentgo.Params{
		ModelName: "gpt-5-mini",
		Prompt:    "Say this is a test",
		EndConditions: []agentgo.EndCondition{
			agentgo.NewMaxStepsEndCondition(5),
		},
		Tools: []models.BaseTool{
			models.NewTool(models.NewToolParams{
				Name:        "get_temperature",
				Description: "Get temperature for a specific location",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{"type": "string", "description": "Return the temperature for this location."},
					},
					"required":             []string{"location"},
					"additionalProperties": false,
				},
				Fn: func(params models.ToolExecuteParams) models.ToolExecuteOutput {
					location := params.Input["location"].(string)
					// Simulate getting the temperature for the location
					temperature := "25°C" // Replace with actual logic to get temperature
					return models.ToolExecuteOutput{
						Output: map[string]any{
							"location":    location,
							"temperature": temperature,
						},
					}
				},
			}),
		},
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
