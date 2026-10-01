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

**AgentGo** is an educational Go SDK for interacting with multiple AI providers (OpenAI, DeepSeek, Gemini, Claude) through a common interface. Key architectural patterns:

- **Provider Interface** (providers/provider.go): All providers must implement `AgentProvider` with `Context()` and `GenerateText(prompt string)`
- **Factory Pattern** (factory.go): `CreateAgentProvider()` maps model name prefixes (gpt-*, deepseek-*, gemini-*, claude-*) to implementations
- **FSM Orchestration** (fsm/): Finite state machine drives AI interactions through defined state transitions (StartState → TextGeneration → ToolResolve → EndState)
- **Streaming** (stream_text.go): Async alternative using `LanguageModelStreamOutput` with channels for real-time parts
- **Object Generation** (generate_object.go, stream_object.go): `GenerateObject[T]` / `StreamObject[T]` produce schema-validated objects with a graph-level repair loop

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

### Live example tests — do not run unless asked

Every test in `examples/` calls a real provider API and spends credits. They are all gated behind the `AGENTGO_LIVE_TESTS` env var and skipped by default, so `go test ./...` is offline and free.

- **Never set `AGENTGO_LIVE_TESTS` on your own.** Run live tests only when the user explicitly asks for them.
- To run when asked: `AGENTGO_LIVE_TESTS=1 go test -v ./examples/ -run TestGenerateTextDeepSeek`
- The guard is `requireLiveTest(t)` from `examples/live_test.go`, and it must stay the first statement in each live test — before `utils.LoadEnv("../.env")`, which calls `log.Fatal` when `.env` is missing.
- New live examples need that same first-line guard. Do not add unguarded live calls.

## Environment & .env

- SDK requires `.env` file in root with keys for the providers you use: `OPENAI_API_KEY`, `DEEPSEEK_API_KEY`, `GEMINI_API_KEY`, `CLAUDE_API_KEY`
- Loaded by `utils.LoadEnv()`, panics if missing (intentional for now)
- Each test sets up mocks to avoid real API calls; examples in examples/ show live usage

## Adding a provider

1. Create `providers/<name>/<name>.go` implementing `AgentProvider`
2. Add model prefix detection to `FindSupportedModel()` in factory.go (e.g., `strings.HasPrefix(modelName, "mistral")`)
3. Add case in `CreateAgentProvider()` to instantiate it
4. Add API key case in `LoadAPIKeyFromEnv()`
5. Update factory.go `ModelType` const and test coverage
6. Honor `params.ResponseFormat` in `GenerateText`/`StreamText` when the provider supports structured output (see `providers/openai/converters.go: convertResponseFormat`)
7. Add a `*_test.go` next to the provider with gomock-based unit tests, plus a `factory_test.go` case for the new model prefix

### Provider status and API families

All four providers are implemented, but they split across two different wire formats. Both are spoken with the `openai-go/v3` client, just against different services and base URLs:

- **Responses API** — `providers/openai` (default base URL) and `providers/deepseek` (`https://api.deepseek.com`). Use the `responses` service and `responses.ResponseNewParams`. Stream termination differs from OpenAI: DeepSeek ends with a `response.completed` event and sends no `[DONE]` sentinel.
- **Chat Completions** — `providers/gemini` (`https://generativelanguage.googleapis.com/v1beta/openai/`) and `providers/claude` (`https://api.anthropic.com/v1/`). Use `client.Chat.Completions` and `openai.ChatCompletionNewParams`. Streaming requires `StreamOptions.IncludeUsage` or the stream carries no token counts.

Each provider package is deliberately self-contained and duplicates its converters rather than sharing a package with the others. Follow that convention when adding a provider.

### Known provider gaps

- **Claude cannot do structured output.** Anthropic's OpenAI compatibility layer documents `response_format` as ignored, so `providers/claude/converters.go: convertResponseFormat` returns a `*models.StructuredOutputUnsupportedError` instead of silently dropping the schema. `GenerateObject`/`StreamObject` therefore fail on Claude by design. Supporting it would mean adding the native `anthropic-sdk-go` dependency and implementing the Messages API.
- **Claude reports no usage details.** Both `usage.prompt_tokens_details` and `usage.completion_tokens_details` are documented as always empty, so cached and reasoning token counts are always zero.
- **Tool calls without an id are dropped** by the Chat Completions converters (claude, gemini). A result that references a call the provider did not identify cannot be correlated, so replaying it would leave the API rejecting the pair. The Responses API converters (openai, deepseek) send such calls as uncorrelatable items instead.

## Message model

`models.Message` content is a slice of typed parts (`models/message_content.go`), not a text blob:

- `TextContentPart` — plain text
- `ToolCallContentPart` — the model's request, carrying `ID()` (the provider-assigned call id), `Name()`, `Input()`
- `ToolResultContentPart` — the execution outcome, carrying `ToolCallID()` that echoes the originating call's id

`Content().Text()` still concatenates the text parts and returns `""` when there are none, so text-only code paths are unaffected. Use `Content().ToolCalls()` / `Content().ToolResults()` / `Content().Parts()` for structured access.

`MessageRoleTool` carries results. Anthropic models these as `tool_result` blocks on a user turn, but they are results rather than user input, so each provider maps the role onto its own wire format.

A tool round-trip appends **three** messages: the human prompt, an assistant turn recording the model's calls (`models.NewAssistantToolCallsMessage`, built in `resolveTool`), then one tool message per result (`archiveToolResult`). Replaying the call is what lets Anthropic's native Messages API validate the pairing; it errors on a `tool_result` with no matching preceding `tool_use`.

The end conditions in `endconditions/` are unaffected — they read the tool archive, not messages.

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
