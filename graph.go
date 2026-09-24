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
	tools             []models.ToolCall
	toolResults       []models.ToolExecuteOutput
	stream            bool
}

type agentState struct {
	toolExecutionsArchive *models.ToolExecutionsArchive
	messages              []models.Message
	totalUsage            models.LanguageModelUsage
	steps                 []step
	currentStep           step
	textGenerated         bool
	shouldEnd             bool
}

type agentStateDelta struct {
	from string

	addStep    *step
	stepAction string

	availableTools    []models.ToolCall
	archiveToolResult *models.ToolExecuteOutput

	textGenerated bool
	stream        bool
	shouldEnd     bool
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
	END_STEP        = "end_step"
)

func checkLoop(ctx context.Context, state agentState) (agentStateDelta, error) {
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
	oldState.currentStep.stream = oldState.currentStep.stream || delta.stream
	oldState.shouldEnd = oldState.shouldEnd || delta.shouldEnd

	if delta.archiveToolResult != nil {
		oldState = archiveToolResult(oldState, delta.archiveToolResult)
	}

	if delta.availableTools != nil {
		logger.Debug("Updating available tools", slog.Any("tools", delta.availableTools))
		oldState.currentStep.tools = delta.availableTools
	}

	if delta.stepAction != "" {
		logger.Info("Updating step action", slog.String("action", delta.stepAction))
		oldState.currentStep.action = delta.stepAction
	}

	return oldState
}

func createGenerateTextGraph() graph.StateGraph[agentState, agentStateDelta] {
	g := graph.New(accumulateAgentState)

	g.AddNode(PREPARE_PROCESS, prepareProcess)
	g.AddNode(END_PROCESS, endProcess)
	g.AddNode(PREPARE_STEP, prepareStep)
	g.AddNode(LOOP_CHECK, checkLoop)
	g.AddNode(RESOLVE_TOOL, resolveTool)
	g.AddWorkerNode(EXECUTE_TOOL, executeTool)
	g.AddNode(PREPARE_TEXT, prepareText)
	g.AddNode(GENERATE_TEXT, generateText)
	g.AddNode(STREAM_TEXT, streamText)
	g.AddNode(END_TEXT, endText)
	g.AddNode(END_STEP, endStep)

	g.AddEdge(graph.START, PREPARE_PROCESS)
	g.AddEdge(PREPARE_PROCESS, PREPARE_STEP)
	g.AddEdge(PREPARE_STEP, LOOP_CHECK)
	g.AddNamedConditionalEdge(LOOP_CHECK, func(state agentState) string {
		return state.currentStep.action
	}, map[string][]graph.Target{
		"tool": graph.IDs(RESOLVE_TOOL),
		"text": graph.IDs(PREPARE_TEXT),
		"end":  graph.IDs(END_STEP),
	})
	g.AddConditionalEdge(RESOLVE_TOOL, func(state agentState) []graph.Target {
		currentStep := state.currentStep
		if len(currentStep.tools) == 0 {
			return graph.IDs(PREPARE_TEXT)
		}

		return utils.Map(currentStep.tools, func(tool models.ToolCall) graph.Target {
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
	g.AddNamedConditionalEdge(END_STEP, func(state agentState) string {
		if state.shouldEnd || state.textGenerated {
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
