package agentgo

import (
	"log"
	"testing"
)

func TestGenerateTextGraph(t *testing.T) {
	g := createAgentGraph()

	log.Println(string(g.Visualize()))
}
