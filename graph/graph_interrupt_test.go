package graph

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphInterrupt(t *testing.T) {
	g := New[int](func(a, b int) int {
		return a + b
	})

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
	})

	g.AddNode("double", func(ctx context.Context, state int) (int, error) {
		i, err := Interrupt[int](ctx, map[string]any{"message": "double the state"})

		if err != nil {
			t.Fatalf("Interrupt failed: %v", err)
		}

		if i != nil {
			logger.Info("Interrupt returned", slog.Any("result", i))
		}
		return state * 2, nil
	})

	g.AddEdge(START, "inc")
	g.AddEdge("inc", "double")
	g.AddEdge("double", END)

	config := InvocationConfig[int]{}
	ctx := context.Background()
	_, err := g.Invoke(ctx, 0, config)

	assert.Error(t, err, "Expected an error due to interrupt")
	assert.IsType(t, &SuperStepExecutionError{}, err, "Expected an InterruptError")

	// superStepErr, _ := err.(*SuperStepExecutionError)
	// interrupts := superStepErr.Interrupts()
	// assert.NotEmpty(t, interrupts, "Expected at least one interrupt")
}
