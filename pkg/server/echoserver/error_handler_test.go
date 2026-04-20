package echoserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestDefaultErrorHandler(t *testing.T) {
	e := echo.New()
	
	t.Run("handles generic error as 500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		err := errors.New("something went wrong")
		DefaultErrorHandler(err, c)
		
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
		
		var resp ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		
		if resp.Error != "Internal Server Error" {
			t.Errorf("expected Error='Internal Server Error', got '%s'", resp.Error)
		}
		if resp.Message != "something went wrong" {
			t.Errorf("expected Message='something went wrong', got '%s'", resp.Message)
		}
	})

	t.Run("handles echo.HTTPError correctly", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		err := echo.NewHTTPError(http.StatusBadRequest, "invalid input")
		DefaultErrorHandler(err, c)
		
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
		
		var resp ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		
		if resp.Error != "Bad Request" {
			t.Errorf("expected Error='Bad Request', got '%s'", resp.Error)
		}
		if resp.Message != "invalid input" {
			t.Errorf("expected Message='invalid input', got '%s'", resp.Message)
		}
	})
}

func TestWithErrorHandler(t *testing.T) {
	handlerCalled := false
	customHandler := func(err error, c echo.Context) {
		handlerCalled = true
		_ = c.String(http.StatusTeapot, "custom error")
	}
	
	s := New(WithErrorHandler(customHandler))
	
	// Register a route that returns an error
	s.RegisterRoutes(NewRoute("GET", "/error", &testHandler{
		handleFunc: func(c echo.Context) error {
			return errors.New("trigger custom handler")
		},
	}))
	
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	s.app.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusTeapot {
		t.Errorf("expected status 418, got %d", rec.Code)
	}
	if !handlerCalled {
		t.Error("expected custom error handler to be called")
	}
}
