package echoserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apperr "github.com/dev-igorcarvalho/don/pkg/errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestDefaultErrorHandler(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "AppError with details",
			err: &apperr.AppError{
				Code:       "TEST_ERROR",
				Message:    "Human message",
				HttpStatus: http.StatusBadRequest,
				Detail:     apperr.ErrorDetail{Detail: "Technical detail"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"code":"TEST_ERROR","message":"Human message","detail":"Technical detail"}` + "\n",
		},
		{
			name: "Wrapped AppError",
			err: (&apperr.AppError{
				Code:       "WRAPPED_ERROR",
				Message:    "Original message",
				HttpStatus: http.StatusConflict,
			}).WithErr(errors.New("db error")),
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"code":"WRAPPED_ERROR","message":"Original message"}` + "\n",
		},
		{
			name:           "Standard echo.HTTPError",
			err:            echo.NewHTTPError(http.StatusNotFound, "Not found"),
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"code":"Not Found","message":"Not found"}` + "\n",
		},
		{
			name:           "Generic error",
			err:            errors.New("something went wrong"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"code":"INTERNAL_ERROR","message":"An unexpected internal error occurred"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			DefaultErrorHandler(tt.err, c)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
		})
	}
}
