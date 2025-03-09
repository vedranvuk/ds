package graph

import (
	"math/rand"
	"sync"
	"testing"
)

func TestGraph(t *testing.T) {
	t.Run("Link/Linked/Unlink", func(t *testing.T) {
		g := NewGraph[int]()

		// Initially not linked
		if g.Linked(1, 2) {
			t.Error("should not be linked")
		}

		// Link and check
		g.Link(1, 2)
		if !g.Linked(1, 2) {
			t.Error("should be linked")
		}
		if g.Linked(2, 1) {
			t.Error("should not be linked")
		}

		// Unlink and check
		g.Unlink(1, 2)
		if g.Linked(1, 2) {
			t.Error("should not be linked")
		}

		// Check wasLinked returns
		linked := g.Link(1, 2)
		if linked {
			t.Error("should return false the first time")
		}
		linked = g.Link(1, 2)
		if !linked {
			t.Error("should return true the second time")
		}
		unlinked := g.Unlink(1, 2)
		if !unlinked {
			t.Error("should return true the first time")
		}
		unlinked = g.Unlink(1, 2)
		if unlinked {
			t.Error("should return false the second time")
		}
	})

	t.Run("Links", func(t *testing.T) {
		g := NewGraph[string]()
		g.Link("a", "b")
		g.Link("a", "c")
		g.Link("a", "b") // Duplicate link, should not affect result

		links := g.Links("a")
		if len(links) != 2 {
			t.Fatalf("expected 2 links, got %d", len(links))
		}

		expected := map[string]bool{"b": true, "c": true}
		for _, link := range links {
			if _, ok := expected[link]; !ok {
				t.Errorf("unexpected link: %s", link)
			}
			delete(expected, link)
		}
		if len(expected) != 0 {
			t.Error("missing links")
		}

		// Test Links returns an empty slice when no links exist
		links = g.Links("d")
		if len(links) != 0 {
			t.Errorf("expected empty slice, got %v", links)
		}
	})

	t.Run("EnumLinks", func(t *testing.T) {
		g := NewGraph[int]()
		g.Link(1, 2)
		g.Link(1, 3)

		seen := make(map[int]bool)
		count := 0
		g.EnumLinks(1, func(link int) bool {
			seen[link] = true
			count++
			return true
		})

		if count != 2 {
			t.Fatalf("expected 2 links, got %d", count)
		}

		if !seen[2] || !seen[3] {
			t.Error("missing links")
		}

		// Test early exit from EnumLinks
		count = 0
		g.EnumLinks(1, func(link int) bool {
			count++
			return false
		})
		if count != 1 {
			t.Errorf("expected 1 call, got %d", count)
		}

		// Test EnumLinks doesn't call the function when no links exist.
		count = 0
		g.EnumLinks(4, func(link int) bool {
			count++
			return true
		})
		if count != 0 {
			t.Errorf("expected 0 call, got %d", count)
		}
	})
}

