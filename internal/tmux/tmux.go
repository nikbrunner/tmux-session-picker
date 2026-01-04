package tmux

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Session represents a tmux session
type Session struct {
	Name         string
	LastActivity time.Time
	Windows      []Window
	Expanded     bool
}

// Window represents a tmux window
type Window struct {
	Index int
	Name  string
}

// CurrentSession returns the name of the current tmux session
func CurrentSession() (string, error) {
	out, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ListSessions returns all tmux sessions sorted by activity (most recent first)
// Excludes the current session and popup sessions
func ListSessions(excludeCurrent string) ([]Session, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_activity} #{session_name}").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return []Session{}, nil
	}

	var sessions []Session
	now := time.Now()

	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		name := parts[1]

		// Skip current session and popup sessions
		if name == excludeCurrent || strings.HasPrefix(name, "_popup_") {
			continue
		}

		activityUnix, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}

		sessions = append(sessions, Session{
			Name:         name,
			LastActivity: time.Unix(activityUnix, 0),
		})

		// Keep now for potential future use
		_ = now
	}

	// Sort by activity (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActivity.After(sessions[j].LastActivity)
	})

	return sessions, nil
}

// ListWindows returns all windows for a given session
func ListWindows(sessionName string) ([]Window, error) {
	out, err := exec.Command("tmux", "list-windows", "-t", sessionName, "-F", "#{window_index}:#{window_name}").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return []Window{}, nil
	}

	var windows []Window
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		index, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		windows = append(windows, Window{
			Index: index,
			Name:  parts[1],
		})
	}

	return windows, nil
}

// KillSession kills a tmux session by name
func KillSession(name string) error {
	return exec.Command("tmux", "kill-session", "-t", name).Run()
}

// KillWindow kills a tmux window
func KillWindow(sessionName string, windowIndex int) error {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	return exec.Command("tmux", "kill-window", "-t", target).Run()
}

// CreateSession creates a new tmux session
func CreateSession(name, dir string) error {
	return exec.Command("tmux", "new-session", "-d", "-s", name, "-c", dir).Run()
}

// SwitchClient switches the tmux client to a session or window
func SwitchClient(target string) error {
	return exec.Command("tmux", "switch-client", "-t", target).Run()
}

// SelectWindow selects a specific window in the current client
func SelectWindow(sessionName string, windowIndex int) error {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	return exec.Command("tmux", "switch-client", "-t", target).Run()
}
