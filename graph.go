package agentgo

import (
	"context"
	"log/slog"

	"github.com/nquangtrung/agentgo/graph"
	"github.com/nquangtrung/agentgo/models"
	"github.com/nquangtrung/agentgo/utils"
)

type step struct {
	index             int
	usage             models.LanguageModelUsage
	prepareStepResult PrepareStepResult
	action            string
	calls             []models.ToolCall
	toolResults       []models.ToolExecuteOutput
	stream            bool
	errors            []error
}

type agentState struct {
	toolExecutionsArchive *models.ToolExecutionsArchive
	messages              []models.Message
	totalUsage            models.LanguageModelUsage
	steps                 []step
	currentStep           step
	textGenerated         bool
	objectGenerated       bool
	objectAttempts        int
	objectError           error
	objectRaw             string
	shouldEnd             bool
}

type agentStateDelta struct {
	from string

	addStep    *step
	stepAction string

	availableTools    []models.ToolCall
	archiveToolResult *models.ToolExecuteOutput

	textGenerated bool
	objectGenerated bool
	stream        bool
	shouldEnd     bool

	// incrementObjectAttempt marks a failed object generation attempt.
	incrementObjectAttempt bool
	// objectError carries the validation/parse error from a failed object
	// attempt so the entry point can surface it after the graph ends.
	objectError error
	// objectRaw carries the raw text of the last failed attempt so the error
	// part can report it.
	objectRaw string
	// appendMessages adds messages to the conversation without archiving a
	// tool result. Used by the object repair loop to feed the model's bad
	// output and the validation error back into the next attempt.
	appendMessages []models.Message
	// usage folds a failed attempt's token usage into the step and total
	// usage without archiving a record.
	usage *models.LanguageModelUsage

	addError error
}

const (
	PREPARE_PROCESS = "prepare_process"
	END_PROCESS     = "end_process"
	PREPARE_STEP    = "prepare_step"
	LOOP_CHECK      = "loop_check"
	RESOLVE_TOOL    = "resolve_tool"
	EXECUTE_TOOL    = "execute_tool"
	PREPARE_TEXT    = "prepare_text"
	GENERATE_TEXT   = "generate_text"
	STREAM_TEXT     = "stream_text"
	END_TEXT        = "end_text"
	PREPARE_OBJECT  = "prepare_object"
	GENERATE_OBJECT = "generate_object"
	STREAM_OBJECT   = "stream_object"
	END_OBJECT      = "end_object"
	END_STEP        = "end_step"
)

func checkLoop(ctx context.Context, state agentState) (agentStateDelta, error) {
	// Object mode: a schema is present in the context. Object generation is a
	// single-shot call with a repair loop, so tools and end conditions do not
	// apply — MaxObjectRetries bounds the loop instead.
	if ctx.Value(models.SchemaContextKey) != nil {
		if state.objectGenerated {
			return agentStateDelta{
				from:       LOOP_CHECK,
				stepAction: "end",
				shouldEnd:  true,
			}, nil
		}
		return agentStateDelta{
			from:       LOOP_CHECK,
			stepAction: "object",
		}, nil
	}

	endConditions := ctx.Value(models.EndConditionsContextKey).([]models.EndCondition)
	tools := ctx.Value(models.ToolsContextKey).([]models.BaseTool)
	var canProceedToNextStep func(archive *models.ToolExecutionsArchive, endConds []models.EndCondition) bool
	canProceedToNextStep = func(archive *models.ToolExecutionsArchive, endConds []models.EndCondition) bool {
		logger.Debug("Checking loop conditions", slog.Int("stepIndex", state.currentStep.index), slog.Int("endConditions", len(endConds)), slog.Int("tools", len(tools)), slog.Any("archive", archive))
		conditions := endConds
		for _, condition := range conditions {
			if condition.Condition(archive) {
				return false
			}
		}
		return true
	}

	var stepAction string = "end"
	var shouldEnd bool = false
	if state.textGenerated {
		stepAction = "end"
		shouldEnd = true
	} else if len(endConditions) == 0 || len(tools) == 0 {
		stepAction = "text"
	} else if canProceedToNextStep(state.toolExecutionsArchive, endConditions) {
		stepAction = "tool"
	} else {
		stepAction = "end"
		shouldEnd = true
	}

	return agentStateDelta{
		from:       LOOP_CHECK,
		stepAction: stepAction,
		shouldEnd:  shouldEnd,
	}, nil
}

