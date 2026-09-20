package graph

import (
	"context"
	"fmt"
	"log"

	"github.com/nquangtrung/agentgo/fsm"
	"github.com/nquangtrung/agentgo/utils"
)

// invokeCtx holds all mutable state threaded through the FSM during Invoke.
type invokeCtx[T any] struct {
	graph        StateGraph[T]
	config       InvocationConfig[T]
	checkpointer Checkpointer[T]
	threadId     string

	currentStep step[T]
	steps       []step[T]

	// err is set by invokeErrorState and read back by Invoke after the FSM exits.
	err error
}

// --- ErrorState ---

// invokeErrorState is the single error-handling state for the Invoke FSM.
// Every other state transitions here on failure instead of returning an error
// to the FSM runner. It checkpoints the last known step, attaches the
// checkpoint error where the error type supports it, stores the error in
// ic.err, and terminates the FSM by returning nil.
type invokeErrorState[T any] struct {
	err error
}

func (s invokeErrorState[T]) Execute(ctx context.Context, ic *invokeCtx[T]) (fsm.State[invokeCtx[T]], error) {
	log.Printf("Invoke FSM error state: %v", s.err)

	cpErr := ic.checkpointer.Checkpoint(ic.threadId, ic.currentStep.ToCheckpoint())
	if cpErr != nil {
		log.Printf("Additionally, checkpointing failed: %v", cpErr)
	}
	if withCp, ok := s.err.(WithCpError); ok {
		withCp.WithCpError(cpErr)
	}

	ic.err = s.err
	return nil, nil // terminate FSM
}

// --- ExecuteState ---

// executeState checks invocation guards then runs the current step's nodes.
type executeState[T any] struct{}

func (s executeState[T]) Execute(ctx context.Context, ic *invokeCtx[T]) (fsm.State[invokeCtx[T]], error) {
	// Guard: no more targets → done
	if len(ic.currentStep.input.targets) == 0 {
		return nil, nil
	}

	// Guard: recursion limit / context cancellation
	if invocationError := ic.graph.detectInvocationError(ctx, ic.steps, ic.config); invocationError != nil {
		return invokeErrorState[T]{err: invocationError}, nil
	}

	log.Printf("Executing step %d with nodes: %v", len(ic.steps),
		utils.Map(ic.currentStep.input.targets, func(n Target) ID { return n.id }))

	ic.currentStep.result = ic.graph.execute(ctx, ic.currentStep.input)

	log.Printf("Step %d executed, results: %v", len(ic.steps), ic.currentStep.result)

	return barrierState[T]{}, nil
}

// --- BarrierState ---

// barrierState reduces results and routes to the next targets.
type barrierState[T any] struct{}

func (s barrierState[T]) Execute(ctx context.Context, ic *invokeCtx[T]) (fsm.State[invokeCtx[T]], error) {
	input, err := ic.graph.barrier(ctx, ic.currentStep.input.state, ic.currentStep.result)
	if err != nil {
		log.Printf("Error in barrier: %v", err)
		return invokeErrorState[T]{err: err}, nil
	}
	return checkpointState[T]{nextInput: input}, nil
}

// --- CheckpointState ---

// checkpointState persists the current step then advances to the next one.
type checkpointState[T any] struct {
	nextInput stepInput[T]
}

func (s checkpointState[T]) Execute(ctx context.Context, ic *invokeCtx[T]) (fsm.State[invokeCtx[T]], error) {
	if cpErr := ic.checkpointer.Checkpoint(ic.threadId, ic.currentStep.ToCheckpoint()); cpErr != nil {
		return invokeErrorState[T]{err: NewInvocationError(fmt.Errorf("failed to checkpoint: %w", cpErr))}, nil
	}

	log.Printf("Step %d checkpointed, advancing to nodes: %v", len(ic.steps),
		utils.Map(s.nextInput.targets, func(n Target) ID { return n.id }))

	ic.currentStep = step[T]{
		input:  s.nextInput,
		result: ic.currentStep.result,
	}
	ic.steps = append(ic.steps, ic.currentStep)

	return executeState[T]{}, nil
}
