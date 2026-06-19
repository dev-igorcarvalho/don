package primitives

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionGetters(t *testing.T) {
	ctx := context.Background()

	// Test defaults/not found
	if _, ok := SessionDir(ctx); ok {
		t.Error("expected SessionDir to return false for empty context")
	}
	if _, ok := SessionName(ctx); ok {
		t.Error("expected SessionName to return false for empty context")
	}
	if _, ok := SessionID(ctx); ok {
		t.Error("expected SessionID to return false for empty context")
	}
	if logger := Logger(ctx); logger != slog.Default() {
		t.Error("expected Logger to return slog.Default() for empty context")
	}

	// Test with values
	testDir := "/tmp/session"
	testName := "test-orchestrator"
	testID := "2023-01-01-uuid-test-orchestrator"
	testLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx = context.WithValue(ctx, sessionDirKey{}, testDir)
	ctx = context.WithValue(ctx, sessionNameKey{}, testName)
	ctx = context.WithValue(ctx, sessionIDKey{}, testID)
	ctx = context.WithValue(ctx, loggerKey{}, testLogger)

	if dir, ok := SessionDir(ctx); !ok || dir != testDir {
		t.Errorf("expected SessionDir to return %s, got %s", testDir, dir)
	}
	if name, ok := SessionName(ctx); !ok || name != testName {
		t.Errorf("expected SessionName to return %s, got %s", testName, name)
	}
	if id, ok := SessionID(ctx); !ok || id != testID {
		t.Errorf("expected SessionID to return %s, got %s", testID, id)
	}
	if logger := Logger(ctx); logger != testLogger {
		t.Error("expected Logger to return the injected logger")
	}
}

func TestInitSession(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "session_test_*")
	if err != nil {
		t.Fatalf("failed to create pocs_automacoes dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change working directory to pocs_automacoes dir
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current wd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change wd: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	orch := &Orchestrator{Name: "test-orch"}
	ctx := context.Background()

	newCtx, logFile, err := orch.initSession(ctx)
	if err != nil {
		t.Fatalf("initSession failed: %v", err)
	}
	defer logFile.Close()

	// Verify context values
	sessionDir, ok := SessionDir(newCtx)
	if !ok {
		t.Error("SessionDir not found in context")
	}
	if !strings.Contains(sessionDir, ".agentic/session") {
		t.Errorf("unexpected sessionDir: %s", sessionDir)
	}

	sessionName, ok := SessionName(newCtx)
	if !ok || sessionName != "test-orch" {
		t.Errorf("unexpected sessionName: %s", sessionName)
	}

	sessionID, ok := SessionID(newCtx)
	if !ok || !strings.HasSuffix(sessionID, "test-orch") {
		t.Errorf("unexpected sessionID: %s", sessionID)
	}

	logger := Logger(newCtx)
	if logger == nil || logger == slog.Default() {
		t.Error("logger should be initialized and not default")
	}

	// Verify directory structure
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		t.Errorf("session directory does not exist: %s", sessionDir)
	}

	logsDir := filepath.Join(sessionDir, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		t.Errorf("logs directory does not exist: %s", logsDir)
	}

	runLog := filepath.Join(logsDir, "run.log")
	if _, err := os.Stat(runLog); os.IsNotExist(err) {
		t.Errorf("run.log does not exist: %s", runLog)
	}

	// Verify logger works and writes to file
	testMessage := "test log message"
	logger.Info(testMessage)

	// Since slog might have buffering or we need to ensure it's written
	// Actually slog.NewTextHandler(w, nil) writes immediately to w if w is an os.File or similar.
	// But let's read the file to be sure.
	content, err := os.ReadFile(runLog)
	if err != nil {
		t.Fatalf("failed to read run.log: %v", err)
	}
	if !strings.Contains(string(content), testMessage) {
		t.Errorf("expected log file to contain %q, got %q", testMessage, string(content))
	}
}

func TestNewUUID(t *testing.T) {
	uuid1 := newUUID()
	uuid2 := newUUID()

	if uuid1 == uuid2 {
		t.Errorf("expected different UUIDs, got both %s", uuid1)
	}

	// Basic UUID v4 format check: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	// where y is 8, 9, a, or b.
	parts := strings.Split(uuid1, "-")
	if len(parts) != 5 {
		t.Fatalf("invalid UUID format: %s", uuid1)
	}

	if parts[2][0] != '4' {
		t.Errorf("expected UUID version 4, got %c", parts[2][0])
	}
}
