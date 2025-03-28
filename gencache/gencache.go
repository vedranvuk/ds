// Copyright 2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package gencache

import (
	"sync"
	"unsafe"
)

// GenCache is a generic cache of any type of value V, keyed by a comparable
// key K.
type GenCache[K comparable, V any] struct {
	zero        V
	used, limit uint32
	maxItems    uint32
	entries     map[K]V
	order       []K
}

// NewGenCache returns a new [GenCache].
func NewGenCache[K comparable, V any](memLimit uint32, itemLimit uint32) *GenCache[K, V] {
	var p = &GenCache[K, V]{
		limit:    memLimit,
		maxItems: itemLimit,
		entries:  make(map[K]V),
		order:    make([]K, itemLimit, itemLimit),
		zero:     *new(V),
	}
	return p
}

// Get retrieves an item from cache by id and true if found. Otherwise returns 
// zero value of V and false.
func (self *GenCache[K, V]) Get(key K) (value V, found bool) {
	if value, found = self.entries[key]; !found {
		return self.zero, false
	}
	return
}

// Put stores buf into cache under id and rotates the cache if storage limit
// has been reached. It returns the old value if one existed at specified id 
// and true or zero value of v and false otherwise.
func (self *GenCache[K, V]) Put(key K, data V) (old V, replaced bool) {
	var dataSize = uint32(unsafe.Sizeof(data))
	for uint32(self.used+dataSize) > self.limit || uint32(len(self.order)) > self.maxItems {
		var (
			delId   = self.order[0]
			delSize = uint32(unsafe.Sizeof(self.entries[delId]))
		)
		self.used -= delSize
		old, replaced = self.entries[delId]
		delete(self.entries, delId)
		self.order = self.order[1:]
	}
	self.used += dataSize
	self.order = append(self.order, key)
	self.entries[key] = data
	return
}

// Delete deletes entry under key from cache if it exists and returns truth if
// it was found and deleted.
func (self *GenCache[K, V]) Delete(key K) (exists bool) {
	var value V
	if value, exists = self.entries[key]; exists {
		self.used -= uint32(unsafe.Sizeof(value))
		delete(self.entries, key)
		return true
	}
	return false
}

// Returns truth if entry under key exists in cache.
func (self *GenCache[K, V]) Exists(key K) (exists bool) {
	_, exists = self.entries[key]
	return
}

// Usage returns current memory usage in bytes.
func (self *GenCache[K, V]) Usage() (used uint32) { return self.used }

// SyncGenCache is the concurrency safe version of [GenCache].
type SyncGenCache[K comparable, V any] struct {
	mutex sync.RWMutex
	GenCache[K, V]
}

// NewSyncGenCache returns a new [SyncGenCache].
func NewSyncGenCache[K comparable, V any](memLimit uint32, itemLimit uint32) *SyncGenCache[K, V] {
	return &SyncGenCache[K, V]{
		GenCache: *NewGenCache[K, V](memLimit, itemLimit),
	}
}

// Get retrieves an item from cache by id and true if found. Otherwise returns 
// zero value of V and false.
func (self *SyncGenCache[K, V]) Get(key K) (value V, found bool) {
	self.mutex.RLock()
	value, found = self.GenCache.Get(key)
	self.mutex.RUnlock()
	return
}

// Put stores buf into cache under id and rotates the cache if storage limit
// has been reached. It returns the old value if one existed at specified id 
// and true or zero value of v and false otherwise.
func (self *SyncGenCache[K, V]) Put(key K, data V) (old V, replaced bool) {
	self.mutex.Lock()
	old, replaced = self.GenCache.Put(key, data)
	self.mutex.Unlock()
	return
}

// Delete deletes entry under key from cache if it exists and returns truth if
// it was found and deleted.
func (self *SyncGenCache[K, V]) Delete(key K) (exists bool) {
	self.mutex.Lock()
	exists = self.GenCache.Delete(key)
	self.mutex.Unlock()
	return
}

// Returns truth if entry under key exists in cache.
func (self *SyncGenCache[K, V]) Exists(key K) (exists bool) {
	self.mutex.RLock()
	exists = self.GenCache.Exists(key)
	self.mutex.RUnlock()
	return
}

// Usage returns current memory usage in bytes.
func (self *SyncGenCache[K, V]) Usage() (used uint32) {
	self.mutex.RLock()
	used = self.GenCache.Usage()
	self.mutex.RUnlock()
	return
}
