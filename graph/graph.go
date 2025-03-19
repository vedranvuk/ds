// Copyright 2025 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package graph implements a generic graph data structure.
package graph

import "sync"

// Graph is a many-to many map of generic comparable entries.
// The graph structure stores links between keys of a comparable type K.
type Graph[K comparable] struct {
	links map[K]map[K]struct{}
}

// NewGraph returns a new [Graph].
//
// Example:
//
//	g := NewGraph[int]()
func NewGraph[K comparable]() *Graph[K] {
	return &Graph[K]{
		links: make(map[K]map[K]struct{}),
	}
}

// Link links a and b if not already linked and returns if a and b were already
// linked prior to this call.
//
// Arguments:
//
//	a: The first key to link.
//	b: The second key to link.
//
// Returns:
//
//	wasLinked: True if a and b were already linked prior to this call, false otherwise.
//
// Example:
//
//	g := NewGraph[string]()
//	wasLinked := g.Link("a", "b") // wasLinked will be false
//	wasLinked = g.Link("a", "b")    // wasLinked will be true
func (self *Graph[K]) Link(a, b K) (wasLinked bool) {
	wasLinked = self.Linked(a, b)
	if _, exists := self.links[a]; !exists {
		self.links[a] = make(map[K]struct{})
	}
	self.links[a][b] = struct{}{}
	return
}

// Unlink breaks the link between a and b if it exists and returns if the link
// existed prior to this call.
//
// Arguments:
//
//	a: The first key to unlink.
//	b: The second key to unlink.
//
// Returns:
//
//	wasLinked: True if the link between a and b existed prior to this call, false otherwise.
//
// Example:
//
//	g := NewGraph[string]()
//	g.Link("a", "b")
//	wasLinked := g.Unlink("a", "b") // wasLinked will be true
//	wasLinked = g.Unlink("a", "b")    // wasLinked will be false
func (self *Graph[K]) Unlink(a, b K) (wasLinked bool) {
	wasLinked = self.Linked(a, b)
	if m, exists := self.links[a]; exists {
		delete(m, b)
		self.links[a] = m
	}
	return
}

// Linked returns if a and b are linked.
//
// Arguments:
//
//	a: The first key.
//	b: The second key.
//
// Returns:
//
//	linked: True if a and b are linked, false otherwise.
//
// Example:
//
//	g := NewGraph[string]()
//	g.Link("a", "b")
//	linked := g.Linked("a", "b") // linked will be true
//	linked = g.Linked("b", "a")    // linked will be false
func (self *Graph[K]) Linked(a, b K) (linked bool) {
	if links, exists := self.links[a]; exists {
		_, linked = links[b]
	}
	return
}

// Links returns names of all keys a key links to.
//
// Arguments:
//
//	key: The key to get the links for.
//
// Returns:
//
//	out: A slice containing all keys that 'key' links to.
//
// Example:
//
//	g := NewGraph[string]()
//	g.Link("a", "b")
//	g.Link("a", "c")
//	links := g.Links("a") // links will be []string{"b", "c"} or []string{"c", "b"} (order not guaranteed)
func (self *Graph[K]) Links(key K) (out []K) {
	if out = make([]K, 0, len(self.links[key])); cap(out) == 0 {
		return
	}
	self.EnumLinks(key, func(k K) bool {
		out = append(out, k)
		return true
	})
	return
}

// EnumLinks calls f for each link a key has. It keeps calling f until all links
// have been enumerated or f returns false.
//
// Arguments:
//
//	key: The key to enumerate links for.
//	f:   The function to call for each link.  The function receives the linked key as argument.
//	     If the function returns false, the enumeration is stopped.
//
// Example:
//
//	g := NewGraph[string]()
//	g.Link("a", "b")
//	g.Link("a", "c")
//	g.EnumLinks("a", func(k string) bool {
//		fmt.Println(k) // Will print "b" and "c" (order not guaranteed)
//		return true
//	})
func (self *Graph[K]) EnumLinks(key K, f func(key K) bool) {
	if m, exists := self.links[key]; exists {
		for key := range m {
			if !f(key) {
				break
			}
		}
	}
}

