package game

// Returns a pointer to the given value (useful for getting a pointer to a temporary).
func ptr[T any](value T) *T {
	return &value
}
