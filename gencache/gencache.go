package gencache

import (
	"errors"
	"reflect"
	"sync"
	"unsafe"
)

// GenCache is a generic cache of any type of value V, keyed by a comparable
// key K.
//
type GenCache[K comparable, V any] struct {
	mutex sync.RWMutex

	zero        V
	valSize     uint32
	used, limit uint32
	maxItems    uint32
	entries     map[K]V
	order       []K
}

// New returns a new [GenCache].
func New[K comparable, V any](memLimit uint32, itemLimit uint32) *GenCache[K, V] {
	var p = &GenCache[K, V]{
		limit:    memLimit,
		maxItems: itemLimit,
		entries:  make(map[K]V),
		order:    make([]K, itemLimit, itemLimit),
		zero:     *new(V),
	}
	p.valSize = uint32(reflect.TypeOf(p.zero).Size())
	return p
}

// ErrCacheMiss is returned by Cache.Get if item is not found in cache.
var ErrCacheMiss = errors.New("cache miss")

// Get retrieves an item from cache by id.
// If the item was not found an ErrCacheMiss is returned.
func (self *GenCache[K, V]) Get(key K) (out V, err error) {
	self.mutex.RLock()
	out, err = self.get(key)
	self.mutex.RUnlock()
	return
}

// Get retrieves an item from cache by id.
// If the item was not found an ErrCacheMiss is returned.
func (self *GenCache[K, V]) get(key K) (out V, err error) {
	var exists bool
	if out, exists = self.entries[key]; !exists {
		return self.zero, ErrCacheMiss
	}
	return
}

// Put stores buf into cache under id and rotates the cache if storage limit
// has been reached.
func (self *GenCache[K, V]) Put(key K, data V) {
	self.mutex.Lock()
	self.put(key, data)
	self.mutex.Unlock()
}

func (self *GenCache[K, V]) put(key K, data V) {
	var dataSize = uint32(unsafe.Sizeof(data))
	for uint32(self.used+dataSize) > self.limit || uint32(len(self.order)) > self.maxItems {
		var (
			delId   = self.order[0]
			delSize = uint32(unsafe.Sizeof(self.entries[delId]))
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
func (self *GenCache[K, V]) Delete(key K) (exists bool) {
	self.mutex.Lock()
	exists = self.delete(key)
	self.mutex.Unlock()
	return
}

// Delete deletes entry under key from cache if it exists and returns truth if
// it was found and deleted.
func (self *GenCache[K, V]) delete(key K) (exists bool) {
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
	self.mutex.RLock()
	_, exists = self.entries[key]
	self.mutex.RUnlock()
	return
}

// Usage returns current memory usage in bytes.
func (self *GenCache[K, V]) Usage() (used uint32) {
	self.mutex.RLock()
	used = self.used
	self.mutex.RUnlock()
	return
}
