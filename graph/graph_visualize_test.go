package graph

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func saveGraph(g StateGraph[int, int], name string) {
	visualization := g.Visualize()
	toBeSaved := fmt.Sprintf("```mermaid\n%s\n```", visualization)
	err := os.WriteFile(
		fmt.Sprintf("%s.md", name),
		[]byte(toBeSaved),
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func TestVisualize(t *testing.T) {
	type test struct {
		name       string
		graphFunc  func() StateGraph[int, int]
		outputName string
	}

	tt := []test{
		{
			name:       "visualize 2 interrupts in same super step",
			graphFunc:  createMockGraphWith2InterruptInSameSuperStep,
			outputName: "graph_visualize_test_2ItrsInSameSuperStep",
		},
		{
			name:       "visualize 2 interrupts in same node",
			graphFunc:  createMockGraphWith2InterruptInSingleNode,
			outputName: "graph_visualize_test_2ItrsInSameNode",
		},
		{
			name:       "visualize 2 interrupts in 2 nodes",
			graphFunc:  createMockGraphWith2InterruptIn2Node,
			outputName: "graph_visualize_test_2ItrsIn2Nodes",
		},
		{
			name:       "visualize simple graph with single node",
			graphFunc:  createSimpleGraphForVisualization,
			outputName: "graph_visualize_test_simple_graph",
		},
		{
			name:       "visualize graph with multiple sequential nodes",
			graphFunc:  createMultipleNodesGraphForVisualization,
			outputName: "graph_visualize_test_multiple_nodes",
		},
		{
			name:       "visualize graph with fan-out",
			graphFunc:  createFanOutGraphForVisualization,
			outputName: "graph_visualize_test_fanout",
		},
		{
			name:       "visualize graph with imbalanced nodes",
			graphFunc:  createImbalanceNodesGraphForVisualization,
			outputName: "graph_visualize_test_imbalanced_nodes",
		},
		{
			name:       "visualize graph with cycle and condition",
			graphFunc:  createCycleGraphForVisualization,
			outputName: "graph_visualize_test_cycle",
		},
		{
			name:       "visualize graph with worker node",
			graphFunc:  createWorkerNodeGraphForVisualization,
			outputName: "graph_visualize_test_worker_node",
		},
		{
			name:       "visualize complex graph",
			graphFunc:  createComplexMapWithNoInterrupt,
			outputName: "graph_visualize_test_complex_map",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := tc.graphFunc()
			saveGraph(g, tc.outputName)
		})
	}
}
