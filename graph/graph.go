// Implementation of a Pregel-like state graph for managing state transitions and node executions in a concurrent environment.
package graph

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/nquangtrung/agentgo/fsm"
	"github.com/nquangtrung/agentgo/utils"
	"github.com/nquangtrung/agentgo/visualizer"
)

const TAG string = "[Graph]"

type ID = string

type Visualizer interface {
	Visualize() []byte
	AddState(state string, description string, conditional bool)
	AddEdge(start string, end string, label string)
}

type StateGraph[T any, D any] struct {
	nodes      map[ID]node[T, D]
	edges      map[ID]stateEdge[T, D]
	reducer    Reducer[T, D]
	visualizer Visualizer

	interrupts map[string]InterruptResult
}

type InvocationConfig[T any, D any] struct {
	RecursonLimit int
	StartNodes    []ID
	ThreadId      string
	Checkpointer  Checkpointer[T, D]

	partialResults map[ID]D
}

type StateGraphConfig[T any] struct {
	RecursonLimit int
	StartNode     ID
}

func (g StateGraph[T, D]) panicIfNodeExists(id ID) {
	if _, exists := g.nodes[id]; exists {
		panic(fmt.Errorf("Node %s exists", id))
	}
}

func (g StateGraph[T, D]) panicIfNodeNotExists(id ID) {
	if _, exists := g.nodes[id]; !exists {
		panic(fmt.Errorf("Node %s does not exist", id))
	}
}

func (g *StateGraph[T, D]) AddNode(name ID, fn NodeFn[T, D]) {
	id := name
	node := &stateNode[T, D]{
		ID: id,
		fn: fn,
	}

	g.add(node)
}

func (g *StateGraph[T, D]) AddWorkerNode(name ID, fn WorkerNodeFn[T, D]) {
	id := name
	node := &workerNode[T, D]{
		ID: id,
		fn: fn,
	}

	g.add(node)
}

func (g *StateGraph[T, D]) add(node node[T, D]) {
	g.panicIfNodeExists(node.id())
	g.nodes[node.id()] = node
}

func (g *StateGraph[T, D]) AddEdge(start ID, end ID) {
	g.panicIfNodeNotExists(start)
	g.panicIfNodeNotExists(end)
	g.edges[start] = stateEdge[T, D]{
		start: start,
		end:   []ID{end},
	}
}

func (g *StateGraph[T, D]) AddConditionalEdge(start ID, router Router[T], ends []ID) {
	g.panicIfNodeNotExists(start)
	for _, end := range ends {
		g.panicIfNodeNotExists(end)
	}

	g.edges[start] = stateEdge[T, D]{
		start:  start,
		end:    ends,
		router: router,
	}
}

func (g *StateGraph[T, D]) AddNamedConditionalEdge(start ID, NamedRouter NamedRouter[T], ends NamedRouterMap) {
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

	g.edges[start] = stateEdge[T, D]{
		start:  start,
		end:    []ID{},
		endMap: ends,
		router: router,
	}
}

func (g *StateGraph[T, D]) FanOut(start ID, ends []ID) {
	g.panicIfNodeNotExists(start)
	for _, end := range ends {
		g.panicIfNodeNotExists(end)
	}
	g.edges[start] = stateEdge[T, D]{
		start: start,
		end:   ends,
	}
}

func (g StateGraph[T, D]) executeAll(ctx context.Context, input stepInput[T, D], channel chan nodeResult[T, D]) {
	var wg sync.WaitGroup
	utils.Each(input.targets, func(target Target) {
		wg.Go(func() {
			cachedResult, found := utils.Find(input.lastResult, func(result nodeResult[T, D]) bool {
				return result.err != nil && result.id == target.id
			})
			if found {
				logger.Info("Using cached result for node", slog.String("id", target.id), slog.Any("state", cachedResult.delta))
				channel <- cachedResult
				return
			}

			logger.Info("Executing node", slog.String("id", target.id))
			state := input.state
			node := g.nodes[target.id]
			delta, err := node.execute(ctx, state, target)

			if err != nil {
				logger.Warn("Error executing node", slog.String("node", target.id), slog.String("err", err.Error()))
				channel <- nodeResult[T, D]{
					id:    target.id,
					delta: delta, // Return the original state in case of error
					err:   err,
				}
				return
			} else {
				logger.Debug("Node executed", slog.String("node", target.id), slog.Any("state", delta))
				channel <- nodeResult[T, D]{
					id:    target.id,
					delta: delta,
				}
			}
		})
	})

	wg.Wait()
	close(channel)
}

