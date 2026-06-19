// ---
// title: Health Handler
// description: Provides a basic health check endpoint to verify the service status.
// last_updated: 2026-05-03
// type: Implementation
// ---

package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Handle(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "up"})
}
