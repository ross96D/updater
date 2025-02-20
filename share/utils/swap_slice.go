package utils

// efficient delete slice implementation with no element order garanties
type SwapSlice[T any] struct {
	buffer []T
}

func (s *SwapSlice[T]) Append(e ...T) {
	s.buffer = append(s.buffer, e...)
}

func (s SwapSlice[T]) Len() int {
	return len(s.buffer)
}

func (s SwapSlice[T]) Slice() []T {
	return s.buffer
}

func (s *SwapSlice[T]) Remove(i int) {
	length := len(s.buffer)
	s.buffer[i] = s.buffer[length-1]
	s.buffer = s.buffer[0 : length-1]
}
