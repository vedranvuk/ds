package trie

import (
	"fmt"
	"slices"
	"strings"
)

// Trie implements a binary prefix tree.
type Trie[V any] struct {
	root *Node[V]
}

// New returns a new *Trie.
func New[V any]() *Trie[V] {
	return &Trie[V]{
		root: new(Node[V]),
	}
}

// Put inserts value under key. If a value already exists at key it is returned
// with true, otherwise a zero value of V is returned and false.
func (self Trie[V]) Put(key string, value V) (old V, replaced bool) {

	if key == "" {
		panic("invalid key, must not be empty")
	}

	var qry = []rune(key)
	var idx, found = self.root.Branches.find(qry[0])

	// fast path, no branch for key starting rune.
	if !found {
		var node = &Node[V]{
			Prefix:   qry,
			Value:    value,
			HasValue: true,
		}
		self.root.Branches = slices.Insert(self.root.Branches, idx, node)
		return *new(V), false
	}

	var node = self.root.Branches[idx]
	var npfx = []rune(node.Prefix)

restart:
	var i = 0
	for {
		// end of query reached
		if i == len(qry) {

			// Node prefix fully matched.
			if i == len(npfx) {
				old, replaced = node.Value, node.HasValue
				node.Value, node.HasValue = value, true
				return
			}

			// node prefix partially matched, split node.
			var newNode = &Node[V]{
				Prefix:   npfx[i:],
				Value:    node.Value,
				HasValue: node.HasValue,
				Branches: node.Branches,
			}
			node.Prefix = qry
			node.Value = value
			node.HasValue = true
			node.Branches = Branches[V]{newNode}

			return *new(V), false
		}

		// end of node prefix reached
		if i == len(npfx) {
			idx, found = node.Branches.find(qry[i])

			// No branches found, insert new.
			if !found {
				var newNode = &Node[V]{
					Prefix:   qry[i:],
					Value:    value,
					HasValue: true,
				}
				node.Branches = slices.Insert(node.Branches, idx, newNode)
				return *new(V), false
			}

			// Set current node to node under matched branch and restart.
			node = node.Branches[idx]
			qry = qry[i:]
			npfx = []rune(node.Prefix)
			goto restart
		}

		// missmatch at the middle of node prefix, split node.
		if qry[i] != npfx[i] {

			var insertNode = &Node[V]{
				Prefix:   qry[i:],
				Value:    value,
				HasValue: true,
			}

			var splitNode = &Node[V]{
				Prefix:   npfx[i:],
				Value:    node.Value,
				HasValue: node.HasValue,
				Branches: node.Branches,
			}

			node.Prefix = qry[:i]
			node.Value = *new(V)
			node.HasValue = false
			if splitNode.Prefix[0] > insertNode.Prefix[0] {
				node.Branches = Branches[V]{insertNode, splitNode}
			} else {
				node.Branches = Branches[V]{splitNode, insertNode}
			}

			return *new(V), false
		}

		i++
	}
}

// Get returns the value at key and true if it exists or a zero value of V and
// false if not found.
func (self Trie[V]) Get(key string) (value V, found bool) {

	if key == "" {
		panic("invalid key, must not be empty")
	}

	var qry = []rune(key)
	var idx int
	if idx, found = self.root.Branches.find(qry[0]); !found {
		return *new(V), false
	}
	var node = self.root.Branches[idx]
	var npfx = []rune(node.Prefix)

restart:
	var i = 0
	for {
		// end of query reached
		if i == len(qry) {
			if node.HasValue {
				return node.Value, true
			}
			return *new(V), false
		}

		// end of node prefix reached
		if i == len(npfx) {
			idx, found = node.Branches.find(qry[i])

			// No branches found, insert new.
			if !found {
				return *new(V), false
			}

			// Set current node to node under matched branch and restart.
			node = node.Branches[idx]
			qry = qry[i:]
			npfx = []rune(node.Prefix)
			goto restart
		}

		// missmatch at the middle of node prefix, split node.
		if qry[i] != npfx[i] {
			return *new(V), false
		}

		i++
	}
}

// Exists returns true if key exists.
func (self Trie[V]) Exists(key string) (exists bool) {
	_, exists = self.Get(key)
	return
}

// Prefixes returns a list of set keys which are a prefix of key.
func (self Trie[V]) Prefixes(key string) []string {
	return nil
}

// HasPrefixes returns true if key has any prefixes.
func (self Trie[V]) HasPrefixes(key string) bool {
	return false
}

func (self Trie[V]) Print() {
	self.print(self.root, 0)
}

func (self Trie[V]) print(n *Node[V], indent int) {
	fmt.Printf("%sNODE \"%s\"", mkindent(indent), string(n.Prefix))
	if n.HasValue {
		fmt.Printf(" VALUE: %v", n.Value)
	}
	fmt.Printf("\n")
	if len(n.Branches) > 0 {
		for _, v := range n.Branches {
			self.print(v, indent+1)
		}
	}
}

func mkindent(depth int) string { return strings.Repeat("\t", depth) }

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

// nodeByRune returns a node from index at which [Node.Prefix][0] matches r.
// Returns nil if no such node.
func (self Branches[V]) nodeByRune(r rune) *Node[V] {

	if idx, match := self.find(r); match {
		return self[idx]
	}

	return nil
}

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
