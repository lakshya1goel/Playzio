package domain

import "fmt"

type HttpError struct {
	StatusCode int
	Message    string
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("status %d: %s", e.StatusCode, e.Message)
}

func NewHttpError(code int, message string) *HttpError {
	return &HttpError{
		StatusCode: code,
		Message:    message,
	}
}