// SyncGraph is the concurrency safe version of [Graph].
// It uses a mutex to protect the underlying Graph from concurrent access.
type SyncGraph[K comparable] struct {
	mu    sync.Mutex
	graph *Graph[K]
}

// NewSyncGraph returns a new concurrency safe [Graph].
//
// Example:
//
//	g := NewSyncGraph[int]()
func NewSyncGraph[K comparable]() *SyncGraph[K] {
	return &SyncGraph[K]{
		graph: NewGraph[K](),
	}
}

// Link links a and b if not already linked and returns if a and b were already
// linked prior to this call.  This method is concurrency safe.
//
// Arguments:
//
//	a: The first key to link.
//	b: The second key to link.
//
// Returns:
//
//	wasLinked: True if a and b were already linked prior to this call, false otherwise.
//
// Example:
//
//	g := NewSyncGraph[string]()
//	wasLinked := g.Link("a", "b") // wasLinked will be false
//	wasLinked = g.Link("a", "b")    // wasLinked will be true
func (self *SyncGraph[K]) Link(a, b K) (wasLinked bool) {
	self.mu.Lock()
	wasLinked = self.graph.Link(a, b)
	self.mu.Unlock()
	return
}

// Unlink breaks the link between a and b if it exists and returns if the link
// existed prior to this call. This method is concurrency safe.
//
// Arguments:
//
//	a: The first key to unlink.
//	b: The second key to unlink.
//
// Returns:
//
//	wasLinked: True if the link between a and b existed prior to this call, false otherwise.
//
// Example:
//
//	g := NewSyncGraph[string]()
//	g.Link("a", "b")
//	wasLinked := g.Unlink("a", "b") // wasLinked will be true
//	wasLinked = g.Unlink("a", "b")    // wasLinked will be false
func (self *SyncGraph[K]) Unlink(a, b K) (wasLinked bool) {
	self.mu.Lock()
	wasLinked = self.graph.Unlink(a, b)
	self.mu.Unlock()
	return
}

// Linked returns if a and b are linked. This method is concurrency safe.
//
// Arguments:
//
//	a: The first key.
//	b: The second key.
//
// Returns:
//
//	linked: True if a and b are linked, false otherwise.
//
// Example:
//
//	g := NewSyncGraph[string]()
//	g.Link("a", "b")
//	linked := g.Linked("a", "b") // linked will be true
//	linked = g.Linked("b", "a")    // linked will be false
func (self *SyncGraph[K]) Linked(a, b K) (linked bool) {
	self.mu.Lock()
	linked = self.graph.Linked(a, b)
	self.mu.Unlock()
	return
}

// Links returns names of all keys a key links to. This method is concurrency safe.
//
// Arguments:
//
//	key: The key to get the links for.
//
// Returns:
//
//	out: A slice containing all keys that 'key' links to.
//
// Example:
//
//	g := NewSyncGraph[string]()
//	g.Link("a", "b")
//	g.Link("a", "c")
//	links := g.Links("a") // links will be []string{"b", "c"} or []string{"c", "b"} (order not guaranteed)
func (self *SyncGraph[K]) Links(key K) (out []K) {
	self.mu.Lock()
	out = self.graph.Links(key)
	self.mu.Unlock()
	return
}

// EnumLinks calls f for each link a key has. It keeps calling f until all links
// have been enumerated or f returns false. This method is concurrency safe.
//
// Arguments:
//
//	key: The key to enumerate links for.
//	f:   The function to call for each link. The function receives the linked key as argument.
//	     If the function returns false, the enumeration is stopped.
//
// Example:
//
//	g := NewSyncGraph[string]()
//	g.Link("a", "b")
//	g.Link("a", "c")
//	g.EnumLinks("a", func(k string) bool {
//		fmt.Println(k) // Will print "b" and "c" (order not guaranteed)
//		return true
//	})
func (self *SyncGraph[K]) EnumLinks(key K, f func(key K) bool) {
	self.mu.Lock()
	self.graph.EnumLinks(key, f)
	self.mu.Unlock()
	return
}
