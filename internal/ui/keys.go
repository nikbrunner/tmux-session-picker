package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the application
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Expand   key.Binding
	Collapse key.Binding
	Select   key.Binding
	Kill     key.Binding
	Create   key.Binding
	Quit     key.Binding
	Cancel   key.Binding
	Confirm  key.Binding
	Jump1    key.Binding
	Jump2    key.Binding
	Jump3    key.Binding
	Jump4    key.Binding
	Jump5    key.Binding
	Jump6    key.Binding
	Jump7    key.Binding
	Jump8    key.Binding
	Jump9    key.Binding
}

// DefaultKeyMap returns the default key bindings
var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j", "down"),
	),
	Expand: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("l", "expand"),
	),
	Collapse: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("h", "collapse"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "switch"),
	),
	Kill: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "kill"),
	),
	Create: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "new"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc"),
		key.WithHelp("q/esc", "quit"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "confirm"),
	),
	Jump1: key.NewBinding(key.WithKeys("1")),
	Jump2: key.NewBinding(key.WithKeys("2")),
	Jump3: key.NewBinding(key.WithKeys("3")),
	Jump4: key.NewBinding(key.WithKeys("4")),
	Jump5: key.NewBinding(key.WithKeys("5")),
	Jump6: key.NewBinding(key.WithKeys("6")),
	Jump7: key.NewBinding(key.WithKeys("7")),
	Jump8: key.NewBinding(key.WithKeys("8")),
	Jump9: key.NewBinding(key.WithKeys("9")),
}

// HelpNormal returns the help text for normal mode
func HelpNormal() string {
	return HelpKeyStyle.Render("1-9") + HelpDescStyle.Render(" jump  ") +
		HelpKeyStyle.Render("j/k") + HelpDescStyle.Render(" nav  ") +
		HelpKeyStyle.Render("h/l") + HelpDescStyle.Render(" expand  ") +
		HelpKeyStyle.Render("x") + HelpDescStyle.Render(" kill  ") +
		HelpKeyStyle.Render("c") + HelpDescStyle.Render(" new")
}

// HelpConfirmKill returns the help text for kill confirmation mode
func HelpConfirmKill() string {
	return HelpKeyStyle.Render("x") + HelpDescStyle.Render(" confirm  ") +
		HelpKeyStyle.Render("esc") + HelpDescStyle.Render(" cancel")
}

// HelpCreate returns the help text for create mode
func HelpCreate() string {
	return HelpKeyStyle.Render("enter") + HelpDescStyle.Render(" create  ") +
		HelpKeyStyle.Render("esc") + HelpDescStyle.Render(" cancel")
}
