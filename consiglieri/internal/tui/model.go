package tui

import (
	"context"
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
)

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

	return MainModel{
		state:    viewDashboard,
		list:     l,
		viewport: vp,
		spinner:  s,
		workDir:  workDir,
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
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

		case "enter":
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

	case tea.WindowSizeMsg:
		h, v := DocStyle.GetFrameSize()
		m.width = msg.Width - h
		m.height = msg.Height - v

		// Calculate sizes based on header/footer and margins
		mainHeight := m.height - 6
		if mainHeight < 10 {
			mainHeight = 10
		}

		// Split 1/3 and 2/3
		leftWidth := m.width / 3
		if leftWidth < 20 {
			leftWidth = 20
		}
		rightWidth := m.width - leftWidth - 4
		if rightWidth < 20 {
			rightWidth = 20
		}

		// Set sizes for list (inside left pane: border/padding subtracts 4 lines/columns)
		m.list.SetSize(leftWidth-4, mainHeight-4)

		// Set sizes for viewport (inside right pane: border/padding subtracts 4; 3 lines for titles/status)
		m.viewport.Width = rightWidth - 4
		m.viewport.Height = mainHeight - 7
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

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Update children based on focus or running state
	var cmd tea.Cmd
	if m.running || len(m.logLines) > 0 {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
