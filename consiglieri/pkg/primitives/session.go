package primitives

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type sessionDirKey struct{}
type sessionNameKey struct{}
type sessionIDKey struct{}
type loggerKey struct{}

// SessionDir returns the session directory path from ctx.
func SessionDir(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(sessionDirKey{}).(string)
	return v, ok
}

// SessionName returns the orchestrator name from ctx.
func SessionName(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(sessionNameKey{}).(string)
	return v, ok
}

// SessionID returns the session ID from ctx.
func SessionID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(sessionIDKey{}).(string)
	return v, ok
}

// Logger returns the slog.Logger stored in ctx, falling back to slog.Default().
func Logger(ctx context.Context) *slog.Logger {
	if v, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return v
	}
	return slog.Default()
}

// initSession creates .agentic/session/<date>-<uuid>-<name>, builds a tee logger
// (stdout + run.log), and injects session dir, name, ID, and logger into the context.
// The caller is responsible for closing the returned log file.
func (o *Orchestrator) initSession(ctx context.Context) (context.Context, *os.File, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return ctx, nil, fmt.Errorf("getwd: %w", err)
	}

	sessionBase := filepath.Join(pwd, ".agentic", "session")
	if err := os.MkdirAll(sessionBase, 0o755); err != nil {
		return ctx, nil, fmt.Errorf("create session base: %w", err)
	}

	sessionID := fmt.Sprintf("%s-%s-%s", time.Now().Format("2006-01-02"), newUUID(), o.Name)
	sessionDir := filepath.Join(sessionBase, sessionID)

	if err := os.Mkdir(sessionDir, 0o755); err != nil {
		return ctx, nil, fmt.Errorf("create session dir: %w", err)
	}

	logsDir := filepath.Join(sessionDir, "logs")
	if err := os.Mkdir(logsDir, 0o755); err != nil {
		return ctx, nil, fmt.Errorf("create logs dir: %w", err)
	}

	logFile, err := os.Create(filepath.Join(logsDir, "run.log"))
	if err != nil {
		return ctx, nil, fmt.Errorf("create log file: %w", err)
	}

	w := io.MultiWriter(os.Stdout, logFile)
	logger := slog.New(slog.NewTextHandler(w, nil))

	ctx = context.WithValue(ctx, sessionDirKey{}, sessionDir)
	ctx = context.WithValue(ctx, sessionNameKey{}, o.Name)
	ctx = context.WithValue(ctx, sessionIDKey{}, sessionID)
	ctx = context.WithValue(ctx, loggerKey{}, logger)
	return ctx, logFile, nil
}

// newUUID returns a random UUID v4 string using crypto/rand.
func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
