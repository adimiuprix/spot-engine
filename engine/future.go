package engine

import (
	"context"
	"fmt"
)

// Future represents a placeholder for an asynchronous operation result
type Future[T any] struct {
	resultChan chan result[T]
	done       bool
	value      T
	err        error
}

type result[T any] struct {
	value T
	err   error
}

// NewFuture creates a new Future
func NewFuture[T any]() *Future[T] {
	return &Future[T]{
		resultChan: make(chan result[T], 1),
	}
}

// Complete completes the future with a value
func (f *Future[T]) Complete(value T) {
	if f.done {
		return
	}
	f.resultChan <- result[T]{value: value}
	f.done = true
	close(f.resultChan)
}

// Fail completes the future with an error
func (f *Future[T]) Fail(err error) {
	if f.done {
		return
	}
	f.resultChan <- result[T]{err: err}
	f.done = true
	close(f.resultChan)
}

// Wait blocks until the future is completed or context is cancelled
// Returns the value and any error
func (f *Future[T]) Wait(ctx context.Context) (T, error) {
	if f.done {
		return f.value, f.err
	}

	select {
	case res := <-f.resultChan:
		f.value = res.value
		f.err = res.err
		return res.value, res.err
	case <-ctx.Done():
		var zero T
		return zero, fmt.Errorf("future wait cancelled: %w", ctx.Err())
	}
}

// IsDone returns true if the future has been completed
func (f *Future[T]) IsDone() bool {
	return f.done
}
