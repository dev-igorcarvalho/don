// Package main serves as the entry point for the Don Caporegime TUI dashboard CLI tool.
// It discovers registered agentic workflows and launches the terminal UI dashboard.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dev-igorcarvalho/don/caporegime/internal/tui"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// defaultWorkflowsDir is the default directory relative to the repository root
// where the TUI dashboard searches for registered workflows.
const defaultWorkflowsDir = tui.DefaultWorkflowsDir

// main is the entry point of the CLI application. It discovers available workflows
// in the default directory and starts the interactive terminal dashboard.
func main() {
	initFlag := flag.Bool("init", false, "Initialize the workspace folder structure and a sample workflow")
	flag.Parse()

	if *initFlag {
		if err := tui.InitializeWorkspace(defaultWorkflowsDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing workspace: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Workspace initialized successfully at %s!\n", defaultWorkflowsDir)
		os.Exit(0)
	}

	items, err := tui.DiscoverWorkflows(defaultWorkflowsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering workflows: %v\n", err)
		os.Exit(1)
	}

	if err := runDashboard(defaultWorkflowsDir, items); err != nil {
		log.Fatalf("Error running dashboard: %v", err)
	}
}

// setupWorkflows discovers and validates agentic workflows inside the target directory.
// It returns a slice of items ready for the dashboard list, or an error if discovery
// fails or if no workflows are found.
func setupWorkflows(dir string) ([]list.Item, error) {
	items, err := tui.DiscoverWorkflows(dir)
	if err != nil {
		return nil, fmt.Errorf("error discovering workflows: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no workflows found in %s.\nPlease add a .go workflow file to that directory.", dir)
	}

	return items, nil
}

// runDashboard initializes and starts the Bubble Tea dashboard TUI with the found workflows.
// It returns an error if the Bubble Tea program fails to run.
func runDashboard(dir string, items []list.Item) error {
	model := tui.NewMainModel(dir, items)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
