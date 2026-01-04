package ui

import "github.com/charmbracelet/lipgloss"

// Colors
var (
	ColorPrimary   = lipgloss.Color("212") // Light blue
	ColorSecondary = lipgloss.Color("241") // Gray
	ColorSuccess   = lipgloss.Color("120") // Green
	ColorWarning   = lipgloss.Color("214") // Orange/Yellow
	ColorError     = lipgloss.Color("196") // Red
	ColorDim       = lipgloss.Color("240") // Dim gray
	ColorClaude    = lipgloss.Color("209") // Claude orange
)

// Styles
var (
	// Container styles
	AppStyle = lipgloss.NewStyle().
			Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1).
			MarginBottom(1)

	FooterStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Padding(0, 1)

	MessageStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Padding(0, 1)

	ErrorMessageStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Padding(0, 1)

	// Session row styles
	SessionStyle = lipgloss.NewStyle().
			Padding(0, 1)

	SelectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	// Window row styles (indented)
	WindowStyle = lipgloss.NewStyle().
			Padding(0, 1).
			PaddingLeft(7)

	WindowSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Padding(0, 1).
				PaddingLeft(7)

	// Text styles
	IndexStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Width(3)

	ExpandedIcon   = lipgloss.NewStyle().Foreground(ColorPrimary).Render("▼")
	CollapsedIcon  = lipgloss.NewStyle().Foreground(ColorSecondary).Render("▶")

	TimeStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	// Claude status styles
	ClaudeNewStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	ClaudeWorkingStyle = lipgloss.NewStyle().
				Foreground(ColorWarning)

	ClaudeWaitingStyle = lipgloss.NewStyle().
				Foreground(ColorSuccess)

	ClaudeLabelStyle = lipgloss.NewStyle().
				Foreground(ColorClaude)

	// Input styles
	InputPromptStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary)

	// Help styles
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorDim)
)

// FormatClaudeStatus formats the Claude status for display
func FormatClaudeStatus(state string) string {
	if state == "" {
		return ""
	}

	label := ClaudeLabelStyle.Render("CC:")

	switch state {
	case "new":
		return "[" + label + " " + ClaudeNewStyle.Render("new") + "]"
	case "working":
		return "[" + label + " " + ClaudeWorkingStyle.Render("working") + "]"
	case "waiting":
		return "[" + label + " " + ClaudeWaitingStyle.Render("waiting") + "]"
	default:
		return ""
	}
}
