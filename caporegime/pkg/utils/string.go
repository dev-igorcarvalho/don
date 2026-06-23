// Package utils provides utility functions and helpers for the caporegime framework.
package utils

import (
	"regexp"
	"strings"
)

// underscoreRegex is a compiled regular expression used to identify one or more consecutive underscores.
var underscoreRegex = regexp.MustCompile(`_+`)

// nameReplacer is a strings.Replacer configured to convert slashes, backslashes,
// spaces, and hyphens into underscores.
var nameReplacer = strings.NewReplacer("/", "_", "\\", "_", " ", "_", "-", "_")

// SanitizeName normalizes the input name string into a clean, lowercase identifier
// suitable for filesystem paths, session directories, or configuration keys.
// It replaces slashes, backslashes, spaces, and hyphens with underscores, collapses
// consecutive underscores into a single underscore, and trims any leading or trailing underscores.
// It returns the sanitized string.
func SanitizeName(name string) string {
	replacedDelimiters := nameReplacer.Replace(name)
	collapsedUnderscores := underscoreRegex.ReplaceAllString(replacedDelimiters, "_")
	return strings.ToLower(strings.Trim(collapsedUnderscores, "_"))
}
