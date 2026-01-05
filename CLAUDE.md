# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

tsm (**t**mux **s**ession **m**anager) is a Go TUI application for managing tmux sessions. It provides fuzzy filtering, session/window navigation, and Claude Code status integration. Built with Bubbletea/Lipgloss.

## Build Commands

```bash
make build          # Build binary to ./tsm
make install        # Build and install to ~/.local/bin/tsm
make test           # Run tests
make fmt            # Format code
make lint           # Run golangci-lint
make tidy           # go mod tidy
```

For quick iteration: `go build -o tsm ./cmd/tsm/ && cp tsm ~/.local/bin/tsm`

## Architecture

```
cmd/tsm/main.go          # Entry point, handles `tsm init` subcommand
internal/
  model/model.go         # Bubbletea Model - main state and Update/View logic
  ui/
    keys.go              # Key bindings (KeyMap) and help text functions
    styles.go            # Lipgloss colors and styles
  config/config.go       # TOML config loading (~/.config/tsm/config.toml)
  tmux/tmux.go           # tmux command wrappers (list sessions, switch, kill)
  claude/status.go       # Claude Code status file parsing
hooks/tsm-hook.sh        # Claude Code hook for status updates
```

### Bubbletea Model Flow

The model (`internal/model/model.go`) has three modes:
- **ModeNormal**: Session list with fuzzy filtering (typing filters, Ctrl+keys navigate)
- **ModeConfirmKill**: Kill confirmation prompt
- **ModeCreate**: Text input for new session name

Key state:
- `sessions []tmux.Session` - Raw session data
- `items []Item` - Flattened view (sessions + expanded windows)
- `filter string` - Current filter text
- `cursor int` - Selected item index

### Key Bindings

Navigation uses Ctrl modifiers to reserve letters for filtering:
- `Ctrl+j/k` or arrows: Navigate
- `Ctrl+h/l` or arrows: Collapse/Expand sessions
- `Ctrl+n`: Create new session
- `Ctrl+x`: Kill (requires `Ctrl+y` to confirm)
- `1-9`: Jump to session (only when no filter active)
- Type letters: Fuzzy filter sessions

## Configuration

Config file: `~/.config/tsm/config.toml`

```toml
layout = "ide"                    # Layout script for new sessions
layout_dir = "~/.config/tmux/layouts"
claude_status_enabled = true      # Show [CC: working/waiting] status
cache_dir = "~/.cache/tsm"
```

Environment variables override config: `TMUX_LAYOUT`, `TMUX_LAYOUTS_DIR`, `TMUX_SESSION_PICKER_CLAUDE_STATUS=1`

## Testing

Must test inside tmux:
```bash
tmux display-popup -w50% -h35% -B -E "./tsm"
```

## Claude Status Integration

The hook (`hooks/tsm-hook.sh`) writes status files to `~/.cache/tsm/<session>.status`. The TUI reads these to show `[CC: new|working|waiting]` badges per session.
