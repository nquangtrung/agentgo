package graph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type InterruptError struct {
	Name string
	NodeExecutionError
}

func (e *InterruptError) Error() string {
	return fmt.Sprintf("InterruptError: %s Thread: %s Node: %s", e.Name, e.ThreadID, e.ID)
}

func NewInterruptError(name string) *InterruptError {
	return &InterruptError{
		Name: name,
		NodeExecutionError: NodeExecutionError{
			withIDBase:       withIDBase{},
			withThreadIDBase: withThreadIDBase{},
		},
	}
}

type InterruptResult struct {
	Name    string
	Payload any
	Result  any
}

func Interrupt[T any](ctx context.Context, payload any) (any, error) {
	graph, ok := ctx.Value("graph").(StateGraph[T])
	if !ok {
		// This should never happens
		panic("graph not found in context")
	}

	interruptUid := uuid.New().String()
	if _, exists := graph.interrupts[interruptUid]; !exists {
		logger.Debug("Interrupt does not exists", "id", interruptUid)
		graph.interrupts[interruptUid] = InterruptResult{
			Name:    interruptUid,
			Payload: payload,
		}

		// This error should end the execution of the graph and return the
		// interrupt control to the caller of the graph execution
		return nil, NewInterruptError(interruptUid)
	}

	existingInterrupt := graph.interrupts[interruptUid]
	if existingInterrupt.Result == nil {
		logger.Debug("Interrupt exists but expecting result", "id", interruptUid)
		// The user has not set the result yet, we should wait for the result to be set
		return nil, NewInterruptError(interruptUid)
	}

	// The user has set the result, we can return it
	return existingInterrupt, nil
}
