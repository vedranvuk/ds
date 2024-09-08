package ttl

import (
	"math/rand"
	"testing"
	"time"

	"github.com/vedranvuk/strutils"
)

func RandomKey() string {
	return strutils.RandomString(true, true, true, 16)
}

func TestTTL(t *testing.T) {
	var numTimeouts = 0
	var ttl = NewTTL[string](func(key string) {
		numTimeouts++
	})
	defer ttl.Stop()

	const numLoops = 1000
	for i := 0; i < numLoops; i++ {
		ttl.Put(RandomKey(), time.Millisecond*time.Duration(rand.Intn(1000)-rand.Intn(1000)))
	}

	var start = time.Now()
	for numTimeouts != numLoops && time.Since(start) < 5*time.Second {
		time.Sleep(1 * time.Millisecond)
	}

	if numTimeouts != numLoops {
		t.Fatal()
	}
}

func TestTTLRepeat(t *testing.T) {
	var numTimeouts = 0
	var ttl = NewTTL(func(key string) {
		numTimeouts++
	})
	defer ttl.Stop()

	var keys []string
	for i := 0; i < 100; i++ {
		keys = append(keys, RandomKey())
	}

	const numLoops = 1000
	for i := 0; i < numLoops; i++ {
		ttl.Put(keys[rand.Intn(100)], time.Millisecond*time.Duration(rand.Intn(500)-250))
	}
}

func BenchmarkPut(b *testing.B) {
	var ttl = NewTTL[string](nil)
	defer ttl.Stop()

	var keys = make([]string, b.N)
	var durs = make([]time.Duration, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = RandomKey()
		durs[i] = time.Millisecond * time.Duration(rand.Intn(1000)-rand.Intn(1000))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ttl.Put(keys[i], durs[i])
	}
	b.StopTimer()
}

func BenchmarkDelete(b *testing.B) {
	var ttl = NewTTL[string](nil)
	defer ttl.Stop()

	var keys = make([]string, b.N)
	var durs = make([]time.Duration, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = RandomKey()
		durs[i] = 10 * time.Second
		ttl.Put(keys[i], durs[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ttl.Delete(keys[i])
	}
	b.StopTimer()
}
