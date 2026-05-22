package service

import "fmt"

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func validationError(format string, args ...any) ValidationError {
	return ValidationError{Message: fmt.Sprintf(format, args...)}
}
