package graph

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/nquangtrung/agentgo/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewGraph(t *testing.T) {
	g := New[int](func(a, b int) int {
		return a + b
	})

	assert.NotNil(t, g.reducer, "Expected reducer to be not nil")
	assert.NotNil(t, g.nodes, "Expected nodes map to be initialized")
	assert.NotNil(t, g.edges, "Expected edges map to be initialized")
	assert.Len(t, utils.Keys(g.nodes), 2, "Expected nodes length to be 2 (START and END nodes)")
	assert.NotNil(t, g.nodes[START], "Expected START to be initialized")
	assert.NotNil(t, g.nodes[END], "Expected END to be initialized")
	assert.Len(t, utils.Keys(g.edges), 0, "Expected edges length to be 0")
}

func TestAddNode(t *testing.T) {
	g := New[int](func(a, b int) int {
		return a + b
	})

	fn := func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	}
	id := "inc"
	g.AddNode(id, fn)

	assert.Contains(t, g.nodes, id, "Expected node to be added to the graph")
}

func TestAddNodePanic(t *testing.T) {
	g := New[int](func(a, b int) int {
		return a + b
	})

	fn := func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	}

	assert.Panics(t, func() {
		g.AddNode(START, fn)
	}, "Expected panic when adding a node with START ID")
	assert.Panics(t, func() {
		g.AddNode(END, fn)
	}, "Expected panic when adding a node with END ID")

	id := "inc"
	g.AddNode(id, fn)
	assert.Panics(t, func() {
		g.AddNode(id, fn)
	}, "Expected panic when adding a duplicate node")
}

func TestAddEdge(t *testing.T) {
	g := New[int](func(a, b int) int {
		return a + b
	})

	g.AddEdge(START, END)

	assert.Contains(t, g.edges, START, "Expected edge to be added to the graph")
	assert.Equal(t, []ID{END}, g.edges[START].end, "Expected edge end to match the provided end ID")
}

func TestAddEdgePanic(t *testing.T) {
	g := New[int](func(a, b int) int {
		return a + b
	})

	assert.Panics(t, func() {
		g.AddEdge("nonexistent", END)
	}, "Expected panic when adding an edge with a non-existent start node")

	assert.Panics(t, func() {
		g.AddEdge(START, "nonexistent")
	}, "Expected panic when adding an edge with a non-existent end node")
}

func TestFanOutPanic(t *testing.T) {
	g := New[int](func(a, b int) int {
		return a + b
	})

	assert.Panics(t, func() {
		g.FanOut("nonexistent", []ID{END})
	}, "Expected panic when fan-out from a non-existent start node")

	assert.Panics(t, func() {
		g.FanOut(START, []ID{"nonexistent", "nonexistent2"})
	}, "Expected panic when fan-out to a non-existent end node")
}

func TestAddConditionalEdge(t *testing.T) {
	g := New[int](func(a, b int) int {
		return a + b
	})

	condition := func(state int) []Target {
		if state > 0 {
			return IDs("positive")
		}
		return IDs("nonpositive")
	}

	g.AddNode("positive", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})
	g.AddNode("nonpositive", func(ctx context.Context, state int) (int, error) {
		return state - 1, nil
	})

	g.AddConditionalEdge(START, condition, []ID{"positive", "nonpositive"})
	g.AddEdge("positive", END)
	g.AddEdge("nonpositive", END)

	ctx := context.Background()
	result, err := g.Invoke(ctx, 1, InvocationConfig[int]{})
	assert.Equal(t, 3, result, "Expected final state to be 2 after running the graph with positive input")
	assert.NoError(t, err, "Expect result without error")

	result, err = g.Invoke(ctx, -1, InvocationConfig[int]{})
	assert.NoError(t, err, "Expect result without error")
	assert.Equal(t, -3, result, "Expected final state to be -2 after running the graph with non-positive input")
}

func TestAddConditionalEdgePanic(t *testing.T) {
	g := New[int](func(a, b int) int {
		return a + b
	})

	condition := func(state int) []Target {
		if state > 0 {
			return IDs("positive")
		}
		return IDs("nonpositive")
	}

	assert.Panics(t, func() {
		g.AddConditionalEdge("nonexistent", condition, []ID{"positive", "nonpositive"})
	}, "Expected panic when adding a conditional edge from a non-existent start node")

	assert.Panics(t, func() {
		g.AddConditionalEdge(START, condition, []ID{"nonexistent", "nonexistent2"})
	}, "Expected panic when adding a conditional edge to a non-existent end node")
}

