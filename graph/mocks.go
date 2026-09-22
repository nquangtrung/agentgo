package graph

import (
	"context"
	"log/slog"
)

func createMockGraphWith2InterruptInSingleNode() StateGraph[int] {
	g := New[int](func(a, b int) int {
		return a + b
	})

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		itr1, err := Interrupt[int](ctx, "inc1-itr", map[string]any{"message": "Should I increment the state?"})
		if err != nil {
			return 0, err
		}

		if itr1.Result != "approved" {
			logger.Info("Interrupt returned", slog.Any("result", itr1))
			return 0, itr1.Reject("User rejected increment node")
		}

		itr2, err := Interrupt[int](ctx, "inc2-itr", map[string]any{"message": "Should I increment the state again?"})
		if err != nil {
			return 0, err
		}

		if itr2.Result != "approved" {
			logger.Info("Interrupt returned", slog.Any("result", itr2))
			return 0, itr2.Reject("User rejected increment node again")
		}

		return state + 1, nil
	})

	g.AddEdge(START, "inc")
	g.AddEdge("inc", END)

	return g
}
func createMockGraphWith2InterruptIn2Node() StateGraph[int] {
	g := New[int](func(a, b int) int {
		return a + b
	})

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		itr1, err := Interrupt[int](ctx, "inc-itr", map[string]any{"message": "Should I increment the state?"})
		if err != nil {
			return 0, err
		}

		if itr1.Result != "approved" {
			logger.Info("Interrupt returned", slog.Any("result", itr1))
			return 0, itr1.Reject("User rejected increment node")
		}

		return state + 1, nil
	})

	g.AddNode("double", func(ctx context.Context, state int) (int, error) {
		itr2, err := Interrupt[int](ctx, "double-itr", map[string]any{"message": "Should I double the state?"})
		if err != nil {
			return 0, err
		}

		if itr2.Result != "approved" {
			logger.Info("Interrupt returned", slog.Any("result", itr2))
			return 0, itr2.Reject("User rejected double node")
		}

		return state * 2, nil
	})

	g.AddEdge(START, "inc")
	g.AddEdge("inc", "double")
	g.AddEdge("double", END)

	return g
}

func createMockGraphWith2InterruptInSameSuperStep() StateGraph[int] {
	g := New[int](func(a, b int) int {
		return a + b
	})

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})
	g.AddNode("inc2", func(ctx context.Context, state int) (int, error) {
		return state + 2, nil
	})
	g.AddNode("double", func(ctx context.Context, state int) (int, error) {
		fromUser, err := Interrupt[int](ctx, "double-itr", map[string]any{"message": "triple the state"})
		if err != nil {
			return 0, err
		}

		if fromUser.Result != "approved" {
			logger.Info("Interrupt returned", slog.Any("result", fromUser))
			return 0, fromUser.Reject("User rejected double nod")
		}

		return state * 2, nil
	})

	g.AddNode("triple", func(ctx context.Context, state int) (int, error) {
		fromUser, err := Interrupt[int](ctx, "triple-itr", map[string]any{"message": "Should I triple the state?"})
		if err != nil {
			return 0, err
		}

		if fromUser.Result != "approved" {
			logger.Info("Interrupt returned", slog.Any("result", fromUser))
			return 0, fromUser.Reject("User rejected triple nod")
		}

		return state * 3, nil
	})

	g.AddEdge(START, "inc")
	g.FanOut("inc", []ID{"inc2", "double", "triple"})
	g.AddEdge("double", END)
	g.AddEdge("triple", END)
	g.AddEdge("inc2", END)

	return g
}

func createComplexMapWithNoInterrupt() StateGraph[int] {
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
	return g
}

// Visualization helper graphs

func createSimpleGraphForVisualization() StateGraph[int] {
	g := New(func(a, b int) int {
		return a + b
	})

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})
	g.AddEdge(START, "inc")
	g.AddEdge("inc", END)

	return g
}

func createMultipleNodesGraphForVisualization() StateGraph[int] {
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

	return g
}

func createFanOutGraphForVisualization() StateGraph[int] {
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

	return g
}

func createImbalanceNodesGraphForVisualization() StateGraph[int] {
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

	return g
}

func createCycleGraphForVisualization() StateGraph[int] {
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

	return g
}

func createWorkerNodeGraphForVisualization() StateGraph[int] {
	g := New(func(a, b int) int {
		return a + b
	})

	g.AddWorkerNode("worker", func(_ context.Context, params any) (int, error) {
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

	return g
}
