package graph

import (
	"context"
	"testing"

	"github.com/nquangtrung/agentgo/utils"
	"github.com/stretchr/testify/assert"
)

func TestSimpleGraph(t *testing.T) {
	type test struct {
		name          string
		initialState  int
		expectedState int
		expectedError bool
	}

	tt := []test{
		{
			name:          "simple increment",
			initialState:  0,
			expectedState: 1,
			expectedError: false,
		},
		{
			name:          "simple increment from 5",
			initialState:  5,
			expectedState: 11,
			expectedError: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := createSimpleGraphForVisualization()

			ctx := context.Background()
			result, err := g.Invoke(ctx, tc.initialState, InvocationConfig[int]{})

			if tc.expectedError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, tc.expectedState, result, "Expected final state to match")
			}
		})
	}
}

func TestGraphWithMultipleNodes(t *testing.T) {
	type test struct {
		name          string
		initialState  int
		expectedState int
		expectedError bool
	}

	tt := []test{
		{
			name:          "inc then double from 0",
			initialState:  0,
			expectedState: 3,
			expectedError: false,
		},
		{
			name:          "inc then double from 5",
			initialState:  5,
			expectedState: 33,
			expectedError: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := createMultipleNodesGraphForVisualization()

			ctx := context.Background()
			result, err := g.Invoke(ctx, tc.initialState, InvocationConfig[int]{})

			if tc.expectedError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, tc.expectedState, result, "Expected final state to match")
			}
		})
	}
}

func TestGraphWithFanOut(t *testing.T) {
	type test struct {
		name          string
		initialState  int
		expectedState int
		expectedError bool
	}

	tt := []test{
		{
			name:          "fanout inc and double from 1",
			initialState:  1,
			expectedState: 5,
			expectedError: false,
		},
		{
			name:          "fanout inc and double from 0",
			initialState:  0,
			expectedState: 1,
			expectedError: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := createFanOutGraphForVisualization()

			ctx := context.Background()
			result, err := g.Invoke(ctx, tc.initialState, InvocationConfig[int]{})

			if tc.expectedError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, tc.expectedState, result, "Expected final state to match")
			}
		})
	}
}

func TestGraphWithImbalanceNodes(t *testing.T) {
	type test struct {
		name          string
		initialState  int
		expectedState int
		expectedError bool
	}

	tt := []test{
		{
			name:          "imbalance nodes from 1",
			initialState:  1,
			expectedState: 18,
			expectedError: false,
		},
		{
			name:          "imbalance nodes from 0",
			initialState:  0,
			expectedState: 9,
			expectedError: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := createImbalanceNodesGraphForVisualization()

			ctx := context.Background()
			result, err := g.Invoke(ctx, tc.initialState, InvocationConfig[int]{})

			if tc.expectedError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, tc.expectedState, result, "Expected final state to match")
			}
		})
	}
}

func TestGraphWithErrorInNode(t *testing.T) {
	type test struct {
		name                string
		initialState        int
		expectedState       int
		expectedErrorCount  int
		expectedErrorNodes  []string
		expectedError       bool
	}

	tt := []test{
		{
			name:               "two nodes panic",
			initialState:       1,
			expectedState:      1,
			expectedErrorCount: 2,
			expectedErrorNodes: []string{"errorNode1", "errorNode2"},
			expectedError:      true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := New(func(a, b int) int {
				return a + b
			})

			g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
				return state + 1, nil
			})
			g.AddNode("errorNode1", func(ctx context.Context, state int) (int, error) {
				panic("intentional error 1")
			})
			g.AddNode("errorNode2", func(ctx context.Context, state int) (int, error) {
				panic("intentional error 2")
			})

			g.FanOut(START, []ID{"inc", "errorNode1", "errorNode2"})
			g.AddEdge("inc", END)
			g.AddEdge("errorNode1", END)
			g.AddEdge("errorNode2", END)

			ctx := context.Background()
			result, err := g.Invoke(ctx, tc.initialState, InvocationConfig[int]{})

			if tc.expectedError {
				assert.Error(t, err, "Expected an error due to the intentional panic")
				assert.IsType(t, &SuperStepExecutionError{}, err, "Expected error to be of type SuperStepExecutionError")

				superStepErr, _ := err.(*SuperStepExecutionError)
				assert.Len(t, superStepErr.Errs, tc.expectedErrorCount, "Expected exact error count in the super step execution error")

				errorNodes := utils.Map(superStepErr.Errs, func(err error) string {
					if nodeErr, ok := err.(*NodeExecutionError); ok {
						return nodeErr.ID
					}
					return "unknown"
				})
				assert.ElementsMatch(t, tc.expectedErrorNodes, errorNodes, "Expected error nodes to match")
			} else {
				assert.NoError(t, err, "Expected no error")
			}
			assert.Equal(t, tc.expectedState, result, "Expected final state to match")
		})
	}
}

