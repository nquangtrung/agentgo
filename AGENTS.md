## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, use the installed graphify skill or instructions before doing anything else.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).

## Project overview

**AgentGo** is an educational Go SDK for interacting with multiple AI providers (OpenAI, Gemini, Claude) through a common interface. Key architectural patterns:

- **Provider Interface** (providers/provider.go): All providers must implement `AgentProvider` with `Context()` and `GenerateText(prompt string)`
- **Factory Pattern** (factory.go): `CreateAgentProvider()` maps model names (gpt-*, gemini-*, claude-*) to implementations
- **FSM Orchestration** (fsm/): Finite state machine drives AI interactions through defined state transitions (StartState → TextGeneration → ToolResolve → EndState)
- **Streaming** (stream_text.go): Async alternative using `LanguageModelStreamOutput` with channels for real-time parts

## Testing

```bash
# Full test suite
go test ./...

# Single package (e.g., providers/openai)
go test ./providers/openai

# Single test by name
go test -run TestGenerateText ./

# Verbose output
go test -v ./...
```

Tests use testify assertions, uber/mock for mocks (generated, pre-committed in mocks/), and context.Background() for test contexts. Mocks are already generated; do not regenerate them unless interface signatures change.

## Environment & .env

- SDK requires `.env` file in root with `OPENAI_API_KEY`, `GEMINI_API_KEY`, `CLAUDE_API_KEY`
- Loaded by `utils.LoadEnv()`, panics if missing (intentional for now)
- Each test sets up mocks to avoid real API calls; examples in examples/ show live usage

## Adding a provider

1. Create `providers/<name>/<name>.go` implementing `AgentProvider`
2. Add model prefix detection to `FindSupportedModel()` in factory.go (e.g., `strings.HasPrefix(modelName, "mistral")`)
3. Add case in `CreateAgentProvider()` to instantiate it
4. Add API key case in `LoadAPIKeyFromEnv()`
5. Update factory.go `ModelType` const and test coverage

## Context patterns

- Execution flows through `context.Context` with typed values: `models.ProviderContextKey`, `models.MachineContextKey`, `models.EndConditionsContextKey`, `models.ToolsContextKey`, `models.StreamContextKey`, `models.PartEmitterContextKey`
- Pass context.Background() in tests; production callers provide cancellable contexts (e.g., from request handlers)
- Do not use context.TODO() except in experimental code

## Error handling

Custom error types in models/error.go: `UnsupportedModelError`, `ExecutionContextError`, `ToolExecutionError`, `ToolNotFoundError`
- Use these instead of generic errors for clear error semantics
- `utils.Must()` panics on error (tests only; avoid in production code)

## Quirks

- FSM logs state transitions to stdout at log level INFO (see generate_text_test.go output); intentional for now, suppress if needed for CI
- Streaming mode yields `models.Part` objects (text, tool, step) via channel; non-streaming collects all in `LanguageModelOutput`
- Provider Context is immutable per execution; stored in `models.LanguageModelContext` (models/llm.go)
