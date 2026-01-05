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

// GetStatus reads the Claude Code status for a session from the given cache directory.
// Returns empty Status if no status file exists.
func GetStatus(sessionName string, cacheDir string) Status {
	statusFile := filepath.Join(cacheDir, sessionName+".status")
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
func CleanupStale(cacheDir string, activeSessions []string) {
	entries, err := os.ReadDir(cacheDir)
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
			os.Remove(filepath.Join(cacheDir, entry.Name()))
		}
	}
}
