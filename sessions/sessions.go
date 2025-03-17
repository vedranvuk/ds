// Copyright 2025 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package sessions implements a generic session manager.
package sessions

import (
	"errors"
	"sync"
	"time"

	"github.com/vedranvuk/ds/ttl"
)

var (
	// ErrNotFound is returned if a session or a user was not found.
	ErrNotFound = errors.New("not found")
	// ErrMaxSessions is returned when a maximum number of sessions
	// cumulatively per user has been reached.
	ErrMaxSessions = errors.New("maximum session count reached")
	// ErrMaxUserSessions is returned when a maximum number of sessions per
	// user has been reached.
	ErrMaxUserSessions = errors.New("maximum user session count reached")
)

// Map maintains a list of keys for a key that time out after a set
// duration. Intended for use as an in-memory user session manager.
type Map[K comparable] struct {
	mu                 sync.Mutex
	z                  K
	newKey             func() K
	userCounts         map[K]int
	userToSessions     map[K]map[K]struct{}
	sessionToUser      map[K]K
	sessionDurations   map[K]time.Duration
	timeouts           *ttl.TTL[K]
	maxSessions        int
	maxSessionsPerUser int
	totalCount         int
}

// New returns new [Map]. maxSessions controls maximum nunber of sessions
// cumulatively and maxSessionsPerUser limits number of sessions per user.
func New[K comparable](maxSessions, maxSessionsPerUser int, newKey func() K) (out *Map[K]) {
	out = &Map[K]{
		z:                  *new(K),
		newKey:             newKey,
		userCounts:         make(map[K]int),
		userToSessions:     make(map[K]map[K]struct{}),
		sessionToUser:      make(map[K]K),
		sessionDurations:   make(map[K]time.Duration),
		maxSessions:        maxSessions,
		maxSessionsPerUser: maxSessionsPerUser,
		totalCount:         0,
	}
	out.timeouts = ttl.New(out.timeout)
	return
}

// Add adds a new session for userID and discards it after duration amount
// unless extended with [Map.Extend].
func (self *Map[K]) Add(userID K, duration time.Duration) (sessionID K, err error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.totalCount == self.maxSessions {
		return self.z, ErrMaxSessions
	}
	if count, ok := self.userCounts[userID]; ok {
		if count == self.maxSessionsPerUser {
			return self.z, ErrMaxUserSessions
		}
	}
	sessionID = self.newKey()
	if _, exists := self.userToSessions[userID]; !exists {
		self.userToSessions[userID] = make(map[K]struct{})
	}
	self.userToSessions[userID][sessionID] = struct{}{}
	self.sessionToUser[sessionID] = userID
	self.userCounts[userID] = self.userCounts[userID] + 1
	self.sessionDurations[sessionID] = duration
	self.timeouts.Put(sessionID, duration)
	self.totalCount += 1
	return
}

// UserID returns a userID by sessionID and true if session with sessionID
// exists. Otherwise returns a zero key and false.
func (self *Map[K]) UserID(sessionID K) (userID K, found bool) {
	self.mu.Lock()
	userID, found = self.sessionToUser[sessionID]
	self.mu.Unlock()
	return
}

// Extend resets the timeout of a session with sessionID to timeout set at
// with [Map.Add] when the session was created.
func (self *Map[K]) Extend(sessionID K) (err error) {
	self.mu.Lock()
	err = self.extend(sessionID)
	self.mu.Unlock()
	return
}

// RemoveSession removes a session by sessionID.
func (self *Map[K]) RemoveSession(sessionID K) (err error) {
	self.mu.Lock()
	err = self.removeSession(sessionID)
	self.mu.Unlock()
	return
}

// RemoveUser removes a user from the session list and all sessions created for
// the user.
func (self *Map[K]) RemoveUser(userID K) (err error) {
	self.mu.Lock()
	if sessions, exists := self.userToSessions[userID]; exists {
		for key := range sessions {
			self.removeSession(key)
		}
	}
	self.mu.Unlock()
	return
}

// SessionCount returns the total active session count.
func (self *Map[K]) SessionCount() (out int) {
	self.mu.Lock()
	out = self.totalCount
	self.mu.Unlock()
	return
}

// UserSessionCount returns number of active sessions for a user with userID.
func (self *Map[K]) UserSessionCount(userID K) (out int) {
	self.mu.Lock()
	out = self.userCounts[userID]
	self.mu.Unlock()
	return
}

// extend extends the session under sesionID.
func (self *Map[K]) extend(sessionID K) (err error) {
	if _, exists := self.sessionToUser[sessionID]; !exists {
		err = ErrNotFound
	} else {
		return self.timeouts.Put(sessionID, self.sessionDurations[sessionID])
	}
	return
}

// removeSession removes a session by sessionID.
func (self *Map[K]) removeSession(sessionID K) error {
	var userID, found = self.sessionToUser[sessionID]
	if !found {
		return ErrNotFound
	}
	if l, ok := self.userCounts[userID]; ok {
		if l == 0 {
			panic("chats: user session count at 0 on removal")
		} else if l == 1 {
			delete(self.userCounts, userID)
		} else {
			self.userCounts[userID] = l - 1
		}
	} else {
		panic("chats: user not in user list on removal")
	}
	if userSessions, exists := self.userToSessions[userID]; exists {
		delete(userSessions, sessionID)
		self.userToSessions[userID] = userSessions
	}
	delete(self.sessionToUser, sessionID)
	delete(self.sessionDurations, sessionID)
	self.totalCount -= 1
	return nil
}

// timeout is an event handler for ttl timeout.
func (self *Map[K]) timeout(key K) {
	self.mu.Lock()
	self.removeSession(key)
	self.mu.Unlock()
	return
}
