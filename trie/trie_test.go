package trie

import (
	"os"
	"testing"

	"github.com/vedranvuk/strutils"
)

const input = "01234567"

var runev rune

func BenchmarkConvertToRunes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runev = []rune(input)[0]
	}
}

var bytev byte

func BenchmarkConvertToByteSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bytev = []byte(input)[0]
	}
}

type Test struct {
	Key      string
	Val      int
	Old      int
	Replaced bool
}

var tests = []Test{
	{"apple", 1, 0, false},
	{"appleseed", 2, 0, false},
	{"app", 3, 0, false},
	{"absolute", 4, 0, false},
	{"ablative", 5, 0, false},
	{"beach", 6, 0, false},
	{"bleach", 7, 0, false},
	{"blue", 8, 0, false},
	{"blueish", 9, 0, false},
}

func TestTrie(t *testing.T) {
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
	tree.Print(os.Stdout)
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
