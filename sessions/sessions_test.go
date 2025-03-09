// Copyright 2025 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package sessions

import (
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type TestKey string

var keyCounter int64

func newTestKey() TestKey {
	atomic.AddInt64(&keyCounter, 1)
	return TestKey(generateRandomString(10) + strconv.FormatInt(keyCounter, 10))
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := time.Now().UnixNano()
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand%int64(len(charset))]
		seededRand /= int64(len(charset))
	}
	return string(b)
}

func assertEqual[T comparable](t *testing.T, expected, actual T, msg string) {
	if actual != expected {
		t.Errorf("%s: expected [%v], actual [%v]", msg, expected, actual)
	}
}

func assertNotNil(t *testing.T, actual interface{}, msg string) {
	if actual == nil {
		t.Errorf("%s: should not be nil", msg)
	}
}

func assertNotEmpty(t *testing.T, actual string, msg string) {
	if actual == "" {
		t.Errorf("%s: should not be empty", msg)
	}
}

func assertTrue(t *testing.T, actual bool, msg string) {
	if !actual {
		t.Errorf("%s: should be true", msg)
	}
}

func assertFalse(t *testing.T, actual bool, msg string) {
	if actual {
		t.Errorf("%s: should be false", msg)
	}
}

func assertNoError(t *testing.T, err error, msg string) {
	if err != nil {
		t.Errorf("%s: unexpected error [%v]", msg, err)
	}
}

func assertErrorIs(t *testing.T, err, target error, msg string) {
	if !errors.Is(err, target) {
		t.Errorf("%s: error [%v is not [%v]", msg, err, target)
	}
}

func assertLessOrEqual(t *testing.T, actual, expected int, msg string) {
	if actual > expected {
		t.Errorf("%s: actual [%d is not less or equal to expected [%d]", msg, actual, expected)
	}
}

func TestNew(t *testing.T) {
	sessions := New[TestKey](10, 3, newTestKey)
	assertNotNil(t, sessions, "New Sessions")
	assertEqual(t, 0, sessions.SessionCount(), "Initial SessionCount")
}

func TestAdd(t *testing.T) {
	sessions := New[TestKey](10, 3, newTestKey)
	userID := newTestKey()

	sessionID, err := sessions.Add(userID, time.Minute)
	assertNoError(t, err, "Add session")
	assertNotEmpty(t, string(sessionID), "SessionID should not be empty")
	assertEqual(t, 1, sessions.SessionCount(), "SessionCount after add")
	assertEqual(t, 1, sessions.UserSessionCount(userID), "UserSessionCount after add")

	_, err = sessions.Add(userID, time.Minute)
	assertNoError(t, err, "Add session 2")
	_, err = sessions.Add(userID, time.Minute)
	assertNoError(t, err, "Add session 3")
	_, err = sessions.Add(userID, time.Minute)
	assertErrorIs(t, err, ErrMaxUserSessions, "Add session 4, max user sessions error")
	assertEqual(t, 3, sessions.SessionCount(), "SessionCount after max user sessions")
	assertEqual(t, 3, sessions.UserSessionCount(userID), "UserSessionCount after max user sessions")

	sessions = New[TestKey](2, 3, newTestKey)
	_, err = sessions.Add(newTestKey(), time.Minute)
	assertNoError(t, err, "Add session 1")
	_, err = sessions.Add(newTestKey(), time.Minute)
	assertNoError(t, err, "Add session 2")
	_, err = sessions.Add(newTestKey(), time.Minute)
	assertErrorIs(t, err, ErrMaxSessions, "Add session 3, max sessions error")
	assertEqual(t, 2, sessions.SessionCount(), "SessionCount after max sessions")
}

func TestUserID(t *testing.T) {
	sessions := New[TestKey](10, 3, newTestKey)
	userID := newTestKey()

	sessionID, _ := sessions.Add(userID, time.Minute)
	retUserID, found := sessions.UserID(sessionID)
	assertTrue(t, found, "UserID found")
	assertEqual(t, userID, retUserID, "UserID returned")

	_, found = sessions.UserID(newTestKey())
	assertFalse(t, found, "UserID not found for unknown session")
}

