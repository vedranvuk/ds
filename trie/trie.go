// Copyright 2025 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

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
// It has fast lookups and can retrieve keys which are a prefix or suffix of 
// some key.
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

// Put inserts value keyed by key.
//
// If an entry under key already exists its value is replaced with value 
// argument and the old value is returned with replaced being true. Otherwise a
// zero value of V is returned and false.
//
// Key must not be empty. If it is, no value is inserted and a zero value of V
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
	var nodeRunes = currentNode.Prefix

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
			if idx, found = currentNode.Branches.find(keyRunes[i]); !found {
				// No branches found, insert new.
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
			nodeRunes = currentNode.Prefix
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

// Get returns the value at key and true if key exists or a zero value of V and
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
	var npfx = node.Prefix

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
			if idx, found = node.Branches.find(qry[i]); !found {
				return self.zero, false
			}
			node = node.Branches[idx]
			qry = qry[i:]
			npfx = node.Prefix
			goto restart
		}

		if qry[i] != npfx[i] {
			return self.zero, false
		}

		i++
	}
}

// Delete deletes an entry by key. It returns true if entry was found and
// deleted and false if not found.
//
// After an entry has been deleted its parent nodes are traversed back towards
// root and deleted if they have no other child branches. This means that a
// single Delete operation may remove additional unused nodes. It can also
// happen that no nodes get deleted in case the node that contains the value
// has child branches of its own.
//
// It merges nodes that have single branches along the parent path to avoid 
// fragmentation.
func (self *Trie[V]) Delete(key string) (value V, deleted bool) {
	if key == "" {
		return self.zero, false
	}

	var qry = []rune(key)
	var branches = &self.root.Branches
	var idx int
	var found bool

	if idx, found = branches.find(qry[0]); !found {
		return self.zero, false
	}

	var node = self.root.Branches[idx]
	var npfx = node.Prefix
	var path []*Node[V]

restart:
	var i = 0
	for {
		if i == len(qry) {
			if !node.HasValue {
				return self.zero, false
			}
			value, deleted = node.Value, true
			node.Value = self.zero
			node.HasValue = false

			// Backtrack and merge/delete nodes
			for len(path) > 0 {
				var parent *Node[V] = path[len(path)-1]
				path = path[:len(path)-1] // Pop the last element

				// Find the index of the current node in the parent's branches
				var childIndex int
				var childFound bool
				if childIndex, childFound = parent.Branches.find(node.Prefix[0]); !childFound {
					panic("child not found in parent branches") // This should not happen
				}

				// If the current node has no branches, delete it from the parent
				if len(node.Branches) == 0 && !node.HasValue {
					parent.Branches = slices.Delete(parent.Branches, childIndex, childIndex+1)
					deleted = true
					// If the parent now has only one branch, merge it with the parent
					if len(parent.Branches) == 1 && !parent.HasValue {
						var remainingChild *Node[V] = parent.Branches[0]
						parent.Prefix = append(parent.Prefix, remainingChild.Prefix...)
						parent.Value = remainingChild.Value
						parent.HasValue = remainingChild.HasValue
						parent.Branches = remainingChild.Branches
					}
				}
				node = parent // Move up to the parent node
			}

			// Special case for root node
			if len(node.Branches) == 0 && !node.HasValue && node != self.root {
				if idx, found = self.root.Branches.find(node.Prefix[0]); !found {
					panic("should not happen")
				}
				self.root.Branches = slices.Delete(self.root.Branches, idx, idx+1)
				deleted = true
			}
			return
		}

		if i == len(npfx) {
			branches = &node.Branches
			if idx, found = branches.find(qry[i]); !found {
				return self.zero, false
			}
			path = append(path, node)
			node = node.Branches[idx]
			qry = qry[i:]
			npfx = node.Prefix
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
// For example, if trie has following keys:
//  - foo
//  - foobar
//  - foobarbaz
// Prefixes then returns ["foo", "foobar"] for key "foobarbaz".
func (self *Trie[V]) Prefixes(key string) (out []string) {

	if key == "" {
		return
	}

	var scanned []rune

	var qry = []rune(key)
	var idx, found = self.root.Branches.find(qry[0])
	if !found {
		return
	}
	var node = self.root.Branches[idx]
	var npfx = node.Prefix

restart:
	var i = 0
	for {
		if i == len(qry) {
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
			npfx = node.Prefix
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

	if key == "" {
		return false
	}
	var qry = []rune(key)
	var idx, found = self.root.Branches.find(qry[0])
	if !found {
		return false
	}
	var node = self.root.Branches[idx]
	var npfx = node.Prefix

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
			npfx = node.Prefix
			goto restart
		}

		if qry[i] != npfx[i] {
			return false
		}

		i++
	}
}

// Suffixes returns a list of set keys which are a suffix of key.
//
// For example, if trie has following keys:
//  - foo
//  - foobar
//  - foobarbaz
// suffixes then returns ["foobar", "foobarbaz"] for key "foo".
func (self *Trie[V]) Suffixes(key string) (out []string) {

	if key == "" {
		return nil
	}

	var qry = []rune(key)
	var idx int
	var found bool
	if idx, found = self.root.Branches.find(qry[0]); !found {
		return nil
	}
	var node = self.root.Branches[idx]
	var npfx = node.Prefix
	var scanned []rune

restart:
	var i = 0
	for {
		if i == len(qry) {
			self.enumKeys(node, "", func(k string) bool {
				if s := string(scanned) + k; s != key {
					out = append(out, string(scanned)+k)
				}
				return true
			})
			return
		}

		if i == len(npfx) {
			scanned = append(scanned, node.Prefix...)
			if idx, found = node.Branches.find(qry[i]); !found {
				return nil
			}
			node = node.Branches[idx]
			qry = qry[i:]
			npfx = node.Prefix
			goto restart
		}

		if qry[i] != npfx[i] {
			return nil
		}

		i++
	}
}

// Suffixes returns a list of set keys which are suffixes of key argument.
func (self *Trie[V]) HasSuffixes(key string) bool {

	if key == "" {
		return false
	}

	var qry = []rune(key)
	var idx int
	var found bool
	if idx, found = self.root.Branches.find(qry[0]); !found {
		return false
	}
	var node = self.root.Branches[idx]
	var npfx = node.Prefix
	var scanned []rune

restart:
	var i = 0
	for {
		if i == len(qry) {
			var has bool
			self.enumKeys(node, "", func(k string) bool {
				if s := string(scanned) + k; s != key {
					has = true
					return false
				}
				return true
			})
			return has
		}

		if i == len(npfx) {
			scanned = append(scanned, node.Prefix...)
			if idx, found = node.Branches.find(qry[i]); !found {
				return false
			}
			node = node.Branches[idx]
			qry = qry[i:]
			npfx = node.Prefix
			goto restart
		}

		if qry[i] != npfx[i] {
			return false
		}

		i++
	}
}

// Enum enumerates all key-value pairs in the Trie.
//
// It calls the provided function 'f' for each key-value pair in the Trie.
// The enumeration stops if 'f' returns false.
func (self *Trie[V]) Enum(f func(key string, value V) bool) {
	self.enum(self.root, "", f)
}

func (self *Trie[V]) enum(node *Node[V], prefix string, f func(key string, value V) bool) {
	var currentKey string
	currentKey = prefix + string(node.Prefix)

	if node.HasValue {
		if !f(currentKey, node.Value) {
			return
		}
	}

	for _, child := range node.Branches {
		self.enum(child, currentKey, f)
	}
}

// EnumKeys enumerates all keys in the Trie.
//
// It calls the provided function 'f' for each key in the Trie.
// The enumeration stops if 'f' returns false.
func (receiver *Trie[V]) EnumKeys(f func(key string) bool) {
	receiver.enumKeys(receiver.root, "", f)
}

func (receiver *Trie[V]) enumKeys(node *Node[V], prefix string, f func(key string) bool) {
	var currentKey string
	currentKey = prefix + string(node.Prefix)

	if node.HasValue {
		if !f(currentKey) {
			return
		}
	}

	for _, child := range node.Branches {
		receiver.enumKeys(child, currentKey, f)
	}
}

// EnumValues enumerates all values in the Trie.
//
// It calls the provided function 'f' for each value in the Trie.
// The enumeration stops if 'f' returns false.
func (self *Trie[V]) EnumValues(f func(value V) bool) {
	self.enumValues(self.root, f)
}

func (receiver *Trie[V]) enumValues(node *Node[V], f func(value V) bool) {
	if node.HasValue {
		if !f(node.Value) {
			return
		}
	}

	for _, child := range node.Branches {
		receiver.enumValues(child, f)
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
		var v = self[i].Prefix[0]
		if r > v {
			return 1
		} else if r < v {
			return -1
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
	self.mutex.RUnlock()
	return
}

// Delete deletes an entry by key. It returns true if entry was found and
// deleted and false if not found.
//
// After an entry has been deleted its parent nodes are traversed back towards
// root and deleted if they have no other child branches. This means that a
// single Delete operation may remove additional unused nodes. It can also
// happen that no nodes get deleted in case the node that contains the value
// has child branches of its own.
//
// It does not merge nodes that have single branches which results in tree
// fragmentation after delete operations.
func (self *SyncTrie[V]) Delete(key string) (value V, deleted bool) {
	self.mutex.Lock()
	value, deleted = self.trie.Delete(key)
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

// Suffixes returns a list of set keys which are suffixes of key argument.
func (self *SyncTrie[V]) Suffixes(key string) (out []string) {
	self.mutex.RLock()
	out = self.trie.Suffixes(key)
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
