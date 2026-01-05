package model

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	ModePickDirectory
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
	killTarget     string // Name of session/window being killed
	config         config.Config
	maxNameWidth   int    // For column alignment
	filter         string // Current filter text for fuzzy matching

	// Directory picker state
	repoDirs     []string // All scanned directories (relative paths like "owner/repo")
	repoFiltered []string // Filtered list based on repoFilter
	repoFilter   string   // Current filter text for directory picker
	repoCursor   int      // Selected item in directory list

	// Scroll state
	scrollOffset     int // Scroll offset for session list
	repoScrollOffset int // Scroll offset for directory picker

	// Window size
	width  int
	height int
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
		m.setError("Error: %v", msg.err)
		return m, nil

	case clearMessageMsg:
		m.message = ""
		m.messageIsError = false
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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
	case ModePickDirectory:
		return m.handlePickDirectoryMode(msg)
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
			m.updateScrollOffset()
		}

	case key.Matches(msg, keys.Down):
		if m.cursor < len(m.items)-1 {
			m.cursor++
			m.updateScrollOffset()
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

	case key.Matches(msg, keys.PickDirectory):
		m.mode = ModePickDirectory
		m.filter = "" // Clear any active filter
		m.repoFilter = ""
		m.repoCursor = 0
		m.repoScrollOffset = 0
		m.repoDirs = m.scanRepoDirectories()
		m.repoFiltered = m.repoDirs
		// Request window size to get proper height for layout
		return m, tea.WindowSize()

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
	case key.Matches(msg, keys.Kill):
		// Double C-x confirms the kill
		return m.killCurrent()
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
			m.setError("Session name cannot be empty")
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

func (m *Model) handlePickDirectoryMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keys := ui.DefaultKeyMap

	switch {
	case key.Matches(msg, keys.Cancel):
		// Clear filter first, then exit on second press
		if m.repoFilter != "" {
			m.repoFilter = ""
			m.repoFiltered = m.repoDirs
			m.repoCursor = 0
			return m, nil
		}
		m.mode = ModeNormal
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.repoCursor > 0 {
			m.repoCursor--
			m.updateRepoScrollOffset()
		}

	case key.Matches(msg, keys.Down):
		if m.repoCursor < len(m.repoFiltered)-1 {
			m.repoCursor++
			m.updateRepoScrollOffset()
		}

	case key.Matches(msg, keys.Select):
		if len(m.repoFiltered) > 0 && m.repoCursor < len(m.repoFiltered) {
			return m.createSessionFromDir(m.repoFiltered[m.repoCursor])
		}

	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case msg.Type == tea.KeyBackspace:
		if len(m.repoFilter) > 0 {
			m.repoFilter = m.repoFilter[:len(m.repoFilter)-1]
			m.filterRepoDirs()
		}

	case msg.Type == tea.KeyRunes:
		m.repoFilter += string(msg.Runes)
		m.filterRepoDirs()
	}

	return m, nil
}

// filterRepoDirs filters the repo directories based on repoFilter
func (m *Model) filterRepoDirs() {
	if m.repoFilter == "" {
		m.repoFiltered = m.repoDirs
	} else {
		filterLower := strings.ToLower(m.repoFilter)
		m.repoFiltered = nil
		for _, dir := range m.repoDirs {
			if fuzzyMatch(dir, filterLower) {
				m.repoFiltered = append(m.repoFiltered, dir)
			}
		}
	}
	// Reset cursor if out of bounds
	if m.repoCursor >= len(m.repoFiltered) {
		m.repoCursor = len(m.repoFiltered) - 1
	}
	if m.repoCursor < 0 {
		m.repoCursor = 0
	}
	m.updateRepoScrollOffset()
}

func (m *Model) createSessionFromDir(relPath string) (tea.Model, tea.Cmd) {
	// relPath is like "owner/repo" - convert to full path
	fullPath := filepath.Join(m.config.ReposDir, relPath)

	// Convert "owner/repo" to "owner-repo" for session name
	name := strings.ReplaceAll(relPath, "/", "-")

	if err := tmux.CreateSession(name, fullPath); err != nil {
		m.setError("Error: %v", err)
		m.mode = ModeNormal
		return m, nil
	}

	// Apply layout if configured
	m.applyLayout(name, fullPath)

	// Switch to the new session
	if err := tmux.SwitchClient(name); err != nil {
		m.setError("Created but failed to switch: %v", err)
		return m, m.loadSessions
	}

	return m, tea.Quit
}

