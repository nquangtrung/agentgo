# Graph Report - agentgo  (2026-08-31)

## Corpus Check
- 82 files · ~40,915 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 612 nodes · 1048 edges · 51 communities (39 shown, 12 thin omitted)
- Extraction: 96% EXTRACTED · 4% INFERRED · 0% AMBIGUOUS · INFERRED: 38 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `44173c78`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- MockAgentProvider
- context.Context
- testing.T
- Part
- ToolExecutionsArchive
- OpenAIProvider
- MockAsPart
- MockopenAIResponsesService
- MockEndPart
- AgentGo - AI SDK Clone in Go 🚀
- error.go
- How to Make an AI SDK Clone in Golang - Part 4 - Tool Calling Loops with FSM
- slice.go
- trontria.com/agentgo
- Part 2: Unit Testing - The Right Way
- What You Must Do When Invoked
- What You Must Do When Invoked
- How to Make an AI SDK Clone in Golang - Part 2 - StreamText & Polymorphic Parts
- MockPart
- MockTool
- graphify reference: extra exports and benchmark
- MockasFunctionCaller
- graphify reference: extra exports and benchmark
- graphify reference: query, path, explain
- go.uber.org/mock/gomock.Controller
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
- Message
- .Execute
- LanguageModelUsage
- LanguageModelContext
- part_step.go
- go.uber.org/mock/gomock.Call
- BaseEndPart
- BaseStartPart
- opencode.json

## God Nodes (most connected - your core abstractions)
1. `Part` - 30 edges
2. `LanguageModelContext` - 21 edges
3. `ToolExecutionsArchive` - 20 edges
4. `ToolExecuteOutput` - 19 edges
5. `Message` - 18 edges
6. `LanguageModelUsage` - 18 edges
7. `GenerateText()` - 17 edges
8. `StreamText()` - 16 edges
9. `How to Make an AI SDK Clone in Golang - Part 4 - Tool Calling Loops with FSM` - 16 edges
10. `AgentContext` - 15 edges

## Surprising Connections (you probably didn't know these)
- `resolveProviderFromParams()` --calls--> `CreateAgentProvider()`  [INFERRED]
  resolver.go → factory.go
- `GenerateText()` --calls--> `mustResolveProviderFromParams()`  [INFERRED]
  generate_text.go → resolver.go
- `GenerateText()` --calls--> `resolveMessages()`  [INFERRED]
  generate_text.go → resolver.go
- `StreamText()` --calls--> `resolveMessages()`  [INFERRED]
  stream_text.go → resolver.go
- `StreamText()` --calls--> `mustResolveProviderFromParams()`  [INFERRED]
  stream_text.go → resolver.go

## Import Cycles
- None detected.

## Communities (51 total, 12 thin omitted)

### Community 1 - "context.Context"
Cohesion: 0.06
Nodes (30): AfterTextGenerationState, AgentContext, EndState, FSM, FSM[T], T, New(), PredicateState (+22 more)

### Community 2 - "testing.T"
Cohesion: 0.10
Nodes (35): TestGenerateTextOpenAI(), TestGenerateTextOpenAIWithInput(), TestStreamTextOpenAI(), TestStreamTextOpenAIWithInput(), TestGenerateTextWithToolOpenAI(), GenerateText(), TestGenerateText(), TestGenerateTextMultipleToolCalls() (+27 more)

### Community 3 - "Part"
Cohesion: 0.18
Nodes (16): PartEmitter, NewEmptyPartEmitter(), NewPartEmitter(), EndPart, Part, StepEndPart, StepStartPart, ToolErrorPart (+8 more)

### Community 4 - "ToolExecutionsArchive"
Cohesion: 0.09
Nodes (22): Params, canProceedToNextStep(), executeToolCall(), resolveToolFromToolCall(), sync.Mutex, EndCondition, MaxStepsEndCondition, ToolExecutionsArchive (+14 more)

### Community 5 - "OpenAIProvider"
Cohesion: 0.18
Nodes (12): AgentProviderFactoryParams, ModelType, CreateAgentProvider(), FindSupportedModel(), LoadAPIKeyFromEnv(), OpenAIProvider, openAIResponsesService, NewOpenAIProvider() (+4 more)

### Community 6 - "MockAsPart"
Cohesion: 0.20
Nodes (5): NewMockAsPart(), MockAsPart, BaseTextPart, ContentPart, TextPart

### Community 7 - "MockopenAIResponsesService"
Cohesion: 0.43
Nodes (3): MockopenAIResponsesService, MockopenAIResponsesServiceMockRecorder, NewMockopenAIResponsesService()

### Community 8 - "MockEndPart"
Cohesion: 0.31
Nodes (3): NewMockEndPart(), MockEndPart, MockEndPartMockRecorder

### Community 9 - "AgentGo - AI SDK Clone in Go 🚀"
Cohesion: 0.05
Nodes (38): 1. **Extensibility**, 2. **Testability**, 3. **Separation of Concerns**, 4. **Flexible API**, 5. **Clean Code**, How to Make an AI SDK Clone in Golang - Part 1 - GenerateText, Meet the OpenAI Provider, The Base Implementation: Avoiding Repetition (+30 more)

### Community 10 - "error.go"
Cohesion: 0.22
Nodes (4): ExecutionContextError, ToolExecutionError, ToolNotFoundError, UnsupportedModelError

### Community 11 - "How to Make an AI SDK Clone in Golang - Part 4 - Tool Calling Loops with FSM"
Cohesion: 0.07
Nodes (29): 1. **Explicit State Transitions**, 1. StartState: Initialization, 2. **Single Responsibility**, 2. StepStartState: Mark Step Boundary, 3. **Extensibility**, 3. PredicateState: The Decision Point ⚡, 4. **Debuggability**, 4. ToolResolveState: Execute Tool Calls (+21 more)

