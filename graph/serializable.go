package graph

import "encoding/json"

type Serializable interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}

type Persister[T Serializable] interface {
	Save(threadId string, key string, data []byte) error
	Load(threadId string, key string) ([]byte, error)
}

type SerializableCheckpointer[T Serializable] struct {
	persister Persister[T]
}

func NewSerializableCheckpointer[T Serializable](persister Persister[T]) *SerializableCheckpointer[T] {
	return &SerializableCheckpointer[T]{persister: persister}
}

func (c *SerializableCheckpointer[T]) Checkpoint(threadId string, cp Checkpoint[T]) error {
	state, err := cp.State.Serialize()
	if err != nil {
		return err
	}

	err = c.persister.Save(threadId, "state", state)
	if err != nil {
		return err
	}

	jsonSteps, err := json.Marshal(cp.Steps)
	if err != nil {
		return err
	}

	err = c.persister.Save(threadId, "steps", jsonSteps)
	if err != nil {
		return err
	}

	return nil
}
func (c *SerializableCheckpointer[T]) restoreState(threadId string) (T, error) {
	var state T = *new(T)
	data, err := c.persister.Load(threadId, "state")
	if err != nil {
		return state, err
	}

	err = state.Deserialize(data)
	if err != nil {
		return state, err
	}

	return state, nil
}

func (c *SerializableCheckpointer[T]) restoreSteps(threadId string) ([]ID, error) {
	var steps []ID
	data, err := c.persister.Load(threadId, "steps")
	if err != nil {
		return steps, err
	}

	err = json.Unmarshal(data, &steps)
	if err != nil {
		return steps, err
	}

	return steps, nil
}

func (c *SerializableCheckpointer[T]) Restore(threadId string) (Checkpoint[T], error) {
	var cp Checkpoint[T]
	state, err := c.restoreState(threadId)
	if err != nil {
		return cp, err
	}
	cp.State = state

	steps, err := c.restoreSteps(threadId)
	if err != nil {
		return cp, err
	}
	cp.Steps = steps

	return cp, nil
}
