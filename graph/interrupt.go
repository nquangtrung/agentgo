package graph

import (
	"context"
	"fmt"
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
	name    string
	payload any
	Result  any
}

func (i InterruptResult) Reject(message string) *InterruptRejectedError {
	return NewInterruptRejectedError(i.name, message)
}

func Interrupt[T any](ctx context.Context, name string, payload any) (InterruptResult, error) {
	graph, ok := ctx.Value("graph").(StateGraph[T])
	if !ok {
		// This should never happens
		panic("graph not found in context")
	}

	if _, exists := graph.interrupts[name]; !exists {
		logger.Info("Interrupt does not exists", "id", name)
		graph.interrupts[name] = InterruptResult{
			name:    name,
			payload: payload,
		}

		// This error should end the execution of the graph and return the
		// interrupt control to the caller of the graph execution
		return graph.interrupts[name], NewInterruptError(name)
	}

	existingInterrupt := graph.interrupts[name]
	if existingInterrupt.Result == nil {
		logger.Debug("Interrupt exists but expecting result", "id", name)
		// The user has not set the result yet, we should wait for the result to be set
		return graph.interrupts[name], NewInterruptError(name)
	}

	logger.Info("Interrupt exists and has result", "id", name, "result", existingInterrupt.Result)

	// The user has set the result, we can return it
	return existingInterrupt, nil
}
