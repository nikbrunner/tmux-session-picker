package ui

import "github.com/charmbracelet/lipgloss"

// ANSI 16 colors - adapts to terminal theme
// 0-7: black, red, green, yellow, blue, magenta, cyan, white
// 8-15: bright variants
var (
	ColorPrimary   = lipgloss.Color("4")  // Blue
	ColorSecondary = lipgloss.Color("7")  // White/light gray
	ColorSuccess   = lipgloss.Color("2")  // Green
	ColorWarning   = lipgloss.Color("3")  // Yellow
	ColorError     = lipgloss.Color("1")  // Red
	ColorDim       = lipgloss.Color("8")  // Bright black (dark gray)
	ColorClaude    = lipgloss.Color("5")  // Magenta (distinctive for Claude)
	ColorMuted     = lipgloss.Color("8")  // Bright black (dark gray)
)

// Styles
var (
	// Container styles
	AppStyle = lipgloss.NewStyle().
			Padding(0, 0)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

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

	// Window row styles (indented)
	WindowStyle = lipgloss.NewStyle().
			Padding(0, 1).
			PaddingLeft(10)

	// Text styles
	IndexStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Width(3)

	IndexSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Width(3)

	SessionNameSelectedStyle = lipgloss.NewStyle().
					Foreground(ColorPrimary).
					Bold(true)

	WindowNameSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	ExpandedIcon  = lipgloss.NewStyle().Foreground(ColorPrimary).Render("▼")
	CollapsedIcon = lipgloss.NewStyle().Foreground(ColorMuted).Render("▶")

	TimeStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	LastIcon = lipgloss.NewStyle().Foreground(ColorWarning).Render("󰒮")

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
			Foreground(ColorPrimary).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	HelpSepStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	// Filter style
	FilterStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)
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
