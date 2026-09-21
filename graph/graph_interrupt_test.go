package graph

import (
	"context"
	"log/slog"
	"testing"

	"github.com/nquangtrung/agentgo/utils"
	"github.com/stretchr/testify/assert"
)

func createGraph() StateGraph[int] {
	g := New[int](func(a, b int) int {
		return a + b
	})

	g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
		return state + 1, nil
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
	g.FanOut("inc", []ID{"double", "triple"})
	g.AddEdge("double", END)

	return g
}

func TestGraphInterrupt(t *testing.T) {
	g := createGraph()

	checkpointer := NewInMemoryCheckpointer[int]()
	config := InvocationConfig[int]{
		Checkpointer: checkpointer,
	}
	ctx := context.Background()
	result, err := g.Invoke(ctx, 0, config)

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
	assert.Equal(t, 1, result, "The state should be until before the interrupt, which is 1 (0 + 1)")
}

func TestGraphInterruptPartialReject(t *testing.T) {
	g := createGraph()

	checkpointer := NewInMemoryCheckpointer[int]()
	config := InvocationConfig[int]{
		Checkpointer: checkpointer,
	}
	ctx := context.Background()
	result, err := g.Invoke(ctx, 0, config)

	superStepErr, _ := err.(*SuperStepExecutionError)
	interrupts := superStepErr.Interrupts()

	logger.Info("Resuming graph execution with interrupt result", slog.Any(interrupts[0].Name, "approved"))
	result, err = g.Resume(ctx, interrupts[0].ThreadID, map[string]any{
		interrupts[0].Name: "approved",
	}, config)

	assert.Error(t, err, "Expected an error due to missing interrupts result")
	superStepErr, _ = err.(*SuperStepExecutionError)
	interrupts = superStepErr.Interrupts()
	assert.Len(t, interrupts, 1, "Expected exact 1 interrupt after resuming")
	assert.Equal(t, 1, result, "The state should not change after resuming with the first interrupt result, which is 1 (0 + 1)")

	logger.Info("Resuming graph execution with interrupt result", slog.Any(interrupts[0].Name, "rejected"))
	result, err = g.Resume(ctx, interrupts[0].ThreadID, map[string]any{
		interrupts[0].Name: "rejected",
	}, config)
	assert.Error(t, err, "Expected an error due to interrupt rejection")
	assert.Equal(t, 1, result, "The state should not change after resuming since 1 interrupt is rejected")
	superStepErr, _ = err.(*SuperStepExecutionError)
	assert.IsType(t, superStepErr.Errs[0], &InterruptRejectedError{}, "Error should be interrupt rejected")
}

func TestGraphInterruptAllApproved(t *testing.T) {
	g := createGraph()

	checkpointer := NewInMemoryCheckpointer[int]()
	config := InvocationConfig[int]{
		Checkpointer: checkpointer,
	}
	ctx := context.Background()
	_, err := g.Invoke(ctx, 0, config)

	superStepErr, _ := err.(*SuperStepExecutionError)
	interrupts := superStepErr.Interrupts()

	logger.Info("Resuming graph execution with interrupt result", slog.Any(interrupts[0].Name, "approved"))
	result, err := g.Resume(ctx, interrupts[0].ThreadID, map[string]any{
		interrupts[0].Name: "approved",
		interrupts[1].Name: "approved",
	}, config)

	assert.NoError(t, err, "Expected no error after resuming with all interrupts approved")
	assert.Equal(t, 6, result, "The state should be fully updated (0 + 1 + 2 + 3)")
}

func TestGraphInterruptAllRejectedTogether(t *testing.T) {
	g := createGraph()

	checkpointer := NewInMemoryCheckpointer[int]()
	config := InvocationConfig[int]{
		Checkpointer: checkpointer,
	}
	ctx := context.Background()
	_, err := g.Invoke(ctx, 0, config)

	superStepErr, _ := err.(*SuperStepExecutionError)
	interrupts := superStepErr.Interrupts()

	logger.Info("Resuming graph execution with interrupt result", slog.Any(interrupts[0].Name, "approved"))
	result, err := g.Resume(ctx, interrupts[0].ThreadID, map[string]any{
		interrupts[0].Name: "rejected",
		interrupts[1].Name: "rejected",
	}, config)

	assert.Equal(t, 1, result, "The state should be the last valid state")
	superStepErr, _ = err.(*SuperStepExecutionError)
	assert.Len(t, superStepErr.Errs, 2, "Expected 2 errors due to interrupt rejection")
	assert.IsType(t, superStepErr.Errs[0], &InterruptRejectedError{}, "Error should be interrupt rejected")
	assert.IsType(t, superStepErr.Errs[1], &InterruptRejectedError{}, "Error should be interrupt rejected")
}

func TestGraphInterruptPartiallyRejectedTogether(t *testing.T) {
	g := createGraph()

	checkpointer := NewInMemoryCheckpointer[int]()
	config := InvocationConfig[int]{
		Checkpointer: checkpointer,
	}
	ctx := context.Background()
	_, err := g.Invoke(ctx, 0, config)

	superStepErr, _ := err.(*SuperStepExecutionError)
	interrupts := superStepErr.Interrupts()

	logger.Info("Resuming graph execution with interrupt result", slog.Any(interrupts[0].Name, "approved"))
	result, err := g.Resume(ctx, interrupts[0].ThreadID, map[string]any{
		interrupts[0].Name: "approved",
		interrupts[1].Name: "rejected",
	}, config)

	assert.Equal(t, 1, result, "The state should be the last valid state")
	superStepErr, _ = err.(*SuperStepExecutionError)
	assert.Len(t, superStepErr.Errs, 1, "Expected 2 errors due to interrupt rejection")
	assert.IsType(t, superStepErr.Errs[0], &InterruptRejectedError{}, "Error should be interrupt rejected")
}
