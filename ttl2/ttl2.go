// Copyright 2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package ttl provides a Time-To-Live (TTL) list for comparable keys with
// expiration callbacks.
//
// This implementation uses a somewhat slower linked list and a map.
package ttl2

import (
	"container/list"
	"errors"
	"runtime"
	"sync"
	"time"
)

// TTL is a Time-To-Live list of comparable keys that exist only for the
// duration specified. When a key expires, TTL fires the callback specified in
// [New].
//
// TTL worker should be stopped manually after use with [TTL.Stop].
//
// It works by maintaining an ascending sorted queue of key timeout times that
// get on the next tick that is updated each time a timeout, update, put, or delete occurs.
// It updates a ticker that ticks at the next timeout time in the queue and
// fires the timeout callback. Precision is "okay"; the error is always a delay.
type TTL[K comparable] struct {
	mu       sync.RWMutex
	z        K
	cb       func(K)
	timeouts *list.List
	keys     map[K]*list.Element
	current  *timeout[K]
	comm     chan cmd

	waitmu  sync.Mutex
	waiters []chan time.Time
}

// timeout stores a key queued for timeout in the ttl queue.
type timeout[K comparable] struct {
	// When is the time when the Key should time out.
	// It is constructed by adding duration to current
	// time of queueing up the item.
	When time.Time
	// Key is the key to time out at When.
	Key K
}

// New returns a new TTL which calls the optional cb each time a key expires.
//
// [TTL] is returned started and should be stopped after use with [TTL.Stop].
//
// Parameters:
//   - cb: A callback function that is called when a key expires. It receives the expired key as an argument.
//
// Returns:
//   - *TTL[K]: A pointer to the newly created TTL instance.
//
// Example:
//
//	ttl := ttl.New[string](func(key string) {
//		fmt.Println("Key expired:", key)
//	})
//	defer ttl.Stop()
func New[K comparable](cb func(key K)) (out *TTL[K]) {

	out = &TTL[K]{
		cb:       cb,
		z:        *new(K),
		comm:     make(chan cmd),
		timeouts: list.New(),
		keys:     make(map[K]*list.Element),
	}

	// Sends the stop command to worker to stop the ticker on garbage collection.
	runtime.SetFinalizer(out, func(ttl *TTL[K]) {
		out.comm <- stop
	})

	go out.worker()
	<-out.comm

	return
}

// Len returns the number of events in the list left to fire.
//
// Returns:
//   - l: The number of events in the queue, including the currently waiting event.
//
// Example:
//
//	length := ttl.Len()
//	fmt.Println("Number of events in queue:", length)
func (self *TTL[K]) Len() (l int) {
	self.mu.RLock()
	l = len(self.keys)
	self.mu.RUnlock()
	return
}

// Wait returns a channel that returns the current time when the TLL queue is
// empty and all events have fired.
//
// Returns:
//   - chan time.Time: A channel that will receive the current time when the TTL queue is empty.
//
// Example:
//
//	done := ttl.Wait()
//	<-done
//	fmt.Println("TTL queue is empty")
func (self *TTL[K]) Wait() (out chan time.Time) {
	self.waitmu.Lock()
	out = make(chan time.Time)
	self.waiters = append(self.waiters, out)
	self.mu.RLock()
	var l = len(self.keys)
	self.mu.RUnlock()
	if l == 0 {
		self.waitmu.Unlock()
		self.notifyWaiters()
		return
	}
	self.waitmu.Unlock()
	return
}

// Put adds the key to ttl which will last for duration.
// If key is already present its timeout is reset to duration.
// Returns ErrNotRunning if ttl is stopped.
//
// Parameters:
//   - key: The key to add to the TTL.
//   - duration: The duration for which the key should remain in the TTL.
//
// Returns:
//   - err: An error if the TTL is not running.
//
// Example:
//
//	err := ttl.Put("mykey", time.Minute)
//	if err != nil {
//		fmt.Println("Error putting key:", err)
//	}
func (self *TTL[K]) Put(key K, duration time.Duration) (err error) {
	// fmt.Println("sput")

	self.mu.Lock()

	var t = &timeout[K]{
		Key:  key,
		When: time.Now().Add(duration),
	}

	// If key is the item currently being waited on, unset it.
	// If not, check that its in queue and remove it.
	if self.current != nil && self.current.Key == key {
		self.current = nil
	} else if e, ok := self.keys[key]; ok {
		self.timeouts.Remove(e)
	}

	// Find latest timestamp to insert after.
	var e *list.Element
	for e = self.timeouts.Front(); e != nil; e = e.Next() {
		if e.Value.(*timeout[K]).When.After(t.When) {
			self.keys[key] = self.timeouts.InsertBefore(t, e)
			break
		}
	}
	// If queue was empty or duration belongs to end of list, push to back.
	if e == nil {
		self.keys[key] = self.timeouts.PushBack(t)
	}

	// Notify worker to wait on the next item if current was unset.
	if self.current == nil {
		self.mu.Unlock()
		self.comm <- next
		return
	}

	self.mu.Unlock()

	return
}

