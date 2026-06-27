package primitives

import (
	"context"
	"log/slog"
)

// sessionDirKey is a context key used to store and retrieve the session's root directory path.
type sessionDirKey struct{}

// artifactDirKey is a context key used to store and retrieve the session's artifacts directory path.
type artifactDirKey struct{}

// sessionNameKey is a context key used to store and retrieve the orchestrator/session name.
type sessionNameKey struct{}

// sessionIDKey is a context key used to store and retrieve the unique session ID.
type sessionIDKey struct{}

// loggerKey is a context key used to store and retrieve the session's slog.Logger.
type loggerKey struct{}

// SessionDir retrieves the absolute path of the current session's root directory from the context.
// It returns the path string and a boolean indicating whether the key was present.
func SessionDir(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	v, ok := ctx.Value(sessionDirKey{}).(string)
	return v, ok
}

// SessionName retrieves the name of the Orchestrator associated with the current session from the context.
// It returns the name string and a boolean indicating whether the key was present.
func SessionName(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	v, ok := ctx.Value(sessionNameKey{}).(string)
	return v, ok
}

// SessionID retrieves the unique session ID (formatted as "timestamp-randHex-name") from the context.
// It returns the session ID string and a boolean indicating whether the key was present.
func SessionID(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	v, ok := ctx.Value(sessionIDKey{}).(string)
	return v, ok
}

// Logger retrieves the structured slog.Logger associated with the current session from the context.
// If no logger is found in the context, it falls back to slog.Default().
func Logger(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if v, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return v
	}
	return slog.Default()
}

// ArtifactDir retrieves the absolute path of the current session's artifacts directory from the context.
// It returns the path string and a boolean indicating whether the key was present.
func ArtifactDir(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	v, ok := ctx.Value(artifactDirKey{}).(string)
	return v, ok
}
