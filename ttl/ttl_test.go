// Copyright 2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ttl

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLatency(t *testing.T) {

	var arrivedAt time.Time
	list := New(func(key int) {
		arrivedAt = time.Now()
	})
	defer list.Stop()

	expectedAt := time.Now().Add(1 * time.Second)
	var err error
	err = list.Put(42, 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	<-list.Wait()
	if arrivedAt.Sub(expectedAt) > 100*time.Millisecond {
		t.Fatalf("list is too late: %v", arrivedAt.Sub(expectedAt))
	}
}

func TestLen(t *testing.T) {
	list := New[int](func(key int) {})
	defer list.Stop()
	var err error
	err = list.Put(1, 1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	err = list.Put(2, 1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	err = list.Put(3, 1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if l := list.Len(); l != 3 {
		t.Fatalf("len failed, expected 3, got %v", l)
	}
}

func TestPut(t *testing.T) {

	var list = New(func(key int) {})
	var err error

	err = list.Put(42, 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	err = list.Put(42, 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	err = list.Put(42, 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	err = list.Put(42, 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	err = list.Put(42, 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if list.Len() != 1 {
		t.Fatal("put failed")
	}
	err = list.Delete(42)
	if err != nil {
		t.Fatal(err)
	}
	err = list.Delete(42)
	if err != ErrNotFound {
		t.Fatal("put failed: expected ErrNotFound")
	}

	err = list.Put(42, 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	<-list.Wait()

	list.Stop()

	err = list.Put(42, 1*time.Second)
	if err != ErrNotRunning {
		t.Fatal("expected ErrNotRunning")
	}
}

func TestDelete(t *testing.T) {
	list := New[int](func(key int) {})
	var err error

	err = list.Put(42, 1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	err = list.Delete(42)
	if err != nil {
		t.Fatal(err)
	}
	err = list.Delete(69)
	if err != ErrNotFound {
		t.Fatal("expected ErrNotFound")
	}
	err = list.Stop()
	if err != nil {
		t.Fatal(err)
	}
	err = list.Delete(42)
	if err != ErrNotRunning {
		t.Fatal("expected ErrNotRunning")
	}
}

func TestPutAscending(t *testing.T) {
	list := New[int](func(key int) {})
	var err error
	for i := 0; i < 10; i++ {
		err = list.Put(i, time.Duration(i)*time.Hour)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestPutDescending(t *testing.T) {
	list := New[int](func(key int) {})
	var err error
	for i := 0; i < 10; i++ {
		err = list.Put(i+1, time.Duration(10-i+1)*time.Hour)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestPutRandom(t *testing.T) {
	list := New[int](func(key int) {})
	keys := rand.Perm(10)
	var err error
	for i := 0; i < 10; i++ {
		err = list.Put(keys[i], time.Duration(keys[i])*time.Hour)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestRandom(t *testing.T) {
	const numLoops int = 1e3
	var wg sync.WaitGroup
	list := New[int](func(key int) {
		wg.Done()
	})
	defer list.Stop()
	keys := rand.Perm(numLoops)
	wg.Add(numLoops)
	var err error
	for i := 0; i < numLoops; i++ {
		err = list.Put(keys[i], time.Duration(keys[i])*time.Millisecond)
		if err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()
}

func BenchmarkPutAscending(b *testing.B) {
	list := New[int](func(key int) {})
	b.ResetTimer()
	var err error
	for i := 0; i < b.N; i++ {
		err = list.Put(i, time.Duration(i)*time.Hour)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPutDescending(b *testing.B) {
	list := New[int](func(key int) {})
	b.ResetTimer()
	var err error
	for i := 0; i < b.N; i++ {
		err = list.Put(i+1, time.Duration(10-i+1)*time.Hour)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPutRandom(b *testing.B) {
	list := New[int](func(key int) {})
	keys := rand.Perm(b.N + 1)
	b.ResetTimer()
	var err error
	for i := 0; i < b.N; i++ {
		err = list.Put(keys[i], time.Duration(keys[i])*time.Hour)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestEmptyTTL(t *testing.T) {
	list := New[int](func(key int) {})
	defer list.Stop()

	waitChan := list.Wait()
	select {
	case <-waitChan:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Wait did not return immediately for an empty TTL")
	}
}

func TestOverwrite(t *testing.T) {
	list := New[int](func(key int) {})
	defer list.Stop()

	var err error
	err = list.Put(1, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	err = list.Put(1, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(150 * time.Millisecond)
	if list.Len() != 1 {
		t.Fatal("Key should still be in TTL")
	}

	time.Sleep(100 * time.Millisecond)
	<-list.Wait()

	if list.Len() != 0 {
		t.Fatal("Key should have expired")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	list := New[int](func(key int) {})
	defer list.Stop()

	var err error
	err = list.Delete(1)
	if err != ErrNotFound {
		t.Fatalf("Expected ErrNotFound, got %v", err)
	}
}

func TestConcurrentPutDelete(t *testing.T) {
	list := New[int](func(key int) {})
	defer list.Stop()

	var wg sync.WaitGroup
	numOps := 100

	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()
			var err error
			err = list.Put(key, 50*time.Millisecond)
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(time.Duration(rand.IntN(10)) * time.Millisecond) // Introduce some delay
			err = list.Delete(key)
			if err != nil && err != ErrNotFound {
				t.Fatal(err)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond) // Give time for all operations to complete
	if list.Len() != 0 {
		t.Fatalf("TTL should be empty, but has length %d", list.Len())
	}
}

func TestMultipleWaiters(t *testing.T) {
	list := New[int](func(key int) {})
	defer list.Stop()

	numWaiters := 5
	waitChans := make([]chan time.Time, numWaiters)
	for i := 0; i < numWaiters; i++ {
		waitChans[i] = list.Wait()
	}

	// All waiters should receive a signal immediately since the TTL is empty
	for i := 0; i < numWaiters; i++ {
		select {
		case <-waitChans[i]:
			// Expected behavior
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Waiter %d did not receive signal", i)
		}
	}
}

func TestZeroDuration(t *testing.T) {
	list := New[int](func(key int) {})
	defer list.Stop()

	var err error
	err = list.Put(1, 0*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	<-time.After(50 * time.Millisecond)
	if list.Len() != 0 {
		t.Fatal("Key with zero duration should have expired immediately")
	}
}

func TestNegativeDuration(t *testing.T) {
	list := New[int](func(key int) {})
	defer list.Stop()

	var err error
	err = list.Put(1, -1*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	<-time.After(50 * time.Millisecond)
	if list.Len() != 0 {
		t.Fatal("Key with negative duration should have expired immediately")
	}
}

func TestNilCallback(t *testing.T) {
	list := New[int](nil)
	defer list.Stop()

	var err error
	err = list.Put(1, 10*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(50 * time.Millisecond)
	if list.Len() != 0 {
		t.Fatal("Key should have expired")
	}
}

func TestAddAfterStop(t *testing.T) {
	list := New[int](nil)
	err := list.Stop()
	if err != nil {
		t.Fatal(err)
	}
	err = list.Put(1, time.Second)
	if err != ErrNotRunning {
		t.Fatal("Expected ErrNotRunning on Put after Stop")
	}
}

func TestDeleteAfterStop(t *testing.T) {
	list := New[int](nil)
	err := list.Stop()
	if err != nil {
		t.Fatal(err)
	}
	err = list.Delete(1)
	if err != ErrNotRunning {
		t.Fatal("Expected ErrNotRunning on Delete after Stop")
	}
}

func TestNewAndStop(t *testing.T) {
	var ttl *TTL[string]
	ttl = New[string](nil)
	defer func() {
		if err := ttl.Stop(); err != ErrNotRunning {
			t.Fatalf("Failed to stop TTL: %v", err)
		}
	}()

	if !ttl.running.Load() {
		t.Error("TTL should be running after New()")
	}

	err := ttl.Stop()
	if err != nil {
		t.Fatalf("Stop returned an error: %v", err)
	}

	if ttl.running.Load() {
		t.Error("TTL should not be running after Stop()")
	}

	err = ttl.Stop()
	if err != ErrNotRunning {
		t.Errorf("Second Stop should return ErrNotRunning, got: %v", err)
	}
}

func TestPut_reset(t *testing.T) {
	var (
		ttl    *TTL[string]
		key    = "testKey"
		dur1   = 100 * time.Millisecond
		dur2   = 200 * time.Millisecond
		exists bool
		err    error
	)
	ttl = New[string](nil)
	defer func() {
		if err = ttl.Stop(); err != nil {
			t.Fatalf("Failed to stop TTL: %v", err)
		}
	}()

	err = ttl.Put(key, dur1)
	if err != nil {
		t.Fatalf("Put returned an error: %v", err)
	}

	time.Sleep(dur1 / 2)

	err = ttl.Put(key, dur2)
	if err != nil {
		t.Fatalf("Put (reset) returned an error: %v", err)
	}

	time.Sleep(dur1 / 2)

	exists = ttl.Exists(key)
	if !exists {
		t.Error("Key should still exist after first timeout period because of reset.")
	}

	time.Sleep(dur2)

	exists = ttl.Exists(key)
	if exists {
		t.Error("Key should not exist after second timeout period.")
	}
}

func TestCallback(t *testing.T) {
	var (
		key     = "testKey"
		dur     = 100 * time.Millisecond
		cbCalled atomic.Bool
		err      error
		wg       sync.WaitGroup
	)

	wg.Add(1)
	var ttl *TTL[string]
	ttl = New[string](func(k string) {
		if k == key {
			cbCalled.Store(true)
			wg.Done()
		}
	})

	defer func() {
		if err = ttl.Stop(); err != nil {
			t.Fatalf("Failed to stop TTL: %v", err)
		}
	}()

	err = ttl.Put(key, dur)
	if err != nil {
		t.Fatalf("Put returned an error: %v", err)
	}

	wg.Wait()

	if !cbCalled.Load() {
		t.Error("Callback should have been called")
	}
}

func TestWait(t *testing.T) {
	var (
		ttl *TTL[string]
		dur = 100 * time.Millisecond
		wg  sync.WaitGroup
		err error
	)

	ttl = New[string](nil)
	defer func() {
		if err = ttl.Stop(); err != nil {
			t.Fatalf("Failed to stop TTL: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		c := ttl.Wait()
		<-c
	}()

	time.Sleep(10 * time.Millisecond)

	err = ttl.Put("key1", dur)
	if err != nil {
		t.Fatalf("Put returned an error: %v", err)
	}

	time.Sleep(dur * 2)

	wg.Wait()
}

func TestConcurrentAccess(t *testing.T) {
	var (
		ttl       *TTL[int]
		numRoutines = 100
		numOps      = 1000
		dur         = 10 * time.Millisecond
		wg          sync.WaitGroup
		err         error
	)

	ttl = New[int](nil)
	defer func() {
		if err = ttl.Stop(); err != nil {
			t.Fatalf("Failed to stop TTL: %v", err)
		}
	}()

	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := rand.IntN(100)
				switch rand.IntN(4) {
				case 0:
					err = ttl.Put(key, dur)
					if err != nil && err != ErrNotRunning {
						t.Errorf("Routine %d: Put returned an error: %v", routineID, err)
					}
				case 1:
					ttl.Exists(key)
				case 2:
					err = ttl.Delete(key)
					if err != nil && err != ErrNotRunning && err != ErrNotFound {
						t.Errorf("Routine %d: Delete returned an error: %v", routineID, err)
					}
				case 3:
					ttl.Len()
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestEarlyTimeout(t *testing.T) {
	var (
		ttl *TTL[string]
		key = "testKey"
		dur = -100 * time.Millisecond // Expired before Put
		cbCalled atomic.Bool
		wg       sync.WaitGroup
		err      error
	)

	wg.Add(1)
	ttl = New[string](func(k string) {
		if k == key {
			cbCalled.Store(true)
			wg.Done()
		}
	})
	defer func() {
		if err = ttl.Stop(); err != nil {
			t.Fatalf("Failed to stop TTL: %v", err)
		}
	}()

	err = ttl.Put(key, dur)
	if err != nil {
		t.Fatalf("Put returned an error: %v", err)
	}

	wg.Wait()

	if !cbCalled.Load() {
		t.Error("Callback should have been called immediately")
	}

	if ttl.Exists(key) {
		t.Error("Key should not exist after immediate timeout")
	}
}

func TestMultipleTimeouts(t *testing.T) {
	var (
		ttl       *TTL[int]
		numKeys   = 10
		dur       = 50 * time.Millisecond
		callbacks = make([]atomic.Bool, numKeys)
		wg        sync.WaitGroup
		err       error
	)

	wg.Add(numKeys)
	ttl = New[int](func(key int) {
		callbacks[key].Store(true)
		wg.Done()
	})
	defer func() {
		if err = ttl.Stop(); err != nil {
			t.Fatalf("Failed to stop TTL: %v", err)
		}
	}()

	for i := 0; i < numKeys; i++ {
		err = ttl.Put(i, dur)
		if err != nil {
			t.Fatalf("Put returned an error: %v", err)
		}
	}

	wg.Wait()

	for i := 0; i < numKeys; i++ {
		if !callbacks[i].Load() {
			t.Errorf("Callback for key %d should have been called", i)
		}
		if ttl.Exists(i) {
			t.Errorf("Key %d should not exist after timeout", i)
		}
	}
}

func TestDeleteBeforeTimeout(t *testing.T) {
	var (
		ttl    *TTL[string]
		key    = "testKey"
		dur    = 200 * time.Millisecond
		exists bool
		err    error
	)

	ttl = New[string](nil)
	defer func() {
		if err = ttl.Stop(); err != nil {
			t.Fatalf("Failed to stop TTL: %v", err)
		}
	}()

	err = ttl.Put(key, dur)
	if err != nil {
		t.Fatalf("Put returned an error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	err = ttl.Delete(key)
	if err != nil {
		t.Fatalf("Delete returned an error: %v", err)
	}

	exists = ttl.Exists(key)
	if exists {
		t.Error("Key should not exist after Delete()")
	}

	time.Sleep(dur)

	// Check that callback was not called (if any)
}

func TestWait_empty(t *testing.T) {
	var (
		ttl *TTL[string]
		err error
	)

	ttl = New[string](nil)
	defer func() {
		if err = ttl.Stop(); err != nil {
			t.Fatalf("Failed to stop TTL: %v", err)
		}
	}()

	c := ttl.Wait()
	select {
	case <-c:
		// ok
	case <-time.After(10 * time.Millisecond):
		t.Fatal("Wait channel should have been immediately closed")
	}
}

func TestWaitMultiple(t *testing.T) {
	var (
		ttl *TTL[string]
		dur = 100 * time.Millisecond
		wg  sync.WaitGroup
		err error
	)

	ttl = New[string](nil)
	defer func() {
		if err = ttl.Stop(); err != nil {
			t.Fatalf("Failed to stop TTL: %v", err)
		}
	}()

	numWaiters := 5
	wg.Add(numWaiters)

	for i := 0; i < numWaiters; i++ {
		go func() {
			defer wg.Done()
			c := ttl.Wait()
			<-c
		}()
	}

	time.Sleep(10 * time.Millisecond)

	err = ttl.Put("key1", dur)
	if err != nil {
		t.Fatalf("Put returned an error: %v", err)
	}

	time.Sleep(dur * 2)

	wg.Wait()
}

func TestDelete_notRunning(t *testing.T) {
	var (
		ttl *TTL[string]
		err error
	)
	ttl = New[string](nil)

	err = ttl.Stop()
	if err != nil {
		t.Fatalf("Stop returned an error: %v", err)
	}

	err = ttl.Delete("key")
	if err != ErrNotRunning {
		t.Fatalf("Expected ErrNotRunning, got: %v", err)
	}
}

func TestPut_notRunning(t *testing.T) {
	var (
		ttl *TTL[string]
		err error
	)
	ttl = New[string](nil)

	err = ttl.Stop()
	if err != nil {
		t.Fatalf("Stop returned an error: %v", err)
	}

	err = ttl.Put("key", time.Second)
	if err != ErrNotRunning {
		t.Fatalf("Expected ErrNotRunning, got: %v", err)
	}
}

func TestInsertTimeout_duplicates(t *testing.T) {
	var (
		ttl *TTL[string]
		dur = 100 * time.Millisecond
		key = "testKey"
		err error
	)

	ttl = New[string](nil)
	defer func() {
		if err = ttl.Stop(); err != nil {
			t.Fatalf("Failed to stop TTL: %v", err)
		}
	}()

	// Add the same key multiple times with different timeouts.
	err = ttl.Put(key, dur*1)
	if err != nil {
		t.Fatalf("Put returned an error: %v", err)
	}

	err = ttl.Put(key, dur*2)
	if err != nil {
		t.Fatalf("Put returned an error: %v", err)
	}

	err = ttl.Put(key, dur*3)
	if err != nil {
		t.Fatalf("Put returned an error: %v", err)
	}

	// The key should only timeout once, at the latest timeout specified.
	time.Sleep(dur * 2)
	if !ttl.Exists(key) {
		t.Fatalf("key %v should exist", key)
	}
	time.Sleep(dur * 2)
	if ttl.Exists(key) {
		t.Fatalf("key %v should not exist", key)
	}
}

func BenchmarkLen(b *testing.B) {
	list := New[int](func(key int) {})
	defer list.Stop()
	var err error
	for i := 0; i < 100; i++ {
		err = list.Put(i, time.Hour)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.Len()
	}
}

func BenchmarkDelete(b *testing.B) {
	list := New[int](func(key int) {})
	defer list.Stop()
	var err error
	for i := 0; i < b.N; i++ {
		err = list.Put(i, time.Hour)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = list.Delete(i)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWait(b *testing.B) {
	list := New[int](func(key int) {})
	defer list.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-list.Wait()
	}
}

func TestExists(t *testing.T) {
	list := New[int](func(key int) {})
	defer list.Stop()
	var err error

	// Test key exists after put
	err = list.Put(1, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !list.Exists(1) {
		t.Fatal("Key should exist")
	}

	// Test key does not exist after delete
	err = list.Delete(1)
	if err != nil {
		t.Fatal(err)
	}
	if list.Exists(1) {
		t.Fatal("Key should not exist")
	}

	// Test key does not exist initially
	if list.Exists(2) {
		t.Fatal("Key should not exist initially")
	}

	// Test after stop
	err = list.Put(3, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Millisecond)

	if list.Exists(3) {
		t.Fatal("Key should not exist after timeout")
	}
}

func BenchmarkExists(b *testing.B) {
	list := New[int](func(key int) {})
	defer list.Stop()
	var err error
	for i := 0; i < b.N; i++ {
		err = list.Put(i, time.Hour)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.Exists(i)
	}
}

func BenchmarkPut(b *testing.B) {
	var (
		ttl *TTL[int]
		err error
	)

	ttl = New[int](nil)
	defer func() {
		if err = ttl.Stop(); err != nil {
			b.Fatalf("Failed to stop TTL: %v", err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = ttl.Put(i, time.Second)
		if err != nil {
			b.Fatalf("Put returned an error: %v", err)
		}
	}
}

func BenchmarkConcurrentPut(b *testing.B) {
	var (
		ttl         *TTL[int]
		numRoutines = 10
		dur         = time.Second
		wg          sync.WaitGroup
		err         error
	)

	ttl = New[int](nil)
	defer func() {
		if err = ttl.Stop(); err != nil {
			b.Fatalf("Failed to stop TTL: %v", err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(numRoutines)
		for j := 0; j < numRoutines; j++ {
			go func(key int) {
				defer wg.Done()
				err = ttl.Put(key, dur)
				if err != nil {
					fmt.Printf("Put returned an error: %v", err)
				}
			}(i*numRoutines + j)
		}
		wg.Wait()
	}
}
