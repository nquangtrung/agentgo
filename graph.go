package agentgo

import (
	"context"

	"github.com/nquangtrung/agentgo/graph"
	"github.com/nquangtrung/agentgo/models"
)

type agentState struct {
	ToolExecutionsArchive *models.ToolExecutionsArchive
	Messages              *[]models.Message
	TotalUsage            models.LanguageModelUsage
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

func createGenerateTextGraph() graph.StateGraph[agentState] {
	g := graph.New[agentState](func(state1 agentState, state2 agentState) agentState {
		return state1
	})

	g.AddNode(PREPARE_PROCESS, func(ctx context.Context, state agentState) (agentState, error) {
		return state, nil
	})
	g.AddNode(END_PROCESS, func(ctx context.Context, state agentState) (agentState, error) {
		return state, nil
	})
	g.AddNode(PREPARE_STEP, func(ctx context.Context, state agentState) (agentState, error) {
		return state, nil
	})
	g.AddNode(LOOP_CHECK, func(ctx context.Context, state agentState) (agentState, error) {
		return state, nil
	})
	g.AddNode(RESOLVE_TOOL, func(ctx context.Context, state agentState) (agentState, error) {
		return state, nil
	})
	g.AddWorkerNode(EXECUTE_TOOL, func(ctx context.Context, payload any) (agentState, error) {
		return agentState{}, nil
	})
	g.AddNode(PREPARE_TEXT, func(ctx context.Context, state agentState) (agentState, error) {
		return state, nil
	})
	g.AddNode(GENERATE_TEXT, func(ctx context.Context, state agentState) (agentState, error) {
		return state, nil
	})
	g.AddNode(STREAM_TEXT, func(ctx context.Context, state agentState) (agentState, error) {
		return state, nil
	})
	g.AddNode(END_TEXT, func(ctx context.Context, state agentState) (agentState, error) {
		return state, nil
	})
	g.AddNode(END_STEP, func(ctx context.Context, state agentState) (agentState, error) {
		return state, nil
	})

	g.AddEdge(graph.START, PREPARE_PROCESS)
	g.AddEdge(PREPARE_PROCESS, PREPARE_STEP)
	g.AddEdge(PREPARE_STEP, LOOP_CHECK)
	g.AddNamedConditionalEdge(LOOP_CHECK, func(state agentState) string {
		return ""
	}, map[string][]graph.Target{
		"tool": graph.IDs(RESOLVE_TOOL),
		"text": graph.IDs(PREPARE_TEXT),
		"end":  graph.IDs(END_STEP),
	})
	g.AddConditionalEdge(RESOLVE_TOOL, func(state agentState) []graph.Target {
		return []graph.Target{
			graph.Send(EXECUTE_TOOL, ""),
		}
	}, []graph.ID{EXECUTE_TOOL})
	g.AddEdge(EXECUTE_TOOL, END_STEP)
	g.AddNamedConditionalEdge(PREPARE_TEXT, func(state agentState) string {
		return ""
	}, map[string][]graph.Target{
		"generate": graph.IDs(GENERATE_TEXT),
		"stream":   graph.IDs(STREAM_TEXT),
	})
	g.AddEdge(STREAM_TEXT, END_TEXT)
	g.AddEdge(GENERATE_TEXT, END_TEXT)
	g.AddEdge(END_TEXT, END_STEP)
	g.AddNamedConditionalEdge(END_STEP, func(state agentState) string {
		return ""
	}, map[string][]graph.Target{
		"loop": graph.IDs(PREPARE_STEP),
		"end":  graph.IDs(END_PROCESS),
	})
	g.AddEdge(END_PROCESS, graph.END)

	return g
}
