package graph

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/nquangtrung/agentgo/fsm"
	"github.com/nquangtrung/agentgo/utils"
	"github.com/nquangtrung/agentgo/visualizer"
)

type ID = string

type Visualizer interface {
	Visualize() string
	AddState(state string, description string, conditional bool)
	AddEdge(start string, end string, label string)
}

type StateGraph[T any] struct {
	nodes      map[ID]node[T]
	edges      map[ID]stateEdge[T]
	reducer    Reducer[T]
	visualizer Visualizer

	interrupts map[string]InterruptResult
}

type InvocationConfig[T any] struct {
	RecursonLimit int
	StartNodes    []ID
	ThreadId      string
	Checkpointer  Checkpointer[T]
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
	node := &stateNode[T]{
		ID: id,
		fn: fn,
	}

	g.add(node)
}

func (g *StateGraph[T]) AddWorkerNode(name ID, fn WorkerNodeFn[T]) {
	id := name
	node := &workerNode[T]{
		ID: id,
		fn: fn,
	}

	g.add(node)
}

func (g *StateGraph[T]) add(node node[T]) {
	g.panicIfNodeExists(node.id())
	g.nodes[node.id()] = node
}

func (g *StateGraph[T]) AddEdge(start ID, end ID) {
	g.panicIfNodeNotExists(start)
	g.panicIfNodeNotExists(end)
	g.edges[start] = stateEdge[T]{
		start: start,
		end:   []ID{end},
	}
}

func (g *StateGraph[T]) AddConditionalEdge(start ID, router Router[T], ends []ID) {
	g.panicIfNodeNotExists(start)
	for _, end := range ends {
		g.panicIfNodeNotExists(end)
	}

	g.edges[start] = stateEdge[T]{
		start:  start,
		end:    ends,
		router: router,
	}
}

func (g *StateGraph[T]) AddNamedConditionalEdge(start ID, NamedRouter NamedRouter[T], ends NamedRouterMap) {
	g.panicIfNodeNotExists(start)
	for _, endList := range ends {
		for _, end := range endList {
			g.panicIfNodeNotExists(end.id)
		}
	}

	router := func(state T) []Target {
		name := NamedRouter(state)
		if targetList, exists := ends[name]; exists {
			return targetList
		}
		return []Target{}
	}

	g.edges[start] = stateEdge[T]{
		start:  start,
		end:    []ID{},
		endMap: ends,
		router: router,
	}
}

func (g *StateGraph[T]) FanOut(start ID, ends []ID) {
	g.panicIfNodeNotExists(start)
	for _, end := range ends {
		g.panicIfNodeNotExists(end)
	}
	g.edges[start] = stateEdge[T]{
		start: start,
		end:   ends,
	}
}

func (g StateGraph[T]) executeAll(ctx context.Context, input stepInput[T], channel chan nodeResult[T]) {
	var wg sync.WaitGroup
	utils.Each(input.targets, func(target Target) {
		wg.Go(func() {
			log.Printf("Executing node %s", target.id)
			state := input.state
			node := g.nodes[target.id]
			newState, err := node.execute(ctx, state, target)

			if err != nil {
				log.Printf("Error executing node %s: %v", target.id, err)
				channel <- nodeResult[T]{
					id:    target.id,
					state: state, // Return the original state in case of error
					err:   err,
				}
				return
			} else {
				log.Printf("Node %s executed, new state: %v", target.id, newState)
				channel <- nodeResult[T]{
					id:    target.id,
					state: newState,
				}
			}
		})
	})

	wg.Wait()
	close(channel)
}

func (g StateGraph[T]) execute(ctx context.Context, input stepInput[T]) []nodeResult[T] {
	channel := make(chan nodeResult[T], len(input.targets))
	result := []nodeResult[T]{}

	go g.executeAll(ctx, input, channel)

	log.Printf("Waiting for results from %d nodes", len(input.targets))
	for {
		select {
		case r, ok := <-channel:
			if !ok {
				log.Printf("Channel closed, all results received")
				return result
			}
			log.Printf("Received result from node %s: %v", r.id, r.state)
			result = append(result, r)
		case <-ctx.Done():
			// The context error will be handled in the caller, we just return the results received so far
			log.Printf("Context done, returning results received so far, the context error: %v", ctx.Err())
			return result
		}
	}
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

func (g StateGraph[T]) route(state T, result []nodeResult[T]) ([]Target, *RouterExecutionError) {
	newTargets := []Target{}
	for _, r := range result {
		edge := g.edges[r.id]
		targets, err := edge.route(state)
		if err != nil {
			return []Target{}, err
		}

		log.Printf("Routing from node %s with state %v to nodes: %v", r.id, r.state, targets)
		newTargets = append(newTargets, targets...)
	}

	return newTargets, nil
}

func (g StateGraph[T]) barrier(_ context.Context, state T, result []nodeResult[T]) (stepInput[T], error) {
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
		targets: nodes,
		state:   newState,
	}, nil
}

