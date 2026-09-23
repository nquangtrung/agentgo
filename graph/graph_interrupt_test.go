package graph

import (
	"context"
	"fmt"
	"testing"

	"github.com/nquangtrung/agentgo/utils"
	"github.com/stretchr/testify/assert"
)

func TestGraphInterruptSameSuperStep(t *testing.T) {
	type expected struct {
		itrs   []string
		error  bool
		result int
	}
	type test struct {
		ops      []map[string]any
		expected []expected
		error    bool
		result   int
	}

	tt := []test{
		{
			ops: []map[string]any{
				{"double-itr": "approved"},
				{"triple-itr": "approved"},
			},
			expected: []expected{
				{itrs: []string{"double-itr", "triple-itr"}, error: true, result: 1},
				{itrs: []string{"triple-itr"}, error: true, result: 1},
			},
			result: 9,
			error:  false,
		},
		{
			ops: []map[string]any{
				{"triple-itr": "approved"},
				{"double-itr": "approved"},
			},
			expected: []expected{
				{itrs: []string{"double-itr", "triple-itr"}, error: true, result: 1},
				{itrs: []string{"double-itr"}, error: true, result: 1},
			},
			result: 9,
			error:  false,
		},
		{
			ops: []map[string]any{
				{"triple-itr": "rejected"},
				{"double-itr": "approved"},
			},
			expected: []expected{
				{itrs: []string{"double-itr", "triple-itr"}, error: true, result: 1},
				{itrs: []string{"double-itr"}, error: true, result: 1},
			},
			result: 1,
			error:  true,
		},
		{
			ops: []map[string]any{
				{"double-itr": "approved"},
				{"triple-itr": "rejected"},
			},
			expected: []expected{
				{itrs: []string{"double-itr", "triple-itr"}, error: true, result: 1},
				{itrs: []string{"triple-itr"}, error: true, result: 1},
			},
			result: 1,
			error:  true,
		},
		{
			ops: []map[string]any{
				{"double-itr": "rejected"},
				{"triple-itr": "approved"},
			},
			expected: []expected{
				{itrs: []string{"double-itr", "triple-itr"}, error: true, result: 1},
				{itrs: []string{"triple-itr"}, error: true, result: 1},
			},
			result: 1,
			error:  true,
		},
		{
			ops: []map[string]any{
				{"double-itr": "rejected"},
				{"triple-itr": "rejected"},
			},
			expected: []expected{
				{itrs: []string{"double-itr", "triple-itr"}, error: true, result: 1},
				{itrs: []string{"triple-itr"}, error: true, result: 1},
			},
			result: 1,
			error:  true,
		},
		{
			ops: []map[string]any{
				{"double-itr": "approved", "triple-itr": "approved"},
			},
			expected: []expected{
				{itrs: []string{"double-itr", "triple-itr"}, error: true, result: 1},
			},
			result: 9,
			error:  false,
		},
		{
			ops: []map[string]any{
				{"double-itr": "approved", "triple-itr": "rejected"},
			},
			expected: []expected{
				{itrs: []string{"double-itr", "triple-itr"}, error: true, result: 1},
			},
			result: 1,
			error:  true,
		},
		{
			ops: []map[string]any{
				{"double-itr": "rejected", "triple-itr": "approved"},
			},
			expected: []expected{
				{itrs: []string{"double-itr", "triple-itr"}, error: true, result: 1},
			},
			result: 1,
			error:  true,
		},
		{
			ops: []map[string]any{
				{"double-itr": "rejected", "triple-itr": "rejected"},
			},
			expected: []expected{
				{itrs: []string{"double-itr", "triple-itr"}, error: true, result: 1},
			},
			result: 1,
			error:  true,
		},
	}

	for _, tc := range tt {
		g := createMockGraphWith2InterruptInSameSuperStep()

		checkpointer := NewInMemoryCheckpointer[int, int]()
		config := InvocationConfig[int, int]{
			Checkpointer: checkpointer,
		}
		ctx := context.Background()
		result, err := g.Invoke(ctx, 0, config)

		for i := range tc.ops {
			assert.Equal(t, tc.expected[i].result, result, "The state should be the last valid state before the interrupt")
			if !tc.expected[i].error {
				assert.NoError(t, err, "Expected no error")
				continue
			}

			assert.Equal(t, tc.expected[i].result, result, "The state should be the last valid state before the interrupt")
			assert.Error(t, err, "Expected an error due to interrupt")
			superStepErr, _ := err.(*SuperStepExecutionError)
			interrupts := superStepErr.Interrupts()
			assert.Len(t, interrupts, len(tc.expected[i].itrs), fmt.Sprintf("Expected exact %d interrupt", len(tc.expected[i].itrs)))
			names := utils.Map(interrupts, func(itr *InterruptError) string {
				return itr.Name
			})
			assert.ElementsMatch(t, tc.expected[i].itrs, names, fmt.Sprintf("Expected interrupts to be from %v", tc.expected[i].itrs))

			result, err = g.Resume(ctx, interrupts[0].ThreadID, tc.ops[i], config)
		}

		if !tc.error {
			assert.NoError(t, err, "Expected no error")
		} else {
			assert.Error(t, err, "Expected an error due to interrupt rejection")
		}
		assert.Equal(t, tc.result, result, "The state should be the last valid state before the interrupt")
	}
}

