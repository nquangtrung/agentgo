package graph

import (
	"fmt"
	"log"
	"sync"

	"github.com/nquangtrung/agentgo/utils"
)

type ID = string

type stepInput[T any] struct {
	nodes []StateNode[T]
	state T
}
type step[T any] struct {
	input        stepInput[T]
	result       []nodeResult[T]
	reducedState T
}

type StateGraph[T any] struct {
	nodes   map[ID]StateNode[T]
	edges   map[ID]StateEdge[T]
	reducer Reducer[T]
}

type InvocationConfig struct {
	RecursonLimit int
	StartNode     ID
}

type StateGraphConfig[T any] struct {
	RecursonLimit int
	StartNode     ID
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

func (g StateGraph[T]) executeAll(input stepInput[T], channel chan nodeResult[T]) {
	var wg sync.WaitGroup
	utils.Each(input.nodes, func(n StateNode[T]) {
		wg.Go(func() {
			log.Printf("Executing node %s", n.ID)
			state := input.state
			newState, err := n.execute(state)

			if err != nil {
				log.Printf("Error executing node %s: %v", n.ID, err)
				channel <- nodeResult[T]{
					id:    n.ID,
					state: state, // Return the original state in case of error
					err:   err,
				}
				return
			} else {
				log.Printf("Node %s executed, new state: %v", n.ID, newState)
				channel <- nodeResult[T]{
					id:    n.ID,
					state: newState,
				}
			}
		})
	})

	wg.Wait()
	close(channel)
}

func (g StateGraph[T]) execute(input stepInput[T]) []nodeResult[T] {
	channel := make(chan nodeResult[T], len(input.nodes))
	result := []nodeResult[T]{}

	go g.executeAll(input, channel)

	log.Printf("Waiting for results from %d nodes", len(input.nodes))
	for r := range channel {
		log.Printf("Received result from node %s: %v", r.id, r.state)
		result = append(result, r)
	}

	return result
}

func (g StateGraph[T]) reduce(state T, result []nodeResult[T]) (T, error) {
	reducedState := state
	for _, value := range result {
		if value.id == END || value.id == START {
			// terminal node, does not contribute to the state
			continue
		}

		// Again, we expect the reducer to handle errors internally
		// and return a valid state, so we don't handle errors here
		newState, reducerErr := executeReducer(g.reducer, reducedState, value.state)
		if reducerErr != nil {
			return reducedState, reducerErr
		}

		reducedState = newState
	}

	log.Printf("reduced state %v", reducedState)

	return reducedState, nil
}

func (g StateGraph[T]) route(state T, result []nodeResult[T]) ([]StateNode[T], *RouterExecutionError) {
	nodes := make(map[ID][]ID)
	for _, r := range result {
		edge := g.edges[r.id]
		if edge.Router == nil {
			utils.Each(g.edges[r.id].End, func(nextNodeId ID) {
				nodes[nextNodeId] = append(nodes[nextNodeId], r.id)
			})
			continue
		}

		routedNodes, err := edge.route(state)
		if err != nil {
			return []StateNode[T]{}, err
		}

		log.Printf("Routing from node %s with state %v to nodes: %v", r.id, r.state, routedNodes)

		utils.Each(routedNodes, func(nextNodeId ID) {
			nodes[nextNodeId] = append(nodes[nextNodeId], r.id)
		})
	}

	return utils.Map(
		utils.Keys(nodes),
		func(id ID) StateNode[T] {
			return g.nodes[id]
		},
	), nil
}

func (g StateGraph[T]) barrier(state T, result []nodeResult[T]) (stepInput[T], error) {
	// check for errors in the results
	executionError := NewSuperStepExecutionErrorFromResults(result)
	if executionError != nil {
		return stepInput[T]{}, executionError
	}

	// reduce result at barrier
	newState, reducerErr := g.reduce(state, result)
	if reducerErr != nil {
		return stepInput[T]{}, reducerErr
	}

	// route to next nodes
	nodes, routerError := g.route(newState, result)
	if routerError != nil {
		return stepInput[T]{}, routerError
	}

	return stepInput[T]{
		nodes: nodes,
		state: newState,
	}, nil
}

func (g StateGraph[T]) resolveStartNode(config InvocationConfig) StateNode[T] {
	if config.StartNode == "" {
		return g.nodes[START]
	}

	g.panicIfNodeNotExists(config.StartNode)
	return g.nodes[config.StartNode]
}

func (g StateGraph[T]) resolveRecursionLimit(config InvocationConfig) int {
	if config.RecursonLimit <= 0 {
		return 25
	}

	return config.RecursonLimit
}

func (g StateGraph[T]) detectInvocationError(steps []step[T], config InvocationConfig) *InvocationError {
	if len(steps) > g.resolveRecursionLimit(config) {
		return &InvocationError{
			fmt.Errorf("Recursion limit exceeded: %d", len(steps)),
		}
	}

	return nil
}

func (g StateGraph[T]) Invoke(initial T, config InvocationConfig) (T, error) {
	currentStep := step[T]{
		input: stepInput[T]{
			state: initial,
			nodes: []StateNode[T]{g.resolveStartNode(config)},
		},
		result: []nodeResult[T]{},
	}
	steps := []step[T]{
		currentStep,
	}

	for len(currentStep.input.nodes) > 0 {
		invocationError := g.detectInvocationError(steps, config)
		if invocationError != nil {
			log.Printf("Invocation error detected: %v", invocationError)
			return currentStep.input.state, invocationError
		}

		log.Printf("Executing step %d with nodes: %v", len(steps), utils.Map(currentStep.input.nodes, func(n StateNode[T]) ID { return n.ID }))
		result := g.execute(currentStep.input)
		currentStep.result = result

		log.Printf("Step %d executed, results: %v", len(steps), result)
		input, err := g.barrier(currentStep.input.state, result)
		if err != nil {
			// If we can't create the next step input,
			// we return the last valid state and the error
			log.Printf("Error creating next step input: %v", err)
			return currentStep.input.state, err
		}

		log.Printf("Step %d created next step input with nodes: %v", len(steps), utils.Map(currentStep.input.nodes, func(n StateNode[T]) ID { return n.ID }))
		currentStep = step[T]{
			input:  input,
			result: result,
		}
		steps = append(steps, currentStep)
	}

	return currentStep.input.state, nil
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
