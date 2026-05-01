// Package must provides helper functions that log a fatal error and exit
// the application when a critical operation fails or a requirement is not met.
package must

import (
	"fmt"
	"log"
)

// Get returns v if err is nil, otherwise logs fatal.
// Use this for mandatory initializations that return a value.
func Get[T any](v T, err error) T {
	if err != nil {
		log.Fatalf("critical failure: %v", err)
	}
	return v
}

// Getf returns v if err is nil, otherwise logs fatal with a formatted message.
// Use this when you want to provide extra context to a critical failure.
func Getf[T any](v T, err error, format string, args ...any) T {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		log.Fatalf("%s: %v", msg, err)
	}
	return v
}

// Succeed logs a fatal error and exits if err is not nil.
// Use this for mandatory operations that only return an error.
func Succeed(err error) {
	if err != nil {
		log.Fatalf("critical error: %v", err)
	}
}

// Succeedf logs a fatal error and exits if err is not nil with a formatted message.
// Use this when you want to provide extra context to a critical operation failure.
func Succeedf(err error, format string, args ...any) {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		log.Fatalf("%s: %v", msg, err)
	}
}

// BeTrue logs a fatal error and exits if ok is false.
// Use this for mandatory boolean checks or "comma-ok" patterns.
func BeTrue(ok bool, msg string) {
	if !ok {
		log.Fatalf("critical requirement missing: %s", msg)
	}
}

// BeTruef logs a fatal error and exits if ok is false with a formatted message.
func BeTruef(ok bool, format string, args ...any) {
	if !ok {
		msg := fmt.Sprintf(format, args...)
		log.Fatalf("critical requirement missing: %s", msg)
	}
}
