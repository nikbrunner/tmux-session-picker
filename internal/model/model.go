package model

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nikbrunner/tsm/internal/claude"
	"github.com/nikbrunner/tsm/internal/config"
	"github.com/nikbrunner/tsm/internal/tmux"
	"github.com/nikbrunner/tsm/internal/ui"
)

// Mode represents the current UI mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeConfirmKill
	ModeCreate
)

// Item represents either a session or a window in the flattened list
type Item struct {
	IsSession    bool
	SessionIndex int // Index in the sessions slice
	WindowIndex  int // Index in the session's windows slice (only for windows)
}

// Model is the main application state
type Model struct {
	sessions       []tmux.Session
	claudeStatuses map[string]claude.Status
	currentSession string
	cursor         int
	items          []Item // Flattened list of visible items
	mode           Mode
	message        string
	messageIsError bool
	input          textinput.Model
	lastKeyTime    time.Time
	lastKey        string
	killTarget     string // Name of session/window being killed
	config         config.Config
	maxNameWidth   int    // For column alignment
	filter         string // Current filter text for fuzzy matching
}

// New creates a new Model
func New(currentSession string, cfg config.Config) Model {
	ti := textinput.New()
	ti.CharLimit = 50

	return Model{
		currentSession: currentSession,
		input:          ti,
		config:         cfg,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return m.loadSessions
}

// loadSessions fetches sessions from tmux
func (m Model) loadSessions() tea.Msg {
	sessions, err := tmux.ListSessions(m.currentSession)
	if err != nil {
		return errMsg{err}
	}
	return sessionsMsg{sessions}
}

type sessionsMsg struct {
	sessions []tmux.Session
}

type errMsg struct {
	err error
}

type clearMessageMsg struct{}

// clearMessageAfter returns a command that clears the message after a delay
func clearMessageAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sessionsMsg:
		m.sessions = msg.sessions
		m.loadClaudeStatuses()
		m.calculateColumnWidths()
		m.rebuildItems()
		if len(m.items) == 0 {
			m.message = "No other sessions. Press c to create one."
		}
		return m, nil

	case errMsg:
		m.message = fmt.Sprintf("Error: %v", msg.err)
		m.messageIsError = true
		return m, nil

	case clearMessageMsg:
		m.message = ""
		m.messageIsError = false
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Handle text input updates in create mode
	if m.mode == ModeCreate {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeNormal:
		return m.handleNormalMode(msg)
	case ModeConfirmKill:
		return m.handleConfirmKillMode(msg)
	case ModeCreate:
		return m.handleCreateMode(msg)
	}
	return m, nil
}

func (m *Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Cancel):
		// Escape: clear filter if active, otherwise quit
		if m.filter != "" {
			m.filter = ""
			m.rebuildItems()
			return m, nil
		}
		return m, tea.Quit

	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, keys.Down):
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}

	case key.Matches(msg, keys.Expand):
		m.expandCurrent()

	case key.Matches(msg, keys.Collapse):
		m.collapseCurrent()

	case key.Matches(msg, keys.Select):
		return m.selectCurrent()

	case key.Matches(msg, keys.Kill):
		return m.confirmKill()

	case key.Matches(msg, keys.Create):
		m.mode = ModeCreate
		m.filter = "" // Clear any active filter
		// Reset input completely
		m.input.Reset()
		m.input.SetValue("")
		m.input.Focus()
		return m, textinput.Blink

	// Number jumps (only when no filter active)
	case m.filter == "" && key.Matches(msg, keys.Jump1):
		return m.handleJump(1)
	case m.filter == "" && key.Matches(msg, keys.Jump2):
		return m.handleJump(2)
	case m.filter == "" && key.Matches(msg, keys.Jump3):
		return m.handleJump(3)
	case m.filter == "" && key.Matches(msg, keys.Jump4):
		return m.handleJump(4)
	case m.filter == "" && key.Matches(msg, keys.Jump5):
		return m.handleJump(5)
	case m.filter == "" && key.Matches(msg, keys.Jump6):
		return m.handleJump(6)
	case m.filter == "" && key.Matches(msg, keys.Jump7):
		return m.handleJump(7)
	case m.filter == "" && key.Matches(msg, keys.Jump8):
		return m.handleJump(8)
	case m.filter == "" && key.Matches(msg, keys.Jump9):
		return m.handleJump(9)

	case msg.Type == tea.KeyBackspace:
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.rebuildItems()
		}

	case msg.Type == tea.KeyRunes:
		// Add typed characters to filter
		m.filter += string(msg.Runes)
		m.rebuildItems()
	}

	return m, nil
}