func (g StateGraph[T]) resolveStartNode(config InvocationConfig[T]) []Target {
	if len(config.StartNodes) == 0 {
		return IDs(START)
	}

	for _, startNode := range config.StartNodes {
		if _, exists := g.nodes[startNode]; !exists {
			panic(fmt.Errorf("Start node %s does not exist", startNode))
		}
	}
	return IDs(config.StartNodes...)
}

func (g StateGraph[T]) resolveRecursionLimit(config InvocationConfig[T]) int {
	if config.RecursonLimit <= 0 {
		return 25
	}

	return config.RecursonLimit
}

func (g StateGraph[T]) detectInvocationError(ctx context.Context, steps []step[T], config InvocationConfig[T]) *InvocationError {
	if ctx.Err() != nil {
		return NewInvocationError(fmt.Errorf("Context error: %v", ctx.Err()))
	}

	if len(steps) > g.resolveRecursionLimit(config) {
		return NewInvocationError(fmt.Errorf("Recursion limit exceeded: %d", len(steps)))
	}

	return nil
}

func (g StateGraph[T]) Resume(ctx context.Context, threadId string, interruptResults map[string]any, config InvocationConfig[T]) (T, error) {
	checkpointer := g.resolveCheckpointer(config)

	for interruptName, result := range interruptResults {
		if existingInterrupt, exists := g.interrupts[interruptName]; exists {
			existingInterrupt.Result = result
			g.interrupts[interruptName] = existingInterrupt
		} else {
			return *new(T), NewInvocationError(fmt.Errorf("Interrupt %s does not exist", interruptName))
		}
	}

	cp, err := checkpointer.Restore(threadId)
	if err != nil {
		return *new(T), NewInvocationError(fmt.Errorf("Failed to restore checkpoint: %v", err))
	}

	invocationConfig := InvocationConfig[T]{
		RecursonLimit: config.RecursonLimit,
		StartNodes:    cp.Steps,
		ThreadId:      threadId,
		Checkpointer:  checkpointer,
	}
	state := cp.State
	return g.Invoke(ctx, state, invocationConfig)
}

func (g StateGraph[T]) resolveThreadId(config InvocationConfig[T]) string {
	if config.ThreadId == "" {
		return uuid.New().String()
	}

	return config.ThreadId
}

func (g StateGraph[T]) resolveCheckpointer(config InvocationConfig[T]) Checkpointer[T] {
	if config.Checkpointer == nil {
		return NewInMemoryCheckpointer[T]()
	}

	return config.Checkpointer
}

func (g StateGraph[T]) Invoke(ctx context.Context, initial T, config InvocationConfig[T]) (T, error) {
	ctx = context.WithValue(ctx, "graph", g)

	threadId := g.resolveThreadId(config)
	checkpointer := g.resolveCheckpointer(config)

	initialStep := step[T]{
		input: stepInput[T]{
			state:   initial,
			targets: g.resolveStartNode(config),
		},
		result: []nodeResult[T]{},
	}

	ic := &invokeCtx[T]{
		graph:        g,
		config:       config,
		checkpointer: checkpointer,
		threadId:     threadId,
		currentStep:  initialStep,
		steps:        []step[T]{initialStep},
	}

	machine := fsm.New[invokeCtx[T]]()
	if fsmErr := machine.Run(ctx, executeState[T]{}, ic); fsmErr != nil {
		return ic.currentStep.input.state, fsmErr
	}

	if ic.err != nil {
		return ic.currentStep.input.state, ic.err
	}

	err := checkpointer.Checkpoint(threadId, ic.currentStep.ToCheckpoint())
	return ic.currentStep.input.state, err
}

// Generate a mermaid diagram of the state graph.
func (g StateGraph[T]) Visualize() string {
	for _, node := range g.nodes {
		g.visualizer.AddState(node.id(), node.id(), false)
	}

	for _, edge := range g.edges {
		if edge.endMap != nil {
			for name, targetList := range edge.endMap {
				for _, target := range targetList {
					g.visualizer.AddEdge(edge.start, target.id, name)
				}
			}
		} else if edge.router != nil {
			routerName := fmt.Sprintf("router_%s", edge.start)
			g.visualizer.AddState(
				routerName,
				fmt.Sprintf("Router for %s", edge.start),
				true,
			)
			g.visualizer.AddEdge(edge.start, routerName, "")
			for _, end := range edge.end {
				g.visualizer.AddEdge(routerName, end, "")
			}
		} else {
			for _, end := range edge.end {
				g.visualizer.AddEdge(edge.start, end, "")
			}
		}
	}
	return g.visualizer.Visualize()
}

func (g *StateGraph[T]) Compile() {
}

func New[T any](reducer Reducer[T]) StateGraph[T] {
	graph := StateGraph[T]{
		nodes:      make(map[ID]node[T]),
		edges:      make(map[ID]stateEdge[T]),
		reducer:    reducer,
		visualizer: visualizer.NewMermaidVisualizer(),
		interrupts: make(map[string]InterruptResult),
	}
	graph.add(newStartNode[T]())
	graph.add(newEndNode[T]())
	return graph
}
