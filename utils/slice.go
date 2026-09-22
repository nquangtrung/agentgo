package utils

func Each[T any](input []T, action func(T)) {
	for _, v := range input {
		action(v)
	}
}

func Find[T any](input []T, predicate func(T) bool) (T, bool) {
	for _, v := range input {
		if predicate(v) {
			return v, true
		}
	}
	return *new(T), false
}

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

func Contains[T any](input []T, value T, same func(T, T) bool) bool {
	for _, v := range input {
		if same(v, value) {
			return true
		}
	}
	return false
}

func Unique[T any](input []T, same func(T, T) bool) []T {
	result := make([]T, 0)
	for _, v := range input {
		if !Contains(result, v, same) {
			result = append(result, v)
		}
	}
	return result
}

func Keys[T any](input map[string]T) []string {
	result := make([]string, 0, len(input))
	for k := range input {
		result = append(result, k)
	}
	return result
}

func Values[T any](input map[string]T) []T {
	result := make([]T, 0, len(input))
	for _, v := range input {
		result = append(result, v)
	}
	return result
}
