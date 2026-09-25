package utils

import "reflect"

func Ternary[T any](cond bool, valueIfTrue T, valueIfFalse T) T {
	if cond {
		return valueIfTrue
	} else {
		return valueIfFalse
	}
}

func IsType(a, b any) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(b)
}
