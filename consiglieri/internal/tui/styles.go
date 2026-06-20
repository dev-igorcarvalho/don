package tui

import "github.com/charmbracelet/lipgloss"

const (
	ColorDarkCharcoal = "#1C1C1E"
	ColorSlateGray    = "#2C2C2E"
	ColorLavender     = "#8E5FE6"
	ColorCoolMint     = "#00F0FF"
	ColorBrightAmber  = "#FFB800"
	ColorSoftWhite    = "#F2F2F7"
	ColorMutedGray    = "#8E8E93"
)

var (
	// Layout & border styles
	DocStyle = lipgloss.NewStyle().Margin(1, 2)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorDarkCharcoal)).
			Background(lipgloss.Color(ColorLavender)).
			Padding(0, 1).
			Bold(true)

	ActivePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorLavender)).
			Padding(1)

	InactivePaneStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorSlateGray)).
				Padding(1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorLavender)).
			Bold(true)

	MutedTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMutedGray))

	// Status line styles
	RunningStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorCoolMint)).
				Bold(true)

	ErrorStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorBrightAmber)).
				Bold(true)

	SuccessStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorCoolMint)).
				Bold(true)
)
