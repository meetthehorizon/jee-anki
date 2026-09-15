package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Brand colors
	ColorPrimary   = lipgloss.Color("#7D56F4")
	ColorSecondary = lipgloss.Color("#2563EB")
	ColorSuccess   = lipgloss.Color("#10B981")
	ColorWarning   = lipgloss.Color("#F59E0B")
	ColorError     = lipgloss.Color("#EF4444")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorWhite     = lipgloss.Color("#FFFFFF")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorPrimary).
			Padding(0, 2).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true).
			MarginBottom(1)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			MarginBottom(1)

	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSuccess)

	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWarning)

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorError)

	HighlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	CodeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary).
			Background(lipgloss.Color("#1E293B")).
			Padding(0, 1)
)
