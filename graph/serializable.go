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

	serializedResults := make(map[ID][]byte)
	for id, result := range cp.Results {
		resultData, err := result.Serialize()
		if err != nil {
			return err
		}
		serializedResults[id] = resultData
	}
	jsonResults, err := json.Marshal(cp.Results)

	err = c.persister.Save(threadId, "results", jsonResults)
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

func (c *SerializableCheckpointer[T]) restoreResults(threadId string) (map[ID]T, error) {
	var results map[ID]T = make(map[ID]T)
	data, err := c.persister.Load(threadId, "results")

	if err != nil {
		return results, err
	}

	jsonResults := make(map[ID][]byte)
	err = json.Unmarshal(data, &jsonResults)
	if err != nil {
		return results, err
	}

	for id, resultData := range jsonResults {
		var result T = *new(T)
		err := result.Deserialize(resultData)
		if err != nil {
			return results, err
		}
		results[id] = result
	}

	return results, nil
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

	results, err := c.restoreResults(threadId)
	if err != nil {
		return cp, err
	}
	cp.Results = results

	return cp, nil
}
