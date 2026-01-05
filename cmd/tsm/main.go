package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nikbrunner/tsm/internal/config"
	"github.com/nikbrunner/tsm/internal/model"
	"github.com/nikbrunner/tsm/internal/tmux"
)

func main() {
	// Handle subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			if err := config.Init(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created config file at %s\n", config.Path())
			return
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			fmt.Println("Usage: tsm [init]")
			os.Exit(1)
		}
	}

	// Check if running inside tmux
	if os.Getenv("TMUX") == "" {
		fmt.Println("Error: tsm must be run from within tmux")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Get current session to exclude from list
	currentSession, err := tmux.CurrentSession()
	if err != nil {
		fmt.Printf("Error getting current session: %v\n", err)
		os.Exit(1)
	}

	// Initialize and run the TUI
	m := model.New(currentSession, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
