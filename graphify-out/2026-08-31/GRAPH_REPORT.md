# Graph Report - agentgo  (2026-08-28)

## Corpus Check
- 65 files · ~36,358 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 513 nodes · 869 edges · 41 communities (31 shown, 10 thin omitted)
- Extraction: 96% EXTRACTED · 4% INFERRED · 0% AMBIGUOUS · INFERRED: 38 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `de4ba562`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- go.uber.org/mock/gomock.Call
- Params
- testing.T
- Part
- ExecutionContext
- LanguageModelContext
- MockAsPart
- MockopenAIResponsesService
- MockEndPart
- AgentGo - AI SDK Clone in Go 🚀
- error.go
- CreateAgentProvider
- Reduce
- trontria.com/agentgo
- Part 2: Unit Testing - The Right Way
- What You Must Do When Invoked
- What You Must Do When Invoked
- How to Make an AI SDK Clone in Golang - Part 2 - StreamText & Polymorphic Parts
- MockPart
- MockTool
- graphify reference: extra exports and benchmark
- go.uber.org/mock/gomock.Controller
- graphify reference: extra exports and benchmark
- graphify reference: query, path, explain
- MockStartPart
- graphify reference: query, path, explain
- graphify reference: add a URL and watch a folder
- graphify reference: commit hook and native CLAUDE.md integration
- graphify reference: incremental update and cluster-only
- graphify reference: add a URL and watch a folder
- graphify reference: commit hook and native CLAUDE.md integration
- graphify reference: incremental update and cluster-only
- graphify reference: GitHub clone and cross-repo merge
- graphify reference: transcribe video and audio
- graphify.js
- graphify reference: GitHub clone and cross-repo merge
- graphify reference: transcribe video and audio
- AGENTS.md
- .copilot/skills/graphify/references/extraction-spec.md
- .opencode/skills/graphify/references/extraction-spec.md

## God Nodes (most connected - your core abstractions)
1. `Part` - 25 edges
2. `LanguageModelContext` - 19 edges
3. `ExecutionContext` - 17 edges
4. `Message` - 16 edges
5. `ToolExecuteOutput` - 16 edges
6. `Params` - 15 edges
7. `AgentGo - AI SDK Clone in Go 🚀` - 14 edges
8. `GenerateText()` - 13 edges
9. `BaseTool` - 13 edges
10. `How to Make an AI SDK Clone in Golang - Part 2 - StreamText & Polymorphic Parts` - 13 edges

## Surprising Connections (you probably didn't know these)
- `resolveProviderFromParams()` --calls--> `CreateAgentProvider()`  [INFERRED]
  resolver.go → factory.go
- `TestGenerateText()` --calls--> `GenerateText()`  [INFERRED]
  generate_text_test.go → generate_text.go
- `TestGenerateTextWithTool()` --calls--> `GenerateText()`  [INFERRED]
  generate_text_test.go → generate_text.go
- `TestGenerateTextOpenAI()` --calls--> `GenerateText()`  [EXTRACTED]
  examples/openai_test.go → generate_text.go
- `TestGenerateTextOpenAIWithInput()` --calls--> `GenerateText()`  [EXTRACTED]
  examples/openai_test.go → generate_text.go

## Import Cycles
- None detected.

## Communities (41 total, 10 thin omitted)

### Community 0 - "go.uber.org/mock/gomock.Call"
Cohesion: 0.19
Nodes (4): go.uber.org/mock/gomock.Call, MockAgentProvider, MockAgentProviderMockRecorder, MockAsPartMockRecorder

### Community 1 - "Params"
Cohesion: 0.08
Nodes (41): Params, canProceedToNextStep(), doLoop(), GenerateText(), context.Context, github.com/openai/openai-go/v3/responses.ResponseNewParamsInputUnion, github.com/openai/openai-go/v3/responses.ToolUnionParam, ContextKey (+33 more)

### Community 2 - "testing.T"
Cohesion: 0.09
Nodes (35): EndCondition, MaxStepsEndCondition, TestGenerateTextOpenAI(), TestGenerateTextOpenAIWithInput(), TestStreamTextOpenAI(), TestStreamTextOpenAIWithInput(), TestGenerateTextWithToolOpenAI(), TestGenerateText() (+27 more)

