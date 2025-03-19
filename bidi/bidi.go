// Copyright 2025 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package bidi provides generic bidirectional one-to-one maps.
package bidi

import "sync"

// Map is an unordered generic bidirectional one-to-one map.
//
// It allows you to look up values by key and keys by value.
// The zero value of the type is not usable, use [New] to create a new map.
//
// Example:
//
//	m := New[string]()
//	m.Put("key1", "value1")
//	val, ok := m.Val("key1") // val == "value1", ok == true
//	key, ok := m.Key("value1") // key == "key1", ok == true
type Map[K comparable] struct {
	z        K
	keyToVal map[K]K
	valToKey map[K]K
}

// New returns a new [Map].
//
// Example:
//
//	m := New[string]()
//	m.Put("key1", "value1")
func New[K comparable]() *Map[K] {
	return &Map[K]{
		z:        *new(K),
		keyToVal: make(map[K]K),
		valToKey: make(map[K]K),
	}
}

// Len returns number of pairs in the map.
//
// Example:
//
//	m := New[string]()
//	m.Put("key1", "value1")
//	m.Put("key2", "value2")
//	l := m.Len() // l == 2
func (self *Map[K]) Len() int { return len(self.keyToVal) }

// KeyExists returns truth if key exists.
//
// Arguments:
//   - key: The key to check for existence.
//
// Returns:
//   - exists: True if the key exists in the map, false otherwise.
//
// Example:
//
//	m := New[string]()
//	m.Put("key1", "value1")
//	exists := m.KeyExists("key1") // exists == true
//	exists := m.KeyExists("key2") // exists == false
func (self *Map[K]) KeyExists(key K) (exists bool) {
	_, exists = self.keyToVal[key]
	return
}

// ValExists returns truth if value exists.
//
// Arguments:
//   - value: The value to check for existence.
//
// Returns:
//   - exists: True if the value exists in the map, false otherwise.
//
// Example:
//
//	m := New[string]()
//	m.Put("key1", "value1")
//	exists := m.ValExists("value1") // exists == true
//	exists := m.ValExists("value2") // exists == false
func (self *Map[K]) ValExists(value K) (exists bool) {
	_, exists = self.valToKey[value]
	return
}

// Key returns the key of the value.
//
// Arguments:
//   - value: The value to look up.
//
// Returns:
//   - key: The key associated with the value.  If the value is not found,
//     the zero value of K is returned.
//   - b: True if the value was found, false otherwise.
//
// Example:
//
//	m := New[string]()
//	m.Put("key1", "value1")
//	key, ok := m.Key("value1") // key == "key1", ok == true
//	key, ok := m.Key("value2") // key == "", ok == false
func (self *Map[K]) Key(value K) (key K, b bool) {
	key, b = self.valToKey[value]
	return
}

// Val returns the value of the key.
//
// Arguments:
//   - key: The key to look up.
//
// Returns:
//   - value: The value associated with the key. If the key is not found, the
//     zero value of K is returned.
//   - b: True if the key was found, false otherwise.
//
// Example:
//
//	m := New[string]()
//	m.Put("key1", "value1")
//	value, ok := m.Val("key1") // value == "value1", ok == true
//	value, ok := m.Val("key2") // value == "", ok == false
func (self *Map[K]) Val(key K) (value K, b bool) {
	value, b = self.keyToVal[key]
	return
}

// Put stores value under key and returns oldValue that was replaced and a
// truth if value existed under key k and was replaced.
//
// Arguments:
//   - key: The key to store the value under.
//   - value: The value to store.
//
// Returns:
//   - oldValue: The value that was previously stored under the key, or the
//     zero value of K if no value was previously stored.
//   - found: True if a value was previously stored under the key and was
//     replaced, false otherwise.
//
// Example:
//
//	m := New[string]()
//	oldValue, found := m.Put("key1", "value1") // oldValue == "", found == false
//	oldValue, found = m.Put("key1", "value2") // oldValue == "value1", found == true
func (self *Map[K]) Put(key K, value K) (oldValue K, found bool) {
	if oldValue, found = self.keyToVal[key]; found {
		delete(self.valToKey, oldValue)
	}
	self.keyToVal[key] = value
	self.valToKey[value] = key
	return self.z, false
}

