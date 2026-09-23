package graph

import "encoding/json"

type Serializable interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}

type Persister interface {
	Save(threadId string, key string, data []byte) error
	Load(threadId string, key string) ([]byte, error)
}

type SerializableCheckpointer[T Serializable, D Serializable] struct {
	persister Persister
}

func NewSerializableCheckpointer[T Serializable, D Serializable](persister Persister) *SerializableCheckpointer[T, D] {
	return &SerializableCheckpointer[T, D]{persister: persister}
}

func (c *SerializableCheckpointer[T, D]) Checkpoint(threadId string, cp Checkpoint[T, D]) error {
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
func (c *SerializableCheckpointer[T, D]) restoreState(threadId string) (T, error) {
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

func (c *SerializableCheckpointer[T, D]) restoreSteps(threadId string) ([]ID, error) {
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

func (c *SerializableCheckpointer[T, D]) restoreResults(threadId string) (map[ID]D, error) {
	var results map[ID]D = make(map[ID]D)
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
		var result D = *new(D)
		err := result.Deserialize(resultData)
		if err != nil {
			return results, err
		}
		results[id] = result
	}

	return results, nil
}

func (c *SerializableCheckpointer[T, D]) Restore(threadId string) (Checkpoint[T, D], error) {
	var cp Checkpoint[T, D]
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
