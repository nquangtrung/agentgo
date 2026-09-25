package graph

import (
	"context"
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
	result, err := g.Invoke(ctx, 1, InvocationConfig[int, int]{})
	assert.Equal(t, 3, result, "Expected final state to be 2 after running the graph with positive input")
	assert.NoError(t, err, "Expect result without error")

	result, err = g.Invoke(ctx, -1, InvocationConfig[int, int]{})
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