func archiveToolResult(oldState agentState, toolResult *models.ToolExecuteOutput) agentState {
	logger.Debug("Archiving tool result", slog.Int("stepIndex", oldState.currentStep.index), slog.Any("tool", toolResult.ToolCall))
	toolName := "text"
	if toolResult.ToolCall != nil {
		toolName = toolResult.ToolCall.ToolName
	}

	archive := oldState.toolExecutionsArchive
	archive.AddToolCallWithResult(toolName, toolResult)

	messages := oldState.messages
	messages = append(messages, models.NewMessageFromToolResult(*toolResult))

	currentStep := oldState.currentStep
	currentStep.usage = models.AccumulateUsage(currentStep.usage, toolResult.Usage)

	oldState.currentStep = currentStep
	oldState.totalUsage = models.AccumulateUsage(oldState.totalUsage, toolResult.Usage)
	oldState.toolExecutionsArchive = archive
	oldState.messages = messages
	oldState.toolExecutionsArchive = archive

	return oldState
}

func accumulateAgentState(oldState agentState, delta agentStateDelta) agentState {
	logger.Debug("Accumulate state", slog.String("delta", delta.from), slog.Int("stepIndex", oldState.currentStep.index))

	if delta.addStep != nil {
		logger.Debug("Adding new step", slog.Int("stepIndex", delta.addStep.index))
		oldState.currentStep = *delta.addStep
		oldState.steps = append(oldState.steps, *delta.addStep)
	}

	oldState.textGenerated = oldState.textGenerated || delta.textGenerated
	oldState.objectGenerated = oldState.objectGenerated || delta.objectGenerated
	// A successful attempt clears any error from a prior failed attempt.
	if delta.objectGenerated {
		oldState.objectError = nil
		oldState.objectRaw = ""
	}
	oldState.currentStep.stream = oldState.currentStep.stream || delta.stream
	oldState.shouldEnd = oldState.shouldEnd || delta.shouldEnd

	if delta.archiveToolResult != nil {
		oldState = archiveToolResult(oldState, delta.archiveToolResult)
	}

	if delta.incrementObjectAttempt {
		oldState.objectAttempts++
	}

	if delta.objectError != nil {
		oldState.objectError = delta.objectError
	}

	if delta.objectRaw != "" {
		oldState.objectRaw = delta.objectRaw
	}

	if len(delta.appendMessages) > 0 {
		oldState.messages = append(oldState.messages, delta.appendMessages...)
	}

	if delta.usage != nil {
		oldState.currentStep.usage = models.AccumulateUsage(oldState.currentStep.usage, *delta.usage)
		oldState.totalUsage = models.AccumulateUsage(oldState.totalUsage, *delta.usage)
	}

	if delta.availableTools != nil {
		logger.Debug("Updating available tools", slog.Any("tools", delta.availableTools))
		oldState.currentStep.calls = delta.availableTools
	}

	if delta.stepAction != "" {
		logger.Info("Updating step action", slog.String("action", delta.stepAction))
		oldState.currentStep.action = delta.stepAction
	}

	if delta.addError != nil {
		logger.Error("Adding error to current step", slog.Any("error", delta.addError))
		oldState.currentStep.errors = append(oldState.currentStep.errors, delta.addError)
	}

	return oldState
}

