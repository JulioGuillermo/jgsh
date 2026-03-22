package logic

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CompletionEngine handles finding potential matches for tab completion.
type CompletionEngine struct{}

// NewCompletionEngine creates a new CompletionEngine instance.
func NewCompletionEngine() *CompletionEngine {
	return &CompletionEngine{}
}

// CheckDependencies verifies if the system has the required completion packages.
func (e *CompletionEngine) CheckDependencies() (bool, string) {
	paths := []string{
		"/usr/share/bash-completion/bash_completion", // Linux (Debian, Arch, Fedora)
		"/usr/local/etc/bash_completion",             // macOS (Homebrew)
		"/etc/bash_completion",                       // Legacy/Other
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return true, ""
		}
	}

	// If not found, determine the install command based on the OS
	warning := "💡 TIP: Install 'bash-completion' for advanced autocomplete (sudo, git, pacman, etc.)"
	return false, warning
}

// Complete returns a list of possible completions for the given input and cursor position.
func (e *CompletionEngine) Complete(input string, pos int, cwd string) ([]string, string) {
	if pos < 0 || pos > len(input) {
		pos = len(input)
	}

	// Find the word being completed
	start := pos
	for start > 0 && !isSeparator(input[start-1]) {
		start--
	}
	word := input[start:pos]

	var matches []string

	// 1. Try advanced programmable completion first (git, pacman, sudo, etc.)
	matches = append(matches, e.compgenProgrammable(input, pos, cwd)...)

	// 2. Determine context for smart fallback
	trimmedInput := strings.TrimSpace(input[:start])
	parts := strings.Fields(trimmedInput)
	firstWord := ""
	if len(parts) > 0 {
		firstWord = parts[0]
	}

	isFirstWord := firstWord == ""
	isCommandWrapper := firstWord == "sudo" || firstWord == "doas" || firstWord == "time" || firstWord == "xargs" || firstWord == "exec"

	// 3. Fallback/Merge with basic sources
	if isFirstWord || isCommandWrapper {
		// Suggest commands, aliases, and builtins if we are at a command position
		matches = append(matches, e.compgen("-abck", word, cwd)...)
	}

	if firstWord == "cd" {
		// Only directories for cd
		matches = append(matches, e.compgen("-d", word, cwd)...)
	} else {
		// Generic file and directory completion
		matches = append(matches, e.compgen("-f", word, cwd)...)
	}

	// Deduplicate matches
	uniqueMatches := make([]string, 0)
	seen := make(map[string]bool)
	for _, m := range matches {
		if !seen[m] {
			// Add trailing slash for directories
			if info, err := os.Stat(filepath.Join(cwd, m)); err == nil && info.IsDir() && !strings.HasSuffix(m, string(os.PathSeparator)) {
				m += string(os.PathSeparator)
			}
			uniqueMatches = append(uniqueMatches, m)
			seen[m] = true
		}
	}

	return uniqueMatches, word
}

// compgenProgrammable attempts to use the full bash-completion system.
func (e *CompletionEngine) compgenProgrammable(line string, pos int, cwd string) []string {
	// This script simulates a bash completion trigger with full-context awareness
	script := `
		# Source main completion script if available
		if [ -f /usr/share/bash-completion/bash_completion ]; then
			. /usr/share/bash-completion/bash_completion
		elif [ -f /etc/bash_completion ]; then
			. /etc/bash_completion
		fi

		COMP_LINE="$1"
		COMP_POINT="$2"
		
		# Robust word splitting
		COMP_WORDS=()
		read -a COMP_WORDS <<< "$COMP_LINE"
		COMP_CWORD=$((${#COMP_WORDS[@]} - 1))
		
		# If the line ends with a space, add an empty word for the next completion
		if [[ "$COMP_LINE" == *" " ]]; then
			COMP_WORDS+=("")
			COMP_CWORD=$((${#COMP_WORDS[@]} - 1))
		fi

		cmd="${COMP_WORDS[0]}"
		if [ -z "$cmd" ]; then exit 0; fi
		
		# Load the specific completion function if not already loaded
		if ! complete -p "$cmd" &>/dev/null; then
			_completion_loader "$cmd" &>/dev/null
		fi

		# Extract the completion function (-F) or action (-A)
		comp_spec=$(complete -p "$cmd" 2>/dev/null)
		
		if [[ $comp_spec == *"-F "* ]]; then
			func=$(echo "$comp_spec" | sed 's/.*-F \([^ ]*\).*/\1/')
			if declare -f "$func" >/dev/null; then
				cur="${COMP_WORDS[$COMP_CWORD]}"
				prev="${COMP_WORDS[$COMP_CWORD-1]}"
				COMP_REPLY=()
				$func "$cmd" "$cur" "$prev"
				printf "%s\n" "${COMP_REPLY[@]}"
			fi
		elif [[ $comp_spec == *"-A "* ]]; then
			action=$(echo "$comp_spec" | sed 's/.*-A \([^ ]*\).*/\1/')
			cur="${COMP_WORDS[$COMP_CWORD]}"
			compgen -A "$action" -- "$cur"
		fi
	`

	cmd := exec.Command("bash", "-c", script, "--", line, fmt.Sprintf("%d", pos))
	cmd.Dir = cwd
	out, _ := cmd.Output()

	lines := strings.Split(string(out), "\n")
	var results []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			results = append(results, l)
		}
	}
	return results
}

// compgen executes the bash builtin compgen to get completions.
func (e *CompletionEngine) compgen(action, prefix, cwd string) []string {
	cmd := exec.Command("bash", "-c", "compgen "+action+" -- \"$1\"", "--", prefix)
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(out), "\n")
	var results []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			results = append(results, line)
		}
	}
	return results
}

func isSeparator(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == ';' || c == '&' || c == '|' || c == '>' || c == '<'
}
