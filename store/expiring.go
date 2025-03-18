package store

import (
	"sync"
	"time"
)

type TimeExpirable interface {
	IsExpired(now time.Time) bool
}

type ExpirableStore[T TimeExpirable] struct {
	mu    sync.Mutex
	items map[string]T
	done  chan struct{}
}

func NewExpiringStore[T TimeExpirable](capacity ...int) *ExpirableStore[T] {
	var finalCap int
	if len(capacity) > 0 {
		finalCap = capacity[0]
	} else {
		finalCap = 64
	}
	store := &ExpirableStore[T]{
		items: make(map[string]T, finalCap),
		done:  make(chan struct{}),
	}
	go store.cleanup()
	return store
}

func (s *ExpirableStore[T]) Set(key string, item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = item
}

func (s *ExpirableStore[T]) Get(key string) (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[key]
	return item, ok
}

func (s *ExpirableStore[T]) Pop(key string) (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[key]
	delete(s.items, key)
	return item, ok
}

func (s *ExpirableStore[T]) PopValidate(key string) (T, bool) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[key]
	if !ok {
		return item, false
	}
	delete(s.items, key)
	return item, !item.IsExpired(now)
}

func (s *ExpirableStore[T]) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
}

func (s *ExpirableStore[T]) Purge() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, item := range s.items {
		if item.IsExpired(now) {
			delete(s.items, key)
		}
	}
}

func (s *ExpirableStore[T]) Close() error {
	close(s.done)
	return nil
}

func (s *ExpirableStore[T]) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.Purge()
		}
	}
}
