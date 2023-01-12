package util

import "sync"

func NewStack[T any](size, cap int) Stack[T] {
	return Stack[T]{
		data: make([]T, size, cap),
	}
}

type Stack[T any] struct {
	data []T
	len  int
	mu   sync.Mutex
}

func (s *Stack[T]) Push(v ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append(s.data, v...)
	s.len += len(v)
}

func (s *Stack[T]) Pop() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var v T
	if s.Len() <= 0 {
		return v, false
	}
	s.len -= 1
	v, s.data = s.data[s.len], s.data[:s.len]
	return v, true
}

func (s *Stack[T]) Len() int {
	return s.len
}
