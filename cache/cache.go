// Copyright 2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"errors"
	"sync"
)

// Cache is a simple rotating cache that maintains byte slices in memory up to
// the defined storage limit, both in memory size and entry count.
type Cache struct {
	mutex sync.RWMutex

	used, limit uint32
	maxItems    uint32
	entries     map[string][]byte
	order       []string
}

// NewCache returns a new cache with the given memory usage limit in bytes and
// maximum entry count.
func NewCache(memLimit uint32, itemLimit uint32) *Cache {
	var p = &Cache{
		limit:    memLimit,
		maxItems: itemLimit,
		entries:  make(map[string][]byte),
		order:    make([]string, itemLimit, itemLimit),
	}
	return p
}

// ErrCacheMiss is returned by Cache.Get if item is not found in cache.
var ErrCacheMiss = errors.New("cache miss")

// Get retrieves an item from cache by id.
// If the item was not found an ErrCacheMiss is returned.
func (self *Cache) Get(key string) (out []byte, err error) {
	self.mutex.RLock()
	out, err = self.get(key)
	self.mutex.RUnlock()
	return
}

// Get retrieves an item from cache by id.
// If the item was not found an ErrCacheMiss is returned.
func (self *Cache) get(key string) (out []byte, err error) {
	var exists bool
	if out, exists = self.entries[key]; !exists {
		return nil, ErrCacheMiss
	}
	return
}

// Put stores buf into cache under id and rotates the cache if storage limit
// has been reached.
func (self *Cache) Put(key string, data []byte) {
	self.mutex.Lock()
	self.put(key, data)
	self.mutex.Unlock()
}

func (self *Cache) put(key string, data []byte) {
	var dataSize = uint32(len(data))
	for uint32(self.used+dataSize) > self.limit || uint32(len(self.order)) > self.maxItems {
		var (
			delId   = self.order[0]
			delSize = uint32(len(self.entries[delId]))
		)
		self.used -= delSize
		delete(self.entries, delId)
		self.order = self.order[1:]
	}
	self.used += dataSize
	self.order = append(self.order, key)
	self.entries[key] = data
}

// Delete deletes entry under key from cache if it exists and returns truth if
// it was found and deleted.
func (self *Cache) Delete(key string) (exists bool) {
	self.mutex.Lock()
	exists = self.delete(key)
	self.mutex.Unlock()
	return
}

// Delete deletes entry under key from cache if it exists and returns truth if
// it was found and deleted.
func (self *Cache) delete(key string) (exists bool) {
	var value []byte
	if value, exists = self.entries[key]; exists {
		self.used -= uint32(len(value))
		delete(self.entries, key)
		return true
	}
	return false
}

// Returns truth if entry under key exists in cache.
func (self *Cache) Exists(key string) (exists bool) {
	self.mutex.RLock()
	_, exists = self.entries[key]
	self.mutex.RUnlock()
	return
}

// Usage returns current memory usage in bytes.
func (self *Cache) Usage() (used uint32) {
	self.mutex.RLock()
	used = self.used
	self.mutex.RUnlock()
	return
}
