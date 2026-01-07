package ui

import (
	"strings"
	"testing"
)

func TestFormatClaudeStatus(t *testing.T) {
	tests := []struct {
		name      string
		state     string
		wantEmpty bool
		contains  string
	}{
		{
			name:      "empty state returns empty",
			state:     "",
			wantEmpty: true,
		},
		{
			name:      "new state returns empty (no visual noise)",
			state:     "new",
			wantEmpty: true,
		},
		{
			name:      "working state",
			state:     "working",
			wantEmpty: false,
			contains:  "⏳ working...",
		},
		{
			name:      "waiting state",
			state:     "waiting",
			wantEmpty: false,
			contains:  "💬 waiting...",
		},
		{
			name:      "unknown state returns empty",
			state:     "unknown",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatClaudeStatus(tt.state)

			if tt.wantEmpty && result != "" {
				t.Errorf("FormatClaudeStatus(%q) = %q, want empty", tt.state, result)
			}

			if !tt.wantEmpty {
				if result == "" {
					t.Errorf("FormatClaudeStatus(%q) returned empty, want non-empty", tt.state)
				}
				if tt.contains != "" && !strings.Contains(result, tt.contains) {
					t.Errorf("FormatClaudeStatus(%q) = %q, should contain %q", tt.state, result, tt.contains)
				}
				if !strings.Contains(result, "Claude Code:") {
					t.Errorf("FormatClaudeStatus(%q) = %q, should contain 'Claude Code:'", tt.state, result)
				}
			}
		})
	}
}

func TestScrollbarChars(t *testing.T) {
	tests := []struct {
		name         string
		totalItems   int
		visibleItems int
		scrollOffset int
		height       int
		wantLen      int
		allSpaces    bool
	}{
		{
			name:         "all items fit - no scrollbar",
			totalItems:   5,
			visibleItems: 10,
			scrollOffset: 0,
			height:       5,
			wantLen:      5,
			allSpaces:    true,
		},
		{
			name:         "exactly fits - no scrollbar",
			totalItems:   5,
			visibleItems: 5,
			scrollOffset: 0,
			height:       5,
			wantLen:      5,
			allSpaces:    true,
		},
		{
			name:         "needs scrollbar",
			totalItems:   20,
			visibleItems: 5,
			scrollOffset: 0,
			height:       5,
			wantLen:      5,
			allSpaces:    false,
		},
		{
			name:         "scrolled down",
			totalItems:   20,
			visibleItems: 5,
			scrollOffset: 10,
			height:       5,
			wantLen:      5,
			allSpaces:    false,
		},
		{
			name:         "zero height",
			totalItems:   20,
			visibleItems: 5,
			scrollOffset: 0,
			height:       0,
			wantLen:      0,
			allSpaces:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ScrollbarChars(tt.totalItems, tt.visibleItems, tt.scrollOffset, tt.height)

			if len(result) != tt.wantLen {
				t.Errorf("len(ScrollbarChars) = %d, want %d", len(result), tt.wantLen)
			}

			if tt.allSpaces {
				for i, ch := range result {
					if ch != " " {
						t.Errorf("result[%d] = %q, want space (no scrollbar needed)", i, ch)
					}
				}
			} else {
				// Should have some non-space characters (the thumb)
				hasThumb := false
				for _, ch := range result {
					if ch != " " {
						hasThumb = true
						break
					}
				}
				if !hasThumb {
					t.Error("Scrollbar should have visible thumb characters")
				}
			}
		})
	}
}

func TestRenderBorder(t *testing.T) {
	tests := []struct {
		width   int
		wantLen int
	}{
		{width: 10, wantLen: 10},
		{width: 0, wantLen: 0},
		{width: 50, wantLen: 50},
	}

	for _, tt := range tests {
		result := RenderBorder(tt.width)
		// The result will have ANSI codes, but should contain the border chars
		if !strings.Contains(result, "─") && tt.width > 0 {
			t.Errorf("RenderBorder(%d) should contain border character", tt.width)
		}
	}
}
