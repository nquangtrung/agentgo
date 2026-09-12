package graph

import (
	"testing"

	"github.com/nquangtrung/agentgo/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewGraph(t *testing.T) {
	g := New[int](func(a, b int) (int, error) {
		return a + b, nil
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
	g := New[int](func(a, b int) (int, error) {
		return a + b, nil
	})

	fn := func(state int) (int, error) {
		return state + 1, nil
	}
	id := "inc"
	g.AddNode(id, fn)

	assert.Contains(t, g.nodes, id, "Expected node to be added to the graph")
}

func TestAddEdge(t *testing.T) {
	g := New[int](func(a, b int) (int, error) {
		return a + b, nil
	})

	g.AddEdge(START, END)

	assert.Contains(t, g.edges, START, "Expected edge to be added to the graph")
	assert.Equal(t, []ID{END}, g.edges[START].End, "Expected edge end to match the provided end ID")
}

func TestSimpleGraph(t *testing.T) {
	g := New(func(a, b int) (int, error) {
		return a + b, nil
	})

	g.AddNode("inc", func(state int) (int, error) {
		return state + 1, nil
	})
	g.AddEdge(START, "inc")
	g.AddEdge("inc", END)

	result := g.Invoke(0)
	assert.Equal(t, 1, result, "Expected final state to be 1 after running the graph")
}

func TestGraphWithMultipleNodes(t *testing.T) {
	g := New(func(a, b int) (int, error) {
		return a + b, nil
	})

	g.AddNode("inc", func(state int) (int, error) {
		return state + 1, nil
	})
	g.AddNode("double", func(state int) (int, error) {
		return state * 2, nil
	})

	g.AddEdge(START, "inc")
	g.AddEdge("inc", "double")
	g.AddEdge("double", END)

	result := g.Invoke(0)
	assert.Equal(t, 2, result, "Expected final state to be 2 after running the graph")
}

func TestGraphWithFanOut(t *testing.T) {
	g := New(func(a, b int) (int, error) {
		return a + b, nil
	})

	g.AddNode("inc", func(state int) (int, error) {
		return state + 1, nil
	})
	g.AddNode("double", func(state int) (int, error) {
		return state * 2, nil
	})

	g.FanOut(START, []ID{"inc", "double"})
	g.AddEdge("inc", END)
	g.AddEdge("double", END)

	result := g.Invoke(1)
	assert.Equal(t, 4, result, "Expected final state to be 3 after running the graph")
}

func TestGraphWithImbalanceNodes(t *testing.T) {
	g := New(func(a, b int) (int, error) {
		return a + b, nil
	})

	g.AddNode("inc", func(state int) (int, error) {
		return state + 1, nil
	})
	g.AddNode("inc2", func(state int) (int, error) {
		return state + 2, nil
	})
	g.AddNode("double", func(state int) (int, error) {
		return state * 2, nil
	})

	g.FanOut(START, []ID{"inc", "inc2"})
	g.AddEdge("inc2", "double")
	g.AddEdge("inc", END)
	g.AddEdge("double", END)

	result := g.Invoke(1)
	assert.Equal(t, 6, result, "Expected final state to be 4 after running the graph")
}
