package utils

func IsTransientError(err error) bool {
	if err == nil {
		return false
	}

	// Check if the error is a network error or a timeout error
	if _, ok := err.(interface{ Timeout() bool }); ok {
		return true
	}
	if _, ok := err.(interface{ Temporary() bool }); ok {
		return true
	}

	return false
}
