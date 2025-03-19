// Copyright 2025 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package queue implements a generic queue data structure.
package queue

import "sync"

// Queue is a generic queue of any.
// Items are Pushed to end of list and popped from the front.
//
// Example:
//
//	q := New[int]()
//	q.Push(1)
//	q.Push(2)
//	v, ok := q.Pop() // v == 1, ok == true
type Queue[V any] struct {
	items []V
}

// New returns a new queue of V.
//
// Example:
//
//	q := New[int]()
func New[V any]() *Queue[V] { return &Queue[V]{} }

// Push pushes v to end of queue.
//
// Example:
//
//	q := New[int]()
//	q.Push(1)
//	q.Push(2)
func (self *Queue[V]) Push(v V) { self.items = append(self.items, v) }

// Pop returns item from the start of the queue and truth if one was found.
// Returned value should be ignored if truth if false, it is the zero value of
// Queue generic type.
//
// Example:
//
//	q := New[int]()
//	q.Push(1)
//	v, ok := q.Pop() // v == 1, ok == true
//	v, ok = q.Pop()  // v == 0, ok == false
func (self *Queue[V]) Pop() (v V, b bool) {
	if b = len(self.items) > 0; !b {
		return
	}
	v = self.items[0]
	self.items = self.items[1:]
	return
}

// SyncQueue is concurrency safe Queue.
//
// Example:
//
//	q := &SyncQueue[int]{q: New[int]()}
//	go q.Push(1)
//	v, ok := q.Pop()
type SyncQueue[V any] struct {
	mu sync.Mutex
	q  *Queue[V]
}

// Push pushes v to end of queue.
//
// Example:
//
//	q := &SyncQueue[int]{q: New[int]()}
//	q.Push(1)
func (self *SyncQueue[V]) Push(v V) {
	self.mu.Lock()
	self.q.Push(v)
	self.mu.Unlock()
}

// Pop returns item from the start of the queue and truth if one was found.
// Returned value should be ignored if truth if false, it is the zero value of
// Queue generic type.
//
// Example:
//
//	q := &SyncQueue[int]{q: New[int]()}
//	q.Push(1)
//	v, ok := q.Pop() // v == 1, ok == true
func (self *SyncQueue[V]) Pop() (v V, b bool) {
	self.mu.Lock()
	v, b = self.q.Pop()
	self.mu.Unlock()
	return
}
