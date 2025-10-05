// maps/maps_test.go
// Copyright 2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package maps

import (
	"math/rand/v2"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestOrderedMap(t *testing.T) {
	t.Run("Basic Operations", func(t *testing.T) {
		m := NewOrderedMap[string, int]()

		// Put and Len
		m.Put("a", 1)
		m.Put("b", 2)
		m.Put("c", 3)
		if m.Len() != 3 {
			t.Errorf("Expected length 3, got %d", m.Len())
		}

		// Exists
		if !m.Exists("b") {
			t.Error("Expected 'b' to exist")
		}
		if m.Exists("d") {
			t.Error("Expected 'd' not to exist")
		}

		// Get
		val, ok := m.Get("a")
		if !ok {
			t.Error("Expected 'a' to be found")
		}
		if val != 1 {
			t.Errorf("Expected value 1 for 'a', got %d", val)
		}
		val, ok = m.Get("d")
		if ok {
			t.Error("Expected 'd' not to be found")
		}
		if val != 0 {
			t.Errorf("Expected zero value for 'd', got %v", val)
		}

		// GetAt
		val, ok = m.GetAt(1)
		if !ok {
			t.Error("Expected index 1 to be found")
		}
		if val != 2 {
			t.Errorf("Expected value 2 at index 1, got %d", val)
		}
		val, ok = m.GetAt(5)
		if ok {
			t.Error("Expected index 5 not to be found")
		}
		if val != 0 {
			t.Errorf("Expected zero value at index 5, got %v", val)
		}

		// Put (replace)
		oldVal, existed := m.Put("b", 4)
		if !existed {
			t.Error("Expected 'b' to be replaced")
		}
		if oldVal != 2 {
			t.Errorf("Expected old value 2 for 'b', got %d", oldVal)
		}
		val, _ = m.Get("b")
		if val != 4 {
			t.Errorf("Expected value 4 for 'b', got %d", val)
		}

		// Delete
		oldVal, existed = m.Delete("a")
		if !existed {
			t.Error("Expected 'a' to be deleted")
		}
		if oldVal != 1 {
			t.Errorf("Expected old value 1 for 'a', got %d", oldVal)
		}
		if m.Exists("a") {
			t.Error("Expected 'a' not to exist after deletion")
		}
		if m.Len() != 2 {
			t.Errorf("Expected length 2 after deletion, got %d", m.Len())
		}

		// DeleteAt
		oldVal, existed = m.DeleteAt(0)
		if !existed {
			t.Error("Expected index 0 to be deleted")
		}
		if oldVal != 4 {
			t.Errorf("Expected old value 4 at index 0, got %d", oldVal)
		}
		if m.Len() != 1 {
			t.Errorf("Expected length 1 after deletion, got %d", m.Len())
		}

		// Delete non existent
		oldVal, existed = m.Delete("foo")
		if existed {
			t.Error("Expected 'foo' not to be deleted")
		}
		if oldVal != 0 {
			t.Errorf("Expected zero value 'foo', got %d", oldVal)
		}

		// DeleteAt out of range
		oldVal, existed = m.DeleteAt(69)
		if existed {
			t.Error("Expected index 69 not to be deleted")
		}
		if oldVal != 0 {
			t.Errorf("Expected zero value at index 69, got %d", oldVal)
		}

		// EnumKeys
		keys := make([]string, 0)
		m.EnumKeys(func(k string) bool {
			keys = append(keys, k)
			return true
		})
		if len(keys) != m.Len() {
			t.Errorf("Expected number of keys to match map size, got %d and %d", len(keys), m.Len())
		}

		// EnumValues
		values := make([]int, 0)
		m.EnumValues(func(v int) bool {
			values = append(values, v)
			return true
		})
		if len(values) != m.Len() {
			t.Errorf("Expected number of values to match map size, got %d and %d", len(values), m.Len())
		}

		// Keys
		keysSlice := m.Keys()
		if len(keysSlice) != m.Len() {
			t.Errorf("Expected number of keys to match map size, got %d and %d", len(keysSlice), m.Len())
		}

		// Values
		valuesSlice := m.Values()
		if len(valuesSlice) != m.Len() {
			t.Errorf("Expected number of values to match map size, got %d and %d", len(valuesSlice), m.Len())
		}
	})

	t.Run("Deletion Edge Cases", func(t *testing.T) {
		m := NewOrderedMap[string, int]()
		m.Put("a", 1)
		m.Put("b", 2)
		m.Put("c", 3)

		// Delete first element
		m.Delete("a")
		if m.Len() != 2 {
			t.Errorf("Expected length 2 after deleting first element, got %d", m.Len())
		}
		if _, exists := m.Get("a"); exists {
			t.Error("Expected 'a' to be deleted")
		}
		if val, _ := m.GetAt(0); val != 2 {
			t.Errorf("Expected 'b' to be at index 0 after deleting 'a'")
		}

		// Delete last element
		m.Put("a", 1)
		m.Delete("c")
		if m.Len() != 2 {
			t.Errorf("Expected length 2 after deleting last element, got %d", m.Len())
		}
		if _, exists := m.Get("c"); exists {
			t.Error("Expected 'c' to be deleted")
		}

		// Delete middle element
		m.Put("c", 3)
		m.Delete("b")
		if m.Len() != 2 {
			t.Errorf("Expected length 2 after deleting middle element, got %d", m.Len())
		}
		if _, exists := m.Get("b"); exists {
			t.Error("Expected 'b' to be deleted")
		}
		if val, _ := m.GetAt(0); val != 1 {
			t.Errorf("Expected 'a' to be at index 0 after deleting 'b'")
		}
		if val, _ := m.GetAt(1); val != 3 {
			t.Errorf("Expected 'c' to be at index 1 after deleting 'b'")
		}

		// Delete all elements one by one using DeleteAt
		m = NewOrderedMap[string, int]()
		m.Put("a", 1)
		m.Put("b", 2)
		m.Put("c", 3)
		m.DeleteAt(0)
		m.DeleteAt(0)
		m.DeleteAt(0)
		if m.Len() != 0 {
			t.Errorf("Expected length 0 after deleting all elements, got %d", m.Len())
		}
	})

	t.Run("Enum Termination", func(t *testing.T) {
		m := NewOrderedMap[string, int]()
		m.Put("a", 1)
		m.Put("b", 2)
		m.Put("c", 3)

		// EnumKeys termination
		count := 0
		m.EnumKeys(func(k string) bool {
			count++
			if k == "b" {
				return false // Terminate early
			}
			return true
		})
		if count != 2 {
			t.Errorf("Expected EnumKeys to terminate after 'b', enumerated %d", count)
		}

		// EnumValues termination
		count = 0
		m.EnumValues(func(v int) bool {
			count++
			if v == 2 {
				return false // Terminate early
			}
			return true
		})
		if count != 2 {
			t.Errorf("Expected EnumValues to terminate after 2, enumerated %d", count)
		}
	})
}

func TestSyncMap(t *testing.T) {
	const numLoops = 1000
	const numRoutines = 4

	t.Run("Concurrent Access", func(t *testing.T) {
		var m = NewOrderedSyncMap[string, int]()
		var wg sync.WaitGroup
		wg.Add(numRoutines)

		for i := 0; i < numRoutines; i++ {
			go func(i int) {
				defer wg.Done()
				for j := 0; j < numLoops; j++ {
					key := strconv.Itoa(i*numLoops + j)
					val := i*numLoops + j

					// Concurrent Put
					m.Put(key, val)
					time.Sleep(time.Duration(rand.IntN(100)) * time.Microsecond)

					// Concurrent Get
					v, ok := m.Get(key)
					if !ok {
						t.Errorf("Routine %d: Key %s not found", i, key)
					} else if v != val {
						t.Errorf("Routine %d: Expected value %d for key %s, got %d", i, val, key, v)
					}
				}
			}(i)
		}
		wg.Wait()

		if m.Len() != numLoops*numRoutines {
			t.Errorf("Expected total count %d, got %d", numLoops*numRoutines, m.Len())
		}
	})

	t.Run("Concurrent Delete", func(t *testing.T) {
		var m = NewOrderedSyncMap[string, int]()
		var wg sync.WaitGroup
		wg.Add(numRoutines)

		// Populate the map first
		for i := 0; i < numLoops*numRoutines; i++ {
			m.Put(strconv.Itoa(i), i)
		}

		// Concurrent Delete routines
		for i := 0; i < numRoutines; i++ {
			go func(i int) {
				defer wg.Done()
				for j := 0; j < numLoops; j++ {
					key := strconv.Itoa(i*numLoops + j)

					// Concurrent Delete
					v, ok := m.Delete(key)
					if !ok {
						t.Errorf("Routine %d: Key %s not found for deletion", i, key)
					} else if v != i*numLoops+j {
						t.Errorf("Routine %d: Wrong value returned on deletion", i)
					}
					time.Sleep(time.Duration(rand.IntN(100)) * time.Microsecond)
				}
			}(i)
		}

		wg.Wait()
		if m.Len() != 0 {
			t.Errorf("Expected length to be zero after all deletions, got %d", m.Len())
		}
	})
}

func BenchmarkSyncMapPut(b *testing.B) {
	m := NewSyncMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Put(k[i], i)
	}
}