### Community 3 - "Part"
Cohesion: 0.10
Nodes (31): time.Time, BaseStepEndPart, BaseStepErrorPart, BaseStepStartPart, BaseToolErrorPart, BaseToolPart, BaseToolResultPart, BaseToolStartPart (+23 more)

### Community 4 - "ExecutionContext"
Cohesion: 0.16
Nodes (9): sync.Mutex, ExecutionContext, NewToolParams, Step, BaseTool, Tool, ToolExecuteOutput, ToolExecuteParams (+1 more)

### Community 5 - "LanguageModelContext"
Cohesion: 0.12
Nodes (13): BasePart, BaseTextPart, ContentPart, LanguageModelContext, PartType, NewPart(), TextPart, OpenAIProvider (+5 more)

### Community 7 - "MockopenAIResponsesService"
Cohesion: 0.21
Nodes (7): github.com/openai/openai-go/v3/packages/ssestream.Stream, github.com/openai/openai-go/v3/responses.Response, github.com/openai/openai-go/v3/responses.ResponseNewParams, github.com/openai/openai-go/v3/responses.ResponseStreamEventUnion, MockopenAIResponsesService, MockopenAIResponsesServiceMockRecorder, NewMockopenAIResponsesService()

### Community 8 - "MockEndPart"
Cohesion: 0.36
Nodes (3): NewMockEndPart(), MockEndPart, MockEndPartMockRecorder

### Community 9 - "AgentGo - AI SDK Clone in Go 🚀"
Cohesion: 0.05
Nodes (38): 1. **Extensibility**, 2. **Testability**, 3. **Separation of Concerns**, 4. **Flexible API**, 5. **Clean Code**, How to Make an AI SDK Clone in Golang - Part 1 - GenerateText, Meet the OpenAI Provider, The Base Implementation: Avoiding Repetition (+30 more)

### Community 10 - "error.go"
Cohesion: 0.22
Nodes (4): ExecutionContextError, ToolExecutionError, ToolNotFoundError, UnsupportedModelError

### Community 11 - "CreateAgentProvider"
Cohesion: 0.43
Nodes (6): AgentProviderFactoryParams, ModelType, CreateAgentProvider(), FindSupportedModel(), LoadAPIKeyFromEnv(), GetEnvVar()

### Community 12 - "Reduce"
Cohesion: 0.53
Nodes (5): U, Filter(), T, Map(), Reduce()

### Community 15 - "Part 2: Unit Testing - The Right Way"
Cohesion: 0.08
Nodes (24): 1. **Always Include System Messages When Setting Behavior**, 2. **Build Messages Incrementally**, 3. **Keep Message History Reasonable**, 4. **Test with Different Message Histories**, Best Practices for Messages, Building a Multi-Turn Conversation, Enter GoMock: Automatic Mocking, Go's Testing Philosophy (+16 more)

### Community 16 - "What You Must Do When Invoked"
Cohesion: 0.08
Nodes (24): For /graphify add and --watch, For /graphify query, For the commit hook and native CLAUDE.md integration, For --update and --cluster-only, /graphify, Honesty Rules, Interpreter guard for subcommands, Part A - Structural extraction for code files (+16 more)

### Community 17 - "What You Must Do When Invoked"
Cohesion: 0.08
Nodes (24): For /graphify add and --watch, For /graphify query, For the commit hook and native CLAUDE.md integration, For --update and --cluster-only, /graphify, Honesty Rules, Interpreter guard for subcommands, Part A - Structural extraction for code files (+16 more)

### Community 18 - "How to Make an AI SDK Clone in Golang - Part 2 - StreamText & Polymorphic Parts"
Cohesion: 0.10
Nodes (19): 1. **Non-blocking**, 2. **Memory Efficient**, 3. **Error Handling Ready**, 4. **Extensible**, 5. **Testable**, Digging Into Part Types, Go Concurrency: The Channel Pattern, Go's Concurrency Philosophy (+11 more)

### Community 19 - "MockPart"
Cohesion: 0.21
Nodes (3): NewMockPart(), MockPart, MockPartMockRecorder

### Community 20 - "MockTool"
Cohesion: 0.27
Nodes (3): NewMockTool(), MockTool, MockToolMockRecorder

