// Package trie implements a string prefix trie.
package trie

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
)

// Trie implements a prefix tree of generic values keyed by a string key.
//
// It has fast lookups and can retrieve keys which are a prefix of some key.
//
// Implemented as a tree of [Node] where each node stores branches in a slice
// where each indice starts with a unique rune and is sorted alphabetically.
// A binary search is used on branch lookups, rest of key is scanned
// sequentially.
//
// Puts require 2 allocations when allocating a new Node in the structure.
// Gets are fast, require no allocations.
type Trie[V any] struct {
	root *Node[V]
	zero V
}

// New returns a new [Trie].
func New[V any]() *Trie[V] {
	return &Trie[V]{
		root: new(Node[V]),
		zero: *new(V),
	}
}

// Put inserts value under key.
//
// If a value already exists at key it is returned with true, otherwise a zero
// value of V is returned and false.
//
// Key must not be empty. If it is no value is inserted and a zero value of V
// and false is returned.
func (self *Trie[V]) Put(key string, value V) (old V, replaced bool) {

	if key == "" {
		return self.zero, false
	}

	var keyRunes = []rune(key)
	var idx, found = self.root.Branches.find(keyRunes[0])

	// fast path, no branch for key starting rune.
	if !found {
		var newNode = &Node[V]{
			Prefix:   keyRunes,
			Value:    value,
			HasValue: true,
		}
		self.root.Branches = slices.Insert(self.root.Branches, idx, newNode)
		return self.zero, false
	}

	var currentNode = self.root.Branches[idx]
	var nodeRunes = []rune(currentNode.Prefix)

restart:
	var i = 0
	for {
		// end of query reached
		if i == len(keyRunes) {

			// Node prefix fully matched.
			if i == len(nodeRunes) {
				old, replaced = currentNode.Value, currentNode.HasValue
				currentNode.Value, currentNode.HasValue = value, true
				return
			}

			// node prefix partially matched, split node.
			var newNode = &Node[V]{
				Prefix:   nodeRunes[i:],
				Value:    currentNode.Value,
				HasValue: currentNode.HasValue,
				Branches: currentNode.Branches,
			}
			currentNode.Prefix = keyRunes
			currentNode.Value = value
			currentNode.HasValue = true
			currentNode.Branches = Branches[V]{newNode}

			return self.zero, false
		}

		// end of node prefix reached
		if i == len(nodeRunes) {
			idx, found = currentNode.Branches.find(keyRunes[i])

			// No branches found, insert new.
			if !found {
				var newNode = &Node[V]{
					Prefix:   keyRunes[i:],
					Value:    value,
					HasValue: true,
				}
				currentNode.Branches = slices.Insert(currentNode.Branches, idx, newNode)
				return self.zero, false
			}

			// Set current node to node under matched branch and restart.
			currentNode = currentNode.Branches[idx]
			keyRunes = keyRunes[i:]
			nodeRunes = []rune(currentNode.Prefix)
			goto restart
		}

		// missmatch at the middle of node prefix, split node.
		if keyRunes[i] != nodeRunes[i] {

			var newNode = &Node[V]{
				Prefix:   keyRunes[i:],
				Value:    value,
				HasValue: true,
			}

			var splitNode = &Node[V]{
				Prefix:   nodeRunes[i:],
				Value:    currentNode.Value,
				HasValue: currentNode.HasValue,
				Branches: currentNode.Branches,
			}

			currentNode.Prefix = keyRunes[:i]
			currentNode.Value = self.zero
			currentNode.HasValue = false
			if splitNode.Prefix[0] > newNode.Prefix[0] {
				currentNode.Branches = Branches[V]{newNode, splitNode}
			} else {
				currentNode.Branches = Branches[V]{splitNode, newNode}
			}

			return self.zero, false
		}

		i++
	}
}

// Get returns the value at key and true if it exists or a zero value of V and
// false if not found.
//
// Key must not be empty. If it is Get returns a zero value and false.
func (self *Trie[V]) Get(key string) (value V, found bool) {

	if key == "" {
		return self.zero, false
	}

	var qry = []rune(key)
	var idx int
	if idx, found = self.root.Branches.find(qry[0]); !found {
		return self.zero, false
	}
	var node = self.root.Branches[idx]
	var npfx = []rune(node.Prefix)

restart:
	var i = 0
	for {
		if i == len(qry) {
			if node.HasValue {
				return node.Value, true
			}
			return self.zero, false
		}

		if i == len(npfx) {
			idx, found = node.Branches.find(qry[i])

			if !found {
				return self.zero, false
			}

			node = node.Branches[idx]
			qry = qry[i:]
			npfx = []rune(node.Prefix)
			goto restart
		}

		if qry[i] != npfx[i] {
			return self.zero, false
		}

		i++
	}
}

// Exists returns true if key exists.
func (self *Trie[V]) Exists(key string) (exists bool) {
	_, exists = self.Get(key)
	return
}

