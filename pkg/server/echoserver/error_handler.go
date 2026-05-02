// ---
// title: Echo Error Handler
// description: Centralized error handling for the Echo server, translating application errors into standardized JSON responses.
// last_updated: 2026-05-02
// type: Utility
// ---

package echoserver

import (
	"errors"
	"fmt"
	"net/http"

	apperr "github.com/dev-igorcarvalho/don/pkg/errors"
	"github.com/labstack/echo/v4"
)

// ErrorResponse is the standard API error structure.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// DefaultErrorHandler is a standard error handler that returns JSON.
func DefaultErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	errorCode := "INTERNAL_ERROR"
	message := "An unexpected internal error occurred"
	var detail string

	var appErr *apperr.AppError
	var he *echo.HTTPError

	switch {
	case apperr.As(err, &appErr):
		if appErr.HttpStatus != 0 {
			code = appErr.HttpStatus
		}
		errorCode = appErr.Code
		message = appErr.Message
		detail = appErr.Detail.Detail

	case errors.As(err, &he):
		code = he.Code
		errorCode = http.StatusText(he.Code)
		message = fmt.Sprintf("%v", he.Message)
	}

	resp := ErrorResponse{
		Code:    errorCode,
		Message: message,
		Detail:  detail,
	}

	if err := c.JSON(code, resp); err != nil {
		c.Logger().Error(err)
	}
}
