package tui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverWorkflows(t *testing.T) {
	// Create a temp directory
	tempDir := t.TempDir()

	// 1. Create a file with comments
	file1Path := filepath.Join(tempDir, "test_workflow_one.go")
	content1 := `package main
// Name: Test Workflow Alpha
// Description: This is a beautiful test description.
func main() {}
`
	if err := os.WriteFile(file1Path, []byte(content1), 0644); err != nil {
		t.Fatalf("failed to write test file 1: %v", err)
	}

	// 2. Create a file without comments (fallback test)
	file2Path := filepath.Join(tempDir, "simple.go")
	content2 := `package main
func main() {}
`
	if err := os.WriteFile(file2Path, []byte(content2), 0644); err != nil {
		t.Fatalf("failed to write test file 2: %v", err)
	}

	// 3. Create a non-go file to check filtering
	file3Path := filepath.Join(tempDir, "readme.md")
	if err := os.WriteFile(file3Path, []byte("# README"), 0644); err != nil {
		t.Fatalf("failed to write readme file: %v", err)
	}

	items, err := DiscoverWorkflows(tempDir)
	if err != nil {
		t.Fatalf("DiscoverWorkflows returned unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected to discover exactly 2 workflow items, got %d", len(items))
	}

	// Find the item by name/filename
	var itemAlpha, itemSimple *WorkflowItem
	for _, item := range items {
		w := item.(WorkflowItem)
		if w.name == "Test Workflow Alpha" {
			itemAlpha = &w
		} else if w.name == "simple" {
			itemSimple = &w
		}
	}

	if itemAlpha == nil {
		t.Errorf("expected to find 'Test Workflow Alpha'")
	} else {
		if itemAlpha.description != "This is a beautiful test description." {
			t.Errorf("expected description 'This is a beautiful test description.', got '%s'", itemAlpha.description)
		}
		if itemAlpha.filePath != file1Path {
			t.Errorf("expected filePath '%s', got '%s'", file1Path, itemAlpha.filePath)
		}
	}

	if itemSimple == nil {
		t.Errorf("expected to find fallback item 'simple'")
	} else if itemSimple.description != "No description provided." {
		t.Errorf("expected default description, got '%s'", itemSimple.description)
	}
}
