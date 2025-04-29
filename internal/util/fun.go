package util

func Range[T any](n int, f func(int) T) []T {
	out := make([]T, n)
	for i := range n {
		out[i] = f(i)
	}
	return out
}

func Map[T, R any](slice []T, f func(T) R) []R {
	out := make([]R, len(slice))
	for i, t := range slice {
		out[i] = f(t)
	}
	return out
}
