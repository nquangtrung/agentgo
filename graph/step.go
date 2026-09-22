package graph

import "github.com/nquangtrung/agentgo/utils"

type stepInput[T any] struct {
	targets    []Target
	state      T
	lastResult []nodeResult[T]
}
type step[T any] struct {
	input        stepInput[T]
	result       []nodeResult[T]
	reducedState T
}

func (s step[T]) toCheckpoint() Checkpoint[T] {
	return Checkpoint[T]{
		State: s.input.state,
		Steps: utils.Map(s.input.targets, func(n Target) ID { return n.id }),
		Results: utils.Reduce(
			s.result,
			func(acc map[ID]T, result nodeResult[T]) map[ID]T {
				acc[result.id] = result.state
				return acc
			},
			map[ID]T{},
		),
	}
}
