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
//	.caporegime/session/<timestamp>-<rand>-<orchestrator_name>/
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
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dev-igorcarvalho/don/caporegime/pkg/utils"
)

// initSession initializes the workspace directories, log files, and logger for an Orchestrator run.
// It injects the session directory, artifact directory, session name, session ID, and logger into the
// context, and returns the modified context and the log file pointer.
// The caller must ensure that the returned log file is closed at the end of the session.
// sessionDirs holds the paths and unique identifier of the initialized session directories.
type sessionDirs struct {
	SessionDir   string
	ArtifactsDir string
	LogsDir      string
	SessionID    string
}

func (o *Orchestrator) initSession(ctx context.Context) (context.Context, *os.File, error) {
	dirs, err := ensureSessionDirs(o.Name)
	if err != nil {
		return ctx, nil, err
	}

	logFile, err := os.Create(filepath.Join(dirs.LogsDir, "run.log"))
	if err != nil {
		return ctx, nil, fmt.Errorf("create log file: %w", err)
	}

	w := io.MultiWriter(os.Stdout, logFile)
	logger := slog.New(slog.NewTextHandler(w, nil))

	ctx = context.WithValue(ctx, sessionDirKey{}, dirs.SessionDir)
	ctx = context.WithValue(ctx, artifactDirKey{}, dirs.ArtifactsDir)
	ctx = context.WithValue(ctx, sessionNameKey{}, o.Name)
	ctx = context.WithValue(ctx, sessionIDKey{}, dirs.SessionID)
	ctx = context.WithValue(ctx, loggerKey{}, logger)
	return ctx, logFile, nil
}

// ensureSessionDirs creates the standard folder hierarchy under the ".caporegime/session" directory in the
// current working directory. It creates a unique subdirectory for the current session, along with
// "logs" and "artifacts" subdirectories.
func ensureSessionDirs(sessionName string) (sessionDirs, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return sessionDirs{}, fmt.Errorf("getwd: %w", err)
	}

	sessionBase := filepath.Join(pwd, ".caporegime", "session")
	if err := os.MkdirAll(sessionBase, 0o755); err != nil {
		return sessionDirs{}, fmt.Errorf("create session base: %w", err)
	}

	sessionID := fmt.Sprintf("%s-%s", newUUID(), utils.SanitizeName(sessionName))
	sessionDir := filepath.Join(sessionBase, sessionID)

	if err := os.Mkdir(sessionDir, 0o755); err != nil {
		return sessionDirs{}, fmt.Errorf("create session dir: %w", err)
	}

	logsDir := filepath.Join(sessionDir, "logs")
	if err := os.Mkdir(logsDir, 0o755); err != nil {
		return sessionDirs{}, fmt.Errorf("create logs dir: %w", err)
	}

	artifactsDir := filepath.Join(sessionDir, "artifacts")
	if err := os.Mkdir(artifactsDir, 0o755); err != nil {
		return sessionDirs{}, fmt.Errorf("create artifact dir: %w", err)
	}

	return sessionDirs{
		SessionDir:   sessionDir,
		ArtifactsDir: artifactsDir,
		LogsDir:      logsDir,
		SessionID:    sessionID,
	}, nil
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