// scanRepoDirectories scans the repos directory at the configured depth
// and returns relative paths like "owner/repo"
func (m *Model) scanRepoDirectories() []string {
	var dirs []string
	baseDir := m.config.ReposDir
	depth := m.config.ReposDepth

	// Walk directories at exactly the target depth
	m.walkAtDepth(baseDir, "", depth, &dirs)

	return dirs
}

// walkAtDepth recursively walks directories and collects paths at the target depth
func (m *Model) walkAtDepth(baseDir, currentPath string, remainingDepth int, dirs *[]string) {
	if remainingDepth == 0 {
		// We've reached the target depth - add this directory
		if currentPath != "" {
			*dirs = append(*dirs, currentPath)
		}
		return
	}

	// Read the current directory
	fullPath := filepath.Join(baseDir, currentPath)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		var nextPath string
		if currentPath == "" {
			nextPath = entry.Name()
		} else {
			nextPath = filepath.Join(currentPath, entry.Name())
		}

		m.walkAtDepth(baseDir, nextPath, remainingDepth-1, dirs)
	}
}

func (m *Model) handleJump(num int) (tea.Model, tea.Cmd) {
	// Check if we're inside an expanded session - numbers switch to windows
	if m.cursor >= 0 && m.cursor < len(m.items) {
		item := m.items[m.cursor]
		session := &m.sessions[item.SessionIndex]

		if session.Expanded {
			// Jump to window number within this session
			for _, w := range session.Windows {
				if w.Index == num {
					target := fmt.Sprintf("%s:%d", session.Name, w.Index)
					if err := tmux.SwitchClient(target); err != nil {
						m.setError("Error: %v", err)
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
			m.setError("Error: %v", err)
			return m, nil
		}
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) expandCurrent() {
	if !m.isCursorValid() {
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
			m.setError("Error loading windows: %v", err)
			return
		}
		session.Windows = windows
	}
	session.Expanded = true
	m.rebuildItems()
}

func (m *Model) collapseCurrent() {
	if !m.isCursorValid() {
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
	if !m.isCursorValid() {
		return m, nil
	}

	target := m.getTargetName(m.items[m.cursor])
	if err := tmux.SwitchClient(target); err != nil {
		m.setError("Error: %v", err)
		return m, nil
	}

	return m, tea.Quit
}

func (m *Model) confirmKill() (tea.Model, tea.Cmd) {
	if !m.isCursorValid() {
		return m, nil
	}

	item := m.items[m.cursor]
	m.killTarget = m.getTargetName(item)

	if item.IsSession {
		m.message = fmt.Sprintf("Kill \"%s\"?", m.killTarget)
	} else {
		m.message = fmt.Sprintf("Kill window \"%s\"?", m.killTarget)
	}

	m.mode = ModeConfirmKill
	return m, nil
}

func (m *Model) killCurrent() (tea.Model, tea.Cmd) {
	if !m.isCursorValid() {
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
		m.setError("Error: %v", err)
	}

	m.mode = ModeNormal
	m.killTarget = ""

	// Reload sessions and clear message after 5 seconds
	return m, tea.Batch(m.loadSessions, clearMessageAfter(5*time.Second))
}

func (m *Model) createSession(name string) (tea.Model, tea.Cmd) {
	homeDir := os.Getenv("HOME")
	if err := tmux.CreateSession(name, homeDir); err != nil {
		m.setError("Error: %v", err)
		m.mode = ModeNormal
		m.input.Blur()
		return m, nil
	}

	// Apply layout if configured
	m.applyLayout(name, homeDir)

	// Switch to the new session
	if err := tmux.SwitchClient(name); err != nil {
		m.setError("Created but failed to switch: %v", err)
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

	// Execute layout script in background
	go func() {
		cmd := exec.Command(scriptPath, sessionName, workingDir)
		cmd.Env = append(os.Environ(),
			"TMUX_SESSION="+sessionName,
			"TMUX_WORKING_DIR="+workingDir,
		)
		_ = cmd.Run()
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
	m.updateScrollOffset()
}

// updateScrollOffset adjusts scroll offset to keep cursor visible in session list
func (m *Model) updateScrollOffset() {
	// If cursor is above visible area, scroll up
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	// If cursor is below visible area, scroll down
	if m.cursor >= m.scrollOffset+m.config.MaxVisibleItems {
		m.scrollOffset = m.cursor - m.config.MaxVisibleItems + 1
	}
	// Ensure scroll offset is not negative
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// contentWidth returns the available width inside the app border/padding
func (m *Model) contentWidth() int {
	if m.width > 0 {
		return m.width - ui.AppBorderOverheadX
	}
	return 56 // Default fallback (60 - 4)
}

// contentHeight returns the available height inside the app border/padding
func (m *Model) contentHeight() int {
	if m.height > 0 {
		return m.height - ui.AppBorderOverheadY
	}
	return 0
}

// borderWidth returns the width to use for internal borders
func (m *Model) borderWidth() int {
	return m.contentWidth()
}

// repoMaxVisibleItems returns the actual number of items that can be shown
// based on window height, matching the View's calculation
func (m *Model) repoMaxVisibleItems() int {
	maxItems := m.config.MaxVisibleItems
	contentH := m.contentHeight()
	if contentH > 0 {
		// Reserve: header(1) + border(1) + footer border(1) + footer(1) = 4 lines
		availableForContent := contentH - 4
		if availableForContent < maxItems && availableForContent > 0 {
			maxItems = availableForContent
		}
	} else {
		// Conservative default when height unknown
		maxItems = 5
	}
	return maxItems
}

// updateRepoScrollOffset adjusts scroll offset to keep cursor visible in repo list
func (m *Model) updateRepoScrollOffset() {
	maxVisible := m.repoMaxVisibleItems()
	// If cursor is above visible area, scroll up
	if m.repoCursor < m.repoScrollOffset {
		m.repoScrollOffset = m.repoCursor
	}
	// If cursor is below visible area, scroll down
	if m.repoCursor >= m.repoScrollOffset+maxVisible {
		m.repoScrollOffset = m.repoCursor - maxVisible + 1
	}
	// Ensure scroll offset is not negative
	if m.repoScrollOffset < 0 {
		m.repoScrollOffset = 0
	}
}

// fuzzyMatch checks if the pattern matches the text (case-insensitive, substring match)
func fuzzyMatch(text, pattern string) bool {
	textLower := strings.ToLower(text)
	return strings.Contains(textLower, pattern)
}

// isCursorValid returns true if cursor points to a valid item
func (m *Model) isCursorValid() bool {
	return m.cursor >= 0 && m.cursor < len(m.items)
}

// getTargetName returns the tmux target name for the given item
func (m *Model) getTargetName(item Item) string {
	if item.IsSession {
		return m.sessions[item.SessionIndex].Name
	}
	session := m.sessions[item.SessionIndex]
	window := session.Windows[item.WindowIndex]
	return fmt.Sprintf("%s:%d", session.Name, window.Index)
}

// setError sets an error message on the model
func (m *Model) setError(format string, args ...any) {
	m.message = fmt.Sprintf(format, args...)
	m.messageIsError = true
}

// View implements tea.Model
func (m Model) View() string {
	if m.mode == ModePickDirectory {
		return m.viewPickDirectory()
	}
	return m.viewSessionList()
}

// viewPickDirectory renders the directory picker view
func (m Model) viewPickDirectory() string {
	var b strings.Builder
	usedLines := 0

	// Header - always show "Select directory", append filter if active
	if m.repoFilter != "" {
		b.WriteString(ui.HeaderStyle.Render("Select directory"))
		b.WriteString("  ")
		b.WriteString(ui.FilterStyle.Render(m.repoFilter))
	} else {
		b.WriteString(ui.HeaderStyle.Render("Select directory"))
	}
	b.WriteString("\n")
	usedLines++

	b.WriteString(ui.RenderBorder(m.borderWidth()))
	b.WriteString("\n")
	usedLines++

	// Use shared helper for consistent visible item calculation
	maxItems := m.repoMaxVisibleItems()

	// Directory list (only visible items)
	endIdx := m.repoScrollOffset + maxItems
	if endIdx > len(m.repoFiltered) {
		endIdx = len(m.repoFiltered)
	}
	visibleCount := endIdx - m.repoScrollOffset

	// Get scrollbar characters for each line
	scrollbar := ui.ScrollbarChars(len(m.repoFiltered), maxItems, m.repoScrollOffset, visibleCount)

	contentLines := 0
	for i := m.repoScrollOffset; i < endIdx; i++ {
		dir := m.repoFiltered[i]
		selected := i == m.repoCursor
		lineIdx := i - m.repoScrollOffset

		// Scrollbar on the left
		if lineIdx < len(scrollbar) {
			b.WriteString(scrollbar[lineIdx])
			b.WriteString(" ")
		}

		if selected {
			b.WriteString(ui.FilterStyle.Render(dir))
		} else {
			b.WriteString(dir)
		}
		b.WriteString("\n")
		contentLines++
	}

	// Empty state
	if len(m.repoFiltered) == 0 {
		if m.repoFilter != "" {
			b.WriteString("  No directories matching filter\n")
		} else {
			b.WriteString("  No directories found\n")
		}
		contentLines++
	}
	usedLines += contentLines

	// Add padding to push footer to bottom
	// Footer = border (1) + help line (1) = 2 lines
	footerLines := 2
	contentH := m.contentHeight()
	if contentH > 0 {
		padding := contentH - usedLines - footerLines
		for i := 0; i < padding; i++ {
			b.WriteString("\n")
		}
	}

	b.WriteString(ui.RenderBorder(m.borderWidth()))
	b.WriteString("\n")
	if m.repoFilter != "" {
		b.WriteString(ui.FooterStyle.Render(ui.HelpFiltering()))
	} else {
		b.WriteString(ui.FooterStyle.Render(ui.HelpPickDirectory()))
	}
	return ui.AppStyle.Render(b.String())
}

// viewSessionList renders the main session list view
func (m Model) viewSessionList() string {
	var b strings.Builder
	usedLines := 0

	// Header with optional filter
	if m.filter != "" {
		b.WriteString(ui.HeaderStyle.Render("tsm"))
		b.WriteString("  ")
		b.WriteString(ui.FilterStyle.Render(m.filter))
	} else {
		b.WriteString(ui.HeaderStyle.Render("tsm"))
	}
	b.WriteString("\n")
	usedLines++

	b.WriteString(ui.RenderBorder(m.borderWidth()))
	b.WriteString("\n")
	usedLines++

	// Scroll indicator (top)
	if m.scrollOffset > 0 {
		b.WriteString(ui.TimeStyle.Render(fmt.Sprintf("  ↑ %d more", m.scrollOffset)))
		b.WriteString("\n")
		usedLines++
	}

	// Session list (only visible items)
	endIdx := m.scrollOffset + m.config.MaxVisibleItems
	if endIdx > len(m.items) {
		endIdx = len(m.items)
	}

	// Calculate session numbers (count sessions before visible area)
	sessionNum := 0
	for i := 0; i < m.scrollOffset && i < len(m.items); i++ {
		if m.items[i].IsSession {
			sessionNum++
		}
	}

	contentLines := 0
	for i := m.scrollOffset; i < endIdx; i++ {
		item := m.items[i]
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
		contentLines++
	}

	// Empty state
	if len(m.items) == 0 {
		if m.filter != "" {
			b.WriteString("  No sessions matching filter\n")
		} else {
			b.WriteString("  No other sessions available\n")
		}
		contentLines++
	}
	usedLines += contentLines

	// Scroll indicator (bottom)
	remaining := len(m.items) - endIdx
	if remaining > 0 {
		b.WriteString(ui.TimeStyle.Render(fmt.Sprintf("  ↓ %d more", remaining)))
		b.WriteString("\n")
		usedLines++
	}

	// Message line (rendered before padding, part of content area)
	messageLines := 0
	var messageContent string
	if m.message != "" {
		if m.messageIsError {
			messageContent = ui.ErrorMessageStyle.Render(m.message)
		} else {
			messageContent = ui.MessageStyle.Render(m.message)
		}
		messageLines = 1
	} else if m.mode == ModeCreate {
		messageContent = ui.InputPromptStyle.Render(" New session: ") + m.input.View()
		messageLines = 1
	}

	// Add padding to push footer to bottom
	// Footer = border (1) + help line (1) = 2 lines
	// Plus any message line
	footerLines := 2 + messageLines
	contentH := m.contentHeight()
	if contentH > 0 {
		padding := contentH - usedLines - footerLines
		for i := 0; i < padding; i++ {
			b.WriteString("\n")
		}
	}

	// Render message if present (now part of fixed footer area)
	if messageContent != "" {
		b.WriteString(messageContent)
		b.WriteString("\n")
	}

	b.WriteString(ui.RenderBorder(m.borderWidth()))
	b.WriteString("\n")

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
