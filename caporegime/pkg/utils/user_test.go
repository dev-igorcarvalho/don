package utils

import (
	"strings"
	"testing"
)

func TestGetDonName(t *testing.T) {
	name := GetDonName()
	if !strings.HasPrefix(name, "Don") {
		t.Errorf("expected name to start with 'Don', got %q", name)
	}
}
