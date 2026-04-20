package domain

import (
	"github.com/dev-igorcarvalho/don/pkg/errors"
)

// Error codes follow the pattern: E[ACRONYM][0000]
// E: Indicates an error.
// DON: The acronym for the Don API.
// 0000: A 4-digit incrementing identifier for the specific error.
var (
	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = &apperr.AppError{
		Code:       "EDON0001",
		Message:    "The requested resource was not found",
		HttpStatus: 404,
	}

	// ErrInternal is returned when an unexpected internal error occurs.
	ErrInternal = &apperr.AppError{
		Code:       "EDON0002",
		Message:    "An unexpected internal error occurred",
		HttpStatus: 500,
	}

	// ErrInvalidInput is returned when the input provided is invalid.
	ErrInvalidInput = &apperr.AppError{
		Code:       "EDON0003",
		Message:    "The provided input is invalid",
		HttpStatus: 400,
	}

	// ErrUnauthorized is returned when the user is not authorized to perform an action.
	ErrUnauthorized = &apperr.AppError{
		Code:       "EDON0004",
		Message:    "You are not authorized to perform this action",
		HttpStatus: 401,
	}

	// ErrForbidden is returned when the user is forbidden from performing an action.
	ErrForbidden = &apperr.AppError{
		Code:       "EDON0005",
		Message:    "You are forbidden from performing this action",
		HttpStatus: 403,
	}

	// ErrConflict is returned when there is a conflict with the current state of the resource.
	ErrConflict = &apperr.AppError{
		Code:       "EDON0006",
		Message:    "There is a conflict with the current state of the resource",
		HttpStatus: 409,
	}
)
