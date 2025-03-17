// Copyright 2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ttl

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"
	"time"
)

func TestLatency(t *testing.T) {

	var arrivedAt time.Time
	list := New(func(key int) {
		arrivedAt = time.Now()
		fmt.Printf("timed out: %d\n", key)
	})
	defer list.Stop()

	expectedAt := time.Now().Add(1 * time.Second)
	list.Put(42, 1*time.Second)

	<-list.Wait()
	fmt.Printf("list is %v late\n", arrivedAt.Sub(expectedAt))
}

func TestLen(t *testing.T) {
	list := New[int](func(key int) {})
	defer list.Stop()
	list.Put(1, 1*time.Hour)
	list.Put(2, 1*time.Hour)
	list.Put(3, 1*time.Hour)
	if l := list.Len(); l != 3 {
		t.Fatalf("len failed, expected 3, got %v", l)
	}
}

func TestPut(t *testing.T) {

	var list = New(func(key int) {})

	if err := list.Put(42, 1*time.Second); err != nil {
		t.Fatal(err)
	}
	if err := list.Put(42, 1*time.Second); err != nil {
		t.Fatal(err)
	}
	if err := list.Put(42, 1*time.Second); err != nil {
		t.Fatal(err)
	}
	if err := list.Put(42, 1*time.Second); err != nil {
		t.Fatal(err)
	}
	if err := list.Put(42, 1*time.Second); err != nil {
		t.Fatal(err)
	}
	if list.Len() != 1 {
		t.Fatal("put failed")
	}
	if err := list.Delete(42); err != nil {
		t.Fatal(err)
	}
	if err := list.Delete(42); err != ErrNotFound {
		t.Fatal("put failed: expected ErrNotFound")
	}

	if err := list.Put(42, 1*time.Second); err != nil {
		t.Fatal(err)
	}
	<-list.Wait()

	list.Stop()

	if err := list.Put(42, 1*time.Second); err != ErrNotRunning {
		t.Fatal("expected ErrNotRunning")
	}
}

func TestDelete(t *testing.T) {
	list := New[int](func(key int) {})

	list.Put(42, 1*time.Hour)
	if err := list.Delete(42); err != nil {
		t.Fatal(err)
	}
	if err := list.Delete(69); err != ErrNotFound {
		t.Fatal("expected ErrNotFound")
	}
	list.Stop()
	if err := list.Delete(42); err != ErrNotRunning {
		t.Fatal("expected ErrNotRunning")
	}
}

func TestPutAscending(t *testing.T) {
	list := New[int](func(key int) {})
	for i := 0; i < 10; i++ {
		list.Put(i, time.Duration(i)*time.Hour)
	}
}

func TestPutDescending(t *testing.T) {
	list := New[int](func(key int) {})
	for i := 0; i < 10; i++ {
		list.Put(i+1, time.Duration(10-i+1)*time.Hour)
	}
}

func TestPutRandom(t *testing.T) {
	list := New[int](func(key int) {})
	keys := rand.Perm(10)
	for i := 0; i < 10; i++ {
		list.Put(keys[i], time.Duration(keys[i])*time.Hour)
	}
}

func TestRandom(t *testing.T) {
	const numLoops int = 1e3
	var wg sync.WaitGroup
	list := New[int](func(key int) {
		fmt.Printf("timed out: %d\n", key)
		wg.Done()
	})
	defer list.Stop()
	keys := rand.Perm(numLoops)
	wg.Add(numLoops) // Increment the counter before putting the keys
	for i := 0; i < numLoops; i++ {
		list.Put(keys[i], time.Duration(keys[i])*time.Millisecond)
	}
	wg.Wait()
}

func BenchmarkPutAscending(b *testing.B) {
	list := New[int](func(key int) {})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.Put(i, time.Duration(i)*time.Hour)
	}
}

func BenchmarkPutDescending(b *testing.B) {
	list := New[int](func(key int) {})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.Put(i+1, time.Duration(10-i+1)*time.Hour)
	}
}

func BenchmarkPutRandom(b *testing.B) {
	list := New[int](func(key int) {})
	keys := rand.Perm(b.N + 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.Put(keys[i], time.Duration(keys[i])*time.Hour)
	}
}

func TestEmptyTTL(t *testing.T) {
	list := New[int](func(key int) {})
	defer list.Stop()

	waitChan := list.Wait()
	select {
	case <-waitChan:
		// Expected behavior
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Wait did not return immediately for an empty TTL")
	}
}

func TestOverwrite(t *testing.T) {
	list := New[int](func(key int) {})
	defer list.Stop()

	list.Put(1, 100*time.Millisecond)
	list.Put(1, 200*time.Millisecond)

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

	err := list.Delete(1)
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
			list.Put(key, 50*time.Millisecond)
			time.Sleep(time.Duration(rand.IntN(10)) * time.Millisecond) // Introduce some delay
			list.Delete(key)
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

	list.Put(1, 0*time.Second)

	<-time.After(50 * time.Millisecond)
	if list.Len() != 0 {
		t.Fatal("Key with zero duration should have expired immediately")
	}
}

func TestNegativeDuration(t *testing.T) {
	list := New[int](func(key int) {})
	defer list.Stop()

	list.Put(1, -1*time.Second)

	<-time.After(50 * time.Millisecond)
	if list.Len() != 0 {
		t.Fatal("Key with negative duration should have expired immediately")
	}
}

func TestNilCallback(t *testing.T) {
	list := New[int](nil)
	defer list.Stop()

	list.Put(1, 10*time.Millisecond)
	<-time.After(50 * time.Millisecond)
	if list.Len() != 0 {
		t.Fatal("Key should have expired")
	}
}

func TestAddAfterStop(t *testing.T) {
    list := New[int](nil)
    list.Stop()
    err := list.Put(1, time.Second)
    if err != ErrNotRunning {
        t.Fatal("Expected ErrNotRunning on Put after Stop")
    }
}

func TestDeleteAfterStop(t *testing.T) {
    list := New[int](nil)
    list.Stop()
    err := list.Delete(1)
    if err != ErrNotRunning {
        t.Fatal("Expected ErrNotRunning on Delete after Stop")
    }
}

