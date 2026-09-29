package agentgo

import (
	"sync"

	"github.com/nquangtrung/agentgo/graph"
)

// CheckpointData is the opaque snapshot passed to a custom Checkpointer.
// It holds everything needed to resume an interrupted execution.
// Do not inspect or modify the internal field.
type CheckpointData struct {
	ThreadID string
	// internal holds the graph-level checkpoint state.
	// It is unexported to prevent external tampering.
	internal graph.Checkpoint[agentState, agentStateDelta]
}

// Checkpointer is the public interface for persisting and restoring
// agentgo execution state. Implement this to use a custom store
// (database, Redis, S3, etc.) for human-in-the-loop or long-running executions.
//
// When a ToolMiddleware calls graph.Interrupt, the execution is paused and
// saved via Save(). Call agentgo.ResumeGenerateText or agentgo.ResumeStreamText
// with the same Checkpointer to resume from that point.
type Checkpointer interface {
	// Save persists the checkpoint. ThreadID is guaranteed to be non-empty.
	Save(data CheckpointData) error

	// Load retrieves a previously saved checkpoint.
	// Return error if the checkpoint does not exist (errors.Is check against
	// graph.CheckpointNotFoundError is recommended).
	Load(threadID string) (CheckpointData, error)
}

// graphStorer is an optional interface that Checkpointer implementations
// can implement to support storing and retrieving graph instances for HITL scenarios.
// This is an internal interface; users should not implement it directly.
type graphStorer interface {
	storeGraphInstance(threadID string, g graph.StateGraph[agentState, agentStateDelta])
	retrieveGraphInstance(threadID string) (graph.StateGraph[agentState, agentStateDelta], bool)
}

// agentCheckpointerAdapter bridges agentgo.Checkpointer → graph.Checkpointer[agentState, agentStateDelta].
type agentCheckpointerAdapter struct {
	outer Checkpointer
}

func (a *agentCheckpointerAdapter) Checkpoint(threadId string, cp graph.Checkpoint[agentState, agentStateDelta]) error {
	return a.outer.Save(CheckpointData{ThreadID: threadId, internal: cp})
}

func (a *agentCheckpointerAdapter) Restore(threadId string) (graph.Checkpoint[agentState, agentStateDelta], error) {
	data, err := a.outer.Load(threadId)
	if err != nil {
		return graph.Checkpoint[agentState, agentStateDelta]{}, err
	}
	return data.internal, nil
}

// NewInMemoryCheckpointer returns the default in-memory Checkpointer.
// This is used automatically when no Checkpointer is set in Params.
// All checkpoints are lost when the Checkpointer is garbage collected.
func NewInMemoryCheckpointer() Checkpointer {
	return &inMemoryCheckpointer{
		store:  make(map[string]CheckpointData),
		graphs: make(map[string]graph.StateGraph[agentState, agentStateDelta]),
		mu:     sync.RWMutex{},
	}
}

type inMemoryCheckpointer struct {
	mu     sync.RWMutex
	store  map[string]CheckpointData
	graphs map[string]graph.StateGraph[agentState, agentStateDelta]
}

func (c *inMemoryCheckpointer) Save(data CheckpointData) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[data.ThreadID] = data
	return nil
}

func (c *inMemoryCheckpointer) Load(threadID string) (CheckpointData, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.store[threadID]
	if !ok {
		return CheckpointData{}, graph.NewCheckpointNotFoundError(threadID)
	}
	return data, nil
}

func (c *inMemoryCheckpointer) storeGraphInstance(threadID string, g graph.StateGraph[agentState, agentStateDelta]) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.graphs[threadID] = g
}

func (c *inMemoryCheckpointer) retrieveGraphInstance(threadID string) (graph.StateGraph[agentState, agentStateDelta], bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	g, ok := c.graphs[threadID]
	return g, ok
}

// resolveGraphCheckpointer returns the graph-level checkpointer to pass into InvocationConfig.
// If the user supplies a Checkpointer, it is wrapped; otherwise an InMemoryCheckpointer is used.
func resolveGraphCheckpointer(c Checkpointer) graph.Checkpointer[agentState, agentStateDelta] {
	if c == nil {
		c = NewInMemoryCheckpointer()
	}
	return &agentCheckpointerAdapter{outer: c}
}

// storeGraphForResume stores a graph instance for later retrieval during Resume.
// This is only needed internally; exported for use by GenerateText and StreamText.
func storeGraphForResume(checkpointer Checkpointer, threadID string, g graph.StateGraph[agentState, agentStateDelta]) {
	if gs, ok := checkpointer.(graphStorer); ok {
		gs.storeGraphInstance(threadID, g)
	}
}

// retrieveGraphForResume retrieves a stored graph instance for Resume.
// This is only needed internally; exported for use by ResumeGenerateText and ResumeStreamText.
func retrieveGraphForResume(checkpointer Checkpointer, threadID string) (graph.StateGraph[agentState, agentStateDelta], bool) {
	if gs, ok := checkpointer.(graphStorer); ok {
		return gs.retrieveGraphInstance(threadID)
	}
	return graph.StateGraph[agentState, agentStateDelta]{}, false
}