func TestSimpleGraph(t *testing.T) {
	g := New(func(a, b int) int {
		return a + b
	})

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})
	g.AddEdge(START, "inc")
	g.AddEdge("inc", END)

	ctx := context.Background()
	result, err := g.Invoke(ctx, 0, InvocationConfig[int]{})
	assert.NoError(t, err, "Expect result without error")
	assert.Equal(t, 1, result, "Expected final state to be 1 after running the graph")
}

func TestGraphWithMultipleNodes(t *testing.T) {
	g := New(func(a, b int) int {
		return a + b
	})

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})
	g.AddNode("double", func(ctx context.Context, state int) (int, error) {
		return state * 2, nil
	})

	g.AddEdge(START, "inc")
	g.AddEdge("inc", "double")
	g.AddEdge("double", END)

	ctx := context.Background()
	result, err := g.Invoke(ctx, 0, InvocationConfig[int]{})
	assert.Equal(t, 3, result, "Expected final state to be 2 after running the graph")
	assert.NoError(t, err, "Expect result without error")
}

func TestGraphWithFanOut(t *testing.T) {
	g := New(func(a, b int) int {
		return a + b
	})

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})
	g.AddNode("double", func(ctx context.Context, state int) (int, error) {
		return state * 2, nil
	})

	g.FanOut(START, []ID{"inc", "double"})
	g.AddEdge("inc", END)
	g.AddEdge("double", END)

	ctx := context.Background()
	result, err := g.Invoke(ctx, 1, InvocationConfig[int]{})

	assert.NoError(t, err, "Expect result without error")
	assert.Equal(t, 5, result, "Expected final state to be 3 after running the graph")
}

func TestGraphWithImbalanceNodes(t *testing.T) {
	g := New(func(a, b int) int {
		return a + b
	})

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})
	g.AddNode("inc2", func(ctx context.Context, state int) (int, error) {
		return state + 2, nil
	})
	g.AddNode("double", func(ctx context.Context, state int) (int, error) {
		return state * 2, nil
	})

	g.FanOut(START, []ID{"inc", "inc2"})
	g.AddEdge("inc2", "double")
	g.AddEdge("inc", END)
	g.AddEdge("double", END)

	ctx := context.Background()
	result, err := g.Invoke(ctx, 1, InvocationConfig[int]{})
	assert.NoError(t, err, "Expect result without error")
	assert.Equal(t, 18, result, "Expected final state to be 4 after running the graph")
}

func TestGraphWithErrorInNode(t *testing.T) {
	g := New(func(a, b int) int {
		return a + b
	})

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})
	g.AddNode("errorNode1", func(ctx context.Context, state int) (int, error) {
		panic("intentional error 1")
	})
	g.AddNode("errorNode2", func(ctx context.Context, state int) (int, error) {
		panic("intentional error 2")
	})

	g.FanOut(START, []ID{"inc", "errorNode1", "errorNode2"})
	g.AddEdge("inc", END)
	g.AddEdge("errorNode1", END)
	g.AddEdge("errorNode2", END)

	ctx := context.Background()
	result, err := g.Invoke(ctx, 1, InvocationConfig[int]{})
	assert.Error(t, err, "Expected an error due to the intentional panic in errorNode")
	assert.IsType(t, &SuperStepExecutionError{}, err, "Expected error to be of type SuperStepExecutionError")

	superStepErr, _ := err.(*SuperStepExecutionError)
	assert.Len(t, superStepErr.Errs, 2, "Expected one error in the super step execution error")

	errorNodes := []string{"errorNode1", "errorNode2"}
	assert.ElementsMatch(t, errorNodes, utils.Map(superStepErr.Errs, func(err error) string {
		if nodeErr, ok := err.(*NodeExecutionError); ok {
			return nodeErr.ID
		}
		return "unknown"
	}))

	assert.Equal(t, 1, result, "Expected final state to be 2 after running the graph with error in one node")
}

func TestGraphWithErrorInReducer(t *testing.T) {
	g := New(func(a, b int) int {
		if b == 0 {
			panic("intentional error in reducer")
		}
		return a + b
	})

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})
	g.AddNode("zero", func(ctx context.Context, state int) (int, error) {
		return 0, nil
	})

	g.FanOut(START, []ID{"inc", "zero"})
	g.AddEdge("inc", END)
	g.AddEdge("zero", END)

	ctx := context.Background()
	result, err := g.Invoke(ctx, 1, InvocationConfig[int]{})
	assert.Error(t, err, "Expected an error due to the intentional panic in reducer")
	assert.IsType(t, &ReducerExecutionError{}, err, "Expected error to be of type ReducerExecutionError")

	reducerErr, _ := err.(*ReducerExecutionError)
	assert.Equal(t, "panic in reducer execution: intentional error in reducer", reducerErr.Err.Error(), "Expected error message to match the intentional panic message")
	assert.Equal(t, 1, result, "Expected final state to be 2 after running the graph with error in reducer")
}

