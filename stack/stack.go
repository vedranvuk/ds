package stack

// Stack[T] is a stack of [Node].
// Used for keeping parent/child relationship during HTML tokenization.
type Stack[T any] []T

// NewStack returns a new [Stack[T]].
func New[T any]() *Stack[T] {
	var val = Stack[T](make([]T, 0))
	return &val
}

// Size returns number of strings on the stack.
func (self *Stack[T]) Size() int { return len([]T(*self)) }

// Peek returns the last node added to the stack or nil if queue is empty.
func (self *Stack[T]) Peek() (out T) {
	if self.Size() < 1 {
		return *new(T)
	}
	return []T(*self)[self.Size()-1]
}

// Push pushes a node to the stack.
func (self *Stack[T]) Push(item T) { *self = append(*self, item) }

// Pop removes a node from the stack and returns it.
func (self *Stack[T]) Pop() (out T) {
	var l = self.Size()
	if l < 1 {
		return *new(T)
	}
	out = []T(*self)[l-1]
	*self = []T(*self)[:l-1]
	return
}
