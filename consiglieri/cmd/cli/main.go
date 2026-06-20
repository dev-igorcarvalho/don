package main

import (
	"fmt"
	"log"
	"os"

	"don_consiglieri/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Discover workflows inside .agentic/workflows
	workflowsDir := ".agentic/workflows"
	items, err := tui.DiscoverWorkflows(workflowsDir)
	if err != nil {
		log.Fatalf("Error discovering workflows: %v", err)
	}

	if len(items) == 0 {
		fmt.Printf("No workflows found in %s.\n", workflowsDir)
		fmt.Println("Please add a .go workflow file to that directory.")
		os.Exit(1)
	}

	model := tui.NewMainModel(workflowsDir, items)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running dashboard: %v", err)
	}
}
