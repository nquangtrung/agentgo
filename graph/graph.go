package graph

import (
	"log"
	"sync"

	"github.com/nquangtrung/agentgo/utils"
)

type NodeFn[T any] = func(state T) (T, error)
type Reducer[T any] = func(state1 T, state2 T) (T, error)

type ID = string

type StateEdge[T any] struct {
	Start ID
	End   []ID
}

type StateNode[T any] struct {
	ID ID
	fn NodeFn[T]
}

type nodeResult[T any] struct {
	id    ID
	state T
}
type stepInput[T any] struct {
	nodes []StateNode[T]
	state map[ID]T
}

type StateGraph[T any] struct {
	nodes   map[ID]StateNode[T]
	edges   map[ID]StateEdge[T]
	reducer Reducer[T]
}

func (g *StateGraph[T]) AddNode(name ID, fn NodeFn[T]) {
	id := name
	node := StateNode[T]{
		ID: id,
		fn: fn,
	}

	g.add(node)
}

func (g *StateGraph[T]) add(node StateNode[T]) {
	g.nodes[node.ID] = node
}

func (g *StateGraph[T]) AddEdge(start ID, end ID) {
	g.edges[start] = StateEdge[T]{
		Start: start,
		End:   []ID{end},
	}
}

func (g *StateGraph[T]) FanOut(start ID, ends []ID) {
	g.edges[start] = StateEdge[T]{
		Start: start,
		End:   ends,
	}
}

func (g *StateGraph[T]) executeSuperStep(input stepInput[T]) map[ID]nodeResult[T] {
	channel := make(chan nodeResult[T], len(input.nodes))
	func() {
		var wg sync.WaitGroup
		utils.Each(input.nodes, func(n StateNode[T]) {
			wg.Go(func() {
				log.Printf("Executing node %s", n.ID)
				state := input.state[n.ID]
				newState, err := n.fn(state)
				if err != nil {
					// TODO handle error
					// revert this step, try again.
				}

				log.Printf("Node %s executed, new state: %v", n.ID, newState)
				channel <- nodeResult[T]{
					id:    n.ID,
					state: newState,
				}
			})
		})
		wg.Wait()
		close(channel)
	}()

	result := make(map[ID]nodeResult[T])
	log.Printf("Waiting for results from %d nodes", len(input.nodes))
	for r := range channel {
		log.Printf("Received result from node %s: %v", r.id, r.state)
		result[r.id] = r
	}

	return result
}

func (g *StateGraph[T]) reduce(result map[ID]nodeResult[T], reduceFrom []ID) T {
	state := result[reduceFrom[0]].state
	for i := 1; i < len(reduceFrom); i++ {
		state, _ = g.reducer(state, result[reduceFrom[i]].state)
	}
	return state
}

func (g *StateGraph[T]) reduceAll(result map[ID]nodeResult[T], reduceFrom map[ID][]ID) map[ID]T {
	reducedState := make(map[ID]T)
	for id, from := range reduceFrom {
		log.Printf("Reducing state for node %s from nodes %v", id, from)
		reducedState[id] = g.reduce(result, from)
	}
	return reducedState
}

func (g *StateGraph[T]) createStepInput(result map[ID]nodeResult[T]) stepInput[T] {
	nodes := make(map[ID][]ID)
	for _, r := range result {
		utils.Each(g.edges[r.id].End, func(nextNodeId ID) {
			nodes[nextNodeId] = append(nodes[nextNodeId], r.id)
		})
	}

	return stepInput[T]{
		nodes: utils.Map(
			utils.Keys(nodes),
			func(id ID) StateNode[T] {
				return g.nodes[id]
			},
		),
		state: g.reduceAll(result, nodes),
	}
}

func (g *StateGraph[T]) Invoke(initial T) T {
	var input stepInput[T] = stepInput[T]{
		nodes: []StateNode[T]{g.nodes[START]},
		state: map[ID]T{
			START: initial,
		},
	}

	step := 0
	for ; len(input.nodes) > 0; step++ {
		log.Printf("Executing step %d with nodes: %v", step, utils.Map(input.nodes, func(n StateNode[T]) ID { return n.ID }))
		result := g.executeSuperStep(input)
		log.Printf("Step %d executed, results: %v", step, utils.Keys(result))
		input = g.createStepInput(result)
		log.Printf("Step %d created next step input with nodes: %v", step, utils.Map(input.nodes, func(n StateNode[T]) ID { return n.ID }))
	}

	return initial
}

func (g *StateGraph[T]) Compile() {
}

func New[T any](reducer Reducer[T]) StateGraph[T] {
	graph := StateGraph[T]{
		nodes:   make(map[ID]StateNode[T]),
		edges:   make(map[ID]StateEdge[T]),
		reducer: reducer,
	}
	graph.add(newStartNode[T]())
	graph.add(newEndNode[T]())
	return graph
}
