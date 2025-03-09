package bidi

import "sync"

// Map is an unordered generic bidirectional one-to-one map.
type Map[K comparable] struct {
	z        K
	keyToVal map[K]K
	valToKey map[K]K
}

// New returns a new [Map].
func New[K comparable]() *Map[K] {
	return &Map[K]{
		z:        *new(K),
		keyToVal: make(map[K]K),
		valToKey: make(map[K]K),
	}
}

// Len returns number of pairs in the map.
func (self *Map[K]) Len() int { return len(self.keyToVal) }

// KeyExists returns truth if key exists.
func (self *Map[K]) KeyExists(key K) (exists bool) {
	_, exists = self.keyToVal[key]
	return
}

// ValExists returns truth if value exists.
func (self *Map[K]) ValExists(value K) (exists bool) {
	_, exists = self.valToKey[value]
	return
}

// Key returns the key of the value.
func (self *Map[K]) Key(value K) (key K, b bool) {
	key, b = self.valToKey[value]
	return
}

// Val returns the value of the key.
func (self *Map[K]) Val(key K) (value K, b bool) {
	value, b = self.keyToVal[key]
	return
}

// Put stores value under key and returns oldValue that was replaced and a
// truth if value existed under key k and was replaced.
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
func (self *Map[K]) DeleteByValue(value K) (deletedKey K, exists bool) {
	if deletedKey, exists = self.valToKey[value]; !exists {
		return self.z, false
	}
	delete(self.keyToVal, deletedKey)
	delete(self.valToKey, value)
	return
}

// EnumKeys enumerates all keys in the map.
func (self *Map[K]) EnumKeys(f func(key K) bool) {
	for k := range self.keyToVal {
		if !f(k) {
			break
		}
	}
}

// EnumValues enumerates all values in the map.
func (self *Map[K]) EnumValues(f func(value K) bool) {
	for v := range self.valToKey {
		if !f(v) {
			break
		}
	}
}

// Keys returns all keys.
func (self *Map[K]) Keys() (out []K) {
	out = make([]K, 0, len(self.keyToVal))
	for _, k := range self.keyToVal {
		out = append(out, k)
	}
	return
}

// Values returns all values.
func (self *Map[K]) Values() (out []K) {
	out = make([]K, 0, len(self.valToKey))
	for _, v := range self.valToKey {
		out = append(out, v)
	}
	return
}

// SyncMap is the concurrency safe version of [Map].
type SyncMap[K comparable] struct {
	mu sync.Mutex
	m  *Map[K]
}

// NewBidiMap returns a new [Map].
func NewSync[K comparable]() *SyncMap[K] {
	return &SyncMap[K]{
		m: New[K](),
	}
}

// Len returns number of entries in the map.
func (self *SyncMap[K]) Len() (out int) {
	self.mu.Lock()
	out = self.m.Len()
	self.mu.Unlock()
	return
}

// Exists returns truth if an entry under key k exists.
func (self *SyncMap[K]) KeyExists(k K) (b bool) {
	self.mu.Lock()
	b = self.m.KeyExists(k)
	self.mu.Unlock()
	return
}

// Exists returns truth if an entry under key k exists.
func (self *SyncMap[K]) ValExists(k K) (b bool) {
	self.mu.Lock()
	b = self.m.ValExists(k)
	self.mu.Unlock()
	return
}

// Get returns the entry value under key k and a truth if found.
// if not found a zero value of entry value under key k is rturned.
func (self *SyncMap[K]) Key(value K) (k K, b bool) {
	self.mu.Lock()
	k, b = self.m.Key(value)
	self.mu.Unlock()
	return
}

// Get returns the entry value under key k and a truth if found.
// if not found a zero value of entry value under key k is rturned.
func (self *SyncMap[K]) Val(key K) (v K, b bool) {
	self.mu.Lock()
	v, b = self.m.Val(key)
	self.mu.Unlock()
	return
}

// Put stores value v under key k and returns a value that was replaced and a
// truth if value existed under key k and was replaced.
func (self *SyncMap[K]) Put(k K, v K) (old K, found bool) {
	self.mu.Lock()
	old, found = self.m.Put(k, v)
	self.mu.Unlock()
	return
}

// DeleteByKey deletes an entry by key and returns value that was bound to that
// key and truth if item was found and deleted.
func (self *SyncMap[K]) DeleteByKey(key K) (deletedValue K, exists bool) {
	self.mu.Lock()
	deletedValue, exists = self.m.DeleteByKey(key)
	self.mu.Unlock()
	return
}

// DeleteByValue deletes an entry by value and returns key that was bound to
// that value and truth if item was found and deleted.
func (self *SyncMap[K]) DeleteByValue(value K) (deletedKey K, exists bool) {
	self.mu.Lock()
	deletedKey, exists = self.m.DeleteByValue(value)
	self.mu.Unlock()
	return
}

// EnumKeys enumerates all keys in the map in order as added.
func (self *SyncMap[K]) EnumKeys(f func(key K) bool) {
	self.mu.Lock()
	self.m.EnumKeys(f)
	self.mu.Unlock()
	return
}

// EnumValues enumerates all values in the map in order as added.
func (self *SyncMap[K]) EnumValues(f func(value K) bool) {
	self.mu.Lock()
	self.m.EnumValues(f)
	self.mu.Unlock()
	return
}

// Keys returns all keys.
func (self *SyncMap[K]) Keys() (out []K) {
	self.mu.Lock()
	out = self.m.Keys()
	self.mu.Unlock()
	return
}

// Values returns all values.
func (self *SyncMap[K]) Values() (out []K) {
	self.mu.Lock()
	out = self.m.Values()
	self.mu.Unlock()
	return
}
