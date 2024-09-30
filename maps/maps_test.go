package maps

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

func enum(m *OrderedMap[string, int]) {

	if !testing.Verbose() {
		return
	}

	m.EnumKeys(func(k string) bool {
		fmt.Printf("Key: %s\n", k)
		return true
	})

	m.EnumValues(func(k int) bool {
		fmt.Printf("Value: %d\n", k)
		return true
	})
}

func TestOrderedMap(t *testing.T) {

	m := MakeOrderedMap[string, int]()
	m.Put("a", 0)
	m.Put("b", 1)
	m.Put("c", 2)
	m.Put("d", 3)
	m.Put("e", 4)

	if m.Len() != 5 {
		t.Fatal("Len failed")
	}

	if !m.Exists("c") {
		t.Fatal("Exists failed")
	}

	if _, exists := m.Delete("foo"); exists {
		t.Fatal("Delete failed to return exists")
	}

	if _, exists := m.DeleteAt(69); exists {
		t.Fatal("DeleteAt failed to return exists")
	}

	if v, exists := m.Delete("b"); !exists {
		t.Fatal("Delete failed to find key")
	} else if v != 1 {
		t.Fatal("Delete failed to return old value")
	}

	enum(m)

	if v, exists := m.DeleteAt(2); !exists {
		t.Fatal("DeleteAt failed to find index")
	} else if v != 3 {
		t.Fatal("DeleteAt failed to return old value")
	}

	enum(m)

	if v, exists := m.Get("c"); !exists {
		t.Fatal("Get failed to return truth")
	} else if v != 2 {
		t.Fatal("Get failed to return value")
	}

	if v, existed := m.Put("c", 2); !existed {
		t.Fatal("Put failed to return truth")
	} else if v != 2 {
		t.Fatal("Put failed to return old value")
	}

	if v, found := m.GetAt(1); !found {
		t.Fatal("Index failed to find item")
	} else if v != 2 {
		t.Fatal("Index failed to return correct value")
	}

	if v, existed := m.Put("g", 6); existed {
		t.Fatal("Put failed to return truth")
	} else if v != 0 {
		t.Fatal("Put failed to return old value")
	}

	enum(m)

	if v, existed := m.Delete("g"); !existed {
		t.Fatal("DeleteAt failed to return existed")
	} else if v != 6 {
		t.Fatal("DeleteAt failed to return old value")
	}

	enum(m)

	if v, existed := m.Delete("e"); !existed {
		t.Fatal("Delete failed to return existed")
	} else if v != 4 {
		t.Fatal("Delete failed to return old value")
	}

	enum(m)

}

func BenchmarkOrderedMapPut(b *testing.B) {
	m := MakeOrderedMap[string, int]()
	k := make([]string, b.N)
	for i := range b.N {
		k[i] = strconv.Itoa(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Put(k[i], i)
	}
}

func BenchmarkOrderedMapGet(b *testing.B) {
	m := MakeOrderedMap[string, int]()
	k := make([]string, b.N)
	for i := range b.N {
		k[i] = strconv.Itoa(i)
	}
	for i := 0; i < b.N; i++ {
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(k[i])
	}
}

func BenchmarkOrderedMapDelete(b *testing.B) {
	m := MakeOrderedMap[string, int]()
	k := make([]string, b.N)
	for i := range b.N {
		k[i] = strconv.Itoa(i)
	}
	for i := 0; i < b.N; i++ {
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Delete(k[i])
	}
}

func BenchmarkOrderedMapDeleteReverse(b *testing.B) {
	m := MakeOrderedMap[string, int]()
	k := make([]string, b.N)
	for i := range b.N {
		k[i] = strconv.Itoa(i)
	}
	for i := 0; i < b.N; i++ {
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := b.N - 1; i > 0; i-- {
		m.Delete(k[i])
	}
}

func BenchmarkOrderedMapDeleteAt(b *testing.B) {
	m := MakeOrderedMap[string, int]()
	k := make([]string, b.N)
	for i := range b.N {
		k[i] = strconv.Itoa(i)
	}
	for i := 0; i < b.N; i++ {
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DeleteAt(i)
	}
}

func BenchmarkOrderedMapDeleteAtReverse(b *testing.B) {
	m := MakeOrderedMap[string, int]()
	k := make([]string, b.N)
	for i := range b.N {
		k[i] = strconv.Itoa(i)
	}
	for i := 0; i < b.N; i++ {
		m.Put(k[i], i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.DeleteAt(b.N - i)
	}
}

func TestSyncMap(t *testing.T) {
	const numLoops = 1000
	const numRoutines = 4
	var m = MakeOrderedSyncMap[string, int]()
	var wg sync.WaitGroup
	wg.Add(numRoutines)
	for i := range numRoutines {
		go func(i int) {
			for i := range numLoops + (i * numLoops) {
				m.Put(strconv.Itoa(i), i)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
	enum(m.m)

	for i := 0; i < numRoutines*numLoops; i++ {
		if v, exists := m.Get(strconv.Itoa(i)); !exists {
			t.Fatal("TestSyncMap failed")
		} else if v != i {
			t.Fatal("TestSyncMap failed")
		}
	}

}
