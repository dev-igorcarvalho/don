package tui

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type viewState int

const (
	viewDashboard viewState = iota
	viewExecution
	viewInit
	viewInitSuccess
)

const keyCtrlC = "ctrl+c"
const keyEnter = "enter"

type workspaceInitMsg struct {
	err error
}

func initializeWorkspaceCmd(dir string) tea.Cmd {
	return func() tea.Msg {
		err := InitializeWorkspace(dir)
		return workspaceInitMsg{err: err}
	}
}

type WorkflowBuildFinishedMsg struct {
	filePath   string
	binaryPath string
	buildLog   string
	err        error
}

func (m MainModel) compileWorkflowCmd(item WorkflowItem) tea.Cmd {
	return func() tea.Msg {
		baseDir := filepath.Dir(m.workDir)
		binDir := filepath.Join(baseDir, filepath.Base(DefaultBinDir))

		binName := strings.TrimSuffix(filepath.Base(item.filePath), ".go")
		binaryPath := filepath.Join(binDir, binName)

		sourceInfo, err1 := os.Stat(item.filePath)
		binaryInfo, err2 := os.Stat(binaryPath)
		if err1 == nil && err2 == nil && binaryInfo.ModTime().After(sourceInfo.ModTime()) {
			return WorkflowBuildFinishedMsg{
				filePath:   item.filePath,
				binaryPath: binaryPath,
				err:        nil,
			}
		}

		// Guarantee that the bin directory exists before compiling
		if err := os.MkdirAll(binDir, 0755); err != nil {
			return WorkflowBuildFinishedMsg{
				filePath: item.filePath,
				err:      fmt.Errorf("failed to create bin dir: %w", err),
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, item.filePath)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		err := cmd.Run()
		return WorkflowBuildFinishedMsg{
			filePath:   item.filePath,
			binaryPath: binaryPath,
			buildLog:   out.String(),
			err:        err,
		}
	}
}

type MainModel struct {
	state    viewState
	list     list.Model
	viewport viewport.Model
	spinner  spinner.Model
	width    int
	height   int
	workDir  string

	// Runner State
	runner     *Runner
	running    bool
	logLines   []string
	runErr     error
	activeItem *WorkflowItem

	// Auto-run State
	autoRunFilePath string

	// Initialization State
	initializing bool
	initErr      error
}

func NewMainModel(workDir string, items []list.Item) MainModel {
	baseDir := filepath.Dir(workDir)
	binDir := filepath.Join(baseDir, filepath.Base(DefaultBinDir))

	for i, item := range items {
		wItem := item.(WorkflowItem)
		binName := strings.TrimSuffix(filepath.Base(wItem.filePath), ".go")
		wItem.binaryPath = filepath.Join(binDir, binName)

		sourceInfo, err1 := os.Stat(wItem.filePath)
		if err1 != nil {
			if wItem.buildStatus == BuildStatusNone {
				wItem.buildStatus = BuildStatusCompiling
			}
		} else {
			binaryInfo, err2 := os.Stat(wItem.binaryPath)
			if err2 == nil && binaryInfo.ModTime().After(sourceInfo.ModTime()) {
				wItem.buildStatus = BuildStatusSuccess
			} else {
				wItem.buildStatus = BuildStatusCompiling
			}
		}
		items[i] = wItem
	}

	// Initialize list
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Workflows"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	// Initialize viewport
	vp := viewport.New(0, 0)
	vp.SetContent("Select a workflow and press Enter to start...")

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = RunningStatusStyle

	state := viewDashboard
	if len(items) == 0 {
		state = viewInit
	}

	return MainModel{
		state:    state,
		list:     l,
		viewport: vp,
		spinner:  s,
		workDir:  workDir,
	}
}

func (m MainModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	hasCompiling := false
	for _, item := range m.list.Items() {
		wItem := item.(WorkflowItem)
		if wItem.buildStatus == BuildStatusCompiling {
			hasCompiling = true
			cmds = append(cmds, m.compileWorkflowCmd(wItem))
		}
	}
	if hasCompiling {
		cmds = append(cmds, m.spinner.Tick)
	}
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (m MainModel) mainHeight() int {
	h := m.height - 6
	if h < 10 {
		return 10
	}
	return h
}

func (m MainModel) leftWidth() int {
	w := m.width / 3
	if w < 20 {
		return 20
	}
	return w
}

func (m MainModel) rightWidth() int {
	w := m.width - m.leftWidth() - 4
	if w < 20 {
		return 20
	}
	return w
}

func (m MainModel) handleKeyMsg(msg tea.KeyMsg) (MainModel, tea.Cmd) {
	switch msg.String() {
	case keyCtrlC:
		if m.runner != nil {
			m.runner.Stop()
		}
		return m, tea.Quit

	case "esc":
		if !m.running && len(m.logLines) > 0 {
			// Clear execution state and go back to dashboard details
			m.logLines = nil
			m.runErr = nil
			m.activeItem = nil
			m.viewport.SetContent("Select a workflow and press Enter to start...")
			return m, nil
		}

	case keyEnter:
		if m.running {
			// Cancel execution
			if m.runner != nil {
				m.runner.Stop()
			}
			m.logLines = append(m.logLines, "\n[TUI] Execution cancelled by user.")
			m.viewport.SetContent(strings.Join(m.logLines, "\n"))
			m.viewport.GotoBottom()
			return m, nil
		}

		// Run selected workflow (either not running, or finished and running again)
		selected := m.list.SelectedItem()
		if selected != nil && !m.running {
			item := selected.(WorkflowItem)

			// Check if source file has changed relative to binary
			sourceInfo, err1 := os.Stat(item.filePath)
			binaryInfo, err2 := os.Stat(item.binaryPath)
			needsCompile := err1 != nil || err2 != nil || sourceInfo.ModTime().After(binaryInfo.ModTime())

			if needsCompile {
				item.buildStatus = BuildStatusCompiling
				m.list.SetItem(m.list.Index(), item)
				m.autoRunFilePath = item.filePath
				m.activeItem = &item
				m.logLines = []string{"[TUI] Source file changed. Recompiling before running..."}
				m.viewport.SetContent(strings.Join(m.logLines, "\n"))
				m.runErr = nil
				m.running = true

				return m, tea.Batch(
					m.spinner.Tick,
					m.compileWorkflowCmd(item),
				)
			}

			if item.buildStatus == BuildStatusCompiling {
				m.logLines = []string{"[TUI] Cannot execute workflow: Compilation is still in progress."}
				m.viewport.SetContent(strings.Join(m.logLines, "\n"))
				return m, nil
			}
			if item.buildStatus == BuildStatusFailed {
				m.logLines = []string{"[TUI] Cannot execute workflow: Compilation failed. Inspect errors in the details view."}
				m.viewport.SetContent(strings.Join(m.logLines, "\n"))
				return m, nil
			}

			m.activeItem = &item
			m.logLines = []string{"[TUI] Spawning process for " + item.name + "..."}
			m.viewport.SetContent(strings.Join(m.logLines, "\n"))
			m.runErr = nil
			m.running = true

			m.runner = NewRunner(item.binaryPath)
			m.runner.Start(context.Background())

			return m, tea.Batch(
				m.spinner.Tick,
				WaitForActivity(m.runner.Channel()),
			)
		}
	}
	return m, nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == viewInitSuccess {
			switch msg.String() {
			case keyCtrlC, "q":
				return m, tea.Quit
			case keyEnter, "i", " ":
				m.state = viewDashboard
				return m, nil
			}
			return m, nil
		}
		if m.state == viewInit {
			switch msg.String() {
			case keyCtrlC, "q":
				return m, tea.Quit
			case "i":
				if !m.initializing {
					m.initializing = true
					m.initErr = nil
					return m, tea.Batch(
						m.spinner.Tick,
						initializeWorkspaceCmd(m.workDir),
					)
				}
			}
			return m, nil
		}
		switch msg.String() {
		case keyCtrlC, "esc", keyEnter:
			return m.handleKeyMsg(msg)
		}

	case tea.WindowSizeMsg:
		h, v := DocStyle.GetFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v

		// Set sizes for list (inside left pane: border/padding subtracts 4 lines/columns)
		m.list.SetSize(m.leftWidth()-4, m.mainHeight()-4)

		// Set sizes for viewport (inside right pane: border/padding subtracts 4; 3 lines for titles/status)
		m.viewport.Width = m.rightWidth() - 4
		m.viewport.Height = m.mainHeight() - 7
		if m.viewport.Height < 3 {
			m.viewport.Height = 3
		}

	// Runner output stream messages
	case LogLineMsg:
		if m.running && m.runner != nil {
			m.logLines = append(m.logLines, string(msg))
			m.viewport.SetContent(strings.Join(m.logLines, "\n"))
			m.viewport.GotoBottom()
			return m, WaitForActivity(m.runner.Channel())
		}

	case ProcessFinishedMsg:
		m.running = false
		m.runErr = msg.Err
		if msg.Err != nil {
			m.logLines = append(m.logLines, "\n[TUI] Process exited with error: "+msg.Err.Error())
		} else {
			m.logLines = append(m.logLines, "\n[TUI] Process completed successfully.")
		}
		m.viewport.SetContent(strings.Join(m.logLines, "\n"))
		m.viewport.GotoBottom()
		return m, nil

	case workspaceInitMsg:
		m.initializing = false
		if msg.err != nil {
			m.initErr = msg.err
			return m, nil
		}
		// Workspace initialized successfully. Load workflows now.
		items, err := DiscoverWorkflows(m.workDir)
		if err != nil {
			m.initErr = err
			return m, nil
		}
		if len(items) == 0 {
			m.initErr = fmt.Errorf("workspace initialized but no workflows found")
			return m, nil
		}

		baseDir := filepath.Dir(m.workDir)
		binDir := filepath.Join(baseDir, filepath.Base(DefaultBinDir))

		var compileCmds []tea.Cmd
		for i, item := range items {
			wItem := item.(WorkflowItem)
			binName := strings.TrimSuffix(filepath.Base(wItem.filePath), ".go")
			wItem.binaryPath = filepath.Join(binDir, binName)

			sourceInfo, err1 := os.Stat(wItem.filePath)
			binaryInfo, err2 := os.Stat(wItem.binaryPath)
			if err1 == nil && err2 == nil && binaryInfo.ModTime().After(sourceInfo.ModTime()) {
				wItem.buildStatus = BuildStatusSuccess
			} else {
				wItem.buildStatus = BuildStatusCompiling
				compileCmds = append(compileCmds, m.compileWorkflowCmd(wItem))
			}
			items[i] = wItem
		}

		// Populate list and transition to success view
		m.list.SetItems(items)
		m.state = viewInitSuccess

		if len(compileCmds) > 0 {
			compileCmds = append(compileCmds, m.spinner.Tick)
			return m, tea.Batch(compileCmds...)
		}
		return m, nil

	case WorkflowBuildFinishedMsg:
		items := m.list.Items()
		var updatedItem *WorkflowItem
		for idx, item := range items {
			wItem := item.(WorkflowItem)
			if wItem.filePath == msg.filePath {
				if msg.err != nil {
					wItem.buildStatus = BuildStatusFailed
					wItem.buildError = msg.buildLog
				} else {
					wItem.buildStatus = BuildStatusSuccess
					wItem.buildError = ""
					wItem.binaryPath = msg.binaryPath
				}
				m.list.SetItem(idx, wItem)
				updatedItem = &wItem
				break
			}
		}

		if m.autoRunFilePath == msg.filePath {
			m.autoRunFilePath = ""
			if msg.err != nil {
				m.running = false
				m.runErr = msg.err
				m.logLines = append(m.logLines, "[TUI] Compilation failed:")
				for _, line := range strings.Split(msg.buildLog, "\n") {
					if strings.TrimSpace(line) != "" {
						m.logLines = append(m.logLines, line)
					}
				}
				m.viewport.SetContent(strings.Join(m.logLines, "\n"))
				m.viewport.GotoBottom()
				return m, nil
			}

			if updatedItem != nil {
				m.activeItem = updatedItem
				m.logLines = append(m.logLines, "[TUI] Compilation successful. Spawning process...")
				m.viewport.SetContent(strings.Join(m.logLines, "\n"))
				m.viewport.GotoBottom()

				m.runner = NewRunner(updatedItem.binaryPath)
				m.runner.Start(context.Background())

				return m, tea.Batch(
					m.spinner.Tick,
					WaitForActivity(m.runner.Channel()),
				)
			}
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Update children based on focus or running state
	var cmd tea.Cmd
	if m.state == viewInit || m.state == viewInitSuccess {
		return m, nil
	}
	if m.running || len(m.logLines) > 0 {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
