# AgentGo Graph Package Skill

## Overview

The **@graph/** package implements a **Pregel-like state graph** for managing state transitions and concurrent node executions in AgentGo. It's a generic, type-safe framework for building directed acyclic graphs (DAGs) with support for:

- **State-driven workflows**: Nodes transform state and pass results downstream
- **Multiple edge types**: Simple edges, fan-out, and conditional routing
- **Interrupts & checkpointing**: Pause execution for human approval, resume later
- **Error handling**: Node, reducer, and router-level error recovery
- **Visualization**: Mermaid diagram generation
- **Worker nodes**: Special nodes for map/reduce-style processing

---

## Architecture

### Core Types

```go
type StateGraph[T any] struct
type ID = string
type NodeFn[T any] = func(ctx context.Context, state T) (T, error)
type WorkerNodeFn[T any] = func(ctx context.Context, state T, item I) (T, error)
type Router[T any] = func(state T) []Target
type Reducer[T any] = func(a, b T) T
type Target = string // or ID
type Checkpointer[T any] interface{} // checkpoint state for resumption
```

### Execution Model

1. **SuperSteps**: Synchronous rounds where all eligible nodes run concurrently
2. **Reducer**: Combines results from parallel node executions into a single state
3. **State Flow**: START → node(s) → [optional router] → next node(s) → END
4. **Thread ID**: Unique ID per invocation, used for checkpointing and interrupt resumption

### Special Nodes

- **START/END**: Built-in sentinel nodes (automatically created)
- **State Nodes**: Normal nodes that transform state
- **Worker Nodes**: Process items in state with parallelism, merge results via reducer

---

## Common Patterns

### 1. Creating a Graph

```go
// Define reducer (how to merge parallel results)
g := graph.New[int](func(a, b int) int {
    return a + b  // sum results from parallel nodes
})
```

**Key decisions:**
- Type parameter `T` is your state type (can be struct, slice, map, etc.)
- Reducer must be associative and commutative for deterministic results
- Reducer panics are caught and wrapped in `ReducerExecutionError`

---

### 2. Adding Nodes

#### Simple State Node
```go
g.AddNode("increment", func(ctx context.Context, state int) (int, error) {
    return state + 1, nil
})
```

**Pattern:** Each node receives full state, returns transformed state + error

#### Worker Node (Map/Reduce)
```go
type State struct {
    Value int
    Items []string
}

g.AddWorkerNode("processItems", func(ctx context.Context, state State, item string) (State, error) {
    state.Value += len(item)  // accumulate
    return state, nil
})
```

**Pattern:** Worker nodes iterate `state.Items`, call function per item, reducer merges results

---

### 3. Connecting Nodes

#### Simple Edge (Linear Flow)
```go
g.AddEdge(graph.START, "increment")
g.AddEdge("increment", "double")
g.AddEdge("double", graph.END)
```

#### Fan-Out (Parallel Execution)
```go
g.FanOut(graph.START, []graph.ID{"inc", "double", "triple"})
// All three run concurrently; results merged by reducer
g.AddEdge("inc", graph.END)
g.AddEdge("double", graph.END)
g.AddEdge("triple", graph.END)
```

#### Conditional Routing
```go
router := func(state int) []graph.Target {
    if state > 0 {
        return graph.IDs("positive")  // helper to convert []string to []Target
    }
    return graph.IDs("nonpositive")
}

g.AddConditionalEdge(graph.START, router, []graph.ID{"positive", "nonpositive"})
g.AddEdge("positive", graph.END)
g.AddEdge("nonpositive", graph.END)
```

**Pattern:** Router decides which next node(s) to execute based on state
- Router panics are caught and wrapped in `RouterExecutionError`
- Must return one or more targets that exist in the graph

#### Cycles & Loops
```go
// Create a loop: START → inc → condition → [back to inc or END]
g.AddEdge(graph.START, "inc")
g.AddConditionalEdge("inc", func(state int) []graph.Target {
    if state < 10 {
        return graph.IDs("inc")  // loop back
    }
    return graph.IDs(graph.END)  // exit
}, []graph.ID{"inc", graph.END})
```

**Safety:** Graph execution caps at `RecursionLimit` (default 25) to detect infinite loops

---

### 4. Invoking a Graph

```go
ctx := context.Background()
result, err := g.Invoke(ctx, initialState, graph.InvocationConfig[int]{
    RecursonLimit: 100,  // optional: override default
    Checkpointer:  checkpointer,  // optional: for interrupts
})

if err != nil {
    // Handle execution errors
    if superErr, ok := err.(*graph.SuperStepExecutionError); ok {
        for _, itr := range superErr.Interrupts() {
            // Handle interrupt: itr.Name, itr.ThreadID
        }
    }
}
```

**Return values:**
- `result`: Final state after all nodes execute
- `err`: Non-nil if any node/reducer/router panics OR interrupts occur

---

### 5. Interrupts & Checkpointing

#### Triggering an Interrupt
```go
func(ctx context.Context, state int) (int, error) {
    itr, err := graph.Interrupt[int](ctx, "approval-required", map[string]any{
        "reason": "manual review needed",
    })
    if err != nil {
        return state, err  // Interrupt was rejected
    }
    
    // Approved: resume execution (itr.Result contains approval data)
    return state + 1, nil
}
```

**Pattern:**
- Call `Interrupt()` with a unique name and payload
- Execution pauses, returning `SuperStepExecutionError` with interrupt list
- Caller inspects `err.Interrupts()` to decide approval/rejection
- Call `g.Resume(ctx, threadID, decisions, config)` to continue

#### Checkpointer Interface
```go
type Checkpointer[T any] interface {
    SaveCheckpoint(ctx context.Context, threadID string, stepIndex int, state T) error
    LoadCheckpoint(ctx context.Context, threadID string, stepIndex int) (T, error)
}
```

**Provided implementations:**
- `NewInMemoryCheckpointer[T]()`: In-memory storage (tests, local only)
- Database/Redis checkpoint implementations can be custom

---

### 6. Error Handling

#### Execution Error Hierarchy
```
error
├── SuperStepExecutionError  // Multiple failures in one super step
│   └── Contains []error:
│       ├── NodeExecutionError      // Node fn panicked
│       ├── ReducerExecutionError   // Reducer panicked
│       ├── RouterExecutionError    // Router fn panicked
│       ├── InterruptRejectedError  // User rejected interrupt
│       └── InterruptError          // Node triggered interrupt
├── InvocationError           // Infinite loop or max recursion
└── (custom errors from node functions)
```

#### Error Inspection
```go
result, err := g.Invoke(ctx, state, config)
if err != nil {
    switch e := err.(type) {
    case *graph.SuperStepExecutionError:
        // Multiple concurrent errors
        for _, nodeErr := range e.Errs {
            if ne, ok := nodeErr.(*graph.NodeExecutionError); ok {
                fmt.Printf("Node %s failed: %v\n", ne.ID, ne.Err)
            }
        }
    case *graph.InvocationError:
        // Infinite loop or recursion limit
        fmt.Printf("Execution failed: %v\n", e.Message)
    default:
        // Custom error from node function
    }
}
```

---

### 7. Visualization

```go
g := graph.New[int](func(a, b int) int { return a + b })
// ... add nodes and edges ...

diagram := g.Visualize()  // returns Mermaid diagram string
fmt.Println(string(diagram))
// Output:
// graph TD
//     START["START"]
//     inc["inc"]
//     END["END"]
//     START -->|/ inc /| inc
//     inc --> END
```

**Use cases:**
- Debugging complex graphs
- Documentation
- Testing correctness of routing logic

---

## Testing Patterns

### Table-Driven Tests
```go
tests := []struct {
    name          string
    initialState  int
    expectedState int
    expectedError bool
}{
    {"simple increment", 0, 1, false},
    {"from 5", 5, 11, false},
}

for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) {
        result, err := g.Invoke(context.Background(), tc.initialState, graph.InvocationConfig[int]{})
        assert.Equal(t, tc.expectedState, result)
        if tc.expectedError {
            assert.Error(t, err)
        } else {
            assert.NoError(t, err)
        }
    })
}
```

### Testing Interrupts
```go
checkpointer := graph.NewInMemoryCheckpointer[int]()
config := graph.InvocationConfig[int]{Checkpointer: checkpointer}

result, err := g.Invoke(ctx, 0, config)
assert.Error(t, err)

superErr, _ := err.(*graph.SuperStepExecutionError)
interrupts := superErr.Interrupts()
assert.Len(t, interrupts, 1)

// Resume with approval
result, err = g.Resume(ctx, interrupts[0].ThreadID, map[string]any{
    interrupts[0].Name: "approved",
}, config)
assert.NoError(t, err)
```

### Testing Error Paths
```go
g.AddNode("willFail", func(ctx context.Context, state int) (int, error) {
    panic("intentional panic")
})

result, err := g.Invoke(ctx, 0, graph.InvocationConfig[int]{})
assert.Error(t, err)
assert.IsType(t, &graph.SuperStepExecutionError{}, err)

superErr, _ := err.(*graph.SuperStepExecutionError)
assert.Len(t, superErr.Errs, 1)
assert.IsType(t, &graph.NodeExecutionError{}, superErr.Errs[0])
```

---

## Best Practices

### State Design
- **Keep state immutable where possible**: Reduce side effects in nodes
- **Use structs for complex state**: Type safety, easier to evolve
- **Minimize state size**: Larger states = slower checkpointing

### Node Design
- **Idempotent nodes**: Safe to retry if interrupted
- **Handle context cancellation**: Respect `ctx.Done()`
- **Return errors, not panics**: Panics are caught but complicate debugging

### Router Design
- **Deterministic routing**: Same state = same route every time
- **Never return empty slice**: At minimum, return `IDs(END)` or a default path

### Reducer Design
- **Associative & commutative**: `(a ⊕ b) ⊕ c = a ⊕ (b ⊕ c)`
- **Identity element**: e.g., `0` for sum, `1` for product, `[]T{}` for append
- **No side effects**: Pure function, no I/O or state mutation

### Recursion & Loops
- **Set `RecursionLimit`** if you expect deep recursion (default 25)
- **Always provide exit condition** in routers
- **Test loops thoroughly**: Infinite loops timeout the test

---

## Context Values

Nodes receive a `context.Context` with pre-populated values:

```go
threadID := ctx.Value("threadId").(graph.ID)          // Invocation ID
g := ctx.Value("graph").(graph.StateGraph[T])         // The graph itself (for Interrupt)
tools := ctx.Value("tools").(map[string]any)          // Optional tool registry
```

---

## Examples from Codebase

### Simple Graph (2 nodes)
- Files: `graph_invoke_test.go:11-48`
- Pattern: START → increment → END

### Fan-Out Graph (parallel execution)
- Files: `graph_invoke_test.go:91-128`
- Pattern: START → {inc, double} → END with reducer merging

### Conditional Routing
- Files: `graph_test.go:101-132`
- Pattern: START → (router: state > 0) → {positive, nonpositive} → END

### Worker Node
- Files: `graph_invoke_test.go:431-462`
- Pattern: Map over state.Items, reducer accumulates results

### Error Handling
- Files: `graph_invoke_test.go:171-236` (node errors)
- Files: `graph_invoke_test.go:238-291` (reducer errors)
- Files: `graph_invoke_test.go:294-348` (router errors)

### Interrupts
- Files: `graph_interrupt_test.go` (all)
- Patterns: 
  - 2 interrupts in same super step
  - 2 interrupts in same node
  - 2 interrupts in different nodes
  - Resume with approval/rejection

---

## Common Pitfalls

### 1. **Panics in Reducer**
```go
// BAD: Can panic
g := graph.New[int](func(a, b int) int {
    return a / b  // panics if b == 0
})

// GOOD: Validate inputs
g := graph.New[int](func(a, b int) int {
    if b == 0 { return a }  // sensible default
    return a / b
})
```

### 2. **Non-Deterministic Routing**
```go
// BAD: Random routing
g.AddConditionalEdge(START, func(state int) []graph.Target {
    if rand.Float64() > 0.5 { return graph.IDs("a") }
    return graph.IDs("b")
}, []graph.ID{"a", "b"})

// GOOD: State-based routing
g.AddConditionalEdge(START, func(state int) []graph.Target {
    if state%2 == 0 { return graph.IDs("even") }
    return graph.IDs("odd")
}, []graph.ID{"even", "odd"})
```

### 3. **Forgetting to Handle Context Cancellation**
```go
// BAD: Ignores cancellation
g.AddNode("slow", func(ctx context.Context, state int) (int, error) {
    time.Sleep(10 * time.Second)
    return state, nil
})

// GOOD: Respects cancellation
g.AddNode("slow", func(ctx context.Context, state int) (int, error) {
    select {
    case <-ctx.Done():
        return state, ctx.Err()
    case <-time.After(10 * time.Second):
        return state, nil
    }
})
```

### 4. **Infinite Loops Without Exit**
```go
// BAD: Always routes back to itself
g.AddConditionalEdge("loop", func(state int) []graph.Target {
    return graph.IDs("loop")  // never exits!
}, []graph.ID{"loop"})

// GOOD: Has exit condition
g.AddConditionalEdge("loop", func(state int) []graph.Target {
    if state >= 10 { return graph.IDs(graph.END) }
    return graph.IDs("loop")
}, []graph.ID{"loop", graph.END})
```

---

## When to Use the Graph Package

✅ **Good fits:**
- Multi-step workflows (e.g., approval chains, data pipelines)
- Conditional branching based on state
- Parallel processing with state merging
- Long-running processes with checkpointing
- Human-in-the-loop via interrupts

❌ **Poor fits:**
- Simple sequential execution (use FSM or direct functions)
- Stateless request routing (use middleware)
- Real-time streaming (consider reactive libraries)

---

## Related Packages

- **@fsm/**: Simpler state machine (no state graph, no parallelism)
- **@provider/**: LLM provider abstraction (uses FSM, not graph)
- **@visualizer/**: Mermaid diagram generation (used by graph internally)
- **@utils/**: Helper functions (Keys, Map, etc.)
