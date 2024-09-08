// Copyright 2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ttl

import (
	"errors"
	"sync/atomic"
	"time"
)

// TTL is a Time-To-Live list of keys that exist only for the duration
// specified. When a key expires TTL fires the callback specified in NewTTL.
//
// TTL worker should be stopped manually after use with Stop().
//
// It works by maintaining an ascending sorted queue of key timeout times and
// updates a ticker that ticks at the next timeout time in the queue and
// fires timeout callback. Precision is "okay", error is always a delay.
type TTL[T comparable] struct {
	running    atomic.Bool
	queue      []timeout[T]
	dict       map[T]time.Time
	addTimeout chan timeout[T]
	delTimeout chan T
	pingWorker chan bool
	cb         func(key T)
}

// timeout stores a key queued for timeout in the ttl queue.
type timeout[T comparable] struct {
	// When is the time when the Key should time out.
	When time.Time
	// Key is the key to time out at When.
	Key T
}

// Returns a new TTL which calls the optional cb each time a key expires.
func NewTTL[T comparable](cb func(key T)) *TTL[T] {
	var p = &TTL[T]{
		dict:       make(map[T]time.Time),
		addTimeout: make(chan timeout[T]),
		delTimeout: make(chan T),
		pingWorker: make(chan bool),
		cb:         cb,
	}
	go p.worker()
	<-p.pingWorker
	return p
}

// ErrNotRunning is returned when a timeout is being added or deleted to/from
// a ttl that has been stopped.
var ErrNotRunning = errors.New("ttl is not running")

// Put adds the key to ttl which will last for duration.
// If key is already present its timeout is reset to duration.
// Returns ErrNotRunning if ttl is stopped.
func (self *TTL[T]) Put(key T, duration time.Duration) error {
	if !self.running.Load() {
		return ErrNotRunning
	}
	self.delTimeout <- key
	self.addTimeout <- timeout[T]{time.Now().Add(duration), key}
	return nil
}

// Delete removes a key from the list.
// Returns ErrNotRunning if ttl is stopped.
func (self *TTL[T]) Delete(key T) error {
	if !self.running.Load() {
		return ErrNotRunning
	}
	self.delTimeout <- key
	return nil
}

// Stop stops the worker. This method should be called on shutdown.
// Returns ErrNotRunning if ttl is already stopped.
func (self *TTL[T]) Stop() error {
	if !self.running.Load() {
		return ErrNotRunning
	}
	self.pingWorker <- true
	return nil
}

// doOnTimeout calls ttl.cb if it's not nil and passes key to it.
func (self *TTL[T]) doOnTimeout(key T) {
	if self.cb != nil {
		self.cb(key)
	}
}

// worker is the main ttl logic. It enqueues, re-queues and deletes timeouts
// to/from the queue using channels in order to be concurency safe. It also
// drains the timeouts queue and fires callbacks for timed out timeouts.
func (self *TTL[T]) worker() {

	self.running.Store(true)
	self.pingWorker <- true

	var (
		zero   = *new(T)
		dur    time.Duration
		key    T
		when   time.Time
		ticker = time.NewTicker(time.Second * 1)
	)

loop:
	for {
		select {
		case <-self.pingWorker:
			ticker.Stop()
			break loop
		case t := <-self.addTimeout:
			// Timeout is somehow before Now(), fire immediately without queue.
			if t.When.Before(time.Now()) {
				self.doOnTimeout(t.Key)
				continue
			}
			// No active timeout, initialize it to received timeout.
			if key == zero {
				if dur = time.Until(t.When); dur > 0 {
					key, when = t.Key, t.When
					ticker.Reset(dur)
					continue
				}
				self.doOnTimeout(key)
				key, when = self.advanceQueue(ticker)
				continue
			}
			// New timeout is before current tick being waited on,
			// replace current tick with the new one and reinsert
			// the tick that was being waited on into timeouts queue.
			if t.When.Before(when) {
				ticker.Stop()
				self.insertTimeout(timeout[T]{when, key})
				if dur = time.Until(t.When); dur > 0 {
					key, when = t.Key, t.When
					ticker.Reset(dur)
				} else {
					self.doOnTimeout(t.Key)
					key, when = self.advanceQueue(ticker)
				}
				continue
			}
			// A ticker is active and new event is after current, queue it.
			self.insertTimeout(t)
		case k := <-self.delTimeout:
			// Timeout being deleted is currently active.
			if k == key {
				ticker.Stop()
				key, when = self.advanceQueue(ticker)
				continue
			}
			// Delete the timeout from queue.
			if idx, found := self.findTimeout(self.dict[key]); found {
				self.queue = append(self.queue[:idx], self.queue[idx+1:]...)
				delete(self.dict, k)
			}
		case <-ticker.C:
			ticker.Stop()
			if key != zero {
				delete(self.dict, key)
				self.doOnTimeout(key)
			}
			key, when = self.advanceQueue(ticker)
		}
	}
	clear(self.queue)
	clear(self.dict)
	self.running.Store(false)
}

// advanceQueue returns the next key and when time from the queue and resets the
// ticker to appropriate duration if a timeout entry exists in the queue.
// If now is past the next timeout in the queue the timeout is immediately
// fired as well as all following timeouts before now until a timeout that is
// after now is found. If the queue is empty the ticker is not reset and an
// empty key and zero time are returned.
func (self *TTL[T]) advanceQueue(ticker *time.Ticker) (key T, when time.Time) {
	var dur time.Duration
	for len(self.queue) > 0 {
		key, when = self.queue[0].Key, self.queue[0].When
		self.queue = self.queue[1:]
		delete(self.dict, key)
		if dur = time.Until(when); dur <= 0 {
			self.doOnTimeout(key)
			continue
		}
		ticker.Reset(dur)
		return
	}
	return *new(T), time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
}

// insertTimeout inserts timeout into the timeouts slice sorted by timeout.When.
// It ignores a possible duplicate of timeout.Key in the queue; other logic
// handles key duplicates, specifically Put().
func (self *TTL[T]) insertTimeout(t timeout[T]) {
	var i, j = 0, len(self.queue)
	for i < j {
		var h = int(uint(i+j) >> 1)
		if self.queue[h].When.Before(t.When) {
			i = h + 1
		} else {
			j = h
		}
	}
	self.queue = append(self.queue, timeout[T]{})
	copy(self.queue[i+1:], self.queue[i:])
	self.queue[i] = t
	self.dict[t.Key] = t.When
}

// findTimeout binary searches the queue slice for a timeout by its when time
// and returns its index and true if found or an invalid index and false if not.
func (self *TTL[T]) findTimeout(when time.Time) (idx int, found bool) {
	var (
		n    = len(self.queue)
		i, j = 0, n
	)
	for i < j {
		h := int(uint(i+j) >> 1)
		switch self.cmp(when, h) {
		case 1:
			i = h + 1
		case 0:
			return h, true
		case -1:
			j = h
		}
	}
	return i, i < n && self.cmp(when, i) == 0
}

// cmpTime compares v with with when time of a timeout in queue at i.
// Returns -1 if v is less, 0 if same, 1 if more.
func (self *TTL[T]) cmp(v time.Time, i int) int {
	if v.Before(self.queue[i].When) {
		return -1
	} else if v.After(self.queue[i].When) {
		return 1
	}
	return 0
}
