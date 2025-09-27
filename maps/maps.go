// Copyright 2025 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package ordered implements generic ordered map.
package maps

import (
	"sync"
)

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
