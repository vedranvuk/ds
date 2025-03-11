// Copyright 2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package trie

import (
	"os"
	"slices"
	"strconv"
	"sync"
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

func TestTriePrefixes(t *testing.T) {
	tr := New[int]()
	tr.Put("/", 0)
	tr.Put("/users", 0)
	tr.Put("/users/vedran", 0)
	tr.Put("/users/vedran/go", 0)
	if !slices.Equal(tr.Prefixes("/users/vedran/go"), []string{"/", "/users", "/users/vedran"}) {
		t.Fatal()
	}
}

func TestPrint(t *testing.T) {
	var trie = New[int]()
	for _, test := range tests2 {
		trie.Put(test.Key, test.Val)
	}
	trie.Print(os.Stdout)
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
	for i := 0; i < b.N; i++ {
		tree.Put(entries[i], i)
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

func TestNew(t *testing.T) {
	tr := New[int]()
	if tr == nil {
		t.Fatal("New should not return nil")
	}
}

func TestNewSyncTrie(t *testing.T) {
	tr := NewSyncTrie[int]()
	if tr == nil {
		t.Fatal("NewSyncTrie should not return nil")
	}
}

func TestTrie_Put_Get_Exists(t *testing.T) {
	tr := New[int]()

	// Empty key
	old, replaced := tr.Put("", 1)
	if replaced || old != 0 {
		t.Errorf("Put empty key should return zero value and false, got (%v, %v)", old, replaced)
	}
	val, found := tr.Get("")
	if found || val != 0 {
		t.Errorf("Get empty key should return zero value and false, got (%v, %v)", val, found)
	}
	if tr.Exists("") {
		t.Errorf("Exists empty key should return false, got true")
	}

	// New key
	old, replaced = tr.Put("key1", 10)
	if replaced || old != 0 {
		t.Errorf("Put new key should return zero value and false, got (%v, %v)", old, replaced)
	}
	val, found = tr.Get("key1")
	if !found || val != 10 {
		t.Errorf("Get existing key should return value and true, got (%v, %v)", val, found)
	}
	if !tr.Exists("key1") {
		t.Errorf("Exists existing key should return true, got false")
	}

	// Existing key replace
	old, replaced = tr.Put("key1", 20)
	if !replaced || old != 10 {
		t.Errorf("Put existing key should return old value and true, got (%v, %v)", old, replaced)
	}
	val, found = tr.Get("key1")
	if !found || val != 20 {
		t.Errorf("Get existing key after replace should return new value and true, got (%v, %v)", val, found)
	}
	if !tr.Exists("key1") {
		t.Errorf("Exists existing key after replace should return true, got false")
	}

	// Non-existent key
	val, found = tr.Get("key2")
	if found || val != 0 {
		t.Errorf("Get non-existent key should return zero value and false, got (%v, %v)", val, found)
	}
	if tr.Exists("key2") {
		t.Errorf("Exists non-existent key should return false, got true")
	}
}

func TestTrie_Delete(t *testing.T) {
	tr := New[int]()

	// Delete non-existent key
	val, deleted := tr.Delete("key1")
	if deleted || val != 0 {
		t.Errorf("Delete non-existent key should return zero value and false, got (%v, %v)", val, deleted)
	}

	// Put and Delete existing key
	tr.Put("key1", 10)
	val, deleted = tr.Delete("key1")
	if !deleted || val != 10 {
		t.Errorf("Delete existing key should return old value and true, got (%v, %v)", val, deleted)
	}
	_, found := tr.Get("key1")
	if found {
		t.Errorf("Get after delete should return false, got true")
	}

	// Test node deletion on path
	tr.Put("key1", 10)
	tr.Put("key2", 20)
	tr.Delete("key1")
	_, found1 := tr.Get("key1")
	_, found2 := tr.Get("key2")
	if found1 {
		t.Errorf("Get key1 after delete should return false, got true")
	}
	if !found2 {
		t.Errorf("Get key2 after delete of key1 should return true, got false")
	}

	// Delete empty key
	val, deleted = tr.Delete("")
	if deleted || val != 0 {
		t.Errorf("Delete empty key should return zero value and false, got (%v, %v)", val, deleted)
	}
}

func TestTrie_Prefixes_HasPrefixes(t *testing.T) {
	tr := New[int]()

	// No prefixes
	prefixes := tr.Prefixes("key")
	if len(prefixes) != 0 {
		t.Errorf("Prefixes with no prefixes should return empty slice, got %v", prefixes)
	}
	if tr.HasPrefixes("key") {
		t.Errorf("HasPrefixes with no prefixes should return false, got true")
	}

	// Add prefixes
	tr.Put("k", 1)
	tr.Put("ke", 2)
	tr.Put("key", 3)
	tr.Put("keys", 4)
	tr.Put("longerkey", 5)

	// Check prefixes
	prefixes = tr.Prefixes("key")
	expectedPrefixes := []string{"k", "ke"}
	if !slices.Equal(prefixes, expectedPrefixes) {
		t.Errorf("Prefixes for 'key' should be %v, got %v", expectedPrefixes, prefixes)
	}
	if !tr.HasPrefixes("key") {
		t.Errorf("HasPrefixes for 'key' should return true, got false")
	}

	// Check prefixes for longer key
	prefixes = tr.Prefixes("longerkey")
	expectedPrefixes = []string{}
	if !slices.Equal(prefixes, expectedPrefixes) {
		t.Errorf("Prefixes for 'longerkey' should be %v, got %v", expectedPrefixes, prefixes)
	}
	if tr.HasPrefixes("longerkey") {
		t.Errorf("HasPrefixes for 'longerkey' should return true, got false")
	}

	// Check prefixes for non-existent key with existing prefixes
	prefixes = tr.Prefixes("keyss")
	expectedPrefixes = []string{"k", "ke", "key", "keys"}
	if !slices.Equal(prefixes, expectedPrefixes) {
		t.Errorf("Prefixes for 'keyss' should be %v, got %v", expectedPrefixes, prefixes)
	}
	if !tr.HasPrefixes("keyss") {
		t.Errorf("HasPrefixes for 'keyss' should return true, got false")
	}

	// Check prefixes for exact match key
	prefixes = tr.Prefixes("keys")
	expectedPrefixes = []string{"k", "ke", "key"}
	if !slices.Equal(prefixes, expectedPrefixes) {
		t.Errorf("Prefixes for 'keys' should be %v, got %v", expectedPrefixes, prefixes)
	}
	if !tr.HasPrefixes("keys") {
		t.Errorf("HasPrefixes for 'keys' should return true, got false")
	}

	// Empty key
	prefixes = tr.Prefixes("")
	if len(prefixes) != 0 {
		t.Errorf("Prefixes for empty key should return empty slice, got %v", prefixes)
	}
	if tr.HasPrefixes("") {
		t.Errorf("HasPrefixes for empty key should return false, got true")
	}
}

func TestSyncTrie_Put_Get_Delete_Exists_Prefixes_HasPrefixes(t *testing.T) {
	str := NewSyncTrie[int]()

	// Test Put, Get, Exists, Delete, Prefixes, HasPrefixes functionalities, same as Trie tests but for SyncTrie

	// Empty key
	_, replaced := str.Put("", 1)
	if replaced {
		t.Error("SyncTrie Put empty key should return false")
	}
	_, found := str.Get("")
	if found {
		t.Error("SyncTrie Get empty key should return false")
	}
	if str.Exists("") {
		t.Error("SyncTrie Exists empty key should return false")
	}
	_, deleted := str.Delete("")
	if deleted {
		t.Error("SyncTrie Delete empty key should return false")
	}
	prefixes := str.Prefixes("")
	if len(prefixes) != 0 {
		t.Errorf("SyncTrie Prefixes empty key should return empty slice, got %v", prefixes)
	}
	if str.HasPrefixes("") {
		t.Error("SyncTrie HasPrefixes empty key should return false")
	}

	// Put and Get
	str.Put("key1", 10)
	val, found := str.Get("key1")
	if !found || val != 10 {
		t.Errorf("SyncTrie Get existing key failed, expected (10, true), got (%v, %v)", val, found)
	}
	if !str.Exists("key1") {
		t.Error("SyncTrie Exists existing key failed, expected true, got false")
	}

	// Replace and Get
	old, replaced := str.Put("key1", 20)
	if !replaced || old != 10 {
		t.Errorf("SyncTrie Put replace existing key failed, expected (10, true), got (%v, %v)", old, replaced)
	}
	val, found = str.Get("key1")
	if !found || val != 20 {
		t.Errorf("SyncTrie Get after replace failed, expected (20, true), got (%v, %v)", val, found)
	}

	// Delete and Get
	oldVal, deleted := str.Delete("key1")
	if !deleted || oldVal != 20 {
		t.Errorf("SyncTrie Delete existing key failed, expected (20, true), got (%v, %v)", oldVal, deleted)
	}
	_, found = str.Get("key1")
	if found {
		t.Error("SyncTrie Get after delete failed, expected false, got true")
	}
	if str.Exists("key1") {
		t.Error("SyncTrie Exists after delete failed, expected false, got true")
	}

	// Prefixes and HasPrefixes
	str.Put("k", 1)
	str.Put("ke", 2)
	str.Put("key", 3)
	prefixes = str.Prefixes("key")
	expectedPrefixes := []string{"k", "ke"}
	if !slices.Equal(prefixes, expectedPrefixes) {
		t.Errorf("SyncTrie Prefixes failed, expected %v, got %v", expectedPrefixes, prefixes)
	}
	if !str.HasPrefixes("key") {
		t.Error("SyncTrie HasPrefixes failed, expected true, got false")
	}
	str.Delete("k")
	str.Delete("ke")
	str.Delete("key")
}

func TestSyncTrie_Concurrent(t *testing.T) {
	str := NewSyncTrie[int]()
	var wg sync.WaitGroup
	numRoutines := 100
	numOps := 100

	// Concurrent Puts
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := "key_" + strconv.Itoa(routineID) + "_" + strconv.Itoa(j)
				str.Put(key, routineID*numOps+j)
			}
		}(i)
	}
	wg.Wait()

	// Concurrent Gets
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := "key_" + strconv.Itoa(routineID) + "_" + strconv.Itoa(j)
				val, found := str.Get(key)
				expectedVal := routineID*numOps + j
				if !found || val != expectedVal {
					t.Errorf("Concurrent Get failed for key %s, expected (%d, true), got (%v, %v)", key, expectedVal, val, found)
				}
			}
		}(i)
	}
	wg.Wait()

	// Concurrent Deletes
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := "key_" + strconv.Itoa(routineID) + "_" + strconv.Itoa(j)
				val, deleted := str.Delete(key)
				expectedVal := routineID*numOps + j
				if !deleted || val != expectedVal {
					t.Errorf("Concurrent Delete failed for key %s, expected (%d, true), got (%v, %v)", key, expectedVal, val, deleted)
				}
			}
		}(i)
	}
	wg.Wait()

	// Verify all deleted
	for i := 0; i < numRoutines; i++ {
		for j := 0; j < numOps; j++ {
			key := "key_" + strconv.Itoa(i) + "_" + strconv.Itoa(j)
			_, found := str.Get(key)
			if found {
				t.Errorf("Key %s should have been deleted but still found", key)
			}
		}
	}
}
