package graph

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/nquangtrung/agentgo/utils"
)

type retryNode[T any, D any] struct {
	maxAttempts     int
	initialInterval time.Duration
	backoffFactor   float64
	maxInterval     time.Duration
	jitter          bool
	shouldRetry     func(err error) bool
	node            node[T, D]
}

func applyJitter(interval time.Duration, jitter bool) time.Duration {
	if !jitter {
		return interval
	}

	jitterFactor := 0.1 // 10% jitter
	jitterAmount := time.Duration(float64(interval) * jitterFactor)
	minInterval := interval - jitterAmount
	maxInterval := interval + jitterAmount

	return time.Duration(minInterval + time.Duration(rand.Int63n(int64(maxInterval-minInterval))))
}

func (n retryNode[T, D]) execute(ctx context.Context, state T, target Target) (delta D, err error) {
	delta, err = n.node.execute(ctx, state, target)
	if err == nil {
		return delta, nil
	}

	attempts := 1
	interval := n.initialInterval

	for attempts < n.maxAttempts {
		if e, ok := err.(*NodeExecutionError); ok {
			err = e.Unwrap()
		}
		if !n.shouldRetry(err) {
			logger.Warn("Error is not retriable", slog.Any("error", err))
			break
		}

		logger.Info("Retrying node", slog.Any("node", n.id()), slog.Int("attempts", attempts), slog.Int("wait", int(interval)))

		select {
		case <-time.After(interval):
			delta, err = n.node.execute(ctx, state, target)
			if err == nil {
				return delta, nil
			}
			attempts++
			interval = time.Duration(float64(interval) * n.backoffFactor)
			interval = min(interval, n.maxInterval)
			interval = applyJitter(interval, n.jitter)
		case <-ctx.Done():
			return delta, ctx.Err()
		}
	}

	return delta, err
}

func (n retryNode[T, D]) id() ID {
	return n.node.id()
}

type RetryOptions struct {
	MaxAttempts     int
	InitialInterval time.Duration
	BackoffFactor   float64
	MaxInterval     time.Duration
	Jitter          bool
	ShouldRetry     func(err error) bool
}

func withRetry[T any, D any](node node[T, D], options RetryOptions) node[T, D] {
	if options.ShouldRetry == nil {
		panic("shouldRetry has to be provided")
	}

	return retryNode[T, D]{
		maxAttempts:     utils.Ternary(options.MaxAttempts > 0, options.MaxAttempts, 3),
		initialInterval: utils.Ternary(options.InitialInterval > 0, options.InitialInterval, 100*time.Millisecond),
		backoffFactor:   utils.Ternary(options.BackoffFactor > 0, options.BackoffFactor, 2.0),
		maxInterval:     utils.Ternary(options.MaxInterval > 0, options.MaxInterval, 10*time.Second),
		jitter:          options.Jitter,
		shouldRetry:     options.ShouldRetry,
		node:            node,
	}
}
