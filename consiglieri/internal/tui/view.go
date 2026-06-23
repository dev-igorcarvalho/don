package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m MainModel) View() string {
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

		rightView = lipgloss.JoinVertical(
			lipgloss.Left,
			TitleStyle.Render(rightTitle),
			"",
			MutedTextStyle.Render("Name: ")+item.name,
			"",
			MutedTextStyle.Render("Path: ")+item.filePath,
			"",
			MutedTextStyle.Render("Description:"),
			item.description,
			"",
			"",
			RunningStatusStyle.Render("▶ Press [Enter] to run this workflow"),
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
		HeaderStyle.Render(" DON CONSIGLIERI WORKFLOW DASHBOARD "),
		"",
		mainLayout,
		"",
		footer,
	))
}