// createAgentGraph builds the shared state graph used by GenerateText,
// StreamText, GenerateObject and StreamObject. The object path is taken when
// a schema is present in the context; otherwise the text path runs.
func createAgentGraph() graph.StateGraph[agentState, agentStateDelta] {
	g := graph.New(accumulateAgentState)

	retry := graph.RetryOptions{
		ShouldRetry: func(err error) bool {
			return utils.IsTransientError(err)
		},
		Jitter: true,
	}

	g.AddNode(PREPARE_PROCESS, prepareProcess)
	g.AddNode(END_PROCESS, endProcess)
	g.AddNode(PREPARE_STEP, prepareStep)
	g.AddNode(LOOP_CHECK, checkLoop)
	g.AddNodeWithRetry(RESOLVE_TOOL, resolveTool, retry)
	g.AddWorkerNodeWithRetry(EXECUTE_TOOL, executeTool, retry)
	g.AddNode(PREPARE_TEXT, prepareText)
	g.AddNodeWithRetry(GENERATE_TEXT, generateText, retry)
	g.AddNodeWithRetry(STREAM_TEXT, streamText, retry)
	g.AddNode(END_TEXT, endText)
	g.AddNode(PREPARE_OBJECT, prepareObject)
	g.AddNodeWithRetry(GENERATE_OBJECT, generateObject, retry)
	g.AddNodeWithRetry(STREAM_OBJECT, streamObject, retry)
	g.AddNode(END_OBJECT, endObject)
	g.AddNode(END_STEP, endStep)

	g.AddEdge(graph.START, PREPARE_PROCESS)
	g.AddEdge(PREPARE_PROCESS, PREPARE_STEP)
	g.AddEdge(PREPARE_STEP, LOOP_CHECK)
	g.AddNamedConditionalEdge(LOOP_CHECK, func(state agentState) string {
		return state.currentStep.action
	}, map[string][]graph.Target{
		"tool":   graph.IDs(RESOLVE_TOOL),
		"text":   graph.IDs(PREPARE_TEXT),
		"object": graph.IDs(PREPARE_OBJECT),
		"end":    graph.IDs(END_STEP),
	})
	g.AddConditionalEdge(RESOLVE_TOOL, func(state agentState) []graph.Target {
		currentStep := state.currentStep
		if len(currentStep.calls) == 0 {
			return graph.IDs(PREPARE_TEXT)
		}

		return utils.Map(currentStep.calls, func(tool models.ToolCall) graph.Target {
			return graph.Send(EXECUTE_TOOL, tool)
		})
	}, []graph.ID{EXECUTE_TOOL, PREPARE_TEXT})
	g.AddEdge(EXECUTE_TOOL, END_STEP)
	g.AddNamedConditionalEdge(PREPARE_TEXT, func(state agentState) string {
		if state.currentStep.stream {
			return "stream"
		} else {
			return "generate"
		}
	}, map[string][]graph.Target{
		"generate": graph.IDs(GENERATE_TEXT),
		"stream":   graph.IDs(STREAM_TEXT),
	})
	g.AddEdge(STREAM_TEXT, END_TEXT)
	g.AddEdge(GENERATE_TEXT, END_TEXT)
	g.AddEdge(END_TEXT, END_STEP)
	g.AddNamedConditionalEdge(PREPARE_OBJECT, func(state agentState) string {
		if state.currentStep.stream {
			return "stream"
		} else {
			return "generate"
		}
	}, map[string][]graph.Target{
		"generate": graph.IDs(GENERATE_OBJECT),
		"stream":   graph.IDs(STREAM_OBJECT),
	})
	g.AddEdge(STREAM_OBJECT, END_OBJECT)
	g.AddEdge(GENERATE_OBJECT, END_OBJECT)
	g.AddEdge(END_OBJECT, END_STEP)
	g.AddNamedConditionalEdge(END_STEP, func(state agentState) string {
		if state.shouldEnd || state.textGenerated || state.objectGenerated {
			return "end"
		}
		return "loop"
	}, map[string][]graph.Target{
		"loop": graph.IDs(PREPARE_STEP),
		"end":  graph.IDs(END_PROCESS),
	})
	g.AddEdge(END_PROCESS, graph.END)

	return g
}
