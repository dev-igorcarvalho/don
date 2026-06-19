package apperr

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Internal test helpers since errors.go does not provide sentinels or a New function.
var (
	errTestNotFound     = &AppError{Code: "not_found", Message: "user not found"}
	errTestUnauthorized = &AppError{Code: "unauthorized", Message: "missing token"}
	errTestInternal     = &AppError{Code: "internal", Message: "internal error"}
	errTestConflict     = &AppError{Code: "conflict", Message: "user exists"}
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expected string
	}{
		{
			name: "full fields",
			err: errTestNotFound.
				WithDetail(ErrorDetail{Op: "UserService.Get", Detail: "id=123"}).
				WithErr(errors.New("db error")),
			expected: "[NOT_FOUND] user not found (id=123) - UserService.Get: db error",
		},
		{
			name:     "simple error",
			err:      errTestUnauthorized,
			expected: "[UNAUTHORIZED] missing token",
		},
		{
			name:     "with wrapped err",
			err:      errTestInternal.WithErr(errors.New("db crash")),
			expected: "[INTERNAL] internal error - db crash",
		},
		{
			name: "with detail only",
			err: errTestConflict.
				WithDetail(ErrorDetail{Detail: "username=igor"}),
			expected: "[CONFLICT] user exists (username=igor)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestAppError_Is(t *testing.T) {
	err1 := &AppError{Code: "not_found"}
	err2 := &AppError{Code: "not_found"}
	err3 := &AppError{Code: "bad_request"}

	assert.True(t, err1.Is(err1))
	assert.True(t, err1.Is(err2))
	assert.False(t, err1.Is(err3))
	assert.False(t, err1.Is(errors.New("generic error")))
}

func TestAppError_WithDetail(t *testing.T) {
	err := &AppError{Code: "not_found", Message: "not found"}
	detail := ErrorDetail{Op: "test", Detail: "detail"}
	updated := err.WithDetail(detail)

	assert.Equal(t, detail, updated.Detail)
	assert.Equal(t, err.Code, updated.Code)
	assert.NotEqual(t, err.Detail, updated.Detail)
}

func TestAppError_WithErr(t *testing.T) {
	err := &AppError{Code: "internal", Message: "internal error"}
	wrapped := errors.New("db error")
	updated := err.WithErr(wrapped)

	assert.Equal(t, wrapped, updated.Err)
	assert.Equal(t, err.Code, updated.Code)
}

func TestAs(t *testing.T) {
	err := &AppError{Code: "not_found", Message: "not found"}

	var ae *AppError
	assert.True(t, As(err, &ae))
	assert.Equal(t, "not_found", ae.Code)
}
