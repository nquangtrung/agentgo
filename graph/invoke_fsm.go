package graph

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nquangtrung/agentgo/fsm"
	"github.com/nquangtrung/agentgo/utils"
)

// invokeCtx holds all mutable state threaded through the FSM during Invoke.
type invokeCtx[T any, D any] struct {
	graph        StateGraph[T, D]
	config       InvocationConfig[T, D]
	checkpointer Checkpointer[T, D]
	threadId     string

	currentStep step[T, D]
	steps       []step[T, D]

	// err is set by invokeErrorState and read back by Invoke after the FSM exits.
	err error
}

// --- ErrorState ---

// invokeErrorState is the single error-handling state for the Invoke FSM.
// Every other state transitions here on failure instead of returning an error
// to the FSM runner. It checkpoints the last known step, attaches the
// checkpoint error where the error type supports it, stores the error in
// ic.err, and terminates the FSM by returning nil.
type invokeErrorState[T any, D any] struct {
	err error
}

func (s invokeErrorState[T, D]) Execute(ctx context.Context, ic *invokeCtx[T, D]) (fsm.State[invokeCtx[T, D]], error) {
	logger.Debug("ErrorState", slog.Any("error", s.err))
	cpErr := ic.checkpointer.Checkpoint(ic.threadId, ic.currentStep.toCheckpoint())
	if cpErr != nil {
		logger.Error("Checkpointing failed", slog.Any("error", cpErr))
	}
	if withCp, ok := s.err.(withCpError); ok {
		withCp.SetCpError(cpErr)
	}

	ic.err = s.err
	return nil, nil // terminate FSM
}

// --- ExecuteState ---

// executeState checks invocation guards then runs the current step's nodes.
type executeState[T any, D any] struct{}

func (s executeState[T, D]) Execute(ctx context.Context, ic *invokeCtx[T, D]) (fsm.State[invokeCtx[T, D]], error) {
	// Guard: no more targets → done
	if len(ic.currentStep.input.targets) == 0 {
		return nil, nil
	}

	// Guard: recursion limit / context cancellation
	if invocationError := ic.graph.detectInvocationError(ctx, ic.steps, ic.config); invocationError != nil {
		return invokeErrorState[T, D]{err: invocationError}, nil
	}

	logger.Debug("Executing step", slog.Int("step", len(ic.steps)), slog.Any("nodes",
		utils.Map(ic.currentStep.input.targets, func(n Target) ID { return n.id })))

	ic.currentStep.result = ic.graph.execute(ctx, ic.currentStep.input)

	logger.Debug("Step executed", slog.Int("step", len(ic.steps)), slog.Any("results", ic.currentStep.result))

	return barrierState[T, D]{}, nil
}

// --- BarrierState ---

// barrierState reduces results and routes to the next targets.
type barrierState[T any, D any] struct{}

func (s barrierState[T, D]) Execute(ctx context.Context, ic *invokeCtx[T, D]) (fsm.State[invokeCtx[T, D]], error) {
	input, err := ic.graph.barrier(ctx, ic.currentStep.input.state, ic.currentStep.result)
	if err != nil {
		logger.Warn("Error in barrier", slog.Any("error", err))
		return invokeErrorState[T, D]{err: err}, nil
	}
	return checkpointState[T, D]{nextInput: input}, nil
}

// --- CheckpointState ---

// checkpointState persists the current step then advances to the next one.
type checkpointState[T any, D any] struct {
	nextInput stepInput[T, D]
}

func (s checkpointState[T, D]) Execute(ctx context.Context, ic *invokeCtx[T, D]) (fsm.State[invokeCtx[T, D]], error) {
	if cpErr := ic.checkpointer.Checkpoint(ic.threadId, ic.currentStep.toCheckpoint()); cpErr != nil {
		return invokeErrorState[T, D]{err: NewInvocationError(fmt.Errorf("failed to checkpoint: %w", cpErr))}, nil
	}

	logger.Debug("Step checkpointed, advancing to nodes", slog.Int("step", len(ic.steps)), slog.Any("nodes",
		utils.Map(s.nextInput.targets, func(n Target) ID { return n.id })))

	ic.currentStep = step[T, D]{
		input:  s.nextInput,
		result: ic.currentStep.result,
	}
	ic.steps = append(ic.steps, ic.currentStep)

	return executeState[T, D]{}, nil
}
