package utils

import (
	"os/user"
	"strings"
)

// GetDonName retrieves the current OS user's name or username, formats it,
// and returns it prefixed with "Don " (e.g., "Don Igor" or "Don igor").
// If the user's name or username cannot be retrieved, it returns "Don".
func GetDonName() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	name := u.Name
	if name == "" {
		name = u.Username
	}

	// Capitalize first letter of name if it's a simple, lowercase single word
	if !strings.Contains(name, " ") && len(name) > 0 {
		name = strings.ToUpper(name[:1]) + name[1:]
	}
	return name
}
