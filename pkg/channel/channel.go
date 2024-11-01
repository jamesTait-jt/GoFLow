package channel

func NewSendOnly[T any](ch chan T) chan<- T {
	return ch
}

func NewReadOnly[T any](ch chan T) <-chan T {
	return ch
}
