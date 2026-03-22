package adapters

import (
	"testing"
)

func TestNewTSBashHighlighter(t *testing.T) {
	h, err := NewTSBashHighlighter()
	if err != nil {
		t.Fatalf("NewTSBashHighlighter() returned error: %v", err)
	}
	if h == nil {
		t.Fatal("NewTSBashHighlighter() returned nil")
	}
}

func TestHighlight(t *testing.T) {
	h, err := NewTSBashHighlighter()
	if err != nil {
		t.Fatalf("NewTSBashHighlighter() returned error: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		contains string // Substring that should be in the output
	}{
		{
			name:     "empty string",
			input:    "",
			contains: "",
		},
		{
			name:     "simple command",
			input:    "ls",
			contains: "ls",
		},
		{
			name:     "command with argument",
			input:    "ls -la",
			contains: "ls",
		},
		{
			name:     "command with path",
			input:    "cd /home",
			contains: "cd",
		},
		{
			name:     "piped command",
			input:    "cat file | grep foo",
			contains: "cat",
		},
		{
			name:     "variable",
			input:    "echo $HOME",
			contains: "$HOME",
		},
		{
			name:     "quoted string",
			input:    "echo \"hello world\"",
			contains: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.Highlight(tt.input)
			if tt.contains != "" && !containsSubstring(result, tt.contains) {
				t.Errorf("Highlight(%q) = %q, should contain %q", tt.input, result, tt.contains)
			}
			// Basic sanity: output should not be empty if input is not empty
			if tt.input != "" && result == "" {
				t.Errorf("Highlight(%q) returned empty string, expected non-empty", tt.input)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
