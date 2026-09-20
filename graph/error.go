package graph

import (
	"fmt"

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
	Errs []*NodeExecutionError
}

func (e *SuperStepExecutionError) Error() string {
	return fmt.Sprintf(
		"Super step execution failed with multiple node errors. Affected nodes: %v",
		utils.Map(e.Errs, func(err *NodeExecutionError) string { return err.ID }),
	)
}

func (e *SuperStepExecutionError) Unwrap() error {
	if len(e.Errs) > 0 {
		return e.Errs[0]
	}
	return nil
}

func NewSuperStepExecutionError(errs []*NodeExecutionError) *SuperStepExecutionError {
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
	errors := []*NodeExecutionError{}
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
