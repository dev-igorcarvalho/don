package utils

import (
	"regexp"
	"strings"
)

var underscoreRegex = regexp.MustCompile(`_+`)
var nameReplacer = strings.NewReplacer("/", "_", "\\", "_", " ", "_", "-", "_")

func SanitizeName(name string) string {
	replacedDelimiters := nameReplacer.Replace(name)
	collapsedUnderscores := underscoreRegex.ReplaceAllString(replacedDelimiters, "_")
	return strings.ToLower(strings.Trim(collapsedUnderscores, "_"))
}