func TestGraphWithErrorInReducer(t *testing.T) {
	type test struct {
		name            string
		initialState    int
		expectedState   int
		expectedError   bool
		expectedErrMsg  string
	}

	tt := []test{
		{
			name:           "reducer panic on zero",
			initialState:   1,
			expectedState:  1,
			expectedError:  true,
			expectedErrMsg: "panic in reducer execution: intentional error in reducer",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := New(func(a, b int) int {
				if b == 0 {
					panic("intentional error in reducer")
				}
				return a + b
			})

			g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
				return state + 1, nil
			})
			g.AddNode("zero", func(ctx context.Context, state int) (int, error) {
				return 0, nil
			})

			g.FanOut(START, []ID{"inc", "zero"})
			g.AddEdge("inc", END)
			g.AddEdge("zero", END)

			ctx := context.Background()
			result, err := g.Invoke(ctx, tc.initialState, InvocationConfig[int]{})

			if tc.expectedError {
				assert.Error(t, err, "Expected an error due to the intentional panic in reducer")
				assert.IsType(t, &ReducerExecutionError{}, err, "Expected error to be of type ReducerExecutionError")

				reducerErr, _ := err.(*ReducerExecutionError)
				assert.Equal(t, tc.expectedErrMsg, reducerErr.Err.Error(), "Expected error message to match")
			} else {
				assert.NoError(t, err, "Expected no error")
			}
			assert.Equal(t, tc.expectedState, result, "Expected final state to match")
		})
	}
}

func TestGraphWithErrorInRouter(t *testing.T) {
	type test struct {
		name            string
		initialState    int
		expectedState   int
		expectedError   bool
		expectedErrMsg  string
	}

	tt := []test{
		{
			name:           "router panic on negative state",
			initialState:   -1,
			expectedState:  -1,
			expectedError:  true,
			expectedErrMsg: "panic in router execution: intentional error in router",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := New(func(a, b int) int {
				return a + b
			})

			router := func(state int) []Target {
				if state < 0 {
					panic("intentional error in router")
				}
				return IDs("inc")
			}

			g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
				return state + 1, nil
			})

			g.AddConditionalEdge(START, router, []ID{"inc"})
			g.AddEdge("inc", END)

			ctx := context.Background()
			result, err := g.Invoke(ctx, tc.initialState, InvocationConfig[int]{})

			if tc.expectedError {
				assert.Error(t, err, "Expected an error due to the intentional panic in router")
				assert.IsType(t, &RouterExecutionError{}, err, "Expected error to be of type RouterExecutionError")

				routerErr, _ := err.(*RouterExecutionError)
				assert.Equal(t, START, routerErr.ID, "Expected error ID to match the START node ID")
				assert.Equal(t, tc.expectedErrMsg, routerErr.Err.Error(), "Expected error message to match")
			} else {
				assert.NoError(t, err, "Expected no error")
			}
			assert.Equal(t, tc.expectedState, result, "Expected final state to match")
		})
	}
}

func TestCycleGraphWithCondition(t *testing.T) {
	type test struct {
		name          string
		initialState  int
		expectedState int
		loopLimit     int
		expectedError bool
	}

	tt := []test{
		{
			name:          "loop until state >= 10",
			initialState:  0,
			expectedState: 15,
			loopLimit:     0,
			expectedError: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := createCycleGraphForVisualization()

			ctx := context.Background()
			result, err := g.Invoke(ctx, tc.initialState, InvocationConfig[int]{})

			if tc.expectedError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, tc.expectedState, result, "Expected final state to match")
			}
		})
	}
}

func TestInfiniteLoopGraph(t *testing.T) {
	type test struct {
		name             string
		initialState     int
		expectedError    bool
		expectedErrorMsg string
	}

	tt := []test{
		{
			name:             "infinite loop detection",
			initialState:     0,
			expectedError:    true,
			expectedErrorMsg: "InvocationError",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := New(func(a, b int) int {
				return a + b
			})
			g.AddNode("inc", func(ctx context.Context, state int) (int, error) {
				return state + 1, nil
			})

			g.AddEdge(START, "inc")
			g.AddConditionalEdge("inc", func(state int) []Target {
				return IDs("inc")
			}, []ID{"inc"})

			ctx := context.Background()
			_, err := g.Invoke(ctx, tc.initialState, InvocationConfig[int]{})

			if tc.expectedError {
				assert.Error(t, err, "Expected an error due to infinite loop in the graph")
				assert.IsType(t, &InvocationError{}, err, "Expected error to be of type InvocationError")
			} else {
				assert.NoError(t, err, "Expected no error")
			}
		})
	}
}

func TestGraphWithWorkerNode(t *testing.T) {
	type test struct {
		name          string
		initialState  int
		expectedState int
		expectedError bool
	}

	tt := []test{
		{
			name:          "worker node processes multiple tasks",
			initialState:  5,
			expectedState: 17,
			expectedError: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := createWorkerNodeGraphForVisualization()

			ctx := context.Background()
			result, err := g.Invoke(ctx, tc.initialState, InvocationConfig[int]{})

			if tc.expectedError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, tc.expectedState, result, "Expected final state to match")
			}
		})
	}
}

func TestGraphVisualize(t *testing.T) {
	type test struct {
		name          string
		initialState  int
		expectedState int
		expectedError bool
	}

	tt := []test{
		{
			name:          "complex graph with visualization",
			initialState:  1,
			expectedState: 26,
			expectedError: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g := createComplexMapWithNoInterrupt()
			saveGraph(g, "graph_visualize_test_complex_map_with_no_interrupt")

			ctx := context.Background()
			result, err := g.Invoke(ctx, tc.initialState, InvocationConfig[int]{})

			if tc.expectedError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, tc.expectedState, result, "Expected final state to match")
			}
		})
	}
}