func TestGraph2InterruptSameNode(t *testing.T) {
	type expected struct {
		itr   string
		error bool
	}
	type test struct {
		ops      []string
		expected []expected
		error    bool
		result   int
	}
	tt := []test{
		{
			ops: []string{"approved", "approved"},
			expected: []expected{
				{itr: "inc1-itr", error: true},
				{itr: "inc2-itr", error: true},
			},
			result: 1,
			error:  false,
		},
		{
			ops: []string{"approved", "rejected"},
			expected: []expected{
				{itr: "inc1-itr", error: true},
				{itr: "inc2-itr", error: true},
			},
			result: 0,
			error:  true,
		},
		{
			ops: []string{"rejected"},
			expected: []expected{
				{itr: "inc1-itr", error: true},
			},
			result: 0,
			error:  true,
		},
	}

	for _, tc := range tt {
		g := createMockGraphWith2InterruptInSingleNode()
		ctx := context.Background()
		checkpointer := NewInMemoryCheckpointer[int, int]()
		config := InvocationConfig[int, int]{
			Checkpointer: checkpointer,
		}
		result, err := g.Invoke(ctx, 0, config)

		for i := range tc.ops {
			if !tc.expected[i].error {
				assert.NoError(t, err, "Expected no error")
				continue
			}

			assert.Equal(t, 0, result, "The state should be the last valid state before the interrupt")
			assert.Error(t, err, "Expected an error due to interrupt")
			superStepErr, _ := err.(*SuperStepExecutionError)
			interrupts := superStepErr.Interrupts()
			assert.Len(t, interrupts, 1, "Expected exact 1 interrupt")
			assert.Equal(t, tc.expected[i].itr, interrupts[0].Name, "Expected interrupt from 'inc' node")

			result, err = g.Resume(ctx, interrupts[0].ThreadID, map[string]any{
				interrupts[0].Name: tc.ops[i],
			}, config)
		}

		if !tc.error {
			assert.NoError(t, err, "Expected no error")
		} else {
			assert.Error(t, err, "Expected an error due to interrupt rejection")
		}
		assert.Equal(t, tc.result, result, "The state should be the last valid state before the interrupt")
	}
}

func TestGraph2InterruptIn2Node(t *testing.T) {
	type expected struct {
		itr   string
		error bool
	}
	type test struct {
		ops      []string
		expected []expected
		error    bool
		result   int
	}
	tt := []test{
		{
			ops: []string{"approved", "approved"},
			expected: []expected{
				{itr: "inc-itr", error: true},
				{itr: "double-itr", error: true},
			},
			result: 3,
			error:  false,
		},
		{
			ops: []string{"approved", "rejected"},
			expected: []expected{
				{itr: "inc-itr", error: true},
				{itr: "double-itr", error: true},
			},
			result: 1,
			error:  true,
		},
		{
			ops: []string{"rejected"},
			expected: []expected{
				{itr: "inc-itr", error: true},
			},
			result: 0,
			error:  true,
		},
	}

	for _, tc := range tt {
		g := createMockGraphWith2InterruptIn2Node()
		ctx := context.Background()
		checkpointer := NewInMemoryCheckpointer[int, int]()
		config := InvocationConfig[int, int]{
			Checkpointer: checkpointer,
		}
		result, err := g.Invoke(ctx, 0, config)

		for i := range tc.ops {
			if !tc.expected[i].error {
				assert.NoError(t, err, "Expected no error")
				continue
			}

			assert.Error(t, err, "Expected an error due to interrupt")
			superStepErr, _ := err.(*SuperStepExecutionError)
			interrupts := superStepErr.Interrupts()
			assert.Len(t, interrupts, 1, "Expected exact 1 interrupt")
			assert.Equal(t, tc.expected[i].itr, interrupts[0].Name, fmt.Sprintf("Expected interrupt from '%s' node", tc.expected[i].itr))

			result, err = g.Resume(ctx, interrupts[0].ThreadID, map[string]any{
				interrupts[0].Name: tc.ops[i],
			}, config)
		}

		if !tc.error {
			assert.NoError(t, err, "Expected no error")
		} else {
			assert.Error(t, err, "Expected an error due to interrupt rejection")
		}
		assert.Equal(t, tc.result, result, "The state should be the expected result after all interrupts")
	}
}