// Prefixes returns a list of set keys which are a prefix of key.
func (self *Trie[V]) Prefixes(key string) (out []string) {

	var scanned []rune

	var qry = []rune(key)
	var idx, found = self.root.Branches.find(qry[0])
	if !found {
		return
	}
	var node = self.root.Branches[idx]
	var npfx = []rune(node.Prefix)

restart:
	var i = 0
	for {
		if i == len(qry) {
			/*
				// Returns the query too.
				//
				if node.HasValue {
					out = append(out, string(append(scanned, node.Prefix...)))
					return
				}
			*/
			return
		}

		if i == len(npfx) {
			scanned = append(scanned, node.Prefix...)
			if node.HasValue {
				out = append(out, string(scanned))
			}

			idx, found = node.Branches.find(qry[i])

			if !found {
				return
			}

			node = node.Branches[idx]
			qry = qry[i:]
			npfx = []rune(node.Prefix)
			goto restart
		}

		if qry[i] != npfx[i] {
			return
		}

		i++
	}
}

// HasPrefixes returns true if key has any prefixes.
func (self *Trie[V]) HasPrefixes(key string) bool {

	var qry = []rune(key)
	var idx, found = self.root.Branches.find(qry[0])
	if !found {
		return false
	}
	var node = self.root.Branches[idx]
	var npfx = []rune(node.Prefix)

restart:
	var i = 0
	for {
		if i == len(qry) {
			return false
		}

		if i == len(npfx) {
			if node.HasValue {
				return true
			}

			idx, found = node.Branches.find(qry[i])
			if !found {
				return false
			}

			node = node.Branches[idx]
			qry = qry[i:]
			npfx = []rune(node.Prefix)
			goto restart
		}

		if qry[i] != npfx[i] {
			return false
		}

		i++
	}
}

// Print writes self to writer w as a multiline string representing the tree
// structure.
//
// It is formatted as one node per line where child nodes are indented with two
// spaces each level and line is in format: <indent><prefix>[,value]
func (self *Trie[V]) Print(w io.Writer) {
	self.print(w, self.root, 0)
}

func (self *Trie[V]) print(w io.Writer, n *Node[V], indent int) {
	fmt.Fprintf(w, "%s%s", mkindent(indent), string(n.Prefix))
	if n.HasValue {
		fmt.Fprintf(w, ",%v", n.Value)
	}
	fmt.Fprintf(w, "\n")
	if len(n.Branches) > 0 {
		for _, v := range n.Branches {
			self.print(w, v, indent+1)
		}
	}
}

func mkindent(depth int) string { return strings.Repeat("  ", depth) }

// Node represents a node in the tree.
type Node[V any] struct {
	Prefix   []rune
	Value    V
	HasValue bool
	Branches[V]
}

// Branches is a slice of node branches.
// Each unique starting utf8 code point is in its own slice.
type Branches[V any] []*Node[V]

// find returns Branches index at which a node whose prefix begins with s and
// true or insert index and false if not found.
func (self Branches[V]) find(r rune) (idx int, match bool) {
	return binSearch(len(self), func(i int) int {
		var v = []rune(self[i].Prefix)[0]
		if r > v {
			return 1
		} else if r < v {
			return -1
		}
		return 0
	})
}

func binSearch(n int, cmp func(int) int) (i int, found bool) {
	i, j := 0, n
	for i < j {
		h := int(uint(i+j) >> 1)
		if cmp(h) > 0 {
			i = h + 1
		} else {
			j = h
		}
	}
	return i, i < n && cmp(i) == 0
}

// SyncTrie is the concurrency safe version of [Trie].
type SyncTrie[V any] struct {
	mutex sync.RWMutex
	trie  Trie[V]
}

// NewSyncTrie returns a new [SyncTrie].
func NewSyncTrie[V any]() *SyncTrie[V] {
	return &SyncTrie[V]{
		trie: *New[V](),
	}
}

// Put inserts value under key.
//
// If a value already exists at key it is returned with true, otherwise a zero
// value of V is returned and false.
//
// Key must not be empty. If it is no value is inserted and a zero value of V
// and false is returned.
func (self *SyncTrie[V]) Put(key string, value V) (old V, replaced bool) {
	self.mutex.Lock()
	old, replaced = self.trie.Put(key, value)
	self.mutex.Unlock()
	return
}

// Get returns the value at key and true if it exists or a zero value of V and
// false if not found.
//
// Key must not be empty. If it is Get returns a zero value and false.
func (self *SyncTrie[V]) Get(key string) (value V, found bool) {
	self.mutex.RLock()
	value, found = self.trie.Get(key)
	self.mutex.Unlock()
	return
}

// Exists returns true if key exists.
func (self *SyncTrie[V]) Exists(key string) (exists bool) {
	self.mutex.RLock()
	exists = self.trie.Exists(key)
	self.mutex.RUnlock()
	return
}

// Prefixes returns a list of set keys which are a prefix of key.
func (self *SyncTrie[V]) Prefixes(key string) (out []string) {
	self.mutex.RLock()
	out = self.trie.Prefixes(key)
	self.mutex.RUnlock()
	return
}

// HasPrefixes returns true if key has any prefixes.
func (self *SyncTrie[V]) HasPrefixes(key string) (truth bool) {
	self.mutex.RLock()
	truth = self.trie.HasPrefixes(key)
	self.mutex.RUnlock()
	return
}

// Print writes self to writer w as a multiline string representing the tree
// structure.
//
// It is formatted as one node per line where child nodes are indented with two
// spaces each level and line is in format: <indent><prefix>[,value]
func (self *SyncTrie[V]) Print(w io.Writer) {
	self.mutex.RLock()
	self.trie.Print(w)
	self.mutex.RUnlock()
}
