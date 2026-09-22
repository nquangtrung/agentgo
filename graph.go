package agentgo

import (
	"github.com/nquangtrung/agentgo/graph"
	"github.com/nquangtrung/agentgo/models"
)

type agentState struct {
	ToolExecutionsArchive *models.ToolExecutionsArchive
	Messages              *[]models.Message
}

func createGenerateTextGraph() graph.StateGraph[agentState] {
	g := graph.New[agentState](func(state1 agentState, state2 agentState) agentState {
		return state1
	})

	return g
}
