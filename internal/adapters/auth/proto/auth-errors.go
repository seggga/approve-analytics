/*
	This file holds supported error codes and corresponding
	messages
*/

package proto

const (
	// NoError means there was no error during authentication
	NoError = 0
	// TokenNotValid ...
	TokenNotValid = 1
	// UserNotFound means that storage has no data about a user with given login
	UserNotFound = 2
	// InternalServerError returns on that storage has no data about a user with given login
	InternalServerError = 3
)

// Error is a wrapper for
func Error(i int) string {

	switch i {
	case TokenNotValid:
		return "token is not valid"
	case UserNotFound:
		return "user with given login has not been found"
	case InternalServerError:
		return "internal server error"
	default:
		return ""
	}
}
