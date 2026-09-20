// Visualize graph
package visualizer

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type visualizerEdge struct {
	start string
	end   string
	label string
}
type visualizerState struct {
	name        string
	description string
	conditional bool
}

// Visualize the state graph using Mermaid syntax
type MermaidVisualizer struct {
	states []visualizerState
	edges  []visualizerEdge
}

func NewMermaidVisualizer() *MermaidVisualizer {
	return &MermaidVisualizer{}
}

func GetFunctionName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func (v *MermaidVisualizer) AddState(state string, description string, conditional bool) {
	v.states = append(v.states, visualizerState{
		name:        state,
		description: description,
		conditional: conditional,
	})
}
func (v *MermaidVisualizer) AddEdge(start string, end string, label string) {
	v.edges = append(v.edges, visualizerEdge{
		start: start,
		end:   end,
		label: label,
	})
}

func (v MermaidVisualizer) Visualize() string {
	var sb strings.Builder
	graphType := "stateDiagram-v2"

	fmt.Fprintf(&sb, "%s\n", graphType)
	for _, node := range v.states {
		if node.conditional {
			fmt.Fprintf(&sb, "    state %s <<choice>>\n", node.name)
		} else {
			fmt.Fprintf(&sb, "    state \"%s\" as %s\n", node.name, node.name)
		}
	}

	for _, edge := range v.edges {
		fmt.Fprintf(&sb, "    %s --> %s : %s\n", edge.start, edge.end, edge.label)
	}

	return sb.String()
}