### Community 21 - "graphify reference: extra exports and benchmark"
Cohesion: 0.22
Nodes (8): graphify reference: extra exports and benchmark, Step 6b - Wiki (only if --wiki flag), Step 7 - Neo4j export (only if --neo4j or --neo4j-push flag), Step 7a - FalkorDB export (only if --falkordb or --falkordb-push flag), Step 7b - SVG export (only if --svg flag), Step 7c - GraphML export (only if --graphml flag), Step 7d - MCP server (only if --mcp flag), Step 8 - Token reduction benchmark (only if total_words > 5000)

### Community 22 - "go.uber.org/mock/gomock.Controller"
Cohesion: 0.33
Nodes (5): github.com/openai/openai-go/v3/responses.ResponseFunctionToolCall, go.uber.org/mock/gomock.Controller, MockasFunctionCaller, MockasFunctionCallerMockRecorder, NewMockasFunctionCaller()

### Community 23 - "graphify reference: extra exports and benchmark"
Cohesion: 0.22
Nodes (8): graphify reference: extra exports and benchmark, Step 6b - Wiki (only if --wiki flag), Step 7 - Neo4j export (only if --neo4j or --neo4j-push flag), Step 7a - FalkorDB export (only if --falkordb or --falkordb-push flag), Step 7b - SVG export (only if --svg flag), Step 7c - GraphML export (only if --graphml flag), Step 7d - MCP server (only if --mcp flag), Step 8 - Token reduction benchmark (only if total_words > 5000)

### Community 24 - "graphify reference: query, path, explain"
Cohesion: 0.33
Nodes (5): For /graphify explain, For /graphify path, graphify reference: query, path, explain, Step 0 — Constrained query expansion (REQUIRED before traversal), Step 1 — Traversal

### Community 25 - "MockStartPart"
Cohesion: 0.53
Nodes (3): NewMockStartPart(), MockStartPart, MockStartPartMockRecorder

### Community 26 - "graphify reference: query, path, explain"
Cohesion: 0.33
Nodes (5): For /graphify explain, For /graphify path, graphify reference: query, path, explain, Step 0 — Constrained query expansion (REQUIRED before traversal), Step 1 — Traversal

### Community 27 - "graphify reference: add a URL and watch a folder"
Cohesion: 0.50
Nodes (3): For /graphify add, For --watch, graphify reference: add a URL and watch a folder

### Community 28 - "graphify reference: commit hook and native CLAUDE.md integration"
Cohesion: 0.50
Nodes (3): For git commit hook, For native CLAUDE.md integration, graphify reference: commit hook and native CLAUDE.md integration

### Community 29 - "graphify reference: incremental update and cluster-only"
Cohesion: 0.50
Nodes (3): For --cluster-only, For --update (incremental re-extraction), graphify reference: incremental update and cluster-only

### Community 30 - "graphify reference: add a URL and watch a folder"
Cohesion: 0.50
Nodes (3): For /graphify add, For --watch, graphify reference: add a URL and watch a folder

### Community 31 - "graphify reference: commit hook and native CLAUDE.md integration"
Cohesion: 0.50
Nodes (3): For git commit hook, For native CLAUDE.md integration, graphify reference: commit hook and native CLAUDE.md integration

### Community 32 - "graphify reference: incremental update and cluster-only"
Cohesion: 0.50
Nodes (3): For --cluster-only, For --update (incremental re-extraction), graphify reference: incremental update and cluster-only

## Knowledge Gaps
- **154 isolated node(s):** `trontria.com/agentgo`, `MessageContent`, `ContextKey`, `Usage`, `What graphify is for` (+149 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **10 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Part` connect `Part` to `Params`, `LanguageModelContext`?**
  _High betweenness centrality (0.044) - this node is a cross-community bridge._
- **Why does `LanguageModelContext` connect `LanguageModelContext` to `go.uber.org/mock/gomock.Call`, `Params`, `Part`, `ExecutionContext`, `MockPart`?**
  _High betweenness centrality (0.037) - this node is a cross-community bridge._
- **Why does `MockAgentProvider` connect `go.uber.org/mock/gomock.Call` to `Params`, `testing.T`, `go.uber.org/mock/gomock.Controller`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **What connects `trontria.com/agentgo`, `MessageContent`, `ContextKey` to the rest of the system?**
  _154 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Params` be split into smaller, more focused modules?**
  _Cohesion score 0.07987012987012987 - nodes in this community are weakly interconnected._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.0851063829787234 - nodes in this community are weakly interconnected._
- **Should `Part` be split into smaller, more focused modules?**
  _Cohesion score 0.09725158562367865 - nodes in this community are weakly interconnected._