// Exists returns if key exists in TTL.
//
// Parameters:
//   - key: The key to check for existence.
//
// Returns:
//   - exists: True if the key exists in the TTL, false otherwise.
//
// Example:
//
//	exists := ttl.Exists("mykey")
//	fmt.Println("Key exists:", exists)
func (self *TTL[K]) Exists(key K) (exists bool) {
	self.mu.RLock()
	_, exists = self.keys[key]
	self.mu.RUnlock()
	return
}

// ErrNotFound is returned when delete does not find the item to be deleted.
var ErrNotFound = errors.New("not found")

// Delete removes a key from the list.
// Returns ErrNotRunning if ttl is stopped.
//
// Parameters:
//   - key: The key to remove from the TTL.
//
// Returns:
//   - err: An error if the TTL is not running or if the key is not found.
//
// Example:
//
//	err := ttl.Delete("mykey")
//	if err != nil {
//		fmt.Println("Error deleting key:", err)
//	}
func (self *TTL[K]) Delete(key K) error {
	self.mu.Lock()
	if e, ok := self.keys[key]; ok {
		self.timeouts.Remove(e)
		delete(self.keys, key)
		if self.current != nil && self.current.Key == key {
			self.current = nil
			self.mu.Unlock()
			self.comm <- next
		} else {
			self.mu.Unlock()
		}
		return nil
	}
	self.mu.Unlock()
	return ErrNotFound
}

// cmd is a command exchanged between worker and instance.
type cmd int

const (
	// none is sent from worker to [New] to signal that the worker has started.
	none cmd = iota
	// next is sent from [TTL.Put] and [TTL.Delete] to advance the queue.
	next
	// Stop is sent from runtime finalizer set in [New].
	stop
)

// worker is the main ttl logic. It enqueues, re-queues and deletes timeouts
// to/from the queue using channels in order to be concurency safe. It also
// drains the timeouts queue and fires callbacks for timed out timeouts.
func (self *TTL[K]) worker() {
	var ticker = time.NewTicker(time.Hour)
	ticker.Stop()
	self.comm <- none
loop:
	for {
		select {
		case <-ticker.C:
			self.mu.Lock()
			self.tick(ticker)
			self.mu.Unlock()
		case cmd := <-self.comm:
			switch cmd {
			case next:
				self.mu.Lock()
				// Advance queue.
				if self.current == nil {
					self.next(ticker)
				}
				// queue was empty, notify waiters.
				if self.current == nil {
					ticker.Stop()
					self.mu.Unlock()
					self.notifyWaiters()
					continue
				}
				self.mu.Unlock()
			case stop:
				ticker.Stop()
				break loop
			}
		}
	}
}

// tick is called on timer tick. It removes currently waited on key and calls
// the timeout callback. Finally advances to next item using [TTL.next] or
// modifies waiters that the TTL emptied.
func (self *TTL[K]) tick(ticker *time.Ticker) {
	// remove key and defer timeout callback.
	if self.current != nil {
		// fmt.Println("clearing current")
		delete(self.keys, self.current.Key)
		defer func(key K) {
			self.doOnTimeout(key)
		}(self.current.Key)
		self.current = nil
	}
	// set ticker to next timeout.
	self.next(ticker)
	// notify waiters if queue is empty.
	if self.current == nil {
		self.notifyWaiters()
	}
}

// next sets the current item to next in queue.
func (self *TTL[K]) next(ticker *time.Ticker) {
	if e := self.timeouts.Front(); e != nil {
		self.timeouts.Remove(e)
		var t = e.Value.(*timeout[K])
		self.current = t
		ticker.Reset(max(time.Nanosecond, time.Until(t.When)))
		return
	}
}

// notifyWaiters notifies all callers that requested a notification of TTL
// emptying.
func (self *TTL[K]) notifyWaiters() {
	self.waitmu.Lock()
	for _, waiter := range self.waiters {
		go func(w chan time.Time) {
			w <- time.Now()
		}(waiter)
	}
	self.waiters = nil
	self.waitmu.Unlock()
}

// doOnTimeout calls ttl.cb if it's not nil and passes key to it.
func (self *TTL[K]) doOnTimeout(key K) {
	// fmt.Println("timeout")
	if self.cb != nil {
		go self.cb(key)
	}
	return
}
