package core

import (
	"context"
	"fmt"
)

type Router[T any] struct {
	q chan T

	workers chan struct{}
	worker  func(T)
}

func NewRouter[T any](qSize, workers int, worker func(T)) *Router[T] {
	r := &Router[T]{
		q:       make(chan T, qSize),
		workers: make(chan struct{}, workers),
		worker:  worker,
	}

	return r
}

func (r *Router[T]) Start(pctx context.Context) {
	ctx, cancel := context.WithCancel(pctx)
	defer cancel()

	for {
		select {
		case msg := <-r.q:
			r.workers <- struct{}{}
			go func(m T) {
				defer func() {
					<-r.workers
				}()
				r.worker(m)
			}(msg)
		case <-ctx.Done():
			return
		}
	}
}

func (r *Router[T]) Handle(msg T) error {
	select {
	case r.q <- msg:
	default:
		return fmt.Errorf("router queue is full, dropping msg")
	}
	return nil
}
