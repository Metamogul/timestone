package system

func ptr[T any](t T) *T {
	return &t
}
