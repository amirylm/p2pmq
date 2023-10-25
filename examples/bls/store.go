package blstest

import (
	"sync"
)

type store[T any] struct {
	lock  sync.RWMutex
	items map[string]T
}

func NewStore[T any]() *store[T] {
	return &store[T]{
		items: make(map[string]T),
	}
}

func (s *store[T]) Add(network string, item T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.items[network] = item
}

func (s *store[T]) Get(network string) (T, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	item, ok := s.items[network]
	return item, ok
}
