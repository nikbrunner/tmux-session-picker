# Bubbletea TUI Design for tmux-session-picker

**Date:** 2025-01-04
**Status:** Approved

## Overview

Rewrite the bash/fzf-based tmux session picker as a proper TUI using Go and Bubbletea. Goals: better UX, more features, learning opportunity.

## Layout

```
┌─────────────────────────────────────────────────┐
│  tmux sessions                                  │
├─────────────────────────────────────────────────┤
│  1  ▼ dotfiles              2m ago   [CC: wait] │
│        1: nvim                                  │
│        2: terminal                              │
│        3: lazygit                               │
│  2  ▶ black-atom-core       15m ago             │
│  3  ▶ nbr-haus              1h ago   [CC: work] │
│  4  ▶ koyo                  3h ago              │
│  5  ▶ bm                    1d ago              │
│                                                 │
├─────────────────────────────────────────────────┤
│  {message line - confirmations, feedback}       │
│  1-9 jump  j/k nav  h/l expand  x kill  c new  │
└─────────────────────────────────────────────────┘
```

## Keybindings

| Key | Action |
|-----|--------|
| `j/k` | Navigate up/down |
| `h` | Collapse expanded session |
| `l` | Expand selected session (shows windows) |
| `1-9` | Jump to session by position, or switch to window if inside expanded session |
| `enter` | Switch to selected session/window |
| `x` | Kill selected (asks confirmation) |
| `xx` | Kill selected instantly (double tap) |
| `c` | Create new session (opens input) |
| `esc/q` | Exit picker |

## Interaction Details

### Hierarchy Behavior
- Sessions show as collapsed (`▶`) by default
- Expanding a session (`l`) reveals its windows indented below
- Only one session can be expanded at a time (auto-collapse others)
- Numbers `1-9` context-sensitive: jump to session normally, switch to window when expanded

### Kill Confirmation
- Single `x`: shows confirmation in message line
- Double `xx`: instant kill, no confirmation
- Confirmation UI: `Kill "dotfiles"?  x confirm  esc cancel`

### Create Session
- Press `c` to open input in message line
- Input: `New session: my-project_`
- Keys: `enter` create, `esc` cancel
- After creation: apply layout script, switch to new session

### Footer Structure
- Line 1: Message line (confirmations, feedback, errors)
- Line 2: Keyboard hints (always visible)

## Visual States

### Session Row
```
│  1  ▶ dotfiles              2m ago   [CC: wait] │  # normal
│  1  ▶ dotfiles              2m ago   [CC: wait] │  # selected (highlighted bg)
│  1  ▼ dotfiles              2m ago   [CC: wait] │  # expanded
```

### Window Row (indented)
```
│        1: nvim                                  │  # normal
│        2: terminal                              │  # selected
```

### Claude Status Colors
- `new` - dim/gray
- `working` - yellow/orange
- `waiting` - green (needs attention)

## Data Model

```go
type App struct {
    Sessions     []Session
    Cursor       int      // index in flattened visible list
    Mode         Mode     // normal | confirm-kill | create
    Input        string   // for create mode
    Message      string   // transient feedback
    MessageTimer int      // countdown to clear message
}

type Session struct {
    Name         string
    LastActivity time.Time
    ClaudeStatus string   // "new" | "working" | "waiting" | ""
    Expanded     bool
    Windows      []Window
}

type Window struct {
    Index int
    Name  string
}

type Mode int
const (
    ModeNormal Mode = iota
    ModeConfirmKill
    ModeCreate
)
```

## Data Sources

- `tmux list-sessions -F '#{session_activity} #{session_name}'` - sessions with timestamps
- `tmux list-windows -t <session> -F '#{window_index}: #{window_name}'` - windows
- `~/.cache/tmux-session-picker/<session>.status` - Claude Code status (optional)

## Error Handling

### Startup
- Not in tmux → exit with message "Must run inside tmux"
- No other sessions → show empty state with hint to press `c`

### Runtime
- Session killed externally → refresh list, show message
- Window killed while expanded → collapse session, refresh
- Create fails (duplicate name) → show error in message line

### Edge Cases
- Session names with spaces → proper quoting
- Long session names → truncate with `…`
- More than 9 sessions → `1-9` for first 9, `j/k` for rest
- Expanded session gets killed → auto-collapse, remove

## File Structure

```
tmux-session-picker/
├── cmd/
│   └── tsp/
│       └── main.go          # entry point
├── internal/
│   ├── model/
│   │   └── model.go         # App state, Update, View
│   ├── tmux/
│   │   └── tmux.go          # tmux command wrappers
│   ├── claude/
│   │   └── status.go        # Claude status file reader
│   └── ui/
│       ├── styles.go        # lipgloss styles
│       └── keys.go          # key bindings
├── go.mod
├── go.sum
├── Makefile                  # build, install targets
└── README.md
```

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - styling
- `github.com/charmbracelet/bubbles` - text input component

## Integration

- Single binary `tsp` installed to `~/.local/bin/`
- Existing tmux keybinding updated to call binary
- Layout scripts work unchanged (`$TMUX_LAYOUTS_DIR/$TMUX_LAYOUT.sh`)