func (g StateGraph[T, D]) execute(ctx context.Context, input stepInput[T, D]) []nodeResult[T, D] {
	channel := make(chan nodeResult[T, D], len(input.targets))
	result := []nodeResult[T, D]{}

	go g.executeAll(ctx, input, channel)

	ids := utils.Map(input.targets, func(target Target) ID {
		return target.id
	})
	logger.Debug("Waiting for results from nodes", slog.Any("ids", ids))
	for {
		select {
		case r, ok := <-channel:
			if !ok {
				logger.Debug("Channel closed, all results received", slog.Bool("ok", ok))
				return result
			}
			logger.Debug("Received result from node", slog.String("id", r.id), slog.Any("state", r.delta))
			result = append(result, r)
		case <-ctx.Done():
			// The context error will be handled in the caller, we just return the results received so far
			logger.Debug("Context done, returning results received so far", slog.Any("error", ctx.Err()))
			return result
		}
	}
}

func (g StateGraph[T, D]) reduce(state T, result []nodeResult[T, D]) (T, error) {
	reducedState := state
	for _, value := range result {
		if value.id == END || value.id == START {
			// terminal node, does not contribute to the state
			continue
		}

		// Again, we expect the reducer to handle errors internally
		// and return a valid state, so we don't handle errors here
		newState, reducerErr := executeReducer(g.reducer, reducedState, value.delta)
		if reducerErr != nil {
			return reducedState, reducerErr
		}

		reducedState = newState
	}

	logger.Debug("Reduced state", slog.Any("state", reducedState))

	return reducedState, nil
}

func (g StateGraph[T, D]) route(threadId ID, state T, result []nodeResult[T, D]) ([]Target, *RouterExecutionError) {
	newTargets := []Target{}
	for _, r := range result {
		edge := g.edges[r.id]
		targets, err := edge.route(threadId, state)
		if err != nil {
			return []Target{}, err
		}

		logger.Debug("Routing from node", slog.String("id", r.id), slog.Any("state", r.delta), slog.Any("targets", targets))
		newTargets = append(newTargets, targets...)
	}

	return newTargets, nil
}

func (g StateGraph[T, D]) barrier(ctx context.Context, state T, result []nodeResult[T, D]) (stepInput[T, D], error) {
	threadId := ctx.Value("threadId").(ID)

	// check for errors in the results
	executionError := NewSuperStepExecutionErrorFromResults(result)
	if executionError != nil {
		// Some errors are not interrupt
		return stepInput[T, D]{
			lastResult: result,
		}, executionError
	}

	// reduce result at barrier
	newState, reducerErr := g.reduce(state, result)
	if reducerErr != nil {
		return stepInput[T, D]{
			lastResult: result,
		}, reducerErr
	}

	// route to next nodes
	nodes, routerError := g.route(threadId, newState, result)
	if routerError != nil {
		return stepInput[T, D]{
			lastResult: result,
		}, routerError
	}

	return stepInput[T, D]{
		targets:    nodes,
		state:      newState,
		lastResult: result,
	}, nil
}