func (m *Model) handleConfirmKillMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Confirm):
		return m.killCurrent(false)
	case key.Matches(msg, keys.Cancel):
		m.mode = ModeNormal
		m.message = ""
		m.killTarget = ""
	}

	return m, nil
}

func (m *Model) handleCreateMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Cancel):
		m.mode = ModeNormal
		m.input.Blur()
		return m, nil

	case msg.Type == tea.KeyEnter:
		name := strings.TrimSpace(m.input.Value())
		if name == "" {
			m.message = "Session name cannot be empty"
			m.messageIsError = true
			return m, nil
		}
		return m.createSession(name)
	}

	// Ignore ctrl key combinations - only pass regular typing to input
	if msg.Type == tea.KeyCtrlN || msg.Type == tea.KeyCtrlO ||
		msg.Type == tea.KeyCtrlJ || msg.Type == tea.KeyCtrlK ||
		msg.Type == tea.KeyCtrlH || msg.Type == tea.KeyCtrlL ||
		msg.Type == tea.KeyCtrlX || msg.Type == tea.KeyCtrlY ||
		msg.Type == tea.KeyCtrlP {
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) handleJump(num int) (tea.Model, tea.Cmd) {
	// Check if we're inside an expanded session - numbers switch to windows
	if m.cursor >= 0 && m.cursor < len(m.items) {
		item := m.items[m.cursor]
		var session *tmux.Session

		if !item.IsSession {
			session = &m.sessions[item.SessionIndex]
		} else {
			session = &m.sessions[item.SessionIndex]
		}

		if session.Expanded {
			// Jump to window number within this session
			for _, w := range session.Windows {
				if w.Index == num {
					target := fmt.Sprintf("%s:%d", session.Name, w.Index)
					if err := tmux.SwitchClient(target); err != nil {
						m.message = fmt.Sprintf("Error: %v", err)
						m.messageIsError = true
						return m, nil
					}
					return m, tea.Quit
				}
			}
		}
	}

	// Session labels: 1, 2, 3... map to session indices 0, 1, 2...
	sessionIdx := num - 1
	if sessionIdx >= 0 && sessionIdx < len(m.sessions) {
		session := m.sessions[sessionIdx]
		if err := tmux.SwitchClient(session.Name); err != nil {
			m.message = fmt.Sprintf("Error: %v", err)
			m.messageIsError = true
			return m, nil
		}
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) expandCurrent() {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return
	}

	item := m.items[m.cursor]
	if !item.IsSession {
		return
	}

	// Collapse all other sessions first
	for i := range m.sessions {
		m.sessions[i].Expanded = false
	}

	session := &m.sessions[item.SessionIndex]
	if len(session.Windows) == 0 {
		// Load windows
		windows, err := tmux.ListWindows(session.Name)
		if err != nil {
			m.message = fmt.Sprintf("Error loading windows: %v", err)
			m.messageIsError = true
			return
		}
		session.Windows = windows
	}
	session.Expanded = true
	m.rebuildItems()
}

func (m *Model) collapseCurrent() {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return
	}

	item := m.items[m.cursor]

	var sessionIdx int
	if item.IsSession {
		sessionIdx = item.SessionIndex
	} else {
		// Collapse parent session
		sessionIdx = item.SessionIndex
		// Move cursor to the session
		for i, it := range m.items {
			if it.IsSession && it.SessionIndex == sessionIdx {
				m.cursor = i
				break
			}
		}
	}

	m.sessions[sessionIdx].Expanded = false
	m.rebuildItems()
}

func (m *Model) selectCurrent() (tea.Model, tea.Cmd) {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return m, nil
	}

	item := m.items[m.cursor]

	var target string
	if item.IsSession {
		target = m.sessions[item.SessionIndex].Name
	} else {
		session := m.sessions[item.SessionIndex]
		window := session.Windows[item.WindowIndex]
		target = fmt.Sprintf("%s:%d", session.Name, window.Index)
	}

	if err := tmux.SwitchClient(target); err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
		m.messageIsError = true
		return m, nil
	}

	return m, tea.Quit
}

