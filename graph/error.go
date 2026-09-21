package graph

import (
	"fmt"
	"log/slog"

	"github.com/nquangtrung/agentgo/utils"
)

type withCpError interface {
	SetCpError(err error)
	GetCpError() error
}
type withCpErrorBase struct {
	cpErr error
}

func (cp *withCpErrorBase) SetCpError(err error) {
	cp.cpErr = err
}

func (cp *withCpErrorBase) GetCpError() error {
	return cp.cpErr
}

type withID interface {
	SetID(id ID)
	GetID() ID
}

type withIDBase struct {
	ID string
}

func (e *withIDBase) SetID(id ID) {
	e.ID = id
}

func (e withIDBase) GetID() ID {
	return e.ID
}

type withThreadID interface {
	SetThreadID(id ID)
	GetThreadID() ID
}

type withThreadIDBase struct {
	ThreadID string
}

func (e *withThreadIDBase) SetThreadID(id ID) {
	e.ThreadID = id
}

func (e withThreadIDBase) GetThreadID() ID {
	return e.ThreadID
}

type NodeExecutionError struct {
	Err error

	withCpErrorBase
	withIDBase
	withThreadIDBase
}

func (e *NodeExecutionError) Error() string {
	return fmt.Sprintf("NodeExecutionError: ID=%s, Err=%v", e.ID, e.Err)
}

func (e *NodeExecutionError) Unwrap() error {
	return e.Err
}

func NewNodeExecutionError(threadId ID, nodeId ID, err error) *NodeExecutionError {
	return &NodeExecutionError{
		Err: err,
		withIDBase: withIDBase{
			ID: nodeId,
		},
		withThreadIDBase: withThreadIDBase{
			ThreadID: threadId,
		},
	}
}

type SuperStepExecutionError struct {
	Errs []error
	withCpErrorBase
}

func (e *SuperStepExecutionError) Error() string {
	return fmt.Sprintf(
		"Super step execution failed with multiple node errors. Affected nodes: %v",
		utils.Map(e.Errs, func(err error) string {
			if nodeErr, ok := err.(withID); ok {
				if nodeErr.GetID() != "" {
					return nodeErr.GetID()
				}
				return "unknown"
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
		if itr, ok := err.(*InterruptError); ok {
			interrupts = append(interrupts, itr)
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
		withIDBase: withIDBase{
			ID: r.id,
		},
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
	NodeExecutionError
}

func (e *RouterExecutionError) Error() string {
	return fmt.Sprintf("RouterExecutionError: ID=%s, Err=%v", e.ID, e.Err)
}

func NewRouterExecutionError(threadId ID, nodeId ID, err error) *RouterExecutionError {
	return &RouterExecutionError{
		NodeExecutionError: *NewNodeExecutionError(threadId, nodeId, err),
	}
}

type InvocationError struct {
	Err error
	withCpErrorBase
	withThreadIDBase
}

func (e *InvocationError) Error() string {
	return fmt.Sprintf("InvocationError: Err=%v", e.Err)
}

func (e *InvocationError) Unwrap() error {
	return e.Err
}

func NewInvocationError(err error) *InvocationError {
	return &InvocationError{
		Err: err,
	}
}

type InterruptRejectedError struct {
	Name    string
	Message string
	NodeExecutionError
}

func (e *InterruptRejectedError) Error() string {
	return fmt.Sprintf("InterruptRejectedError: %s Thread: %s Node: %s Message: %s", e.Name, e.ThreadID, e.ID, e.Message)
}

func NewInterruptRejectedError(name, message string) *InterruptRejectedError {
	return &InterruptRejectedError{
		Name:    name,
		Message: message,
		NodeExecutionError: NodeExecutionError{
			withIDBase:       withIDBase{},
			withThreadIDBase: withThreadIDBase{},
		},
	}
}
