package echoserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dev-igorcarvalho/don/pkg/server"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultErrorHandler(t *testing.T) {
	e := echo.New()

	t.Run("handles generic error as 500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := errors.New("something went wrong")
		DefaultErrorHandler(err, c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var resp ErrorResponse
		unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, unmarshalErr)

		assert.Equal(t, "Internal Server Error", resp.Error)
		assert.Equal(t, "something went wrong", resp.Message)
	})

	t.Run("handles echo.HTTPError correctly", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := echo.NewHTTPError(http.StatusBadRequest, "invalid input")
		DefaultErrorHandler(err, c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp ErrorResponse
		unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, unmarshalErr)

		assert.Equal(t, "Bad Request", resp.Error)
		assert.Equal(t, "invalid input", resp.Message)
	})
}

func TestWithErrorHandler(t *testing.T) {
	handlerCalled := false
	customHandler := func(err error, c echo.Context) {
		handlerCalled = true
		_ = c.String(http.StatusTeapot, "custom error")
	}

	s := New(WithErrorHandler(customHandler))

	s.RegisterRoutes(server.NewRoute("GET", "/error", &testHandler{
		handleFunc: func(c echo.Context) error {
			return errors.New("trigger custom handler")
		},
	}))

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	s.app.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.True(t, handlerCalled)
}
