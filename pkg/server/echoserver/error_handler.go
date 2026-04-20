package echoserver

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorResponse is the standard API error structure.
// todo: improve this error response type
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// DefaultErrorHandler is a standard error handler that returns JSON.
// todo: improve this error response handler
func DefaultErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	message := err.Error()

	var he *echo.HTTPError
	if errors.As(err, &he) {
		code = he.Code
		message = fmt.Sprintf("%v", he.Message)
	}

	resp := ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
	}

	if err := c.JSON(code, resp); err != nil {
		c.Logger().Error(err)
	}
}
