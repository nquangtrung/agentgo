package graph

import "github.com/nquangtrung/agentgo/utils"

type stepInput[T any, D any] struct {
	targets    []Target
	state      T
	lastResult []nodeResult[T, D]
}
type step[T any, D any] struct {
	input        stepInput[T, D]
	result       []nodeResult[T, D]
	reducedState T
}

func (s step[T, D]) toCheckpoint() Checkpoint[T, D] {
	return Checkpoint[T, D]{
		State: s.input.state,
		Steps: utils.Map(s.input.targets, func(n Target) ID { return n.id }),
		Results: utils.Reduce(
			s.result,
			func(acc map[ID]D, result nodeResult[T, D]) map[ID]D {
				acc[result.id] = result.delta
				return acc
			},
			map[ID]D{},
		),
	}
}