func BenchmarkSyncMapGet(b *testing.B) {
	m := NewSyncMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(k[i])
	}
}

func BenchmarkSyncMapDelete(b *testing.B) {
	m := NewSyncMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Delete(k[i])
	}
}

func BenchmarkSyncMapDeleteReverse(b *testing.B) {
	m := NewSyncMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := b.N - 1; i >= 0; i-- {
		m.Delete(k[i])
	}
}

func BenchmarkOrderedMapPut(b *testing.B) {
	m := NewOrderedMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Put(k[i], i)
	}
}

func BenchmarkOrderedMapGet(b *testing.B) {
	m := NewOrderedMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(k[i])
	}
}

func BenchmarkOrderedMapDelete(b *testing.B) {
	m := NewOrderedMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Delete(k[i])
	}
}

func BenchmarkOrderedMapDeleteReverse(b *testing.B) {
	m := NewOrderedMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := b.N - 1; i >= 0; i-- { // Corrected loop condition
		m.Delete(k[i])
	}
}

func BenchmarkOrderedMapDeleteAt(b *testing.B) {
	m := NewOrderedMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DeleteAt(0) // Always delete the first element
	}
}

func BenchmarkOrderedMapDeleteAtReverse(b *testing.B) {
	m := NewOrderedMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DeleteAt(m.Len() - 1) // Always delete the last element
	}
}

func BenchmarkOrderedSyncMapPut(b *testing.B) {
	m := NewOrderedSyncMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Put(k[i], i)
	}
}

func BenchmarkOrderedSyncMapGet(b *testing.B) {
	m := NewOrderedSyncMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(k[i])
	}
}

func BenchmarkOrderedSyncMapDelete(b *testing.B) {
	m := NewOrderedSyncMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Delete(k[i])
	}
}

func BenchmarkOrderedSyncMapDeleteReverse(b *testing.B) {
	m := NewOrderedSyncMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := b.N - 1; i >= 0; i-- {
		m.Delete(k[i])
	}
}

func BenchmarkOrderedSyncMapDeleteAt(b *testing.B) {
	m := NewOrderedSyncMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DeleteAt(0)
	}
}

func BenchmarkOrderedSyncMapDeleteAtReverse(b *testing.B) {
	m := NewOrderedSyncMap[string, int]()
	k := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		k[i] = strconv.Itoa(i)
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DeleteAt(m.Len() - 1)
	}
}
