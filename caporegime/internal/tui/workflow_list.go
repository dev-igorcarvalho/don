package tui

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

// WorkflowItem implements list.Item for charmbracelet/list.
type WorkflowItem struct {
	name        string
	filePath    string
	description string
}

func (w WorkflowItem) Title() string       { return w.name }
func (w WorkflowItem) Description() string { return w.description }
func (w WorkflowItem) FilterValue() string { return w.name }

// DiscoverWorkflows reads files in the target directory and parses workflows.
func DiscoverWorkflows(dir string) ([]list.Item, error) {
	var items []list.Item

	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return items, nil
		}
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		fullPath := filepath.Join(dir, file.Name())
		defaultName := strings.TrimSuffix(file.Name(), ".go")
		defaultName = strings.ReplaceAll(defaultName, " ", "_")

		name, description := parseWorkflowMetadata(fullPath, defaultName)

		items = append(items, WorkflowItem{
			name:        name,
			filePath:    fullPath,
			description: description,
		})
	}

	return items, nil
}

// parseWorkflowMetadata reads name and description comments from a Go file.
func parseWorkflowMetadata(filePath string, defaultName string) (name string, description string) {
	name = defaultName
	description = "No description provided."

	f, err := os.Open(filePath)
	if err != nil {
		return name, description
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "//") {
			comment := strings.TrimSpace(strings.TrimPrefix(line, "//"))
			if strings.HasPrefix(strings.ToLower(comment), "name:") {
				name = strings.TrimSpace(comment[5:])
			} else if strings.HasPrefix(strings.ToLower(comment), "description:") {
				description = strings.TrimSpace(comment[12:])
			}
		}
	}
	return name, description
}