func TestGraphWithErrorInRouter(t *testing.T) {
	g := New(func(a, b int) int {
		return a + b
	})

	router := func(state int) []Target {
		if state < 0 {
			panic("intentional error in router")
		}
		return IDs("inc")
	}

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})

	g.AddConditionalEdge(START, router, []ID{"inc"})
	g.AddEdge("inc", END)

	ctx := context.Background()
	result, err := g.Invoke(ctx, -1, InvocationConfig[int]{})
	assert.Error(t, err, "Expected an error due to the intentional panic in router")
	assert.IsType(t, &RouterExecutionError{}, err, "Expected error to be of type RouterExecutionError")

	routerErr, _ := err.(*RouterExecutionError)
	assert.Equal(t, START, routerErr.ID, "Expected error ID to match the START node ID")
	assert.Equal(t, "panic in router execution: intentional error in router", routerErr.Err.Error(), "Expected error message to match the intentional panic message")
	assert.Equal(t, -1, result, "Expected final state to be -1 after running the graph with error in router")
}

func TestCycleGraphWithCondition(t *testing.T) {
	g := New(func(a, b int) int {
		return a + b
	})
	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})

	g.AddEdge(START, "inc")
	g.AddConditionalEdge("inc", func(state int) []Target {
		if state < 10 {
			return IDs("inc")
		}
		return IDs(END)
	}, []ID{"inc", END})

	ctx := context.Background()
	result, err := g.Invoke(ctx, 0, InvocationConfig[int]{})
	assert.NoError(t, err, "Expect result without error")
	assert.Equal(t, 15, result, "Expected final state to be 1 after running the graph with a cycle")
}

func TestInfiniteLoopGraph(t *testing.T) {
	g := New(func(a, b int) int {
		return a + b
	})
	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})

	g.AddEdge(START, "inc")
	g.AddConditionalEdge("inc", func(state int) []Target {
		return IDs("inc")
	}, []ID{"inc"})

	ctx := context.Background()
	_, err := g.Invoke(ctx, 0, InvocationConfig[int]{})
	assert.Error(t, err, "Expected an error due to infinite loop in the graph")
	assert.IsType(t, &InvocationError{}, err, "Expected error to be of type RouterExecutionError")
}

func TestGraphWithWorkerNode(t *testing.T) {
	g := New(func(a, b int) int {
		return a + b
	})

	g.AddWorkerNode("worker", func(params any) (int, error) {
		if p, ok := params.(int); ok {
			return p * 2, nil
		}
		return 0, nil
	})

	g.AddConditionalEdge(START, func(state int) []Target {
		return []Target{
			Send("worker", 1),
			Send("worker", 2),
			Send("worker", 3),
		}
	}, []ID{"worker"})

	g.AddEdge("worker", END)

	ctx := context.Background()
	result, err := g.Invoke(ctx, 5, InvocationConfig[int]{})
	assert.NoError(t, err, "Expect result without error")
	assert.Equal(t, 17, result, "Expected final state to be 10 after running the graph with worker node")
}

func TestGraphVisualize(t *testing.T) {
	g := New(func(a, b int) int {
		return a + b
	})

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})
	g.AddNode("double", func(ctx context.Context, state int) (int, error) {
		return state * 2, nil
	})
	g.AddNode("end_orphaned", func(ctx context.Context, state int) (int, error) {
		return state * 3, nil
	})
	g.AddNode("orphaned", func(ctx context.Context, state int) (int, error) {
		return state * 3, nil
	})
	g.AddNode("start_orphaned", func(ctx context.Context, state int) (int, error) {
		return state * 3, nil
	})

	condition := func(state int) []Target {
		if state > 0 {
			return IDs("inc", "double")
		}
		return IDs("triple")
	}

	g.AddConditionalEdge(START, condition, []ID{"inc", "double", "end_orphaned"})

	g.FanOut("double", []ID{"end_orphaned", END})
	g.AddEdge("start_orphaned", END)
	g.AddNamedConditionalEdge("inc", func(state int) string {
		if state < 10 {
			return "too_small"
		}
		return "large_enough"
	}, NamedRouterMap{
		"too_small":    IDs("inc"),
		"large_enough": IDs(END),
	})

	visualization := g.Visualize()
	toBeSaved := fmt.Sprintf("```mermaid\n%s\n```", visualization)
	err := os.WriteFile(
		"test.md",
		[]byte(toBeSaved),
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	result, err := g.Invoke(ctx, 1, InvocationConfig[int]{})

	assert.NoError(t, err, "Expect result without error")
	assert.Equal(t, 26, result, "Expected final state to be 26 after running the graph with visualization")
}
