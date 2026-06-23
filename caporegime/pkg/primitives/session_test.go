package primitives

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dev-igorcarvalho/don/caporegime/pkg/utils"
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
	if _, ok := ArtifactDir(ctx); ok {
		t.Error("expected ArtifactDir to return false for empty context")
	}
	if logger := Logger(ctx); logger != slog.Default() {
		t.Error("expected Logger to return slog.Default() for empty context")
	}

	// Test with values
	testDir := "/tmp/session"
	testName := "test-orchestrator"
	testID := "2023-01-01-uuid-test-orchestrator"
	testArtifactDir := "/tmp/session/artifacts"
	testLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx = context.WithValue(ctx, sessionDirKey{}, testDir)
	ctx = context.WithValue(ctx, sessionNameKey{}, testName)
	ctx = context.WithValue(ctx, sessionIDKey{}, testID)
	ctx = context.WithValue(ctx, artifactDirKey{}, testArtifactDir)
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
	if dir, ok := ArtifactDir(ctx); !ok || dir != testArtifactDir {
		t.Errorf("expected ArtifactDir to return %s, got %s", testArtifactDir, dir)
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
	if !strings.Contains(sessionDir, ".caporegime/session") {
		t.Errorf("unexpected sessionDir: %s", sessionDir)
	}

	sessionName, ok := SessionName(newCtx)
	if !ok || sessionName != "test-orch" {
		t.Errorf("unexpected sessionName: %s", sessionName)
	}

	sessionID, ok := SessionID(newCtx)
	if !ok || !strings.HasSuffix(sessionID, utils.SanitizeName(orch.Name)) {
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

	// Format check: yyyy-mm-dd-milisecs-randomHex
	parts := strings.Split(uuid1, "-")
	if len(parts) != 5 {
		t.Fatalf("invalid unique ID format: %s", uuid1)
	}

	// parts should be:
	// parts[0]: year (length 4)
	// parts[1]: month (length 2)
	// parts[2]: day (length 2)
	// parts[3]: milliseconds (length 1-3)
	// parts[4]: random hex (length 8)
	if len(parts[0]) != 4 || len(parts[1]) != 2 || len(parts[2]) != 2 {
		t.Errorf("unexpected date format in unique ID: %s", uuid1)
	}
	if len(parts[4]) != 8 {
		t.Errorf("unexpected random hex length in unique ID: %s (got length %d, want 8)", uuid1, len(parts[4]))
	}
}

func TestInitSession_Errors(t *testing.T) {
	// 1. MkdirAll of sessionBase fails
	t.Run("sessionBase creation fails", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "session_base_fail")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a file named .caporegime
		err = os.WriteFile(filepath.Join(tmpDir, ".caporegime"), []byte("blocker"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = os.Chdir(tmpDir)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = os.Chdir(oldWd)
		}()

		orch := &Orchestrator{Name: "test-orch"}
		_, _, err = orch.initSession(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	// 2. Mkdir of sessionDir fails because of long filename
	t.Run("sessionDir creation fails", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "session_base_test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		origBase := os.Getenv("AGENTIC_SESSION_BASE")
		os.Setenv("AGENTIC_SESSION_BASE", tmpDir)
		defer os.Setenv("AGENTIC_SESSION_BASE", origBase)

		orch := &Orchestrator{Name: strings.Repeat("a", 300)}
		_, _, err = orch.initSession(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	// 3. Getwd fails
	t.Run("getwd fails", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "session_getwd_fail")
		if err != nil {
			t.Fatal(err)
		}
		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = os.Chdir(tmpDir)
		if err != nil {
			t.Fatal(err)
		}
		// Delete the current working directory to trigger getwd error
		os.RemoveAll(tmpDir)
		defer func() {
			_ = os.Chdir(oldWd)
		}()

		orch := &Orchestrator{Name: "test-orch"}
		_, _, err = orch.initSession(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	// Helper function inside tests to create deep directory path for exact length
	createLongWd := func(t *testing.T, targetLen int) string {
		tmpDir, err := os.MkdirTemp("", "session_test_long")
		if err != nil {
			t.Fatal(err)
		}
		current := tmpDir
		for {
			remaining := targetLen - len(current) - 1
			if remaining <= 0 {
				break
			}
			segmentLen := remaining
			if segmentLen > 200 {
				segmentLen = 200
			}
			segment := strings.Repeat("a", segmentLen)
			next := filepath.Join(current, segment)
			if err := os.Mkdir(next, 0755); err != nil {
				t.Fatal(err)
			}
			current = next
		}
		return current
	}

	// 4. logsDir creation fails (logsDir path length > PATH_MAX (4096))
	t.Run("logsDir creation fails", func(t *testing.T) {
		pwd := createLongWd(t, 4032)
		defer os.RemoveAll(strings.Split(pwd, "/session_test_long")[0] + "/session_test_long*")

		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = os.Chdir(pwd)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = os.Chdir(oldWd)
		}()

		orch := &Orchestrator{Name: "test-orch"}
		_, _, err = orch.initSession(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	// 5. artifactsDir creation fails (artifactsDir path length > PATH_MAX (4096))
	t.Run("artifactsDir creation fails", func(t *testing.T) {
		pwd := createLongWd(t, 4029)
		defer os.RemoveAll(strings.Split(pwd, "/session_test_long")[0] + "/session_test_long*")

		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = os.Chdir(pwd)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = os.Chdir(oldWd)
		}()

		orch := &Orchestrator{Name: "test-orch"}
		_, _, err = orch.initSession(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	// 6. logFile creation fails (logFile path length > PATH_MAX (4096))
	t.Run("logFile creation fails", func(t *testing.T) {
		pwd := createLongWd(t, 4024)
		defer os.RemoveAll(strings.Split(pwd, "/session_test_long")[0] + "/session_test_long*")

		oldWd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		err = os.Chdir(pwd)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			_ = os.Chdir(oldWd)
		}()

		orch := &Orchestrator{Name: "test-orch"}
		_, f, err := orch.initSession(context.Background())
		if f != nil {
			f.Close()
		}
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
