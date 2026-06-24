package utils

import (
	"os/user"
	"strings"
	"testing"
)

func TestGetDonName(t *testing.T) {
	u, err := user.Current()
	if err != nil {
		t.Skip("skipping test; current user cannot be retrieved")
	}

	expected := u.Name
	if expected == "" {
		expected = u.Username
	}

	// Capitalize first letter of name if it's a simple, lowercase single word
	if !strings.Contains(expected, " ") && len(expected) > 0 {
		expected = strings.ToUpper(expected[:1]) + expected[1:]
	}

	name := GetDonName()
	if name != expected {
		t.Errorf("expected name to be %q, got %q", expected, name)
	}
}
