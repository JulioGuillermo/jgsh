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

	// Remove carriage returns and other control chars that mess up TUI rendering
	result = strings.ReplaceAll(result, "\r", "")

	// Remove other common ESC sequences like ESC = (application keypad)
	result = regexp.MustCompile(`\x1b[=>]`).ReplaceAllString(result, "")

	return result
}

// DetectPrompt checks if the given buffer ends with a shell prompt.
func DetectPrompt(buffer []byte) bool {
	// 1. Check for our special GLOCK prompt first
	sRaw := string(buffer)
	if strings.Contains(sRaw, "GLOCK> ") {
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
	endings := []string{"$", "#", "%", "", ">", "➜", "❯"}
	for _, e := range endings {
		if strings.HasSuffix(s, e) {
			return true
		}
	}

	return false
}

// StripPrompt removes trailing prompt lines from the output.
func StripPrompt(output string) string {
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
