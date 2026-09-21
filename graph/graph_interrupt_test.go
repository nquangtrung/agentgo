package graph

import (
	"context"
	"log/slog"
	"testing"

	"github.com/nquangtrung/agentgo/utils"
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
			return 0, err
		}

		if i != nil {
			logger.Info("Interrupt returned", slog.Any("result", i))
		}
		return state * 2, nil
	})
	g.AddNode("triple", func(ctx context.Context, state int) (int, error) {
		i, err := Interrupt[int](ctx, map[string]any{"message": "triple the state"})
		if err != nil {
			return 0, err
		}

		if i != nil {
			logger.Info("Interrupt returned", slog.Any("result", i))
		}
		return state * 3, nil
	})

	g.AddEdge(START, "inc")
	g.FanOut("inc", []ID{"double", "triple"})
	g.AddEdge("double", END)

	checkpointer := NewInMemoryCheckpointer[int]()
	config := InvocationConfig[int]{
		Checkpointer: checkpointer,
	}
	ctx := context.Background()
	_, err := g.Invoke(ctx, 0, config)

	assert.Error(t, err, "Expected an error due to interrupt")
	assert.IsType(t, &SuperStepExecutionError{}, err, "Expected an InterruptError")

	superStepErr, _ := err.(*SuperStepExecutionError)
	interrupts := superStepErr.Interrupts()
	assert.Len(t, interrupts, 2, "Expected exact 2 interrupts")

	ids := utils.Map(interrupts, func(itr *InterruptError) ID {
		return itr.ID
	})
	assert.ElementsMatch(t, []ID{"double", "triple"}, ids, "Expected interrupts from 'double' and 'triple' nodes")
	assert.NotEqual(t, "", interrupts[0].ThreadID, "Expected non-empty ThreadID for the 1st interrupt")
	assert.NotEqual(t, "", interrupts[1].ThreadID, "Expected non-empty ThreadID for the 2nd interrupt")

	logger.Info("Resuming graph execution with interrupt result", slog.Any("interrupts", interrupts))
	_, err = g.Resume(ctx, interrupts[0].ThreadID, map[string]any{
		interrupts[0].Name: "approved",
	}, config)

	assert.Error(t, err, "Expected an error due to missing interrupts result")
	superStepErr, _ = err.(*SuperStepExecutionError)
	interrupts = superStepErr.Interrupts()
	assert.Len(t, interrupts, 1, "Expected exact 1 interrupt after resuming")
}
