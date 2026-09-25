package agentgo

import (
	"log"
	"testing"
)

func TestGenerateTextGraph(t *testing.T) {
	g := createGenerateTextGraph()

	log.Println(string(g.Visualize()))
}