// DeleteByKey deletes an entry by key and returns value that was bound to that
// key and truth if item was found and deleted.
//
// Arguments:
//   - key: The key of the entry to delete.
//
// Returns:
//   - deletedValue: The value that was stored under the key, or the zero
//     value of K if the key was not found.
//   - exists: True if the key was found and the entry was deleted, false
//     otherwise.
//
// Example:
//
//	m := New[string]()
//	m.Put("key1", "value1")
//	deletedValue, exists := m.DeleteByKey("key1") // deletedValue == "value1", exists == true
//	deletedValue, exists = m.DeleteByKey("key2") // deletedValue == "", exists == false
func (self *Map[K]) DeleteByKey(key K) (deletedValue K, exists bool) {
	if deletedValue, exists = self.keyToVal[key]; !exists {
		return self.z, false
	}
	delete(self.valToKey, deletedValue)
	delete(self.keyToVal, key)
	return
}

// DeleteByValue deletes an entry by value and returns key that was bound to
// that value and truth if item was found and deleted.
//
// Arguments:
//   - value: The value of the entry to delete.
//
// Returns:
//   - deletedKey: The key that was associated with the value, or the zero
//     value of K if the value was not found.
//   - exists: True if the value was found and the entry was deleted, false
//     otherwise.
//
// Example:
//
//	m := New[string]()
//	m.Put("key1", "value1")
//	deletedKey, exists := m.DeleteByValue("value1") // deletedKey == "key1", exists == true
//	deletedKey, exists = m.DeleteByValue("value2") // deletedKey == "", exists == false
func (self *Map[K]) DeleteByValue(value K) (deletedKey K, exists bool) {
	if deletedKey, exists = self.valToKey[value]; !exists {
		return self.z, false
	}
	delete(self.keyToVal, deletedKey)
	delete(self.valToKey, value)
	return
}

// EnumKeys enumerates all keys in the map.
//
// Arguments:
//   - f: A function that will be called for each key in the map.  The
//     function should return true to continue enumeration, or false to stop.
//
// Example:
//
//	m := New[string]()
//	m.Put("key1", "value1")
//	m.Put("key2", "value2")
//	var keys []string
//	m.EnumKeys(func(key string) bool {
//		keys = append(keys, key)
//		return true
//	})
func (self *Map[K]) EnumKeys(f func(key K) bool) {
	for k := range self.keyToVal {
		if !f(k) {
			break
		}
	}
}

// EnumValues enumerates all values in the map.
//
// Arguments:
//   - f: A function that will be called for each value in the map.  The
//     function should return true to continue enumeration, or false to stop.
//
// Example:
//
//	m := New[string]()
//	m.Put("key1", "value1")
//	m.Put("key2", "value2")
//	var values []string
//	m.EnumValues(func(value string) bool {
//		values = append(values, value)
//		return true
//	})
func (self *Map[K]) EnumValues(f func(value K) bool) {
	for v := range self.valToKey {
		if !f(v) {
			break
		}
	}
}

// Keys returns all keys.
//
// Returns:
//   - out: A slice containing all the keys in the map.  The order of the keys
//     is not guaranteed.
//
// Example:
//
//	m := New[string]()
//	m.Put("key1", "value1")
//	m.Put("key2", "value2")
//	keys := m.Keys() // keys == []string{"value1", "value2"} (order not guaranteed)
func (self *Map[K]) Keys() (out []K) {
	out = make([]K, 0, len(self.keyToVal))
	for k := range self.keyToVal {
		out = append(out, k)
	}
	return
}

// Values returns all values.
//
// Returns:
//   - out: A slice containing all the values in the map.  The order of the
//     values is not guaranteed.
//
// Example:
//
//	m := New[string]()
//	m.Put("key1", "value1")
//	m.Put("key2", "value2")
//	values := m.Values() // values == []string{"key1", "key2"} (order not guaranteed)
func (self *Map[K]) Values() (out []K) {
	out = make([]K, 0, len(self.valToKey))
	for v := range self.valToKey {
		out = append(out, v)
	}
	return
}

