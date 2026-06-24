package tui

import (
	"context"
	"fmt"
	"strings"

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

	// Initialization State
	initializing bool
	initErr      error
}

func NewMainModel(workDir string, items []list.Item) MainModel {
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
	s.Spinner = spinner.Dot
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
			m.activeItem = &item
			m.logLines = []string{"[TUI] Spawning process for " + item.name + "..."}
			m.viewport.SetContent(strings.Join(m.logLines, "\n"))
			m.runErr = nil
			m.running = true

			m.runner = NewRunner(item.filePath)
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
		if m.running {
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
		// Populate list and transition to success view
		m.list.SetItems(items)
		m.state = viewInitSuccess
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
