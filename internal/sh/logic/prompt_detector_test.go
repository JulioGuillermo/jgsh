package logic

import (
	"testing"
)

func TestStripAnsi(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no ansi codes",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "simple color code",
			input:    "\x1b[31mred\x1b[0m",
			expected: "red",
		},
		{
			name:     "multiple codes",
			input:    "\x1b[1;31mbold red\x1b[0m",
			expected: "bold red",
		},
		{
			name:     "complex prompt",
			input:    "\x1b]133;A\x1b\\",
			expected: "",
		},
		{
			name:     "prompt with text",
			input:    "\x1b[0;32m\x1b]133;A\x1b\\\x1b[0m jg ",
			expected: " jg ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripAnsi(tt.input)
			if result != tt.expected {
				t.Errorf("StripAnsi(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetectPrompt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "ftcs prompt mark",
			input:    "\x1b]133;A\x1b\\",
			expected: true,
		},
		{
			name:     "simple dollar prompt",
			input:    "ls -la\n$ ",
			expected: true,
		},
		{
			name:     "simple hash prompt",
			input:    "whoami\n# ",
			expected: true,
		},
		{
			name:     "powerline prompt stripped",
			input:    "\x1b[0;32m\x1b]133;A\x1b\\\x1b[0m jg ",
			expected: true,
		},
		{
			name:     "prompt with ansi and powerline symbol",
			input:    "\x1b[0m\x1b[38;5;75m\x1b[48;5;234m jg  \x1b[0m",
			expected: true,
		},
		{
			name:     "command output only",
			input:    "file1.txt\nfile2.txt",
			expected: false,
		},
		{
			name:     "single word",
			input:    "hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectPrompt([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("DetectPrompt(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
