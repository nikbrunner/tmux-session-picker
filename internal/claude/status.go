package claude

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Status represents Claude Code status for a session
type Status struct {
	State     string    // "new", "working", "waiting", or ""
	Timestamp time.Time // When the status was last updated
}

// StatusDir is the directory where Claude status files are stored
var StatusDir = filepath.Join(os.Getenv("HOME"), ".cache", "tmux-session-picker")

// GetStatus reads the Claude Code status for a session
// Returns empty Status if no status file exists or feature is disabled
func GetStatus(sessionName string) Status {
	// Check if feature is enabled
	if os.Getenv("TMUX_SESSION_PICKER_CLAUDE_STATUS") != "1" {
		return Status{}
	}

	statusFile := filepath.Join(StatusDir, sessionName+".status")
	content, err := os.ReadFile(statusFile)
	if err != nil {
		return Status{}
	}

	// Parse format: "state:timestamp"
	parts := strings.SplitN(strings.TrimSpace(string(content)), ":", 2)
	if len(parts) != 2 {
		return Status{}
	}

	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return Status{}
	}

	return Status{
		State:     parts[0],
		Timestamp: time.Unix(timestamp, 0),
	}
}

// CleanupStale removes status files for sessions that no longer exist
func CleanupStale(activeSessions []string) {
	if os.Getenv("TMUX_SESSION_PICKER_CLAUDE_STATUS") != "1" {
		return
	}

	entries, err := os.ReadDir(StatusDir)
	if err != nil {
		return
	}

	activeSet := make(map[string]bool)
	for _, s := range activeSessions {
		activeSet[s] = true
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".status") {
			continue
		}

		sessionName := strings.TrimSuffix(entry.Name(), ".status")
		if !activeSet[sessionName] {
			os.Remove(filepath.Join(StatusDir, entry.Name()))
		}
	}
}
