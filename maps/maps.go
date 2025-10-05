// Copyright 2025 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package ordered implements generic ordered map.
package maps

import (
	"sync"
)

// SyncMap is a generic thread-safe map which supports comparable keys
// and any type of value. It wraps a standard Go map with a mutex for
// concurrent access. Unlike OrderedMap, it does not maintain insertion order.
//
// Example:
//
//	m := NewSyncMap[string, int]()<
//	m.Put("one", 1)
//	m.Put("two", 2)
//	m.Put("three", 3)
//	fmt.Println(m.Len())   // Output: 3
//	v, ok := m.Get("two")
//	fmt.Println(v, ok)     // Output: 2 true
type SyncMap[K comparable, V any] struct {
	mu sync.Mutex // mutex protecting the map
	z  V          // zero value of type V, used for return when not found
	m  map[K]V    // the underlying map
}

// NewSyncMap returns a new SyncMap of key K and value V.
//
// Example:
//
//	m := NewSyncMap[string, int]()
func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{
		z: *new(V),
		m: make(map[K]V),
	}
}

// Len returns the number of elements in the map.
//
// Example:
//
//	m := NewSyncMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	fmt.Println(m.Len()) // Output: 2
func (self *SyncMap[K, V]) Len() (length int) {
	self.mu.Lock()
	length = len(self.m)
	self.mu.Unlock()
	return
}

// Exists returns true if the given key is present in the map.
//
// Example:
//
//	m := NewSyncMap[string, int]()
//	m.Put("one", 1)
//	fmt.Println(m.Exists("one"))   // Output: true
//	fmt.Println(m.Exists("three")) // Output: false
func (self *SyncMap[K, V]) Exists(key K) (exists bool) {
	self.mu.Lock()
	_, exists = self.m[key]
	self.mu.Unlock()
	return
}

// Get returns the value associated with the given key and a boolean indicating
// if the key was found. If the key is not found, it returns the zero value
// of the value type.
//
// Example:
//
//	m := NewSyncMap[string, int]()
//	m.Put("one", 1)
//	v, b := m.Get("one")
//	fmt.Println(v, b)     // Output: 1 true
//	v, b = m.Get("three")
//	fmt.Println(v, b)     // Output: 0 false
func (self *SyncMap[K, V]) Get(key K) (value V, exists bool) {
	self.mu.Lock()
	value, exists = self.m[key]
	self.mu.Unlock()
	return
}

// Put inserts or updates a key-value pair in the map. It returns the old value
// (if any) and a boolean indicating if the value was replaced.
//
// Example:
//
//	m := NewSyncMap[string, int]()
//	old, found := m.Put("one", 1)
//	fmt.Println(old, found) // Output: 0 false
//	old, found = m.Put("one", 11)
//	fmt.Println(old, found) // Output: 1 true
func (self *SyncMap[K, V]) Put(key K, value V) (oldValue V, replaced bool) {
	self.mu.Lock()
	oldValue, replaced = self.m[key]
	self.m[key] = value
	self.mu.Unlock()
	if !replaced {
		return self.z, false
	}
	return
}

// Delete removes a key-value pair from the map. It returns the deleted value
// and a boolean indicating if the key was found.
//
// Example:
//
//	m := NewSyncMap[string, int]()
//	m.Put("one", 1)
//	v, b := m.Delete("one")
//	fmt.Println(v, b)     // Output: 1 true
//	v, b = m.Delete("one")
//	fmt.Println(v, b)     // Output: 0 false
func (self *SyncMap[K, V]) Delete(key K) (value V, exists bool) {
	self.mu.Lock()
	value, exists = self.m[key]
	if exists {
		delete(self.m, key)
	}
	self.mu.Unlock()
	if !exists {
		return self.z, false
	}
	return
}

// Keys returns a slice containing all the keys in the map.
// Note: The order of keys is not guaranteed.
//
// Example:
//
//	m := NewSyncMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	keys := m.Keys()
//	fmt.Println(len(keys)) // Output: 2
func (self *SyncMap[K, V]) Keys() (keys []K) {
	self.mu.Lock()
	keys = make([]K, 0, len(self.m))
	for k := range self.m {
		keys = append(keys, k)
	}
	self.mu.Unlock()
	return
}

// Values returns a slice containing all the values in the map.
// Note: The order of values is not guaranteed.
//
// Example:
//
//	m := NewSyncMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	values := m.Values()
//	fmt.Println(len(values)) // Output: 2
func (self *SyncMap[K, V]) Values() (values []V) {
	self.mu.Lock()
	values = make([]V, 0, len(self.m))
	for _, v := range self.m {
		values = append(values, v)
	}
	self.mu.Unlock()
	return
}