func TestExtend(t *testing.T) {
	sessions := New[TestKey](10, 3, newTestKey)
	userID := newTestKey()
	sessionID, _ := sessions.Add(userID, time.Millisecond*100)

	time.Sleep(time.Millisecond * 50)
	err := sessions.Extend(sessionID)
	assertNoError(t, err, "Extend session")

	time.Sleep(time.Millisecond * 50)
	_, found := sessions.UserID(sessionID)
	assertTrue(t, found, "Session should be extended") // This line was failing

	time.Sleep(time.Second)
	_, found = sessions.UserID(sessionID)
	assertFalse(t, found, "Session should timeout eventually")

	err = sessions.Extend(newTestKey())
	assertErrorIs(t, err, ErrNotFound, "Extend unknown session")
}

func TestRemoveSession(t *testing.T) {
	sessions := New[TestKey](10, 3, newTestKey)
	userID := newTestKey()
	sessionID, _ := sessions.Add(userID, time.Minute)

	err := sessions.RemoveSession(sessionID)
	assertNoError(t, err, "RemoveSession")
	assertEqual(t, 0, sessions.SessionCount(), "SessionCount after RemoveSession")
	assertEqual(t, 0, sessions.UserSessionCount(userID), "UserSessionCount after RemoveSession")

	err = sessions.RemoveSession(newTestKey())
	assertErrorIs(t, err, ErrNotFound, "RemoveSession unknown session")
}

func TestRemoveUser(t *testing.T) {
	sessions := New[TestKey](10, 3, newTestKey)
	userID1 := newTestKey()
	userID2 := newTestKey()

	sessionID1, _ := sessions.Add(userID1, time.Minute)
	sessionID2, _ := sessions.Add(userID1, time.Minute)
	sessionID3, _ := sessions.Add(userID2, time.Minute)

	err := sessions.RemoveUser(userID1)
	assertNoError(t, err, "RemoveUser userID1")
	assertEqual(t, 1, sessions.SessionCount(), "SessionCount after RemoveUser userID1")
	assertEqual(t, 0, sessions.UserSessionCount(userID1), "UserSessionCount after RemoveUser userID1")
	assertEqual(t, 1, sessions.UserSessionCount(userID2), "UserSessionCount for userID2 after RemoveUser userID1")
	_, found := sessions.UserID(sessionID1)
	assertFalse(t, found, "SessionID1 after RemoveUser userID1")
	_, found = sessions.UserID(sessionID2)
	assertFalse(t, found, "SessionID2 after RemoveUser userID1")
	_, found = sessions.UserID(sessionID3)
	assertTrue(t, found, "SessionID3 after RemoveUser userID1")

	err = sessions.RemoveUser(userID1) // Removing non-existent user should be nil error as per implementation
	assertNoError(t, err, "RemoveUser non-existent userID1 should be no error")

	err = sessions.RemoveUser(userID2)
	assertNoError(t, err, "RemoveUser userID2")
	assertEqual(t, 0, sessions.SessionCount(), "SessionCount after RemoveUser userID2")
	assertEqual(t, 0, sessions.UserSessionCount(userID2), "UserSessionCount after RemoveUser userID2")
	_, found = sessions.UserID(sessionID3)
	assertFalse(t, found, "SessionID3 after RemoveUser userID2")
}

func TestSessionCount(t *testing.T) {
	sessions := New[TestKey](10, 3, newTestKey)
	assertEqual(t, 0, sessions.SessionCount(), "Initial SessionCount")

	sessions.Add(newTestKey(), time.Minute)
	sessions.Add(newTestKey(), time.Minute)
	assertEqual(t, 2, sessions.SessionCount(), "SessionCount after adding 2 sessions")

	sessions.RemoveSession(func() TestKey {
		for k := range sessions.sessionToUser {
			return k
		}
		return ""
	}())
	assertEqual(t, 1, sessions.SessionCount(), "SessionCount after removing 1 session")
}

func TestUserSessionCount(t *testing.T) {
	sessions := New[TestKey](10, 3, newTestKey)
	userID := newTestKey()
	assertEqual(t, 0, sessions.UserSessionCount(userID), "Initial UserSessionCount")

	sessions.Add(userID, time.Minute)
	sessions.Add(userID, time.Minute)
	assertEqual(t, 2, sessions.UserSessionCount(userID), "UserSessionCount after adding 2 sessions")

	sessionID, found := func() (TestKey, bool) {
		for k, u := range sessions.sessionToUser {
			if u == userID {
				return k, true
			}
		}
		return "", false
	}()
	assertTrue(t, found, "SessionID found for userID")
	sessions.RemoveSession(sessionID)
	assertEqual(t, 1, sessions.UserSessionCount(userID), "UserSessionCount after removing 1 session")
}

