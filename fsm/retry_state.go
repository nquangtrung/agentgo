package fsm

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"
)

type RetryState struct {
	OriginalState State[AgentContext]
}

func (s *RetryState) Execute(ctx context.Context, fsmCtx *AgentContext) (State[AgentContext], error) {
	if fsmCtx.RetryCount >= 3 {
		return &ErrorState{
			Err: errors.New("Max retry count reached"),
		}, nil
	}
	timeout := time.After(time.Second * 5)
	if fsmCtx.LastRetryWait > 0 {
		timeout = time.After(fsmCtx.LastRetryWait*2 + time.Duration(rand.IntN(10000))*time.Millisecond)
	}

	<-timeout
	return s.OriginalState, nil
}
