package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nikbrunner/tmux-session-picker/internal/model"
	"github.com/nikbrunner/tmux-session-picker/internal/tmux"
)

func main() {
	// Check if running inside tmux
	if os.Getenv("TMUX") == "" {
		fmt.Println("Error: tsp must be run from within tmux")
		os.Exit(1)
	}

	// Get current session to exclude from list
	currentSession, err := tmux.CurrentSession()
	if err != nil {
		fmt.Printf("Error getting current session: %v\n", err)
		os.Exit(1)
	}

	// Initialize and run the TUI
	m := model.New(currentSession)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