### Community 12 - "slice.go"
Cohesion: 0.48
Nodes (6): U, Each(), Filter(), T, Map(), Reduce()

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
Cohesion: 0.19
Nodes (3): NewMockPart(), MockPart, MockPartMockRecorder

### Community 20 - "MockTool"
Cohesion: 0.27
Nodes (3): NewMockTool(), MockTool, MockToolMockRecorder

### Community 21 - "graphify reference: extra exports and benchmark"
Cohesion: 0.22
Nodes (8): graphify reference: extra exports and benchmark, Step 6b - Wiki (only if --wiki flag), Step 7 - Neo4j export (only if --neo4j or --neo4j-push flag), Step 7a - FalkorDB export (only if --falkordb or --falkordb-push flag), Step 7b - SVG export (only if --svg flag), Step 7c - GraphML export (only if --graphml flag), Step 7d - MCP server (only if --mcp flag), Step 8 - Token reduction benchmark (only if total_words > 5000)

### Community 22 - "MockasFunctionCaller"
Cohesion: 0.36
Nodes (4): github.com/openai/openai-go/v3/responses.ResponseFunctionToolCall, MockasFunctionCaller, MockasFunctionCallerMockRecorder, NewMockasFunctionCaller()

### Community 23 - "graphify reference: extra exports and benchmark"
Cohesion: 0.22
Nodes (8): graphify reference: extra exports and benchmark, Step 6b - Wiki (only if --wiki flag), Step 7 - Neo4j export (only if --neo4j or --neo4j-push flag), Step 7a - FalkorDB export (only if --falkordb or --falkordb-push flag), Step 7b - SVG export (only if --svg flag), Step 7c - GraphML export (only if --graphml flag), Step 7d - MCP server (only if --mcp flag), Step 8 - Token reduction benchmark (only if total_words > 5000)

### Community 24 - "graphify reference: query, path, explain"
Cohesion: 0.33
Nodes (5): For /graphify explain, For /graphify path, graphify reference: query, path, explain, Step 0 — Constrained query expansion (REQUIRED before traversal), Step 1 — Traversal

### Community 25 - "go.uber.org/mock/gomock.Controller"
Cohesion: 0.48
Nodes (4): go.uber.org/mock/gomock.Controller, NewMockStartPart(), MockStartPart, MockStartPartMockRecorder

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

### Community 41 - "Message"
Cohesion: 0.20
Nodes (16): AccumulateToolCallResult(), BaseMessage, BaseMessageContent, Message, NewAssistantStringMessage(), NewHumanStringMessage(), NewMessageFromToolResult(), NewStringMessage() (+8 more)

### Community 42 - ".Execute"
Cohesion: 0.23
Nodes (9): ToolResolveState, BaseToolErrorPart, BaseToolPart, BaseToolResultPart, BaseToolStartPart, TestToolErrorPart(), NewToolErrorPart(), NewToolResultPart() (+1 more)

### Community 43 - "LanguageModelUsage"
Cohesion: 0.24
Nodes (9): AccumulateUsage(), ContextKey, LanguageModelUsageInputTokensDetails, LanguageModelUsageOutputTokensDetails, LanguageModelStreamOutput, LanguageModelUsage, NewLanguageModelOutput(), NewLanguageModelStreamOutput() (+1 more)

### Community 44 - "LanguageModelContext"
Cohesion: 0.27
Nodes (7): BasePart, BaseProcessStartPart, LanguageModelContext, NewLanguageModelContext(), PartType, NewPart(), NewProcessStartPart()

### Community 45 - "part_step.go"
Cohesion: 0.33
Nodes (7): BaseStepEndPart, BaseStepErrorPart, BaseStepStartPart, NewStepEndPart(), NewStepErrorPart(), NewStepStartPart(), StepPartImpl

### Community 47 - "BaseEndPart"
Cohesion: 0.60
Nodes (5): BaseEndPart, BaseProcessEndPart, FinishReason, NewEndPart(), NewProcessEndPart()

### Community 49 - "opencode.json"
Cohesion: 0.50
Nodes (3): plugin, $schema, .opencode/plugins/graphify.js

## Knowledge Gaps
- **181 isolated node(s):** `$schema`, `.opencode/plugins/graphify.js`, `trontria.com/agentgo`, `MessageContent`, `ContextKey` (+176 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **12 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `LanguageModelContext` connect `LanguageModelContext` to `MockAgentProvider`, `testing.T`, `ToolExecutionsArchive`, `OpenAIProvider`, `.Execute`, `LanguageModelUsage`, `part_step.go`, `BaseEndPart`, `MockPart`?**
  _High betweenness centrality (0.040) - this node is a cross-community bridge._
- **Why does `Part` connect `Part` to `testing.T`, `MockAsPart`, `LanguageModelUsage`, `LanguageModelContext`, `BaseEndPart`, `BaseStartPart`?**
  _High betweenness centrality (0.033) - this node is a cross-community bridge._
- **Why does `ToolExecuteOutput` connect `ToolExecutionsArchive` to `context.Context`, `.Execute`, `LanguageModelUsage`, `Message`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **What connects `$schema`, `.opencode/plugins/graphify.js`, `trontria.com/agentgo` to the rest of the system?**
  _181 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `context.Context` be split into smaller, more focused modules?**
  _Cohesion score 0.05902980713033314 - nodes in this community are weakly interconnected._
- **Should `testing.T` be split into smaller, more focused modules?**
  _Cohesion score 0.09951690821256039 - nodes in this community are weakly interconnected._
- **Should `ToolExecutionsArchive` be split into smaller, more focused modules?**
  _Cohesion score 0.08970099667774087 - nodes in this community are weakly interconnected._