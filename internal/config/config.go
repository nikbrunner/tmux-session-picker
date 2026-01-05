package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds all configuration options for tsm
type Config struct {
	// Layout script name to apply when creating new sessions
	Layout string `toml:"layout"`

	// Directory containing layout scripts
	LayoutDir string `toml:"layout_dir"`

	// Enable Claude Code status integration
	ClaudeStatusEnabled bool `toml:"claude_status_enabled"`

	// Directory for status cache files
	CacheDir string `toml:"cache_dir"`
}

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() Config {
	home := os.Getenv("HOME")
	return Config{
		Layout:              "",
		LayoutDir:           filepath.Join(home, ".config", "tmux", "layouts"),
		ClaudeStatusEnabled: false,
		CacheDir:            filepath.Join(home, ".cache", "tsm"),
	}
}

// Path returns the path to the config file
func Path() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".config", "tsm", "config.toml")
}

// Load reads configuration from file and environment variables.
// Priority: env vars > config file > defaults
func Load() (Config, error) {
	cfg := DefaultConfig()

	// Load from config file if it exists
	configPath := Path()
	if _, err := os.Stat(configPath); err == nil {
		if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Expand ~ in paths
	cfg.LayoutDir = expandPath(cfg.LayoutDir)
	cfg.CacheDir = expandPath(cfg.CacheDir)

	// Environment variables override config file
	if val := os.Getenv("TMUX_LAYOUT"); val != "" {
		cfg.Layout = val
	}
	if val := os.Getenv("TMUX_LAYOUTS_DIR"); val != "" {
		cfg.LayoutDir = expandPath(val)
	}
	if os.Getenv("TMUX_SESSION_PICKER_CLAUDE_STATUS") == "1" {
		cfg.ClaudeStatusEnabled = true
	}

	return cfg, nil
}

// Init creates a new config file with commented defaults
func Init() error {
	configPath := Path()
	configDir := filepath.Dir(configPath)

	// Create directory if needed
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	// Write default config with comments
	content := `# tsm configuration
# Environment variables override these settings

# Layout script name to apply when creating new sessions
# layout = "ide"

# Directory containing layout scripts
# layout_dir = "~/.config/tmux/layouts"

# Enable Claude Code status integration
# claude_status_enabled = false

# Directory for status cache files
# cache_dir = "~/.cache/tsm"
`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home := os.Getenv("HOME")
		return filepath.Join(home, path[1:])
	}
	return path
}
