// Copyright 2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package trie

import (
	"testing"

	"github.com/vedranvuk/strutils"
)

const input = "01234567"

var runev rune

func TestTriePut(t *testing.T) {
    tr := New[int]()
    old, replaced := tr.Put("key", 1)
    if old != 0 || replaced {
        t.Errorf("expected (0, false), got (%v, %v)", old, replaced)
    }
    old, replaced = tr.Put("key", 2)
    if old != 1 || !replaced {
        t.Errorf("expected (1, true), got (%v, %v)", old, replaced)
    }
}

func TestTrieGet(t *testing.T) {
    tr := New[int]()
    tr.Put("key", 1)
    value, found := tr.Get("key")
    if value != 1 || !found {
        t.Errorf("expected (1, true), got (%v, %v)", value, found)
    }
    value, found = tr.Get("nonexistent")
    if value != 0 || found {
        t.Errorf("expected (0, false), got (%v, %v)", value, found)
    }
}

func TestTrieDelete(t *testing.T) {
    tr := New[int]()
    tr.Put("key", 1)
    value, deleted := tr.Delete("key")
    if value != 1 || !deleted {
        t.Errorf("expected (1, true), got (%v, %v)", value, deleted)
    }
    value, deleted = tr.Delete("nonexistent")
    if value != 0 || deleted {
        t.Errorf("expected (0, false), got (%v, %v)", value, deleted)
    }
}

type Test struct {
	Key      string
	Val      int
	Old      int
	Replaced bool
}

var tests1 = []Test{
	{"apple", 1, 0, false},
	{"appleseed", 2, 0, false},
	{"app", 3, 0, false},
	{"absolute", 4, 0, false},
	{"ablative", 5, 0, false},
	{"beach", 6, 0, false},
	{"bleach", 7, 0, false},
	{"blue", 8, 0, false},
	{"blueish", 9, 0, false},
	{"blueberry", 10, 0, false},
	{"bluebird", 11, 0, false},
	{"bluebell", 12, 0, false},
	{"bluebonnet", 13, 0, false},
}

var tests2 = []Test{
	{"/", 1, 0, false},
	{"/home", 2, 0, false},
	{"/home/user", 3, 0, false},
	{"/home/user/documents", 4, 0, false},
	{"/home/user/downloads", 5, 0, false},
	{"/home/user/music", 6, 0, false},
	{"/home/user/pictures", 7, 0, false},
	{"/home/user/videos", 8, 0, false},
	{"/home/user/.config", 9, 0, false},
	{"/home/user/.local", 10, 0, false},
	{"/home/user/.cache", 11, 0, false},
}

func TestTrie(t *testing.T) {
	runTests(t, tests1)
	runTests(t, tests2)
}

func runTests(t *testing.T, tests []Test) {
	tree := New[int]()
	for _, v := range tests {
		var old, replaced = tree.Put(v.Key, v.Val)
		if old != v.Old {
			t.Fatalf("Put %s failed, Expected old=%v, got old=%v", v.Key, v.Old, old)
		}
		if replaced != v.Replaced {
			t.Fatalf("Put %s failed, Expected replaced=%v, got replaced=%v", v.Key, v.Replaced, replaced)
		}
	}
	for _, v := range tests {
		var val, found = tree.Get(v.Key)
		if val != v.Val {
			t.Fatalf("Get %s failed, Expected val=%v, got val=%v", v.Key, v.Val, val)
		}
		if !found {
			t.Fatalf("Get %s failed, Expected found=true, got found=%v", v.Key, found)
		}
	}
	for _, v := range tests {
		tree.Delete(v.Key)
	}
}

func BenchmarkPut(b *testing.B) {
	tree := New[int]()
	foo := strutils.NewFoo()
	entries := make([]string, 0, b.N)
	for i := 0; i < b.N; i++ {
		entries = append(entries, foo.Name())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Put(entries[i], i)
	}
}

func BenchmarkGet(b *testing.B) {
	tree := New[int]()
	foo := strutils.NewFoo()
	entries := make([]string, 0, b.N)
	for i := 0; i < b.N; i++ {
		entries = append(entries, foo.Name())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Get(entries[i])
	}
}

func BenchmarkDelete(b *testing.B) {
	tree := New[int]()
	foo := strutils.NewFoo()
	entries := make([]string, 0, b.N)
	for i := 0; i < b.N; i++ {
		entries = append(entries, foo.Name())
	}
	for i := 0; i < b.N; i++ {
		tree.Put(entries[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Delete(entries[i])
	}
}