func (m *Model) confirmKill() (tea.Model, tea.Cmd) {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return m, nil
	}

	item := m.items[m.cursor]

	if item.IsSession {
		m.killTarget = m.sessions[item.SessionIndex].Name
		m.message = fmt.Sprintf("Kill \"%s\"?", m.killTarget)
	} else {
		session := m.sessions[item.SessionIndex]
		window := session.Windows[item.WindowIndex]
		m.killTarget = fmt.Sprintf("%s:%d", session.Name, window.Index)
		m.message = fmt.Sprintf("Kill window \"%s\"?", m.killTarget)
	}

	m.mode = ModeConfirmKill
	return m, nil
}

func (m *Model) killCurrent(instant bool) (tea.Model, tea.Cmd) {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return m, nil
	}

	item := m.items[m.cursor]
	var err error

	if item.IsSession {
		session := m.sessions[item.SessionIndex]
		err = tmux.KillSession(session.Name)
		if err == nil {
			m.message = fmt.Sprintf("Killed \"%s\"", session.Name)
		}
	} else {
		session := m.sessions[item.SessionIndex]
		window := session.Windows[item.WindowIndex]
		err = tmux.KillWindow(session.Name, window.Index)
		if err == nil {
			m.message = fmt.Sprintf("Killed window %d", window.Index)
		}
	}

	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
		m.messageIsError = true
	}

	m.mode = ModeNormal
	m.killTarget = ""

	// Reload sessions and clear message after 5 seconds
	return m, tea.Batch(m.loadSessions, clearMessageAfter(5*time.Second))
}

func (m *Model) createSession(name string) (tea.Model, tea.Cmd) {
	homeDir := os.Getenv("HOME")
	if err := tmux.CreateSession(name, homeDir); err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
		m.messageIsError = true
		m.mode = ModeNormal
		m.input.Blur()
		return m, nil
	}

	// Apply layout if configured
	m.applyLayout(name, homeDir)

	// Switch to the new session
	if err := tmux.SwitchClient(name); err != nil {
		m.message = fmt.Sprintf("Created but failed to switch: %v", err)
		m.messageIsError = true
		return m, m.loadSessions
	}

	return m, tea.Quit
}

func (m *Model) applyLayout(sessionName, workingDir string) {
	if m.config.Layout == "" {
		return
	}

	scriptPath := fmt.Sprintf("%s/%s.sh", m.config.LayoutDir, m.config.Layout)
	if _, err := os.Stat(scriptPath); err != nil {
		return
	}

	// Execute layout script (fire and forget)
	go func() {
		cmd := fmt.Sprintf("%s %s %s", scriptPath, sessionName, workingDir)
		_ = os.Setenv("TMUX_SESSION", sessionName)
		_ = os.Setenv("TMUX_WORKING_DIR", workingDir)
		// Note: In production, you'd want proper error handling here
		_, _ = os.StartProcess("/bin/sh", []string{"/bin/sh", "-c", cmd}, &os.ProcAttr{})
	}()
}

func (m *Model) loadClaudeStatuses() {
	m.claudeStatuses = make(map[string]claude.Status)
	if !m.config.ClaudeStatusEnabled {
		return
	}
	for _, s := range m.sessions {
		status := claude.GetStatus(s.Name, m.config.CacheDir)
		if status.State != "" {
			m.claudeStatuses[s.Name] = status
		}
	}
}

func (m *Model) calculateColumnWidths() {
	m.maxNameWidth = 0
	for _, s := range m.sessions {
		if len(s.Name) > m.maxNameWidth {
			m.maxNameWidth = len(s.Name)
		}
	}
}

