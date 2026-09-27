package agentgo

import (
	"context"
	"errors"

	"github.com/nquangtrung/agentgo/graph"
	"github.com/nquangtrung/agentgo/models"
)

// ToolMiddlewareContext carries context passed to tool middleware hooks.
// Result is nil in Before hook, populated in After hook.
type ToolMiddlewareContext struct {
	ToolName string
	Params   map[string]any
	Result   *models.ToolExecuteOutput // nil in Before; populated in After
}

// ToolMiddleware intercepts tool execution.
// Before: called before tool.Execute(); returning an error (including graph.Interrupt)
//         skips execution. For Interrupt errors, execution is paused for human-in-the-loop.
//         For other errors, they are included in the tool result and execution continues.
// After:  called after tool.Execute() only on success (tool.Error == nil);
//         can validate output. Errors propagate the same way as Before.
type ToolMiddleware struct {
	Before func(ctx context.Context, mwCtx ToolMiddlewareContext) error
	After  func(ctx context.Context, mwCtx ToolMiddlewareContext) error
}

// StepMiddlewareContext carries context passed to step middleware hooks.
// All fields are computed from agentState and step, never exposing raw internals.
type StepMiddlewareContext struct {
	StepIndex        int      // 1-based index of the current step
	ToolChoice       string   // name of forced tool choice, empty if none
	ActiveTools      []string // filtered tool names for this step, nil if no filter
	TotalToolsCalled int      // total tool calls archived across all prior steps
	HasErrors        bool     // whether any prior step produced errors
	PriorStepCount   int      // number of completed steps before this one
}

// StepMiddleware intercepts step execution.
// Before: called after step is created in prepareStep; returning an error aborts
//         the entire pipeline. If the error is an Interrupt, execution pauses.
// After:  called in endStep only on success; can audit accumulated step state.
//         Errors abort the pipeline the same way.
type StepMiddleware struct {
	Before func(ctx context.Context, mwCtx StepMiddlewareContext) error
	After  func(ctx context.Context, mwCtx StepMiddlewareContext) error
}

// isInterruptError checks if an error is a graph.InterruptError.
func isInterruptError(err error) bool {
	var interruptErr *graph.InterruptError
	return errors.As(err, &interruptErr)
}

// buildStepMiddlewareContext constructs a StepMiddlewareContext from agentState and step.
func buildStepMiddlewareContext(state agentState, s step) StepMiddlewareContext {
	toolChoice := ""
	if s.prepareStepResult.ToolChoice != nil {
		toolChoice = s.prepareStepResult.ToolChoice.Name
	}

	var activeTools []string
	if s.prepareStepResult.ActiveTools != nil {
		activeTools = *s.prepareStepResult.ActiveTools
	}

	// Count total tool calls archived
	totalToolsCalled := state.toolExecutionsArchive.RecordCount()

	// Check if any prior step has errors
	hasErrors := false
	for _, err := range s.errors {
		if err != nil {
			hasErrors = true
			break
		}
	}

	return StepMiddlewareContext{
		StepIndex:        s.index,
		ToolChoice:       toolChoice,
		ActiveTools:      activeTools,
		TotalToolsCalled: totalToolsCalled,
		HasErrors:        hasErrors,
		PriorStepCount:   len(state.steps) - 1, // current step is already in steps
	}
}
