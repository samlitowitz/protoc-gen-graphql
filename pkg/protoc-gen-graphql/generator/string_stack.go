package generator

// An EmptyStackError occurs when attempting to preform an operation
// which is not viable with an empty stack, e.g. `Pop`.
type EmptyStackError struct{}

func (e *EmptyStackError) Error() string {
	return "value expected, stack is empty"
}

// A StringStack is stack of string elements.
type StringStack []string

// Empty determines if a stack contains no elements
func (s StringStack) Empty() bool {
	return len(s) == 0
}

// Peek returns the first element without altering the stack.
func (s StringStack) Peek() string {
	l := len(s)
	return s[l-1]
}

// Push adds a new element to the top of the stack
func (s StringStack) Push(v string) StringStack {
	return append(s, v)
}

// Pop removes the top element of the stack, returning the modified stack,
// the element, and an EmptyStackError if the stack contains no elements.
func (s StringStack) Pop() (StringStack, string, error) {
	l := len(s)
	if l == 0 {
		return s, "", &EmptyStackError{}
	}
	return s[:l-1], s[l-1], nil
}