// Clear removes all key-value pairs from the map.
//
// Example:
//
//	m := NewSyncMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	m.Clear()
//	fmt.Println(m.Len()) // Output: 0
func (self *SyncMap[K, V]) Clear() {
	self.mu.Lock()
	self.m = make(map[K]V)
	self.mu.Unlock()
}

// OrderedMap is a generic ordered map which supports comparable keys
// and any type of value.  It maintains the order in which keys were
// inserted.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	m.Put("three", 3)
//	fmt.Println(m.Keys())   // Output: [one two three]
//	fmt.Println(m.Values()) // Output: [1 2 3]
type OrderedMap[K comparable, V any] struct {
	z        V              // zero value of type V, used for return when not found
	indexMap map[K]int      // key -> index lookup
	valueMap map[K]V      // key -> value lookup
	keySlice []K              // maintains insertion order
}

// NewOrderedMap returns a new OrderedMap of key K and value V.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		z:        *new(V),
		indexMap: make(map[K]int),
		valueMap: make(map[K]V),
		keySlice: nil,
	}
}

// Len returns the number of elements in the map.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	fmt.Println(m.Len()) // Output: 2
func (self *OrderedMap[K, V]) Len() (length int) {
	length = len(self.valueMap)
	return
}

// Exists returns true if the given key is present in the map.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
//	m.Put("one", 1)
//	fmt.Println(m.Exists("one"))   // Output: true
//	fmt.Println(m.Exists("three")) // Output: false
func (self *OrderedMap[K, V]) Exists(key K) (exists bool) {
	_, exists = self.valueMap[key]
	return
}

// Get returns the value associated with the given key and a boolean indicating
// if the key was found. If the key is not found, it returns the zero value
// of the value type.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
//	m.Put("one", 1)
//	v, b := m.Get("one")
//	fmt.Println(v, b)     // Output: 1 true
//	v, b = m.Get("three")
//	fmt.Println(v, b)     // Output: 0 false
func (self *OrderedMap[K, V]) Get(key K) (value V, exists bool) {
	value, exists = self.valueMap[key]
	return
}

// GetAt returns the value at the given index and a boolean indicating if the
// index is valid. If the index is out of bounds, it returns the zero value
// of the value type.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	v, b := m.GetAt(0)
//	fmt.Println(v, b)     // Output: 1 true
//	v, b = m.GetAt(2)
//	fmt.Println(v, b)     // Output: 0 false
func (self *OrderedMap[K, V]) GetAt(index int) (value V, exists bool) {
	if index < 0 || index >= len(self.keySlice) {
		return self.z, false
	}
	value, exists = self.valueMap[self.keySlice[index]]
	return
}

// Put inserts or updates a key-value pair in the map. It returns the old value
// (if any) and a boolean indicating if the value was replaced.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
//	old, found := m.Put("one", 1)
//	fmt.Println(old, found) // Output: 0 false
//	old, found = m.Put("one", 11)
//	fmt.Println(old, found) // Output: 1 true
func (self *OrderedMap[K, V]) Put(key K, value V) (oldValue V, replaced bool) {
	if oldValue, replaced = self.valueMap[key]; replaced {
		self.valueMap[key] = value
		return
	}

	self.keySlice = append(self.keySlice, key)
	self.indexMap[key] = len(self.keySlice) - 1
	self.valueMap[key] = value
	return self.z, false
}

// Delete removes a key-value pair from the map. It returns the deleted value
// and a boolean indicating if the key was found.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
//	m.Put("one", 1)
//	v, b := m.Delete("one")
//	fmt.Println(v, b)     // Output: 1 true
//	v, b = m.Delete("one")
//	fmt.Println(v, b)     // Output: 0 false
func (self *OrderedMap[K, V]) Delete(key K) (value V, exists bool) {
	if value, exists = self.valueMap[key]; !exists {
		return self.z, false
	}

	var index = self.indexMap[key]
	var updateKeys = self.keySlice[index+1:]

	for _, k := range updateKeys {
		self.indexMap[k]--
	}

	delete(self.indexMap, key)
	delete(self.valueMap, key)
	self.keySlice = append(self.keySlice[:index], updateKeys...)

	return
}

