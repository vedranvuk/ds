package queue

import "sync"

// Queue is a generic queue of any.
// Items are Pushed to end of list and popped from the front.
type Queue[V any] struct {
	items []V
}

// New returns a new queue of V.
func New[V any]() *Queue[V] { return &Queue[V]{} }

// Push pushes v to end of queue.
func (self *Queue[V]) Push(v V) { self.items = append(self.items, v) }

// Pop returns item from the start of the queue and truth if one was found.
// Returned value should be ignored if truth if false, it is the zero value of
// Queue generic type.
func (self *Queue[V]) Pop() (v V, b bool) {
	if b = len(self.items) > 0; !b {
		return
	}
	v = self.items[0]
	self.items = self.items[1:]
	return
}

// SyncQueue is concurrency safe Queue.
type SyncQueue[V any] struct {
	mu sync.Mutex
	q  *Queue[V]
}

func (self *SyncQueue[V]) Push(v V) {
	self.mu.Lock()
	self.q.Push(v)
	self.mu.Unlock()
}

func (self *SyncQueue[V]) Pop() (v V, b bool) {
	self.mu.Lock()
	v, b = self.q.Pop()
	self.mu.Unlock()
	return
}
