// Copyright 2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package maps

import (
	"sync"
)

// OrderedMap is a generic ordered map which supports comparable keys
// and any type of value.
type OrderedMap[K comparable, V any] struct {
	z        V
	indexMap map[K]int
	valueMap map[K]V
	keySlice []K
}

// MakeOrderedMap returns a new OrderedMap of key K and value V.
func MakeOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		*new(V),
		make(map[K]int),
		make(map[K]V),
		nil,
	}
}

// Len returns number of entries in the map.
func (self *OrderedMap[K, V]) Len() int { return len(self.valueMap) }

// Exists returns truth if an entry under key k exists.
func (self *OrderedMap[K, V]) Exists(k K) (b bool) {
	_, b = self.valueMap[k]
	return
}

// Get returns the entry value under key k and a truth if found.
// if not found a zero value of entry value under key k is rturned.
func (self *OrderedMap[K, V]) Get(k K) (v V, b bool) {
	v, b = self.valueMap[k]
	return 
}

// GetAt returns value at index i and a truth if found/index is within range.
func (self *OrderedMap[K, V]) GetAt(i int) (v V, b bool) {
	v, b = self.valueMap[self.keySlice[i]]
	return
}

// Put stores value v under key k and returns a value that was replaced and a
// truth if value existed under key k and was replaced.
func (self *OrderedMap[K, V]) Put(k K, v V) (old V, found bool) {
	if old, found = self.valueMap[k]; found {
		self.valueMap[k] = v
		return
	}
	self.keySlice = append(self.keySlice, k)
	self.indexMap[k] = len(self.keySlice) - 1
	self.valueMap[k] = v
	return self.z, false
}

// Delete deletes an entry by key and returns value that was at that key and
// truth if item was found and deleted.
func (self *OrderedMap[K, V]) Delete(k K) (value V, exists bool) {

	if value, exists = self.valueMap[k]; !exists {
		return self.z, false
	}

	var index = self.indexMap[k]
	var updateKeys = self.keySlice[index+1:]
	for _, key := range updateKeys {
		self.indexMap[key] = self.indexMap[key] - 1
	}
	delete(self.indexMap, k)
	delete(self.valueMap, k)
	self.keySlice = append(self.keySlice[:index], updateKeys...)
	return
}

// Delete deletes entry at index i and returns value that was at that index and
// truth if item was found and deleted.
func (self *OrderedMap[K, V]) DeleteAt(i int) (value V, exists bool) {

	if i >= len(self.keySlice) {
		return self.z, false
	}
	var key = self.keySlice[i]

	if value, exists = self.valueMap[key]; !exists {
		return self.z, false
	}

	var updateKeys = self.keySlice[i+1:]
	for _, key := range updateKeys {
		self.indexMap[key] = self.indexMap[key] - 1
	}
	delete(self.valueMap, key)
	delete(self.indexMap, key)
	self.keySlice = append(self.keySlice[:i], updateKeys...)

	return
}

// EnumKeys enumerates all keys in the map in order as added.
func (self *OrderedMap[K, V]) EnumKeys(f func(k K) bool) {
	for i := 0; i < len(self.keySlice); i++ {
		if !f(self.keySlice[i]) {
			break
		}
	}
}

// EnumValues enumerates all values in the map in order as added.
func (self *OrderedMap[K, V]) EnumValues(f func(v V) bool) {
	for i := 0; i < len(self.keySlice); i++ {
		if !f(self.valueMap[self.keySlice[i]]) {
			break
		}
	}
}

// Keys returns all keys.
func (self *OrderedMap[K, V]) Keys() (out []K) {
	out = make([]K, 0, len(self.keySlice))
	for _, k := range self.keySlice {
		out = append(out, k)
	}
	return
}

// Values returns all values.
func (self *OrderedMap[K, V]) Values() (out []V) {
	out = make([]V, 0, len(self.keySlice))
	for _, i := range self.keySlice {
		out = append(out, self.valueMap[i])
	}
	return
}

// OrderedSyncMap is a [OrderedMap] with a mutext that protects all operations.
type OrderedSyncMap[K comparable, V any] struct {
	mu sync.Mutex
	m  *OrderedMap[K, V]
}

// MakeOrderedSyncMap returns a new OrderedSyncMap of key K and value V.
func MakeOrderedSyncMap[K comparable, V any]() *OrderedSyncMap[K, V] {
	return &OrderedSyncMap[K, V]{
		m: MakeOrderedMap[K, V](),
	}
}

func (self *OrderedSyncMap[K, V]) Len() (l int) {
	self.mu.Lock()
	l = self.m.Len()
	self.mu.Unlock()
	return
}

func (self *OrderedSyncMap[K, V]) Exists(k K) (b bool) {
	self.mu.Lock()
	b = self.m.Exists(k)
	self.mu.Unlock()
	return
}

func (self *OrderedSyncMap[K, V]) Get(k K) (v V, b bool) {
	self.mu.Lock()
	v, b = self.m.Get(k)
	self.mu.Unlock()
	return
}

func (self *OrderedSyncMap[K, V]) GetAt(i int) (v V, b bool) {
	self.mu.Lock()
	v, b = self.m.GetAt(i)
	self.mu.Unlock()
	return
}

func (self *OrderedSyncMap[K, V]) Put(k K, v V) {
	self.mu.Lock()
	self.m.Put(k, v)
	self.mu.Unlock()
}

func (self *OrderedSyncMap[K, V]) DeleteAt(i int) (v V, b bool) {
	self.mu.Lock()
	v, b = self.m.DeleteAt(i)
	self.mu.Unlock()
	return
}

// EnumKeys enumerates all keys in the map in order as added.
func (self *OrderedSyncMap[K, V]) EnumKeys(f func(k K) bool) {
	self.mu.Lock()
	self.m.EnumKeys(f)
	self.mu.Unlock()
}

// EnumValues enumerates all values in the map in order as added.
func (self *OrderedSyncMap[K, V]) EnumValues(f func(v V) bool) {
	self.mu.Lock()
	self.m.EnumValues(f)
	self.mu.Unlock()
}

// Keys returns all keys.
func (self *OrderedSyncMap[K, V]) Keys() (out []K) {
	self.mu.Lock()
	out = self.m.Keys()
	self.mu.Unlock()
	return
}

// Values returns all values.
func (self *OrderedSyncMap[K, V]) Values() (out []V) {
	self.mu.Lock()
	out = self.m.Values()
	self.mu.Unlock()
	return
}
