#!/usr/bin/env bash
# Claude Code hook - writes status to ~/.cache/tmux-session-picker/
# Used by tmux-session-picker to display Claude status per session

STATUS_DIR="$HOME/.cache/tmux-session-picker"
mkdir -p "$STATUS_DIR"

# Read JSON from stdin (required by Claude Code hooks)
cat > /dev/null

# Get tmux session name
TMUX_SESSION=$(tmux display-message -p '#{session_name}' 2>/dev/null)
[[ -z "$TMUX_SESSION" ]] && exit 0

HOOK_TYPE="$1"
STATUS_FILE="$STATUS_DIR/${TMUX_SESSION}.status"
TIMESTAMP=$(date +%s)

case "$HOOK_TYPE" in
    "SessionStart")
        echo "new:$TIMESTAMP" > "$STATUS_FILE"
        ;;
    "PreToolUse")
        echo "working:$TIMESTAMP" > "$STATUS_FILE"
        ;;
    "Stop"|"SubagentStop"|"Notification")
        echo "waiting:$TIMESTAMP" > "$STATUS_FILE"
        # Play notification sound (macOS)
        if command -v afplay &>/dev/null; then
            afplay /System/Library/Sounds/Pop.aiff 2>/dev/null &
        fi
        ;;
esac

exit 0