func (g StateGraph[T, D]) resolveStartNode(config InvocationConfig[T, D]) []Target {
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

func (g StateGraph[T, D]) resolveRecursionLimit(config InvocationConfig[T, D]) int {
	if config.RecursonLimit <= 0 {
		return 100
	}

	return config.RecursonLimit
}

func (g StateGraph[T, D]) detectInvocationError(ctx context.Context, steps []step[T, D], config InvocationConfig[T, D]) *InvocationError {
	if ctx.Err() != nil {
		return NewInvocationError(fmt.Errorf("Context error: %v", ctx.Err()))
	}

	if len(steps) > g.resolveRecursionLimit(config) {
		return NewInvocationError(fmt.Errorf("Recursion limit exceeded: %d", len(steps)))
	}

	return nil
}

func (g StateGraph[T, D]) Resume(ctx context.Context, threadId string, interruptResults map[string]any, config InvocationConfig[T, D]) (T, error) {
	logger.Info("Resuming graph execution", slog.String("threadId", threadId), slog.Any("interruptResults", interruptResults))
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
	logger.Info("Restored checkpoint", slog.String("threadId", threadId), slog.Any("checkpoint", cp))
	if err != nil {
		return *new(T), NewInvocationError(fmt.Errorf("Failed to restore checkpoint: %v", err))
	}

	invocationConfig := InvocationConfig[T, D]{
		RecursonLimit:  config.RecursonLimit,
		StartNodes:     cp.Steps,
		ThreadId:       threadId,
		Checkpointer:   checkpointer,
		partialResults: cp.Results,
	}
	state := cp.State
	return g.Invoke(ctx, state, invocationConfig)
}

func (g StateGraph[T, D]) resolveThreadId(config InvocationConfig[T, D]) string {
	if config.ThreadId == "" {
		return uuid.New().String()
	}

	return config.ThreadId
}

func (g StateGraph[T, D]) resolveCheckpointer(config InvocationConfig[T, D]) Checkpointer[T, D] {
	if config.Checkpointer == nil {
		return NewInMemoryCheckpointer[T, D]()
	}

	return config.Checkpointer
}

func (g StateGraph[T, D]) resolveNodeResultFromPartial(config InvocationConfig[T, D]) []nodeResult[T, D] {
	if config.partialResults == nil {
		return []nodeResult[T, D]{}
	}

	partialResults := []nodeResult[T, D]{}
	for key, value := range config.partialResults {
		partialResults = append(partialResults, nodeResult[T, D]{
			id:    key,
			delta: value,
		})
	}

	return partialResults
}

func (g StateGraph[T, D]) Invoke(ctx context.Context, initial T, config InvocationConfig[T, D]) (T, error) {
	logger.Info("Invoking graph execution", slog.Any("initialState", initial), slog.Any("config", config))
	threadId := g.resolveThreadId(config)
	checkpointer := g.resolveCheckpointer(config)

	initialStep := step[T, D]{
		input: stepInput[T, D]{
			state:      initial,
			targets:    g.resolveStartNode(config),
			lastResult: g.resolveNodeResultFromPartial(config),
		},
		result: []nodeResult[T, D]{},
	}

	ic := &invokeCtx[T, D]{
		graph:        g,
		config:       config,
		checkpointer: checkpointer,
		threadId:     threadId,
		currentStep:  initialStep,
		steps:        []step[T, D]{initialStep},
	}

	machine := fsm.New[invokeCtx[T, D]]()

	ctx = context.WithValue(ctx, "graph", g)
	ctx = context.WithValue(ctx, "threadId", threadId)
	if fsmErr := machine.Run(ctx, executeState[T, D]{}, ic); fsmErr != nil {
		return ic.currentStep.input.state, fsmErr
	}

	if ic.err != nil {
		return ic.currentStep.input.state, ic.err
	}

	err := checkpointer.Checkpoint(threadId, ic.currentStep.toCheckpoint())
	if err != nil {
		logger.Error("Checkpoint error", slog.Any("error", err))
		return ic.currentStep.input.state, err
	}

	return ic.currentStep.input.state, nil
}

// Generate a mermaid diagram of the state graph.
func (g StateGraph[T, D]) Visualize() []byte {
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

func (g *StateGraph[T, D]) Compile() {
}

func New[T any, D any](reducer Reducer[T, D]) StateGraph[T, D] {
	graph := StateGraph[T, D]{
		nodes:      make(map[ID]node[T, D]),
		edges:      make(map[ID]stateEdge[T, D]),
		reducer:    reducer,
		visualizer: visualizer.NewMermaidVisualizer(),
		interrupts: make(map[string]InterruptResult),
	}
	graph.add(newStartNode[T, D]())
	graph.add(newEndNode[T, D]())
	return graph
}