// SyncMap is the concurrency safe version of [Map].
//
// It provides methods for concurrent access to the map.
//
// Example:
//
//	m := NewSync[string]()
//	go m.Put("key1", "value1")
//	go m.Val("key1")
type SyncMap[K comparable] struct {
	mu sync.Mutex
	m  *Map[K]
}

// NewSync returns a new [SyncMap].
//
// Example:
//
//	m := NewSync[string]()
//	m.Put("key1", "value1")
func NewSync[K comparable]() *SyncMap[K] {
	return &SyncMap[K]{
		m: New[K](),
	}
}

// Len returns number of entries in the map.
//
// Example:
//
//	m := NewSync[string]()
//	m.Put("key1", "value1")
//	m.Put("key2", "value2")
//	l := m.Len() // l == 2
func (self *SyncMap[K]) Len() (out int) {
	self.mu.Lock()
	out = self.m.Len()
	self.mu.Unlock()
	return
}

// KeyExists returns truth if an entry under key k exists.
//
// Arguments:
//   - k: The key to check for existence.
//
// Returns:
//   - b: True if the key exists in the map, false otherwise.
//
// Example:
//
//	m := NewSync[string]()
//	m.Put("key1", "value1")
//	exists := m.KeyExists("key1") // exists == true
//	exists := m.KeyExists("key2") // exists == false
func (self *SyncMap[K]) KeyExists(k K) (b bool) {
	self.mu.Lock()
	b = self.m.KeyExists(k)
	self.mu.Unlock()
	return
}

// ValExists returns truth if an entry under value k exists.
//
// Arguments:
//   - k: The value to check for existence.
//
// Returns:
//   - b: True if the value exists in the map, false otherwise.
//
// Example:
//
//	m := NewSync[string]()
//	m.Put("key1", "value1")
//	exists := m.ValExists("value1") // exists == true
//	exists := m.ValExists("value2") // exists == false
func (self *SyncMap[K]) ValExists(k K) (b bool) {
	self.mu.Lock()
	b = self.m.ValExists(k)
	self.mu.Unlock()
	return
}

// Key returns the entry key under value and a truth if found.
// if not found a zero value of entry key under value is returned.
//
// Arguments:
//   - value: The value to look up.
//
// Returns:
//   - k: The key associated with the value.  If the value is not found,
//     the zero value of K is returned.
//   - b: True if the value was found, false otherwise.
//
// Example:
//
//	m := NewSync[string]()
//	m.Put("key1", "value1")
//	key, ok := m.Key("value1") // key == "key1", ok == true
//	key, ok := m.Key("value2") // key == "", ok == false
func (self *SyncMap[K]) Key(value K) (k K, b bool) {
	self.mu.Lock()
	k, b = self.m.Key(value)
	self.mu.Unlock()
	return
}

// Val returns the entry value under key k and a truth if found.
// if not found a zero value of entry value under key k is returned.
//
// Arguments:
//   - key: The key to look up.
//
// Returns:
//   - v: The value associated with the key. If the key is not found, the
//     zero value of K is returned.
//   - b: True if the key was found, false otherwise.
//
// Example:
//
//	m := NewSync[string]()
//	m.Put("key1", "value1")
//	value, ok := m.Val("key1") // value == "value1", ok == true
//	value, ok := m.Val("key2") // value == "", ok == false
func (self *SyncMap[K]) Val(key K) (v K, b bool) {
	self.mu.Lock()
	v, b = self.m.Val(key)
	self.mu.Unlock()
	return
}

// Put stores value v under key k and returns a value that was replaced and a
// truth if value existed under key k and was replaced.
//
// Arguments:
//   - k: The key to store the value under.
//   - v: The value to store.
//
// Returns:
//   - old: The value that was previously stored under the key, or the
//     zero value of K if no value was previously stored.
//   - found: True if a value was previously stored under the key and was
//     replaced, false otherwise.
//
// Example:
//
//	m := NewSync[string]()
//	oldValue, found := m.Put("key1", "value1") // oldValue == "", found == false
//	oldValue, found = m.Put("key1", "value2") // oldValue == "value1", found == true
func (self *SyncMap[K]) Put(k K, v K) (old K, found bool) {
	self.mu.Lock()
	old, found = self.m.Put(k, v)
	self.mu.Unlock()
	return
}

