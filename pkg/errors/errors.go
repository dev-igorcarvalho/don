// Package apperr provides a custom error type for the application,
// supporting machine-readable codes, human-readable messages,
// and contextual details including operation tracing.
package apperr

import (
	"errors"
	"strings"
)

// ErrorDetail contains contextual information about an error,
// including the operation where it occurred and specific details.
type ErrorDetail struct {
	Detail string `json:"detail,omitempty"`
	Op     string `json:"op,omitempty"`
}

// AppError represents an application-specific error.
// It carries a code for machine-level handling, a message for humans,
// and an optional underlying error.
type AppError struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Detail     ErrorDetail `json:"detail,omitempty"`
	Err        error       `json:"-"`
	HttpStatus int         `json:"-"`
}

// Copy returns a deep copy of the AppError.
func (e *AppError) Copy() *AppError {
	newErr := *e
	return &newErr
}

// WithDetail returns a copy of the error with the provided ErrorDetail.
func (e *AppError) WithDetail(detail ErrorDetail) *AppError {
	newErr := e.Copy()
	newErr.Detail = detail
	return newErr
}

// Is reports whether the error has the same code as the target error.
// It implements the errors.Is interface.
func (e *AppError) Is(err error) bool {
	var ae *AppError
	if errors.As(err, &ae) {
		return e.Code == ae.Code
	}
	return false
}

// Error returns a formatted string representation of the error.
// The pattern is: [CODE] Message (Detail) - Op: WrappedErr
func (e *AppError) Error() string {
	var sb strings.Builder

	if e.Code != "" {
		sb.WriteString("[")
		sb.WriteString(strings.ToUpper(e.Code))
		sb.WriteString("] ")
	}

	sb.WriteString(e.Message)

	if e.Detail.Detail != "" {
		sb.WriteString(" (")
		sb.WriteString(e.Detail.Detail)
		sb.WriteString(")")
	}

	if e.Detail.Op != "" || e.Err != nil {
		sb.WriteString(" -")
		if e.Detail.Op != "" {
			sb.WriteString(" ")
			sb.WriteString(e.Detail.Op)
			sb.WriteString(":")
		}
		if e.Err != nil {
			sb.WriteString(" ")
			sb.WriteString(e.Err.Error())
		}
	}

	return sb.String()
}

// Unwrap returns the underlying error.
// It implements the errors.Unwrap interface.
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithErr returns a copy of the error with a new underlying error attached.
func (e *AppError) WithErr(wrappedErr error) *AppError {
	newErr := e.Copy()
	newErr.Err = wrappedErr
	return newErr
}

// As finds the first error in err's chain that matches target.
// It is a helper that wraps errors.As.
func As(err error, target any) bool {
	return errors.As(err, target)
}