// DeleteAt removes the key-value pair at the given index. It returns the
// deleted value and a boolean indicating if the index was valid.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	v, b := m.DeleteAt(0)
//	fmt.Println(v, b)     // Output: 1 true
//	v, b = m.DeleteAt(2)
//	fmt.Println(v, b)     // Output: 0 false
func (self *OrderedMap[K, V]) DeleteAt(index int) (value V, exists bool) {
	if index < 0 || index >= len(self.keySlice) {
		return self.z, false
	}

	var key = self.keySlice[index]
	if value, exists = self.valueMap[key]; !exists {
		return self.z, false
	}

	var updateKeys = self.keySlice[index+1:]
	for _, k := range updateKeys {
		self.indexMap[k]--
	}

	delete(self.valueMap, key)
	delete(self.indexMap, key)
	self.keySlice = append(self.keySlice[:index], updateKeys...)

	return
}

// EnumKeys iterates over the keys in the map in insertion order, calling the
// provided function for each key.  Iteration stops if the function returns false.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	m.EnumKeys(func(k string) bool {
//		fmt.Println(k)
//		return true
//	})
//
// Output:
//
//	one
//	two
func (self *OrderedMap[K, V]) EnumKeys(callback func(key K) (cont bool)) {
	for _, key := range self.keySlice {
		if !callback(key) {
			break
		}
	}
}

// EnumValues iterates over the values in the map in insertion order, calling
// the provided function for each value. Iteration stops if the function
// returns false.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	m.EnumValues(func(v int) bool {
//		fmt.Println(v)
//		return true
//	})
//
// Output:
//
//	1
//	2
func (self *OrderedMap[K, V]) EnumValues(callback func(value V) (cont bool)) {
	for _, key := range self.keySlice {
		if !callback(self.valueMap[key]) {
			break
		}
	}
}

// Keys returns a slice containing all the keys in the map, in insertion order.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	fmt.Println(m.Keys()) // Output: [one two]
func (self *OrderedMap[K, V]) Keys() (keys []K) {
	keys = make([]K, len(self.keySlice))
	copy(keys, self.keySlice)
	return
}

// Values returns a slice containing all the values in the map, in insertion order.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	fmt.Println(m.Values()) // Output: [1 2]
func (self *OrderedMap[K, V]) Values() (values []V) {
	values = make([]V, len(self.keySlice))
	for i, key := range self.keySlice {
		values[i] = self.valueMap[key]
	}
	return
}

// Clear removes all key-value pairs from the map.
//
// Example:
//
//	m := NewOrderedMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	m.Clear()
//	fmt.Println(m.Len()) // Output: 0
func (self *OrderedMap[K, V]) Clear() {
	self.indexMap = make(map[K]int)
	self.valueMap = make(map[K]V)
	self.keySlice = nil
}

// OrderedSyncMap is a thread-safe version of [OrderedMap] using a mutex to protect concurrent access.
type OrderedSyncMap[K comparable, V any] struct {
	mu sync.Mutex // mutex protecting the map
	m  *OrderedMap[K, V] // the underlying ordered map
}

// NewOrderedSyncMap returns a new, empty, thread-safe OrderedSyncMap.
//
// Example:
//
//	m := NewOrderedSyncMap[string, int]()
func NewOrderedSyncMap[K comparable, V any]() *OrderedSyncMap[K, V] {
	return &OrderedSyncMap[K, V]{
		m: NewOrderedMap[K, V](),
	}
}

// Len returns the number of elements in the map.
//
// Example:
//
//	m := NewOrderedSyncMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	fmt.Println(m.Len()) // Output: 2
func (self *OrderedSyncMap[K, V]) Len() (length int) {
	self.mu.Lock()
	length = self.m.Len()
	self.mu.Unlock()
	return
}

// Exists returns true if the given key is present in the map.
//
// Example:
//
//	m := NewOrderedSyncMap[string, int]()
//	m.Put("one", 1)
//	fmt.Println(m.Exists("one"))   // Output: true
//	fmt.Println(m.Exists("three")) // Output: false
func (self *OrderedSyncMap[K, V]) Exists(key K) (exists bool) {
	self.mu.Lock()
	exists = self.m.Exists(key)
	self.mu.Unlock()
	return
}

// Get returns the value associated with the given key and a boolean indicating
// if the key was found. If the key is not found, it returns the zero value
// of the value type.
//
// Example:
//
//	m := NewOrderedSyncMap[string, int]()
//	m.Put("one", 1)
//	v, b := m.Get("one")
//	fmt.Println(v, b)     // Output: 1 true
//	v, b = m.Get("three")
//	fmt.Println(v, b)     // Output: 0 false
func (self *OrderedSyncMap[K, V]) Get(key K) (value V, exists bool) {
	self.mu.Lock()
	value, exists = self.m.Get(key)
	self.mu.Unlock()
	return
}

