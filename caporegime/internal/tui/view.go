package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/dev-igorcarvalho/don/caporegime/pkg/utils"
)

func (m MainModel) View() string {
	if m.state == viewInitSuccess {
		donName := utils.GetDonName()
		initSuccessView := lipgloss.JoinVertical(
			lipgloss.Left,
			TitleStyle.Render(fmt.Sprintf("🕴️ Welcome to the Family, %s.", donName)),
			"",
			"\"A man who doesn't spend time with his family can never be a real man.\"",
			"",
			SuccessStatusStyle.Render("Workspace initialized with respect at:"),
			"  • "+m.workDir+"/",
			"",
			"The foundation of our Cosa Nostra is laid.",
			"",
			RunningStatusStyle.Render("▶ Press [Enter] to assume control of the dashboard"),
		)

		mainHeight := m.mainHeight()
		w := m.width - 4
		if w < 20 {
			w = 20
		}
		h := mainHeight - 4
		if h < 10 {
			h = 10
		}

		pane := ActivePaneStyle.
			Width(w).
			Height(h).
			Render(initSuccessView)

		return DocStyle.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			HeaderStyle.Render(" DON CAPOREGIME WORKSPACE INITIALIZED "),
			"",
			pane,
			"",
			MutedTextStyle.Render("enter: continue to dashboard • q: quit"),
		))
	}

	if m.state == viewInit {
		var initView string
		switch {
		case m.initializing:
			initView = lipgloss.JoinVertical(
				lipgloss.Left,
				TitleStyle.Render("Initializing Workspace..."),
				"",
				m.spinner.View()+" Creating directory structure and sample workflow...",
			)
		case m.initErr != nil:
			initView = lipgloss.JoinVertical(
				lipgloss.Left,
				TitleStyle.Render("Workspace Initialization Error"),
				"",
				ErrorStatusStyle.Render("Error: "+m.initErr.Error()),
				"",
				RunningStatusStyle.Render("▶ Press [i] to try again"),
				MutedTextStyle.Render("▶ Press [q] or [ctrl+c] to quit"),
			)
		default:
			initView = lipgloss.JoinVertical(
				lipgloss.Left,
				TitleStyle.Render("Welcome to Don Caporegime!"),
				"",
				"No agentic workflows were found in your workspace.",
				"To get started, we need to initialize the workspace directory structure.",
				"",
				MutedTextStyle.Render("This will create:"),
				"  • "+m.workDir+"/",
				"  • "+m.workDir+"/hello.go (a sample Hello Workflow)",
				"",
				RunningStatusStyle.Render("▶ Press [i] to initialize the workspace"),
				MutedTextStyle.Render("▶ Press [q] or [ctrl+c] to quit"),
			)
		}

		mainHeight := m.mainHeight()
		w := m.width - 4
		if w < 20 {
			w = 20
		}
		h := mainHeight - 4
		if h < 10 {
			h = 10
		}

		pane := ActivePaneStyle.
			Width(w).
			Height(h).
			Render(initView)

		return DocStyle.Render(lipgloss.JoinVertical(
			lipgloss.Left,
			HeaderStyle.Render(" DON CAPOREGIME WORKFLOW SETUP "),
			"",
			pane,
			"",
			MutedTextStyle.Render("ctrl+c/q: quit • i: initialize workspace"),
		))
	}

	// 1. Left View: Workflows List
	leftView := m.list.View()

	// 2. Right View: Details or Logs
	var rightView string
	var rightTitle string

	selectedItem := m.list.SelectedItem()
	mainHeight := m.mainHeight()

	switch {
	case m.running || len(m.logLines) > 0:
		rightTitle = "EXECUTION LOGS: "
		if m.activeItem != nil {
			rightTitle += m.activeItem.name
		}

		var statusStr string
		switch {
		case m.running:
			statusStr = RunningStatusStyle.Render(m.spinner.View() + " Running...")
		case m.runErr != nil:
			statusStr = ErrorStatusStyle.Render("❌ Failed: " + m.runErr.Error())
		default:
			statusStr = SuccessStatusStyle.Render("✅ Completed successfully")
		}

		content := m.viewport.View()

		rightView = lipgloss.JoinVertical(
			lipgloss.Left,
			TitleStyle.Render(rightTitle),
			statusStr,
			"",
			content,
		)
	case selectedItem != nil:
		item := selectedItem.(WorkflowItem)
		rightTitle = "WORKFLOW DETAILS:"

		var statusInfo string
		var actionHelp string

		switch item.buildStatus {
		case BuildStatusCompiling:
			statusInfo = RunningStatusStyle.Render("Status: " + m.spinner.View() + " Compiling background binary...")
			actionHelp = MutedTextStyle.Render("▶ [Please wait for compilation to complete]")
		case BuildStatusFailed:
			statusInfo = ErrorStatusStyle.Render("Status: ❌ Compilation failed")
			errMsg := item.buildError
			if len(errMsg) > 300 {
				errMsg = errMsg[:300] + "... (truncated)"
			}
			statusInfo = fmt.Sprintf("%s\n\n%s\n%s", statusInfo, MutedTextStyle.Render("Compiler Output:"), errMsg)
			actionHelp = ErrorStatusStyle.Render("▶ [Cannot run: Fix syntax or dependency errors]")
		default:
			statusInfo = SuccessStatusStyle.Render("Status: ✅ Ready to run")
			actionHelp = RunningStatusStyle.Render("▶ Press [Enter] to run this workflow")
		}

		rightView = lipgloss.JoinVertical(
			lipgloss.Left,
			TitleStyle.Render(rightTitle),
			"",
			MutedTextStyle.Render("Name: ")+item.name,
			"",
			MutedTextStyle.Render("Path: ")+item.filePath,
			"",
			statusInfo,
			"",
			MutedTextStyle.Render("Description:"),
			item.description,
			"",
			actionHelp,
		)
	default:
		rightView = "No workflow selected."
	}

	// Adjust pane dimensions
	leftWidth := m.leftWidth()
	rightWidth := m.rightWidth()

	leftPane := InactivePaneStyle.
		Width(leftWidth - 4).
		Height(mainHeight - 4).
		Render(leftView)

	var rightPane string
	switch {
	case m.running:
		rightPane = ActivePaneStyle.
			Width(rightWidth - 4).
			Height(mainHeight - 4).
			Render(rightView)
	default:
		rightPane = InactivePaneStyle.
			Width(rightWidth - 4).
			Height(mainHeight - 4).
			Render(rightView)
	}

	mainLayout := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, "  ", rightPane)

	// Help Text
	var helpText string
	switch {
	case m.running:
		helpText = "ctrl+c: quit • enter: stop execution • pgup/pgdn: scroll logs"
	case len(m.logLines) > 0:
		helpText = "ctrl+c: quit • esc: back to details • pgup/pgdn: scroll logs • enter: run again"
	default:
		helpText = "ctrl+c: quit • enter: run workflow • ↑/↓: navigate • /: search"
	}

	footer := MutedTextStyle.Render(helpText)

	// Stitch everything together inside DocStyle margin
	return DocStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		HeaderStyle.Render(" DON CAPOREGIME WORKFLOW DASHBOARD "),
		"",
		mainLayout,
		"",
		footer,
	))
}
