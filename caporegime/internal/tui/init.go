package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const helloWorkflowTemplate = `// name: Hello Workflow
// description: A simple workflow that runs a pipeline and sleeps, demonstrating a successful workflow execution.
package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/dev-igorcarvalho/don/caporegime/pkg/primitives"
)

func main() {
	// Create a simple pipeline that does something
	p := primitives.NewPipeline("Hello Workflow", func(ctx context.Context) error {
		logger := primitives.Logger(ctx)
		logger.Info("Starting Hello Workflow...")
		for i := 1; i <= 5; i++ {
			logger.Info("Doing work...", "step", i)
			time.Sleep(500 * time.Millisecond)
		}
		logger.Info("Hello Workflow completed successfully!")
		return nil
	})

	// Create an orchestrator and add the pipeline
	orch := primitives.NewOrchestrator("Sample Orchestrator")
	orch.AddWorkflow(p)

	// Run the orchestrator. It handles session initialization.
	if err := orch.Run(context.Background()); err != nil {
		log.Fatalf("Orchestrator failed: %v", err)
	}
}
`

// InitializeWorkspace creates all directories defined in consts.go (resolved relative to the parent of the workflows directory)
// and writes a sample hello.go workflow if none exist.
func InitializeWorkspace(dir string) error {
	baseDir := filepath.Dir(dir)

	// Resolve target directories using the basename of the constants in consts.go
	dirs := []string{
		dir, // the workflows dir (e.g. .caporegime/workflows)
		filepath.Join(baseDir, filepath.Base(DefaultAgentsDir)),
		filepath.Join(baseDir, filepath.Base(DefaultPromptsDir)),
		filepath.Join(baseDir, filepath.Base(DefaultSessionDir)),
		filepath.Join(baseDir, filepath.Base(DefaultBinDir)),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	// Write a .gitignore inside the bin directory to ignore all compiled binaries
	binDir := filepath.Join(baseDir, filepath.Base(DefaultBinDir))
	gitignorePath := filepath.Join(binDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("*\n"), 0644); err != nil {
		return fmt.Errorf("failed to write gitignore for bin dir: %w", err)
	}

	// Write a sample workflow file if the directory is empty or has no Go workflows
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	hasWorkflows := false
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			hasWorkflows = true
			break
		}
	}

	if !hasWorkflows {
		templatePath := filepath.Join(dir, "hello.go")
		if err := os.WriteFile(templatePath, []byte(helloWorkflowTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write sample workflow: %w", err)
		}
	}
	return nil
}