func TestTimeout(t *testing.T) {
	sessions := New[TestKey](10, 3, newTestKey)
	userID := newTestKey()
	sessionID, _ := sessions.Add(userID, time.Millisecond*100)

	time.Sleep(time.Millisecond * 200) // Wait for timeout
	_, found := sessions.UserID(sessionID)
	assertFalse(t, found, "UserID after timeout")
	assertEqual(t, 0, sessions.SessionCount(), "SessionCount after timeout")
	assertEqual(t, 0, sessions.UserSessionCount(userID), "UserSessionCount after timeout")
}

func TestConcurrentAddExtendTimeout(t *testing.T) {
	sessions := New[TestKey](100, 10, newTestKey)
	userID := newTestKey()
	var sessionIDs []TestKey
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sessionID, _ := sessions.Add(userID, time.Millisecond*200)
			sessionIDs = append(sessionIDs, sessionID)
			time.Sleep(time.Millisecond * time.Duration(generateRandomNumber(100))) // Simulate activity
			sessions.Extend(sessionID)
		}()
	}
	wg.Wait()
	time.Sleep(time.Second) // Wait for timeouts
	assertLessOrEqual(t, sessions.SessionCount(), 50, "SessionCount in concurrent test") // Some sessions might have timed out
}

func generateRandomNumber(max int) int {
	seededRand := time.Now().UnixNano()
	return int(seededRand % int64(max))
}

func BenchmarkAdd(b *testing.B) {
	sessions := New[TestKey](b.N, 1, newTestKey)
	userID := newTestKey()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessions.Add(userID, time.Minute)
	}
}

func BenchmarkUserID(b *testing.B) {
	sessions := New[TestKey](b.N, 1, newTestKey)
	sessionIDs := make([]TestKey, b.N)
	userID := newTestKey()
	for i := 0; i < b.N; i++ {
		sessionIDs[i], _ = sessions.Add(userID, time.Minute)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessions.UserID(sessionIDs[i])
	}
}

func BenchmarkExtend(b *testing.B) {
	sessions := New[TestKey](b.N, 1, newTestKey)
	sessionIDs := make([]TestKey, b.N)
	userID := newTestKey()
	for i := 0; i < b.N; i++ {
		sessionIDs[i], _ = sessions.Add(userID, time.Minute)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessions.Extend(sessionIDs[i])
	}
}

func BenchmarkRemoveSession(b *testing.B) {
	sessions := New[TestKey](b.N, 1, newTestKey)
	sessionIDs := make([]TestKey, b.N)
	userID := newTestKey()
	for i := 0; i < b.N; i++ {
		sessionIDs[i], _ = sessions.Add(userID, time.Minute)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessions.RemoveSession(sessionIDs[i])
	}
}

func BenchmarkAddConcurrent(b *testing.B) {
	sessions := New[TestKey](b.N, 1, newTestKey)
	userID := newTestKey()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sessions.Add(userID, time.Minute)
		}
	})
}

func BenchmarkUserIDConcurrent(b *testing.B) {
	sessions := New[TestKey](b.N, 1, newTestKey)
	sessionIDs := make([]TestKey, b.N)
	userID := newTestKey()
	for i := 0; i < b.N; i++ {
		sessionIDs[i], _ = sessions.Add(userID, time.Minute)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sessionID := sessionIDs[generateRandomNumber(b.N)] // Randomly pick a session
			sessions.UserID(sessionID)
		}
	})
}

func BenchmarkExtendConcurrent(b *testing.B) {
	sessions := New[TestKey](b.N, 1, newTestKey)
	sessionIDs := make([]TestKey, b.N)
	userID := newTestKey()
	for i := 0; i < b.N; i++ {
		sessionIDs[i], _ = sessions.Add(userID, time.Minute)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sessionID := sessionIDs[generateRandomNumber(b.N)] // Randomly pick a session
			sessions.Extend(sessionID)
		}
	})
}

func BenchmarkRemoveSessionConcurrent(b *testing.B) {
	sessions := New[TestKey](b.N, 1, newTestKey)
	sessionIDs := make([]TestKey, b.N)
	userID := newTestKey()
	for i := 0; i < b.N; i++ {
		sessionIDs[i], _ = sessions.Add(userID, time.Minute)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sessionID := sessionIDs[generateRandomNumber(b.N)] // Randomly pick a session
			sessions.RemoveSession(sessionID)
		}
	})
}