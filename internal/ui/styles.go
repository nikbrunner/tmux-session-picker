package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ANSI 16 colors - adapts to terminal theme
// 0-7: black, red, green, yellow, blue, magenta, cyan, white
// 8-15: bright variants
var (
	ColorPrimary   = lipgloss.Color("4") // Blue
	ColorSecondary = lipgloss.Color("7") // White/light gray
	ColorSuccess   = lipgloss.Color("2") // Green
	ColorWarning   = lipgloss.Color("3") // Yellow
	ColorError     = lipgloss.Color("1") // Red
	ColorDim       = lipgloss.Color("8") // Bright black (dark gray)
	ColorClaude    = lipgloss.Color("5") // Magenta (distinctive for Claude)
)

// Border and padding overhead for the app container
const (
	// AppBorderOverhead is the total cells used by border + padding per axis
	AppBorderOverheadX = 4 // left border + left padding + right padding + right border
	AppBorderOverheadY = 2 // top border + bottom border (no vertical padding)
)

// Styles
var (
	// Container styles
	AppStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorDim).
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

	// Window row styles (indented)
	WindowStyle = lipgloss.NewStyle().
			Padding(0, 1).
			PaddingLeft(10)

	// Text styles
	IndexStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Width(3)

	IndexSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true).
				Width(3)

	SessionNameSelectedStyle = lipgloss.NewStyle().
					Foreground(ColorWarning).
					Bold(true)

	WindowNameSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorWarning).
				Bold(true)

	ExpandedIcon  = lipgloss.NewStyle().Foreground(ColorPrimary).Render("▼")
	CollapsedIcon = lipgloss.NewStyle().Foreground(ColorDim).Render("▶")

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
			Foreground(ColorDim)

	HelpSepStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	// Filter style
	FilterStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	// Border style
	BorderStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	// Statusline style
	StatuslineStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Padding(0, 1)
)

// RenderBorder returns a horizontal border line
func RenderBorder(width int) string {
	return BorderStyle.Render(strings.Repeat("─", width))
}

// FormatClaudeStatus formats the Claude status for display
func FormatClaudeStatus(state string) string {
	if state == "" {
		return ""
	}

	label := ClaudeLabelStyle.Render("Claude Code:")

	switch state {
	case "new":
		// Don't show badge for "new" - it's just noise
		return ""
	case "working":
		return "[" + label + " " + ClaudeWorkingStyle.Render("⏳ working...") + "]"
	case "waiting":
		return "[" + label + " " + ClaudeWaitingStyle.Render("💬 waiting...") + "]"
	default:
		return ""
	}
}

// ScrollbarChars returns scrollbar characters for each visible line
// totalItems: total number of items in the list
// visibleItems: number of items currently visible
// scrollOffset: current scroll position (first visible item index)
// height: number of lines to render scrollbar for
func ScrollbarChars(totalItems, visibleItems, scrollOffset, height int) []string {
	result := make([]string, height)

	// No scrollbar needed if all items fit
	if totalItems <= visibleItems || height <= 0 {
		for i := range result {
			result[i] = " "
		}
		return result
	}

	// Calculate thumb size (minimum 1 line)
	thumbSize := (visibleItems * height) / totalItems
	if thumbSize < 1 {
		thumbSize = 1
	}

	// Calculate thumb position
	scrollRange := totalItems - visibleItems
	trackRange := height - thumbSize
	thumbPos := 0
	if scrollRange > 0 && trackRange > 0 {
		thumbPos = (scrollOffset * trackRange) / scrollRange
	}

	// Build scrollbar
	trackChar := BorderStyle.Render("│")
	thumbChar := lipgloss.NewStyle().Foreground(ColorSecondary).Render("┃")

	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			result[i] = thumbChar
		} else {
			result[i] = trackChar
		}
	}

	return result
}
