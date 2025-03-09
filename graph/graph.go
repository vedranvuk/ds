package graph

import "sync"

// Graph is a many-to many map of generic comparable entries.
type Graph[K comparable] struct {
	links map[K]map[K]struct{}
}

// New returns a new [Graph].
func NewGraph[K comparable]() *Graph[K] {
	return &Graph[K]{
		links: make(map[K]map[K]struct{}),
	}
}

// Link links a and b if not already linked and returns if a and b were already
// linked prior to this call.
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
func (self *Graph[K]) Unlink(a, b K) (wasLinked bool) {
	wasLinked = self.Linked(a, b)
	if m, exists := self.links[a]; exists {
		delete(m, b)
		self.links[a] = m
	}
	return
}

// Link returns if a and b are linked.
func (self *Graph[K]) Linked(a, b K) (linked bool) {
	if links, exists := self.links[a]; exists {
		_, linked = links[b]
	}
	return
}

// Links returns names of all keys a key links to.
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
type SyncGraph[K comparable] struct {
	mu    sync.Mutex
	graph *Graph[K]
}

// New returns a new [Graph].
func NewSyncGraph[K comparable]() *SyncGraph[K] {
	return &SyncGraph[K]{
		graph: NewGraph[K](),
	}
}

// Link links a and b if not already linked and returns if a and b were already
// linked prior to this call.
func (self *SyncGraph[K]) Link(a, b K) (wasLinked bool) {
	self.mu.Lock()
	wasLinked = self.graph.Link(a, b)
	self.mu.Unlock()
	return
}

// Unlink breaks the link between a and b if it exists and returns if the link
// existed prior to this call.
func (self *SyncGraph[K]) Unlink(a, b K) (wasLinked bool) {
	self.mu.Lock()
	wasLinked = self.graph.Unlink(a, b)
	self.mu.Unlock()
	return
}

// Link returns if a and b are linked.
func (self *SyncGraph[K]) Linked(a, b K) (linked bool) {
	self.mu.Lock()
	linked = self.graph.Linked(a, b)
	self.mu.Unlock()
	return
}

// Links returns names of all keys a key links to.
func (self *SyncGraph[K]) Links(key K) (out []K) {
	self.mu.Lock()
	out = self.graph.Links(key)
	self.mu.Unlock()
	return
}

// EnumLinks calls f for each link a key has. It keeps calling f until all links
// have been enumerated or f returns false.
func (self *SyncGraph[K]) EnumLinks(key K, f func(key K) bool) {
	self.mu.Lock()
	self.graph.EnumLinks(key, f)
	self.mu.Unlock()
	return
}
