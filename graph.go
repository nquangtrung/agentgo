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
	messages              *[]models.Message
	totalUsage            models.LanguageModelUsage
	steps                 []step
	currentStep           step
	textGenerated         bool
	shouldEnd             bool
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

func checkLoop(ctx context.Context, state agentState) (agentState, error) {
	endConditions := ctx.Value(models.EndConditionsContextKey).([]models.EndCondition)
	tools := ctx.Value(models.ToolsContextKey).([]models.BaseTool)
	var canProceedToNextStep func(context *models.ToolExecutionsArchive, endConds []models.EndCondition) bool
	canProceedToNextStep = func(context *models.ToolExecutionsArchive, endConds []models.EndCondition) bool {
		conditions := endConds
		for _, condition := range conditions {
			if condition.Condition(context) {
				return false
			}
		}
		return true
	}

	currentStep := state.currentStep
	if state.textGenerated {
		currentStep.action = "end"
	} else if len(endConditions) == 0 || len(tools) == 0 {
		currentStep.action = "text"
	} else if canProceedToNextStep(state.toolExecutionsArchive, endConditions) {
		currentStep.action = "tool"
	} else {
		currentStep.action = "end"
	}

	return agentState{
		currentStep: currentStep,
	}, nil
}

func accumulateAgentState(oldState agentState, newState agentState) agentState {
	logger.Info("Accumulate state", slog.Any("old", oldState), slog.Any("new", newState))

	return newState
}

func createGenerateTextGraph() graph.StateGraph[agentState] {
	g := graph.New(accumulateAgentState)

	g.AddNode(PREPARE_PROCESS, prepareProcess)
	g.AddNode(END_PROCESS, endProcess)
	g.AddNode(PREPARE_STEP, prepareStep)
	g.AddNode(LOOP_CHECK, checkLoop)
	g.AddNode(RESOLVE_TOOL, resolveTool)
	g.AddWorkerNode(EXECUTE_TOOL, executeTool)
	g.AddNode(PREPARE_TEXT, prepareText)
	g.AddNode(GENERATE_TEXT, func(ctx context.Context, state agentState) (agentState, error) {
		return state, nil
	})
	g.AddNode(STREAM_TEXT, func(ctx context.Context, state agentState) (agentState, error) {
		return state, nil
	})
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
		return utils.Map(currentStep.tools, func(tool models.ToolCall) graph.Target {
			return graph.Send(EXECUTE_TOOL, tool)
		})
	}, []graph.ID{EXECUTE_TOOL})
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
