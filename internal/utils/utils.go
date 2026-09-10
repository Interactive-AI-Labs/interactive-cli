package utils

// ToPtr returns a pointer to the given value.
func ToPtr[T any](v T) *T { return &v }

// NilIfZero returns a pointer to v, or nil when v is the zero value.
func NilIfZero[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}
