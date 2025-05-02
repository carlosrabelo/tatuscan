package service

import "fmt"

// Error is a service-layer error with HTTP status hint.
type Error struct {
	Message    string
	StatusCode int
}

func (e *Error) Error() string { return e.Message }

// Validation returns a 400 error.
func Validation(msg string) *Error {
	return &Error{Message: msg, StatusCode: 400}
}

// NotFound returns a 404 error.
func NotFound(resource, id string) *Error {
	return &Error{Message: fmt.Sprintf("%s not found: %s", resource, id), StatusCode: 404}
}

// Database returns a 500 error.
func Database(msg string) *Error {
	return &Error{Message: msg, StatusCode: 500}
}
