package graph

import (
	"fmt"
	"log/slog"

	"github.com/nquangtrung/agentgo/utils"
)

type WithCpError interface {
	WithCpError(err error)
}

type NodeExecutionError struct {
	ID  string
	Err error
}

func (e *NodeExecutionError) Error() string {
	return fmt.Sprintf("NodeExecutionError: ID=%s, Err=%v", e.ID, e.Err)
}

func (e *NodeExecutionError) Unwrap() error {
	return e.Err
}

func NewNodeExecutionError(id string, err error) *NodeExecutionError {
	return &NodeExecutionError{
		ID:  id,
		Err: err,
	}
}

type SuperStepExecutionError struct {
	Errs []error
}

func (e *SuperStepExecutionError) Error() string {
	return fmt.Sprintf(
		"Super step execution failed with multiple node errors. Affected nodes: %v",
		utils.Map(e.Errs, func(err error) string {
			if nodeErr, ok := err.(*NodeExecutionError); ok {
				return nodeErr.ID
			}
			return "unknown"
		}),
	)
}

func (e *SuperStepExecutionError) Unwrap() error {
	if len(e.Errs) > 0 {
		return e.Errs[0]
	}
	return nil
}

func (e *SuperStepExecutionError) Interrupts() []*InterruptError {
	logger.Debug("Finding interrupts", slog.Any("error", e))
	interrupts := []*InterruptError{}
	for _, err := range e.Errs {
		if _, ok := err.(*NodeExecutionError); !ok {
			continue
		}

		err := err.(*NodeExecutionError)
		logger.Debug("Checking NodeExecutionError for interrupts", slog.String("id", err.ID), slog.Any("error", err.Err))
		if invocationErr, ok := err.Err.(*InvocationError); ok {
			if interrupt, ok := invocationErr.Err.(*InterruptError); ok {
				interrupts = append(interrupts, interrupt)
			}
		} else if interrupt, ok := err.Err.(*InterruptError); ok {
			interrupts = append(interrupts, interrupt)
		}
	}
	return interrupts
}

func NewSuperStepExecutionError(errs []error) *SuperStepExecutionError {
	return &SuperStepExecutionError{
		Errs: errs,
	}
}

func NewNodeExecutionErrorFromResult[T any](r nodeResult[T]) *NodeExecutionError {
	return &NodeExecutionError{
		ID:  r.id,
		Err: r.err,
	}
}

func NewSuperStepExecutionErrorFromResults[T any](results []nodeResult[T]) *SuperStepExecutionError {
	errors := []error{}
	for _, r := range results {
		if r.err != nil {
			errors = append(errors, r.err)
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return &SuperStepExecutionError{
		Errs: errors,
	}
}

type ReducerExecutionError struct {
	Err error
}

func (e *ReducerExecutionError) Error() string {
	return fmt.Sprintf("ReducerExecutionError: Err=%v", e.Err)
}

func (e *ReducerExecutionError) Unwrap() error {
	return e.Err
}

func NewReducerExecutionError(err error) *ReducerExecutionError {
	return &ReducerExecutionError{
		Err: err,
	}
}

type RouterExecutionError struct {
	ID  ID
	Err error
}

func (e *RouterExecutionError) Error() string {
	return fmt.Sprintf("RouterExecutionError: ID=%s, Err=%v", e.ID, e.Err)
}

func (e *RouterExecutionError) Unwrap() error {
	return e.Err
}

func NewRouterExecutionError(id ID, err error) *RouterExecutionError {
	return &RouterExecutionError{
		ID:  id,
		Err: err,
	}
}

type InvocationError struct {
	Err     error
	cpError error
}

func (e *InvocationError) Error() string {
	return fmt.Sprintf("InvocationError: Err=%v", e.Err)
}

func (e *InvocationError) WithCpError(err error) {
	e.cpError = err
}

func (e *InvocationError) Unwrap() error {
	return e.Err
}

func NewInvocationError(err error) *InvocationError {
	return &InvocationError{
		Err: err,
	}
}
