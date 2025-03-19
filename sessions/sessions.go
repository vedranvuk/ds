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
	// ErrMaxSessions is returned when a global maximum number of sessions
	// has been reached.
	ErrMaxSessions = errors.New("maximum session count reached")
	// ErrMaxUserSessions is returned when a maximum number of sessions per
	// user has been reached.
	ErrMaxUserSessions = errors.New("maximum user session count reached")
)

// Map is a concurrent-safe session manager that associates sessions with users,
// enforcing limits on total sessions and sessions per user. It leverages a TTL
// (Time-To-Live) mechanism for automatic session expiration.
//
// The zero value is not usable; use New to create a valid instance.
//
// Generic Type Parameter:
//
//   - K: The comparable type used for session and user identifiers (keys).
//
// Core Features:
//
//   - Session Creation: Generates unique session identifiers and associates them with optional user IDs.
//   - User Linking:  Associates sessions with users, enabling per-user session management.
//   - Timeout Management: Automatically expires sessions after a configurable duration.
//   - Concurrent Safety: Protects internal state with a mutex, ensuring safe concurrent access.
//   - Session Limits: Enforces maximum limits on the total number of active sessions and the number of sessions per user.
//
// Usage:
//
//   - Use New to initialize a Map instance with the desired session limits and a key generation function.
//   - Create sessions using Create or CreateLinked, specifying a timeout duration.
//   - Link sessions to users using Link to enable per-user session management and limits.
//   - Extend the lifetime of a session using Extend.
//   - Remove sessions explicitly using RemoveSession or implicitly through timeout.
//   - Retrieve a user ID associated with a session using UserID.
//   - Get session counts using SessionCount and UserSessionCount.
//
// Error Handling:
//
//   - Returns ErrNotFound if a session or user is not found.
//   - Returns ErrMaxSessions if the maximum number of sessions is reached.
//   - Returns ErrMaxUserSessions if the maximum number of sessions per user is reached.
//
// Internal Data Structures:
//
//   - mu: A mutex to protect concurrent access to the map's internal state.
//   - userCounts:  A map[K]int storing the number of active sessions per user.
//   - userToSessions: A map[K]map[K]struct{} that maps a user ID to a set of session IDs owned by that user.
//   - sessionToUser: A map[K]K that maps a session ID to the user ID that owns the session.
//   - sessionDurations: A map[K]time.Duration that stores the original duration/TTL of each session.
//   - timeouts:  A ttl.TTL[K] instance that manages session timeouts.
//   - maxSessions: The maximum number of total sessions allowed.
//   - maxSessionsPerUser:  The maximum number of sessions allowed per user.
//   - totalCount: The current number of active sessions.
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

// New returns a new [Map] instance.
//
// Parameters:
//
//   - maxSessions: The maximum number of sessions allowed cumulatively.
//   - maxSessionsPerUser: Limits the number of sessions per user.
//   - newKey: A function that generates unique session keys.
//
// Returns:
//
//   - out: A pointer to the newly created [Map] instance.
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

// Create creates a new session valid for the specified duration.
//
// Parameters:
//
//   - duration: The duration for which the session is valid.
//
// Returns:
//
//   - sessionID: The newly generated session ID.
//   - err: An error, if any. Returns [ErrMaxSessions] if the maximum number of sessions has been reached.
func (self *Map[K]) Create(duration time.Duration) (sessionID K, err error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.totalCount == self.maxSessions {
		return self.z, ErrMaxSessions
	}
	sessionID = self.newKey()
	self.sessionDurations[sessionID] = duration
	self.timeouts.Put(sessionID, duration)
	self.totalCount += 1
	return
}

// Link links a session to a user. The session must exist beforehand.
//
// Parameters:
//
//   - sessionID: The ID of the session to link.
//   - userID: The ID of the user to link the session to.
//   - extend: If true, the session timeout is extended to the original duration.
//
// Returns:
//
//   - err: An error, if any. Returns [ErrNotFound] if the session is not found.
//     Returns [ErrMaxUserSessions] if the maximum number of sessions per user has been reached.
func (self *Map[K]) Link(sessionID, userID K, extend bool) (err error) {
	self.mu.Lock()
	defer self.mu.Unlock()

	if _, exists := self.sessionDurations[sessionID]; !exists {
		return ErrNotFound
	}

	if count, ok := self.userCounts[userID]; ok {
		if count == self.maxSessionsPerUser {
			return ErrMaxUserSessions
		}
	}

	if _, exists := self.userToSessions[userID]; !exists {
		self.userToSessions[userID] = make(map[K]struct{})
	}
	self.userToSessions[userID][sessionID] = struct{}{}
	self.sessionToUser[sessionID] = userID
	self.userCounts[userID] = self.userCounts[userID] + 1

	if extend {
		self.timeouts.Put(sessionID, self.sessionDurations[sessionID])
	}

	return
}

