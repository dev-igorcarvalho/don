package utils

import (
	"regexp"
	"strings"
)

var underscoreRegex = regexp.MustCompile(`_+`)

func SanitizeName(name string) string {
	r := strings.NewReplacer("/", "_", "\\", "_", " ", "_", "-", "_")
	replaced := r.Replace(name)
	collapsed := underscoreRegex.ReplaceAllString(replaced, "_")
	return strings.ToLower(strings.Trim(collapsed, "_"))
}
