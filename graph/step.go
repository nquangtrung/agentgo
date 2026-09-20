package graph

import "github.com/nquangtrung/agentgo/utils"

type stepInput[T any] struct {
	targets []Target
	state   T
}
type step[T any] struct {
	input        stepInput[T]
	result       []nodeResult[T]
	reducedState T
}

func (s step[T]) ToCheckpoint() Checkpoint[T] {
	return Checkpoint[T]{
		State: s.input.state,
		Steps: utils.Map(s.input.targets, func(n Target) ID { return n.id }),
	}
}