// CreateLinked creates a new session valid for the specified duration and
// immediately links it to the specified user.
//
// Parameters:
//
//   - userID: The ID of the user to link the session to.
//   - duration: The duration for which the session is valid.
//
// Returns:
//
//   - sessionID: The newly generated session ID.
//   - err: An error, if any. Returns [ErrMaxSessions] if the maximum number of sessions has been reached.
//     Returns [ErrMaxUserSessions] if the maximum number of sessions per user has been reached.
func (self *Map[K]) CreateLinked(userID K, duration time.Duration) (sessionID K, err error) {
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

// UserID returns the user ID associated with the specified session ID.
//
// Parameters:
//
//   - sessionID: The ID of the session.
//
// Returns:
//
//   - userID: The ID of the user associated with the session.  If session does not exist returns zero value of type K.
//   - found: True if a session with sessionID exists, false otherwise.
func (self *Map[K]) UserID(sessionID K) (userID K, found bool) {
	self.mu.Lock()
	userID, found = self.sessionToUser[sessionID]
	self.mu.Unlock()
	return
}

// Extend resets the timeout of a session with the specified session ID to the
// timeout set when the session was created.
//
// Parameters:
//
//   - sessionID: The ID of the session to extend.
//
// Returns:
//
//   - err: An error, if any. Returns [ErrNotFound] if the session is not found.
func (self *Map[K]) Extend(sessionID K) (err error) {
	self.mu.Lock()
	err = self.extend(sessionID)
	self.mu.Unlock()
	return
}

// RemoveSession removes a session by its session ID.
//
// Parameters:
//
//   - sessionID: The ID of the session to remove.
//
// Returns:
//
//   - err: An error, if any.
func (self *Map[K]) RemoveSession(sessionID K) (err error) {
	self.mu.Lock()
	err = self.removeSession(sessionID, false)
	self.mu.Unlock()
	return
}

// RemoveUser removes a user and all of their sessions.
//
// Parameters:
//
//   - userID: The ID of the user to remove.
//
// Returns:
//
//   - err: An error, if any.
func (self *Map[K]) RemoveUser(userID K) (err error) {
	self.mu.Lock()
	if sessions, exists := self.userToSessions[userID]; exists {
		for key := range sessions {
			self.removeSession(key, false)
		}
	}
	self.mu.Unlock()
	return
}

// SessionCount returns the total number of active sessions.
//
// Returns:
//
//   - out: The total number of active sessions.
func (self *Map[K]) SessionCount() (out int) {
	self.mu.Lock()
	out = self.totalCount
	self.mu.Unlock()
	return
}

// UserSessionCount returns the number of active sessions for a specific user.
//
// Parameters:
//
//   - userID: The ID of the user.
//
// Returns:
//
//   - out: The number of active sessions for the user.
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
func (self *Map[K]) removeSession(sessionID K, timedOut bool) (err error) {

	self.totalCount -= 1

	if !timedOut {
		self.timeouts.Delete(sessionID)
	}
	delete(self.sessionDurations, sessionID)

	var userID K
	var found bool
	userID, found = self.sessionToUser[sessionID]
	if !found {
		return
	}
	delete(self.sessionToUser, sessionID)

	if l, ok := self.userCounts[userID]; ok {
		if l == 0 {
			err = errors.New("bug: user session count at 0 on removal")
		} else if l == 1 {
			delete(self.userCounts, userID)
		} else {
			self.userCounts[userID] = l - 1
		}
	} else {
		err = errors.New("bug: user not in user list on removal")
	}

	var userSessions map[K]struct{}
	var exists bool
	userSessions, exists = self.userToSessions[userID]
	if exists {
		delete(userSessions, sessionID)
		if len(userSessions) == 0 {
			delete(self.userToSessions, userID)
		} else {
			self.userToSessions[userID] = userSessions
		}
	}

	return
}

// timeout is an event handler for ttl timeout.
func (self *Map[K]) timeout(key K) {
	self.mu.Lock()
	self.removeSession(key, true)
	self.mu.Unlock()
	return
}