func TestSyncGraph(t *testing.T) {
	t.Run("Link/Linked/Unlink", func(t *testing.T) {
		g := NewSyncGraph[int]()

		// Initially not linked
		if g.Linked(1, 2) {
			t.Error("should not be linked")
		}

		// Link and check
		g.Link(1, 2)
		if !g.Linked(1, 2) {
			t.Error("should be linked")
		}
		if g.Linked(2, 1) {
			t.Error("should not be linked")
		}

		// Unlink and check
		g.Unlink(1, 2)
		if g.Linked(1, 2) {
			t.Error("should not be linked")
		}

		// Check wasLinked returns
		linked := g.Link(1, 2)
		if linked {
			t.Error("should return false the first time")
		}
		linked = g.Link(1, 2)
		if !linked {
			t.Error("should return true the second time")
		}
		unlinked := g.Unlink(1, 2)
		if !unlinked {
			t.Error("should return true the first time")
		}
		unlinked = g.Unlink(1, 2)
		if unlinked {
			t.Error("should return false the second time")
		}
	})

	t.Run("Links", func(t *testing.T) {
		g := NewSyncGraph[string]()
		g.Link("a", "b")
		g.Link("a", "c")
		g.Link("a", "b") // Duplicate link, should not affect result

		links := g.Links("a")
		if len(links) != 2 {
			t.Fatalf("expected 2 links, got %d", len(links))
		}

		expected := map[string]bool{"b": true, "c": true}
		for _, link := range links {
			if _, ok := expected[link]; !ok {
				t.Errorf("unexpected link: %s", link)
			}
			delete(expected, link)
		}
		if len(expected) != 0 {
			t.Error("missing links")
		}

		// Test Links returns an empty slice when no links exist
		links = g.Links("d")
		if len(links) != 0 {
			t.Errorf("expected empty slice, got %v", links)
		}
	})

	t.Run("EnumLinks", func(t *testing.T) {
		g := NewSyncGraph[int]()
		g.Link(1, 2)
		g.Link(1, 3)

		seen := make(map[int]bool)
		count := 0
		g.EnumLinks(1, func(link int) bool {
			seen[link] = true
			count++
			return true
		})

		if count != 2 {
			t.Fatalf("expected 2 links, got %d", count)
		}

		if !seen[2] || !seen[3] {
			t.Error("missing links")
		}

		// Test early exit from EnumLinks
		count = 0
		g.EnumLinks(1, func(link int) bool {
			count++
			return false
		})
		if count != 1 {
			t.Errorf("expected 1 call, got %d", count)
		}

		// Test EnumLinks doesn't call the function when no links exist.
		count = 0
		g.EnumLinks(4, func(link int) bool {
			count++
			return true
		})
		if count != 0 {
			t.Errorf("expected 0 call, got %d", count)
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		g := NewSyncGraph[int]()
		var wg sync.WaitGroup
		const numRoutines = 100
		const numOps = 100

		wg.Add(numRoutines)

		for i := 0; i < numRoutines; i++ {
			go func(routineID int) {
				defer wg.Done()
				r := rand.New(rand.NewSource(int64(routineID))) // Create a local random source

				for j := 0; j < numOps; j++ {
					a := r.Intn(10)
					b := r.Intn(10)

					// Randomly choose between Link, Unlink, and Linked
					op := r.Intn(3)
					switch op {
					case 0:
						g.Link(a, b)
					case 1:
						g.Unlink(a, b)
					case 2:
						g.Linked(a, b)
					}
				}
			}(i)
		}
		wg.Wait()
	})
}

func BenchmarkGraph_Link(b *testing.B) {
	g := NewGraph[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Link(i, i+1)
	}
}

func BenchmarkGraph_Linked(b *testing.B) {
	g := NewGraph[int]()
	for i := 0; i < 1000; i++ {
		g.Link(i, i+1)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Linked(i%1000, (i%1000)+1)
	}
}

func BenchmarkGraph_Unlink(b *testing.B) {
	g := NewGraph[int]()
	for i := 0; i < 1000; i++ {
		g.Link(i, i+1)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Unlink(i%1000, (i%1000)+1)
	}
}

func BenchmarkGraph_Links(b *testing.B) {
	g := NewGraph[int]()
	for i := 0; i < 1000; i++ {
		g.Link(0, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Links(0)
	}
}

func BenchmarkGraph_EnumLinks(b *testing.B) {
	g := NewGraph[int]()
	for i := 0; i < 1000; i++ {
		g.Link(0, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.EnumLinks(0, func(key int) bool {
			return true
		})
	}
}

func BenchmarkSyncGraph_Link(b *testing.B) {
	g := NewSyncGraph[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Link(i, i+1)
	}
}

func BenchmarkSyncGraph_Linked(b *testing.B) {
	g := NewSyncGraph[int]()
	for i := 0; i < 1000; i++ {
		g.Link(i, i+1)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Linked(i%1000, (i%1000)+1)
	}
}

func BenchmarkSyncGraph_Unlink(b *testing.B) {
	g := NewSyncGraph[int]()
	for i := 0; i < 1000; i++ {
		g.Link(i, i+1)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Unlink(i%1000, (i%1000)+1)
	}
}

func BenchmarkSyncGraph_Links(b *testing.B) {
	g := NewSyncGraph[int]()
	for i := 0; i < 1000; i++ {
		g.Link(0, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Links(0)
	}
}

func BenchmarkSyncGraph_EnumLinks(b *testing.B) {
	g := NewSyncGraph[int]()
	for i := 0; i < 1000; i++ {
		g.Link(0, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.EnumLinks(0, func(key int) bool {
			return true
		})
	}
}

func BenchmarkSyncGraph_Concurrent(b *testing.B) {
	g := NewSyncGraph[int]()
	var wg sync.WaitGroup

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		const numRoutines = 100
		wg.Add(numRoutines)

		for i := 0; i < numRoutines; i++ {
			go func(routineID int) {
				defer wg.Done()
				r := rand.New(rand.NewSource(int64(routineID)))
				for j := 0; j < 10; j++ { // Reduced iterations for benchmark
					a := r.Intn(100)
					b := r.Intn(100)

					op := r.Intn(3)
					switch op {
					case 0:
						g.Link(a, b)
					case 1:
						g.Unlink(a, b)
					case 2:
						g.Linked(a, b)
					}
				}
			}(i)
		}
		wg.Wait()
	}
}
