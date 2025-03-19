// Copyright 2025 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cache

import (
	"errors"
	"sync"
)

// Cache is a simple rotating cache that maintains byte slices in memory up to
// the defined storage limit, both in memory size and entry count.
//
// The cache is safe for concurrent access.
type Cache struct {
	mutex sync.RWMutex

	used, limit uint32
	maxItems    uint32
	entries     map[string][]byte
	order       []string
}

// NewCache returns a new cache with the given memory usage limit in bytes and
// maximum entry count.
//
// Arguments:
//
//   - memLimit: The maximum memory usage of the cache in bytes.
//   - itemLimit: The maximum number of items that can be stored in the cache.
//
// Returns:
//
//   - A pointer to a new Cache instance.
//
// Example:
//
//	cache := NewCache(1024*1024, 100) // 1MB limit, 100 items max
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

// Get retrieves an item from cache by key.
//
// Arguments:
//
//   - key: The key of the item to retrieve.
//
// Returns:
//
//   - out: The byte slice stored under the given key, if found.
//   - err: An error. ErrCacheMiss if the item was not found.
//
// Example:
//
//	data, err := cache.Get("my_item")
//	if err != nil {
//		if errors.Is(err, ErrCacheMiss) {
//			fmt.Println("Item not found in cache")
//		} else {
//			fmt.Println("Error getting item:", err)
//		}
//		return
//	}
//	fmt.Println("Item found:", string(data))
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

// Put stores data into cache under key and rotates the cache if storage limit
// has been reached.  If an item with the same key already exists, it will be overwritten.
//
// Arguments:
//
//   - key: The key under which to store the data.
//   - data: The byte slice to store in the cache.
//
// Example:
//
//	data := []byte("some data to cache")
//	cache.Put("my_item", data)
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

// Delete deletes entry under key from cache if it exists and returns true if
// it was found and deleted, false otherwise.
//
// Arguments:
//
//   - key: The key of the item to delete.
//
// Returns:
//
//   - exists: True if the item was found and deleted, false otherwise.
//
// Example:
//
//	deleted := cache.Delete("my_item")
//	if deleted {
//		fmt.Println("Item deleted from cache")
//	} else {
//		fmt.Println("Item not found in cache")
//	}
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

// Exists returns true if an entry under key exists in cache, false otherwise.
//
// Arguments:
//
//   - key: The key of the item to check for.
//
// Returns:
//
//   - exists: True if the item exists in the cache, false otherwise.
//
// Example:
//
//	exists := cache.Exists("my_item")
//	if exists {
//		fmt.Println("Item exists in cache")
//	} else {
//		fmt.Println("Item does not exist in cache")
//	}
func (self *Cache) Exists(key string) (exists bool) {
	self.mutex.RLock()
	_, exists = self.entries[key]
	self.mutex.RUnlock()
	return
}

// Usage returns current memory usage in bytes.
//
// Returns:
//
//   - used: The current memory usage of the cache in bytes.
//
// Example:
//
//	usage := cache.Usage()
//	fmt.Println("Cache usage:", usage, "bytes")
func (self *Cache) Usage() (used uint32) {
	self.mutex.RLock()
	used = self.used
	self.mutex.RUnlock()
	return
}
