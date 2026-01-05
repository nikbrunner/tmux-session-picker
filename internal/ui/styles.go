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

// Border characters (brutalist)
var (
	BorderColor = lipgloss.Color("238")
	LineHoriz   = lipgloss.NewStyle().Foreground(BorderColor).Render("─")
	LineVert    = lipgloss.NewStyle().Foreground(BorderColor).Render("│")
	CornerTL    = lipgloss.NewStyle().Foreground(BorderColor).Render("┌")
	CornerTR    = lipgloss.NewStyle().Foreground(BorderColor).Render("┐")
	CornerBL    = lipgloss.NewStyle().Foreground(BorderColor).Render("└")
	CornerBR    = lipgloss.NewStyle().Foreground(BorderColor).Render("┘")
	TeeLeft     = lipgloss.NewStyle().Foreground(BorderColor).Render("├")
	TeeRight    = lipgloss.NewStyle().Foreground(BorderColor).Render("┤")
)

// Styles
var (
	// Container styles
	AppStyle = lipgloss.NewStyle().
			Padding(0, 0)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(0, 1)

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

	SelectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("240")).
			Foreground(lipgloss.Color("255")).
			Bold(true).
			Padding(0, 1)

	// Window row styles (indented)
	WindowStyle = lipgloss.NewStyle().
			Padding(0, 1).
			PaddingLeft(7)

	WindowSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("240")).
				Foreground(lipgloss.Color("255")).
				Bold(true).
				Padding(0, 1).
				PaddingLeft(7)

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
	CollapsedIcon = lipgloss.NewStyle().Foreground(ColorSecondary).Render("▶")

	TimeStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	LastIcon = lipgloss.NewStyle().Foreground(ColorPrimary).Render("󰒮")

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

// HorizontalLine creates a horizontal line of specified width
func HorizontalLine(width int) string {
	line := ""
	for i := 0; i < width; i++ {
		line += "─"
	}
	return lipgloss.NewStyle().Foreground(BorderColor).Render(line)
}

// BoxTop creates a top border with optional title
func BoxTop(width int, title string) string {
	if title == "" {
		return CornerTL + HorizontalLine(width-2) + CornerTR
	}
	titleStyled := " " + HeaderStyle.Render(title) + " "
	// Account for ANSI codes in title length calculation
	titleLen := len(title) + 2
	leftLen := 2
	rightLen := width - leftLen - titleLen - 2
	if rightLen < 0 {
		rightLen = 0
	}
	return CornerTL + HorizontalLine(leftLen) + titleStyled + HorizontalLine(rightLen) + CornerTR
}

// BoxBottom creates a bottom border
func BoxBottom(width int) string {
	return CornerBL + HorizontalLine(width-2) + CornerBR
}

// BoxSeparator creates a separator line
func BoxSeparator(width int) string {
	return TeeLeft + HorizontalLine(width-2) + TeeRight
}