// DeleteByKey deletes an entry by key and returns value that was bound to that
// key and truth if item was found and deleted.
//
// Arguments:
//   - key: The key of the entry to delete.
//
// Returns:
//   - deletedValue: The value that was stored under the key, or the zero
//     value of K if the key was not found.
//   - exists: True if the key was found and the entry was deleted, false
//     otherwise.
//
// Example:
//
//	m := NewSync[string]()
//	m.Put("key1", "value1")
//	deletedValue, exists := m.DeleteByKey("key1") // deletedValue == "value1", exists == true
//	deletedValue, exists = m.DeleteByKey("key2") // deletedValue == "", exists == false
func (self *SyncMap[K]) DeleteByKey(key K) (deletedValue K, exists bool) {
	self.mu.Lock()
	deletedValue, exists = self.m.DeleteByKey(key)
	self.mu.Unlock()
	return
}

// DeleteByValue deletes an entry by value and returns key that was bound to
// that value and truth if item was found and deleted.
//
// Arguments:
//   - value: The value of the entry to delete.
//
// Returns:
//   - deletedKey: The key that was associated with the value, or the zero
//     value of K if the value was not found.
//   - exists: True if the value was found and the entry was deleted, false
//     otherwise.
//
// Example:
//
//	m := NewSync[string]()
//	m.Put("key1", "value1")
//	deletedKey, exists := m.DeleteByValue("value1") // deletedKey == "key1", exists == true
//	deletedKey, exists = m.DeleteByValue("value2") // deletedKey == "", exists == false
func (self *SyncMap[K]) DeleteByValue(value K) (deletedKey K, exists bool) {
	self.mu.Lock()
	deletedKey, exists = self.m.DeleteByValue(value)
	self.mu.Unlock()
	return
}

// EnumKeys enumerates all keys in the map in order as added.
//
// Arguments:
//   - f: A function that will be called for each key in the map.  The
//     function should return true to continue enumeration, or false to stop.
//
// Example:
//
//	m := NewSync[string]()
//	m.Put("key1", "value1")
//	m.Put("key2", "value2")
//	var keys []string
//	m.EnumKeys(func(key string) bool {
//		keys = append(keys, key)
//		return true
//	})
func (self *SyncMap[K]) EnumKeys(f func(key K) bool) {
	self.mu.Lock()
	self.m.EnumKeys(f)
	self.mu.Unlock()
	return
}

// EnumValues enumerates all values in the map in order as added.
//
// Arguments:
//   - f: A function that will be called for each value in the map.  The
//     function should return true to continue enumeration, or false to stop.
//
// Example:
//
//	m := NewSync[string]()
//	m.Put("key1", "value1")
//	m.Put("key2", "value2")
//	var values []string
//	m.EnumValues(func(value string) bool {
//		values = append(values, value)
//		return true
//	})
func (self *SyncMap[K]) EnumValues(f func(value K) bool) {
	self.mu.Lock()
	self.m.EnumValues(f)
	self.mu.Unlock()
	return
}

// Keys returns all keys.
//
// Returns:
//   - out: A slice containing all the keys in the map.  The order of the keys
//     is not guaranteed.
//
// Example:
//
//	m := NewSync[string]()
//	m.Put("key1", "value1")
//	m.Put("key2", "value2")
//	keys := m.Keys() // keys == []string{"value1", "value2"} (order not guaranteed)
func (self *SyncMap[K]) Keys() (out []K) {
	self.mu.Lock()
	out = self.m.Keys()
	self.mu.Unlock()
	return
}

// Values returns all values.
//
// Returns:
//   - out: A slice containing all the values in the map.  The order of the
//     values is not guaranteed.
//
// Example:
//
//	m := NewSync[string]()
//	m.Put("key1", "value1")
//	m.Put("key2", "value2")
//	values := m.Values() // values == []string{"key1", "key2"} (order not guaranteed)
func (self *SyncMap[K]) Values() (out []K) {
	self.mu.Lock()
	out = self.m.Values()
	self.mu.Unlock()
	return
}
