package stack

import (
	"testing"

)

func TestStack(t *testing.T) {
	s := New[int]()

	// Test Size on empty stack
	if s.Size() != 0 {
		t.Errorf("Expected size 0, got %d", s.Size())
	}

	// Test Peek on empty stack
	if s.Peek() != 0 {
		t.Errorf("Expected peek 0, got %d", s.Peek())
	}

	// Test Push
	s.Push(1)
	if s.Size() != 1 {
		t.Errorf("Expected size 1, got %d", s.Size())
	}
	if s.Peek() != 1 {
		t.Errorf("Expected peek 1, got %d", s.Peek())
	}

	s.Push(2)
	if s.Size() != 2 {
		t.Errorf("Expected size 2, got %d", s.Size())
	}
	if s.Peek() != 2 {
		t.Errorf("Expected peek 2, got %d", s.Peek())
	}

	// Test Pop
	val := s.Pop()
	if val != 2 {
		t.Errorf("Expected pop 2, got %d", val)
	}
	if s.Size() != 1 {
		t.Errorf("Expected size 1, got %d", s.Size())
	}
	if s.Peek() != 1 {
		t.Errorf("Expected peek 1, got %d", s.Peek())
	}

	val = s.Pop()
	if val != 1 {
		t.Errorf("Expected pop 1, got %d", val)
	}
	if s.Size() != 0 {
		t.Errorf("Expected size 0, got %d", s.Size())
	}
	if s.Peek() != 0 {
		t.Errorf("Expected peek 0, got %d", s.Peek())
	}

	// Test Pop on empty stack
	val = s.Pop()
	if val != 0 {
		t.Errorf("Expected pop 0, got %d", val)
	}
	if s.Size() != 0 {
		t.Errorf("Expected size 0, got %d", s.Size())
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New[int]()
	}
}

func BenchmarkPush(b *testing.B) {
	s := New[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(i)
	}
}

func BenchmarkPop(b *testing.B) {
	s := New[int]()
	for i := 0; i < b.N; i++ {
		s.Push(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Pop()
	}
}

func BenchmarkPeek(b *testing.B) {
	s := New[int]()
	s.Push(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Peek()
	}
}

func BenchmarkSize(b *testing.B) {
	s := New[int]()
	s.Push(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Size()
	}
}
