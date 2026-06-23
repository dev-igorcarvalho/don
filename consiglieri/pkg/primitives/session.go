// Package primitives provides the core agentic primitives, such as Agent,
// Pipeline, Orchestrator, and Session.
//
// A session represents the isolated workspace and logging context for a single run
// of an Orchestrator. When an Orchestrator runs, it initializes a session,
// creates the necessary directory structure for logs and artifacts, and injects
// the session parameters (ID, name, directories, and logger) into the context.
//
// Files under the session directory are structured as follows:
//
//	.agentic/session/<timestamp>-<rand>-<orchestrator_name>/
//	├── logs/
//	│   └── run.log
//	└── artifacts/
//
// Injected values can be retrieved from the context using the provided functions
// (e.g., SessionDir, SessionID, Logger, ArtifactDir).
package primitives

import (
	"context"
	"crypto/rand"
	"don_consiglieri/pkg/utils"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
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
	v, ok := ctx.Value(sessionDirKey{}).(string)
	return v, ok
}

// SessionName retrieves the name of the Orchestrator associated with the current session from the context.
// It returns the name string and a boolean indicating whether the key was present.
func SessionName(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(sessionNameKey{}).(string)
	return v, ok
}

// SessionID retrieves the unique session ID (formatted as "timestamp-randHex-name") from the context.
// It returns the session ID string and a boolean indicating whether the key was present.
func SessionID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(sessionIDKey{}).(string)
	return v, ok
}

// Logger retrieves the structured slog.Logger associated with the current session from the context.
// If no logger is found in the context, it falls back to slog.Default().
func Logger(ctx context.Context) *slog.Logger {
	if v, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return v
	}
	return slog.Default()
}

// ArtifactDir retrieves the absolute path of the current session's artifacts directory from the context.
// It returns the path string and a boolean indicating whether the key was present.
func ArtifactDir(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(artifactDirKey{}).(string)
	return v, ok
}

// initSession initializes the workspace directories, log files, and logger for an Orchestrator run.
// It injects the session directory, artifact directory, session name, session ID, and logger into the
// context, and returns the modified context and the log file pointer.
// The caller must ensure that the returned log file is closed at the end of the session.
func (o *Orchestrator) initSession(ctx context.Context) (context.Context, *os.File, error) {
	sessionDir, artifactsDir, logsDir, sessionID, err := ensureSessionDirs(o.Name)
	if err != nil {
		return ctx, nil, err
	}

	logFile, err := os.Create(filepath.Join(logsDir, "run.log"))
	if err != nil {
		return ctx, nil, fmt.Errorf("create log file: %w", err)
	}

	w := io.MultiWriter(os.Stdout, logFile)
	logger := slog.New(slog.NewTextHandler(w, nil))

	ctx = context.WithValue(ctx, sessionDirKey{}, sessionDir)
	ctx = context.WithValue(ctx, artifactDirKey{}, artifactsDir)
	ctx = context.WithValue(ctx, sessionNameKey{}, o.Name)
	ctx = context.WithValue(ctx, sessionIDKey{}, sessionID)
	ctx = context.WithValue(ctx, loggerKey{}, logger)
	return ctx, logFile, nil
}

// ensureSessionDirs creates the standard folder hierarchy under the ".agentic/session" directory in the
// current working directory. It creates a unique subdirectory for the current session, along with
// "logs" and "artifacts" subdirectories.
func ensureSessionDirs(sessionName string) (sessionDir, artifactsDir, logsDir, sessionID string, err error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", "", "", "", fmt.Errorf("getwd: %w", err)
	}

	sessionBase := filepath.Join(pwd, ".agentic", "session")
	if err := os.MkdirAll(sessionBase, 0o755); err != nil {
		return "", "", "", "", fmt.Errorf("create session base: %w", err)
	}

	sessionID = fmt.Sprintf("%s-%s", newUUID(), utils.SanitizeName(sessionName))
	sessionDir = filepath.Join(sessionBase, sessionID)

	if err := os.Mkdir(sessionDir, 0o755); err != nil {
		return "", "", "", "", fmt.Errorf("create session dir: %w", err)
	}

	logsDir = filepath.Join(sessionDir, "logs")
	if err := os.Mkdir(logsDir, 0o755); err != nil {
		return "", "", "", "", fmt.Errorf("create logs dir: %w", err)
	}

	artifactsDir = filepath.Join(sessionDir, "artifacts")
	if err := os.Mkdir(artifactsDir, 0o755); err != nil {
		return "", "", "", "", fmt.Errorf("create artifact dir: %w", err)
	}

	return sessionDir, artifactsDir, logsDir, sessionID, nil
}

// newUUID generates a time-based unique identifier string using the current local date,
// the Unix millisecond timestamp, and four cryptographically secure random bytes.
// The output format is "YYYY-MM-DD-timestamp-randHex".
func newUUID() string {
	now := time.Now()
	milli := now.UnixMilli()
	timeStr := fmt.Sprintf("%s-%d", now.Format("2006-01-02"), milli)
	var b [4]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%s-%x", timeStr, b[:])
}