// GetAt returns the value at the given index and a boolean indicating if the
// index is valid. If the index is out of bounds, it returns the zero value
// of the value type.
//
// Example:
//
//	m := NewOrderedSyncMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	v, b := m.GetAt(0)
//	fmt.Println(v, b)     // Output: 1 true
//	v, b = m.GetAt(2)
//	fmt.Println(v, b)     // Output: 0 false
func (self *OrderedSyncMap[K, V]) GetAt(index int) (value V, exists bool) {
	self.mu.Lock()
	value, exists = self.m.GetAt(index)
	self.mu.Unlock()
	return
}

// Put inserts or updates a key-value pair in the map.
// It overwrites the existing value, if any.
//
// Example:
//
//	m := NewOrderedSyncMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	m.Put("one", 11)
//	fmt.Println(m.Get("one")) // Output: 11 true
func (self *OrderedSyncMap[K, V]) Put(key K, value V) {
	self.mu.Lock()
	self.m.Put(key, value)
	self.mu.Unlock()
}

// Delete removes a key-value pair from the map. It returns the deleted value
// and a boolean indicating if the key was found.
//
// Example:
//
//	m := NewOrderedSyncMap[string, int]()
//	m.Put("one", 1)
//	v, b := m.Delete("one")
//	fmt.Println(v, b)     // Output: 1 true
//	v, b = m.Delete("one")
//	fmt.Println(v, b)     // Output: 0 false
func (self *OrderedSyncMap[K, V]) Delete(key K) (value V, exists bool) {
	self.mu.Lock()
	value, exists = self.m.Delete(key)
	self.mu.Unlock()
	return
}

// DeleteAt removes the key-value pair at the given index. It returns the
// deleted value and a boolean indicating if the index was valid.
//
// Example:
//
//	m := NewOrderedSyncMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	v, b := m.DeleteAt(0)
//	fmt.Println(v, b)     // Output: 1 true
//	v, b = m.DeleteAt(2)
//	fmt.Println(v, b)     // Output: 0 false
func (self *OrderedSyncMap[K, V]) DeleteAt(index int) (value V, exists bool) {
	self.mu.Lock()
	value, exists = self.m.DeleteAt(index)
	self.mu.Unlock()
	return
}

// EnumKeys iterates over the keys in the map in insertion order, calling the
// provided function for each key.  Iteration stops if the function returns false.
//
// Example:
//
//	m := NewOrderedSyncMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	m.EnumKeys(func(k string) bool {
//		fmt.Println(k)
//		return true
//	})
//
// Output:
//
//	one
//	two
func (self *OrderedSyncMap[K, V]) EnumKeys(callback func(key K) (cont bool)) {
	self.mu.Lock()
	self.m.EnumKeys(callback)
	self.mu.Unlock()
}

// EnumValues iterates over the values in the map in insertion order, calling
// the provided function for each value. Iteration stops if the function
// returns false.
//
// Example:
//
//	m := NewOrderedSyncMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	m.EnumValues(func(v int) bool {
//		fmt.Println(v)
//		return true
//	})
//
// Output:
//
//	1
//	2
func (self *OrderedSyncMap[K, V]) EnumValues(callback func(value V) (cont bool)) {
	self.mu.Lock()
	self.m.EnumValues(callback)
	self.mu.Unlock()
}

// Keys returns a slice containing all the keys in the map, in insertion order.
//
// Example:
//
//	m := NewOrderedSyncMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	fmt.Println(m.Keys()) // Output: [one two]
func (self *OrderedSyncMap[K, V]) Keys() (keys []K) {
	self.mu.Lock()
	keys = self.m.Keys()
	self.mu.Unlock()
	return
}

// Values returns a slice containing all the values in the map, in insertion order.
//
// Example:
//
//	m := NewOrderedSyncMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	fmt.Println(m.Values()) // Output: [1 2]
func (self *OrderedSyncMap[K, V]) Values() (values []V) {
	self.mu.Lock()
	values = self.m.Values()
	self.mu.Unlock()
	return
}

// Clear removes all key-value pairs from the map.
//
// Example:
//
//	m := NewOrderedSyncMap[string, int]()
//	m.Put("one", 1)
//	m.Put("two", 2)
//	m.Clear()
//	fmt.Println(m.Len()) // Output: 0
func (self *OrderedSyncMap[K, V]) Clear() {
	self.mu.Lock()
	self.m.Clear()
	self.mu.Unlock()
}
