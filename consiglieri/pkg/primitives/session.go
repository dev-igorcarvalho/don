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

type sessionDirKey struct{}
type artifactDirKey struct{}
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

// ArtifactDir returns the artifact directory path from ctx.
func ArtifactDir(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(artifactDirKey{}).(string)
	return v, ok
}

// initSession creates .agentic/session/<date>-<uuid>-<name>, builds a tee logger
// (stdout + run.log), and injects session dir, name, ID, and logger into the context.
// The caller is responsible for closing the returned log file.
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

// ensureSessionDirs handles directory creation and validation for the session.
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

// newUUID returns a random unique ID string based on timestamp and crypto/rand bytes.
func newUUID() string {
	now := time.Now()
	milli := now.UnixMilli()
	timeStr := fmt.Sprintf("%s-%d", now.Format("2006-01-02"), milli)
	var b [4]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%s-%x", timeStr, b[:])
}
