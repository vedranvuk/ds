package bidi

import (
	"math/rand"
	"sync"
	"testing"
)

func TestBidiMap(t *testing.T) {
	m := New[int]()

	// Test Put and Len
	m.Put(1, 10)
	m.Put(2, 20)
	if m.Len() != 2 {
		t.Errorf("Len() should be 2, got %d", m.Len())
	}

	// Test KeyExists and ValExists
	if !m.KeyExists(1) {
		t.Errorf("KeyExists(1) should be true")
	}
	if !m.ValExists(20) {
		t.Errorf("ValExists(20) should be true")
	}
	if m.KeyExists(3) {
		t.Errorf("KeyExists(3) should be false")
	}
	if m.ValExists(30) {
		t.Errorf("ValExists(30) should be false")
	}

	// Test Val and Key
	val, ok := m.Val(1)
	if !ok || val != 10 {
		t.Errorf("Val(1) should be (10, true), got (%d, %t)", val, ok)
	}
	key, ok := m.Key(20)
	if !ok || key != 2 {
		t.Errorf("Key(20) should be (2, true), got (%d, %t)", key, ok)
	}

	// Test DeleteByKey
	deletedValue, exists := m.DeleteByKey(1)
	if !exists || deletedValue != 10 {
		t.Errorf("DeleteByKey(1) should be (10, true), got (%d, %t)", deletedValue, exists)
	}
	if m.Len() != 1 {
		t.Errorf("Len() should be 1, got %d", m.Len())
	}

	// Test DeleteByValue
	deletedKey, exists := m.DeleteByValue(20)
	if !exists || deletedKey != 2 {
		t.Errorf("DeleteByValue(20) should be (2, true), got (%d, %t)", deletedKey, exists)
	}
	if m.Len() != 0 {
		t.Errorf("Len() should be 0, got %d", m.Len())
	}

	// Test EnumKeys and EnumValues
	m.Put(1, 10)
	m.Put(2, 20)
	keys := []int{}
	m.EnumKeys(func(key int) bool {
		keys = append(keys, key)
		return true
	})
	if len(keys) != 2 {
		t.Errorf("EnumKeys should return 2 keys, got %d", len(keys))
	}
	values := []int{}
	m.EnumValues(func(value int) bool {
		values = append(values, value)
		return true
	})
	if len(values) != 2 {
		t.Errorf("EnumValues should return 2 values, got %d", len(values))
	}

	// Test Keys and Values
	ks := m.Keys()
	if len(ks) != 2 {
		t.Errorf("Keys should return 2 keys, got %d", len(ks))
	}
	vs := m.Values()
	if len(vs) != 2 {
		t.Errorf("Values should return 2 values, got %d", len(vs))
	}
}

func TestSyncBidiMap(t *testing.T) {
	m := NewSync[int]()

	// Test Put and Len
	m.Put(1, 10)
	m.Put(2, 20)
	if m.Len() != 2 {
		t.Errorf("Len() should be 2, got %d", m.Len())
	}

	// Test KeyExists and ValExists
	if !m.KeyExists(1) {
		t.Errorf("KeyExists(1) should be true")
	}
	if !m.ValExists(20) {
		t.Errorf("ValExists(20) should be true")
	}
	if m.KeyExists(3) {
		t.Errorf("KeyExists(3) should be false")
	}
	if m.ValExists(30) {
		t.Errorf("ValExists(30) should be false")
	}

	// Test Val and Key
	val, ok := m.Val(1)
	if !ok || val != 10 {
		t.Errorf("Val(1) should be (10, true), got (%d, %t)", val, ok)
	}
	key, ok := m.Key(20)
	if !ok || key != 2 {
		t.Errorf("Key(20) should be (2, true), got (%d, %t)", key, ok)
	}

	// Test DeleteByKey
	deletedValue, exists := m.DeleteByKey(1)
	if !exists || deletedValue != 10 {
		t.Errorf("DeleteByKey(1) should be (10, true), got (%d, %t)", deletedValue, exists)
	}
	if m.Len() != 1 {
		t.Errorf("Len() should be 1, got %d", m.Len())
	}

	// Test DeleteByValue
	deletedKey, exists := m.DeleteByValue(20)
	if !exists || deletedKey != 2 {
		t.Errorf("DeleteByValue(20) should be (2, true), got (%d, %t)", deletedKey, exists)
	}
	if m.Len() != 0 {
		t.Errorf("Len() should be 0, got %d", m.Len())
	}

	// Test EnumKeys and EnumValues
	m.Put(1, 10)
	m.Put(2, 20)
	keys := []int{}
	m.EnumKeys(func(key int) bool {
		keys = append(keys, key)
		return true
	})
	if len(keys) != 2 {
		t.Errorf("EnumKeys should return 2 keys, got %d", len(keys))
	}
	values := []int{}
	m.EnumValues(func(value int) bool {
		values = append(values, value)
		return true
	})
	if len(values) != 2 {
		t.Errorf("EnumValues should return 2 values, got %d", len(values))
	}

	// Test Keys and Values
	ks := m.Keys()
	if len(ks) != 2 {
		t.Errorf("Keys should return 2 keys, got %d", len(ks))
	}
	vs := m.Values()
	if len(vs) != 2 {
		t.Errorf("Values should return 2 values, got %d", len(vs))
	}
}

func BenchmarkBidiMapPut(b *testing.B) {
	m := New[int]()
	for i := 0; i < b.N; i++ {
		m.Put(i, i*10)
	}
}

func BenchmarkBidiMapGet(b *testing.B) {
	m := New[int]()
	for i := 0; i < 1000; i++ {
		m.Put(i, i*10)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Val(rand.Intn(1000))
	}
}

func BenchmarkSyncBidiMapPut(b *testing.B) {
	m := NewSync[int]()
	for i := 0; i < b.N; i++ {
		m.Put(i, i*10)
	}
}

func BenchmarkSyncBidiMapGet(b *testing.B) {
	m := NewSync[int]()
	for i := 0; i < 1000; i++ {
		m.Put(i, i*10)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Val(rand.Intn(1000))
	}
}

func BenchmarkSyncBidiMapParallelPut(b *testing.B) {
	m := NewSync[int]()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Put(i, i*10)
			i++
		}
	})
}

func BenchmarkSyncBidiMapParallelGet(b *testing.B) {
	m := NewSync[int]()
	for i := 0; i < 1000; i++ {
		m.Put(i, i*10)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Val(rand.Intn(1000))
		}
	})
}

func TestSyncBidiMapConcurrency(t *testing.T) {
	m := NewSync[int]()
	var wg sync.WaitGroup

	// Number of concurrent operations
	numOps := 100

	// Number of goroutines for puts and gets
	numPutters := 10
	numGetters := 10

	wg.Add(numPutters + numGetters)

	// Concurrent Putters
	for i := 0; i < numPutters; i++ {
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := routineID*numOps + j
				m.Put(key, key*10)
			}
		}(i)
	}

	// Concurrent Getters
	for i := 0; i < numGetters; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := rand.Intn(numPutters * numOps) // Get a random key that might exist
				m.Val(key)                           // Just get, don't assert here to avoid race conditions in tests
			}
		}()
	}

	wg.Wait()

	// Optional: Verify the final state of the map (add assertions here if needed)
	expectedSize := numPutters * numOps
	if m.Len() != expectedSize {
		t.Errorf("Expected map size %d, got %d", expectedSize, m.Len())
	}
}