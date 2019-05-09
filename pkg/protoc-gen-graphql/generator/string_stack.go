package generator

type EmptyStackError struct{}

func (e *EmptyStackError) Error() string {
	return "value expected, stack is empty"
}

type StringStack []string

func (s StringStack) Empty() bool {
	return len(s) == 0
}

func (s StringStack) Peek() string {
	l := len(s)
	return s[l-1]
}

func (s StringStack) Push(v string) StringStack {
	return append(s, v)
}

func (s StringStack) Pop() (StringStack, string, error) {
	l := len(s)
	if l == 0 {
		return s, "", &EmptyStackError{}
	}
	return s[:l-1], s[l-1], nil
}
