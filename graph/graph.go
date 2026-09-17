package graph

import (
	"fmt"
	"log"
	"sync"

	"github.com/nquangtrung/agentgo/utils"
)

type NodeFn[T any] = func(state T) (T, error)
type Reducer[T any] = func(state1 T, state2 T) (T, error)
type Router[T any] = func(state T) ([]ID, error)

type ID = string

type StateEdge[T any] struct {
	Start  ID
	End    []ID
	Router Router[T]
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
	state T
}
type step[T any] struct {
	input  stepInput[T]
	result map[ID]nodeResult[T]
}

type StateGraph[T any] struct {
	nodes   map[ID]StateNode[T]
	edges   map[ID]StateEdge[T]
	reducer Reducer[T]
}

func (g StateGraph[T]) panicIfNodeExists(id ID) {
	if _, exists := g.nodes[id]; exists {
		panic(fmt.Errorf("Node %s exists", id))
	}
}

func (g StateGraph[T]) panicIfNodeNotExists(id ID) {
	if _, exists := g.nodes[id]; !exists {
		panic(fmt.Errorf("Node %s does not exist", id))
	}
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
	g.panicIfNodeExists(node.ID)
	g.nodes[node.ID] = node
}

func (g *StateGraph[T]) AddEdge(start ID, end ID) {
	g.panicIfNodeNotExists(start)
	g.panicIfNodeNotExists(end)
	g.edges[start] = StateEdge[T]{
		Start: start,
		End:   []ID{end},
	}
}

func (g *StateGraph[T]) AddConditionalEdge(start ID, router Router[T], ends []ID) {
	g.panicIfNodeNotExists(start)
	for _, end := range ends {
		g.panicIfNodeNotExists(end)
	}

	g.edges[start] = StateEdge[T]{
		Start:  start,
		End:    ends,
		Router: router,
	}
}

func (g *StateGraph[T]) FanOut(start ID, ends []ID) {
	g.panicIfNodeNotExists(start)
	for _, end := range ends {
		g.panicIfNodeNotExists(end)
	}
	g.edges[start] = StateEdge[T]{
		Start: start,
		End:   ends,
	}
}

func (g StateGraph[T]) executeSuperStep(input stepInput[T]) map[ID]nodeResult[T] {
	channel := make(chan nodeResult[T], len(input.nodes))
	func() {
		var wg sync.WaitGroup
		utils.Each(input.nodes, func(n StateNode[T]) {
			wg.Go(func() {
				log.Printf("Executing node %s", n.ID)
				state := input.state
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

func (g StateGraph[T]) reduce(state T, result map[ID]nodeResult[T]) T {
	reducedState := state
	for key, value := range result {
		if key == END || key == START {
			// terminal node, does not contribute to the state
			continue
		}

		newState, err := g.reducer(reducedState, value.state)
		if err != nil {
			// TODO handle error
		}

		reducedState = newState
	}
	return reducedState
}

func (g StateGraph[T]) createStepInput(state T, result map[ID]nodeResult[T]) stepInput[T] {
	nodes := make(map[ID][]ID)
	for _, r := range result {
		edge := g.edges[r.id]
		if edge.Router == nil {
			utils.Each(g.edges[r.id].End, func(nextNodeId ID) {
				nodes[nextNodeId] = append(nodes[nextNodeId], r.id)
			})
			continue
		}

		routedNodes, err := edge.Router(r.state)
		log.Printf("Routing from node %s with state %v to nodes: %v", r.id, r.state, routedNodes)
		if err != nil {
			// TODO handle error
		}
		utils.Each(routedNodes, func(nextNodeId ID) {
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
		state: g.reduce(state, result),
	}
}

func (g StateGraph[T]) Invoke(initial T) T {
	steps := []step[T]{
		step[T]{
			input: stepInput[T]{
				state: initial,
				nodes: []StateNode[T]{g.nodes[START]},
			},
			result: make(map[ID]nodeResult[T]),
		},
	}

	lastStep := steps[0]
	for len(lastStep.input.nodes) > 0 {
		log.Printf("Executing step %d with nodes: %v", len(steps), utils.Map(lastStep.input.nodes, func(n StateNode[T]) ID { return n.ID }))
		result := g.executeSuperStep(lastStep.input)
		lastStep.result = result
		log.Printf("Step %d executed, results: %v", len(steps), result)
		input := g.createStepInput(lastStep.input.state, result)
		log.Printf("Step %d created next step input with nodes: %v", len(steps), utils.Map(lastStep.input.nodes, func(n StateNode[T]) ID { return n.ID }))

		lastStep = step[T]{
			input:  input,
			result: result,
		}
		steps = append(steps, lastStep)
	}

	return lastStep.result[END].state
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
