package logic

import (
	"regexp"
	"strings"
)

// StripAnsi removes ANSI escape sequences and problematic control characters from a string.
func StripAnsi(str string) string {
	// CSI sequences: \x1b[...<letter>
	ansiCSI := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	// OSC sequences: \x1b]...<ST> (\x1b\\) or \x1b]...<BEL> (\x07)
	ansiOSC := regexp.MustCompile(`\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)`)

	result := ansiCSI.ReplaceAllString(str, "")
	result = ansiOSC.ReplaceAllString(result, "")

	// Remove other common ESC sequences like ESC = (application keypad)
	result = regexp.MustCompile(`\x1b[=>]`).ReplaceAllString(result, "")

	return result
}

// IsPasswordPrompt checks if the given output contains a password prompt.
func IsPasswordPrompt(output string) bool {
	lower := strings.ToLower(StripAnsi(output))
	patterns := []string{
		"password:",
		"password for ",
		"passphrase:",
		"[sudo] password",
	}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// FoldCarriageReturns processes a string and simulates terminal behavior for \r.
func FoldCarriageReturns(input string) string {
	if !strings.Contains(input, "\r") {
		return input
	}

	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		if !strings.Contains(line, "\r") {
			result = append(result, line)
			continue
		}

		// Handle \r by keeping only the text after the last \r
		// but only if it's not at the very end.
		parts := strings.Split(line, "\r")
		// The actual terminal behavior is more complex (overwriting character by character),
		// but for progress bars, the last part is almost always what we want.
		lastPart := ""
		for i := len(parts) - 1; i >= 0; i-- {
			if strings.TrimSpace(parts[i]) != "" || i == 0 {
				lastPart = parts[i]
				break
			}
		}
		result = append(result, lastPart)
	}

	return strings.Join(result, "\n")
}

// DetectPrompt checks if the given buffer ends with a shell prompt.
func DetectPrompt(buffer []byte) bool {
	// 1. Check for our special JGSH prompt first - this is the MOST reliable way
	sRaw := string(buffer)
	if strings.Contains(sRaw, "JGSH> ") {
		return true
	}

	// 2. Semantic marks (FTCS_Prompt): \x1b]133;A\x1b\\
	if strings.Contains(sRaw, "\x1b]133;A\x1b\\") {
		return true
	}

	// 3. Strip ANSI and check for prompt symbols
	s := strings.TrimSpace(StripAnsi(sRaw))
	if s == "" {
		return false
	}

	// Common prompt endings (multibyte safe)
	// We REMOVE '%' from this list because it's used in progress bars (pacman, wget, etc.)
	endings := []string{"$", "#", "", ">", "➜", "❯"}
	for _, e := range endings {
		if strings.HasSuffix(s, e) {
			// Ensure it's not just a character in the middle of a string
			// We only want to detect prompts that are at the beginning of a line or after a space
			idx := strings.LastIndex(s, e)
			if idx == 0 || (idx > 0 && (s[idx-1] == ' ' || s[idx-1] == '\n')) {
				return true
			}
		}
	}

	return false
}

// StripPrompt removes trailing prompt lines from the output.
func StripPrompt(output string) string {
	// First, specifically remove our custom prompt if it exists anywhere in the trailing lines
	output = strings.ReplaceAll(output, "JGSH> ", "")

	lines := strings.Split(strings.TrimRight(output, "\n "), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if DetectPrompt([]byte(lines[i])) {
			lines = lines[:i]
		} else {
			break
		}
	}
	return strings.Join(lines, "\n")
}

// StripEcho removes everything from the output up to and including the echoed command.
func StripEcho(output, command string) string {
	if command == "" {
		return output
	}

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		stripped := StripAnsi(line)
		// If any of the first few lines contains our command,
		// assume everything up to and including this line is prompt/echo.
		if strings.Contains(stripped, command) {
			if i+1 < len(lines) {
				return strings.Join(lines[i+1:], "\n")
			}
			return ""
		}
	}
	return output
}