func (m *Model) rebuildItems() {
	m.items = nil
	filterLower := strings.ToLower(m.filter)

	for i, session := range m.sessions {
		// Apply fuzzy filter if active
		if m.filter != "" && !fuzzyMatch(session.Name, filterLower) {
			continue
		}

		m.items = append(m.items, Item{
			IsSession:    true,
			SessionIndex: i,
		})

		if session.Expanded {
			for j := range session.Windows {
				m.items = append(m.items, Item{
					IsSession:    false,
					SessionIndex: i,
					WindowIndex:  j,
				})
			}
		}
	}

	// Ensure cursor is in bounds
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// fuzzyMatch checks if the pattern matches the text (case-insensitive, substring match)
func fuzzyMatch(text, pattern string) bool {
	textLower := strings.ToLower(text)
	return strings.Contains(textLower, pattern)
}

// View implements tea.Model
func (m Model) View() string {
	var b strings.Builder

	// Header with optional filter
	if m.filter != "" {
		b.WriteString(ui.HeaderStyle.Render("tsm"))
		b.WriteString("  ")
		b.WriteString(ui.FilterStyle.Render("/" + m.filter))
	} else {
		b.WriteString(ui.HeaderStyle.Render("tsm"))
	}
	b.WriteString("\n\n")

	// Session list
	sessionNum := 0
	for i, item := range m.items {
		selected := i == m.cursor

		if item.IsSession {
			session := m.sessions[item.SessionIndex]
			sessionNum++
			isFirst := sessionNum == 1
			b.WriteString(m.renderSessionWithLabel(session, sessionNum, isFirst, selected))
		} else {
			session := m.sessions[item.SessionIndex]
			window := session.Windows[item.WindowIndex]
			b.WriteString(m.renderWindow(window, selected))
		}
		b.WriteString("\n")
	}

	// Empty state
	if len(m.items) == 0 {
		if m.filter != "" {
			b.WriteString("  No sessions matching filter\n")
		} else {
			b.WriteString("  No other sessions available\n")
		}
	}

	b.WriteString("\n")

	// Message line
	if m.message != "" {
		if m.messageIsError {
			b.WriteString(ui.ErrorMessageStyle.Render(m.message))
		} else {
			b.WriteString(ui.MessageStyle.Render(m.message))
		}
		b.WriteString("\n")
	} else if m.mode == ModeCreate {
		b.WriteString(ui.InputPromptStyle.Render(" New session: "))
		b.WriteString(m.input.View())
		b.WriteString("\n")
	}

	// Help line
	switch m.mode {
	case ModeNormal:
		if m.filter != "" {
			b.WriteString(ui.FooterStyle.Render(ui.HelpFiltering()))
		} else {
			b.WriteString(ui.FooterStyle.Render(ui.HelpNormal()))
		}
	case ModeConfirmKill:
		b.WriteString(ui.FooterStyle.Render(ui.HelpConfirmKill()))
	case ModeCreate:
		b.WriteString(ui.FooterStyle.Render(ui.HelpCreate()))
	}

	return ui.AppStyle.Render(b.String())
}

func (m Model) renderSessionWithLabel(session tmux.Session, num int, isFirst bool, selected bool) string {
	// Build the row with fixed-width columns
	var b strings.Builder

	// Number label
	label := fmt.Sprintf("%d", num)
	if selected {
		b.WriteString(ui.IndexSelectedStyle.Render(label))
	} else {
		b.WriteString(ui.IndexStyle.Render(label))
	}
	b.WriteString(" ")

	// Last session icon (fixed width column)
	if isFirst {
		b.WriteString(ui.LastIcon)
	} else {
		b.WriteString(" ")
	}
	b.WriteString(" ")

	// Expand icon
	if session.Expanded {
		b.WriteString(ui.ExpandedIcon)
	} else {
		b.WriteString(ui.CollapsedIcon)
	}
	b.WriteString(" ")

	// Session name (padded to max width)
	namePadded := fmt.Sprintf("%-*s", m.maxNameWidth, session.Name)
	if selected {
		b.WriteString(ui.SessionNameSelectedStyle.Render(namePadded))
	} else {
		b.WriteString(namePadded)
	}
	b.WriteString("  ")

	// Time ago (fixed width 8)
	timeAgo := formatTimeAgo(session.LastActivity)
	timePadded := fmt.Sprintf("%-8s", timeAgo)
	b.WriteString(ui.TimeStyle.Render(timePadded))

	// Claude status
	if status, ok := m.claudeStatuses[session.Name]; ok {
		b.WriteString(" ")
		b.WriteString(ui.FormatClaudeStatus(status.State))
	}

	return ui.SessionStyle.Render(b.String())
}

func (m Model) renderWindow(window tmux.Window, selected bool) string {
	var b strings.Builder

	// Window index and name
	windowText := fmt.Sprintf("%d: %s", window.Index, window.Name)
	if selected {
		b.WriteString(ui.WindowNameSelectedStyle.Render(windowText))
	} else {
		b.WriteString(windowText)
	}

	return ui.WindowStyle.Render(b.String())
}

func formatTimeAgo(t time.Time) string {
	d := time.Since(t)

	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}
