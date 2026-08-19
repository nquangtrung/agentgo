package utils

func Filter[T any](input []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, v := range input {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}
func Map[T any, U any](input []T, mapper func(T) U) []U {
	result := make([]U, len(input))
	for i, v := range input {
		result[i] = mapper(v)
	}
	return result
}
func Reduce[T any, U any](input []T, reducer func(U, T) U, initial U) U {
	result := initial
	for _, v := range input {
		result = reducer(result, v)
	}
	return result
}
