# AgentGo - AI SDK Clone in Go 🚀

Welcome to **AgentGo**, an educational project where we build our own AI SDK in Go! This is a great way to learn Go best practices while creating something useful that can interact with multiple AI providers.

## What is AgentGo?

AgentGo is a clean, extensible SDK for working with different AI language models (OpenAI, Gemini, Claude, etc.) using a common interface. Instead of hardcoding OpenAI-specific logic throughout your app, you define a contract (an interface) that any AI provider can implement.

This project demonstrates professional Go patterns like:
- **Interface-based design** - Work with any provider through a common contract
- **Composition over inheritance** - Reuse code through embedding
- **Factory pattern** - Hide complexity behind simple functions
- **Separation of concerns** - Each provider handles its own details

## Quick Start

### Installation

```bash
# Clone the repository
git clone <repo-url>
cd agentgo

# Install dependencies
go mod download
```

### Setup API Keys

Create a `.env` file in the root directory:

```env
OPENAI_API_KEY=your-openai-api-key-here
GEMINI_API_KEY=your-gemini-api-key-here
CLAUDE_API_KEY=your-claude-api-key-here
```

### Basic Usage

```go
package main

import (
	"fmt"
	"trontria.com/agentgo"
)

func main() {
	// Generate text using a model
	output, err := agentgo.GenerateText(agentgo.GenerateTextParams{
		ModelName: "gpt-4",
		Prompt:    "What is Go good at?",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Response:", output.Text)
	fmt.Println("Model:", output.ModelName)
	fmt.Println("Input tokens:", output.Usage.InputTokens)
	fmt.Println("Output tokens:", output.Usage.OutputTokens)
}
```

That's it! Just pass a model name and a prompt. The SDK figures out which provider to use and handles the rest.

## How It Works

The magic is in the **Provider Pattern** + **Factory Pattern**:

1. **Define a contract** - The `AgentProvider` interface says what all providers must do
2. **Implement for each service** - OpenAI, Gemini, Claude each implement the interface differently
3. **Use the factory** - `CreateAgentProvider()` automatically creates the right provider
4. **Write once** - Your code works with any provider that follows the contract

```go
// Your code doesn't care which provider it's using!
output, _ := provider.GenerateText("Hello, AI!")
```

## Project Structure

```
agentgo/
├── models/           # Data structures (Output, Context, Usage, etc.)
├── providers/        # Provider implementations (OpenAI, etc.)
├── utils/            # Utilities (env loading, etc.)
├── examples/         # Example usage
├── blogs/            # Educational blog posts
├── main.go           # SDK entry point
└── go.mod            # Go module definition
```

## Learning Resources

We're building this project step-by-step with educational blog posts:

- **[Part 1: GenerateText](blogs/part1-generatetext.md)** - Learn about providers, interfaces, and the factory pattern
- **Part 2: Streaming Responses** (Coming soon!) - Handle long-running AI tasks
- **More parts coming** - Error handling, multiple providers, building agents, and more!

Each blog post explains the "why" behind the design decisions, not just the "how."

## Supported Providers

Currently implemented:
- ✅ **OpenAI** - All GPT models (gpt-3.5, gpt-4, etc.)

Coming soon:
- 🔜 **Google Gemini** - Google's language models
- 🔜 **Anthropic Claude** - Claude models
- 🔜 **Custom providers** - Implement your own!

## Examples

Check the `examples/` folder for complete working examples:

```bash
# Run the OpenAI example
go run examples/openai.go
```

## Architecture Highlights

### The Interface (The Contract)
```go
type AgentProvider interface {
	GetContext() models.LanguageModelContext
	GenerateText(prompt string) (models.LanguageModelOutput, error)
}
```

Any provider that implements these methods works with our SDK!

### The Base Implementation (No Repetition)
```go
type AgentProviderImpl struct {
	Context models.LanguageModelContext
}

func (p AgentProviderImpl) GetContext() models.LanguageModelContext {
	return p.Context
}
```

All providers embed this, so they don't repeat the common code.

### The Factory (Hide Complexity)
```go
provider, err := CreateAgentProvider(AgentProviderFactoryParams{
	ModelName: "gpt-4",
})
```

The factory detects which provider you need and sets everything up. No API key management in your business logic!

## Development

### Running Tests

```bash
go test ./...
```

### Adding a New Provider

To add support for a new AI service:

1. Create a new file in `providers/` (e.g., `gemini.go`)
2. Implement the `AgentProvider` interface
3. Add a case in `CreateAgentProvider()` to create it
4. Update `FindSupportedModel()` to detect your model names

That's it! Your new provider works everywhere automatically.

## Best Practices Used

This project demonstrates:
- ✨ Clean architecture with clear separation of concerns
- 🎯 SOLID principles (especially Interface Segregation)
- 🧪 Testable code (easy to mock providers)
- 📦 Go idiomatic patterns (interfaces, composition, embedding)
- 📝 Clear, maintainable code with good naming

These aren't theoretical concepts—they're what real-world Go codebases look like!

## Contributing

This is an educational project! Contributions are welcome:
- Add new providers
- Improve existing code
- Write more blog posts
- Add tests
- Submit issues or suggestions

## License

MIT - Feel free to use this for learning!

## Next Steps

1. Read [Part 1: GenerateText](blogs/part1-generatetext.md) to understand the architecture
2. Check out `examples/openai.go` to see it in action
3. Try adding your own provider!
4. Keep an eye out for Part 2 on streaming responses

Happy coding! 🎉
