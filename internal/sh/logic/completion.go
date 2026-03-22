package logic

import (
	"os"
	"path/filepath"
	"strings"
)

// CompletionEngine handles finding potential matches for tab completion.
type CompletionEngine struct{}

// NewCompletionEngine creates a new CompletionEngine instance.
func NewCompletionEngine() *CompletionEngine {
	return &CompletionEngine{}
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

	// If it's the first word, search for commands too
	isFirstWord := true
	for i := 0; i < start; i++ {
		if !isSeparator(input[i]) {
			isFirstWord = false
			break
		}
	}

	if isFirstWord && !strings.Contains(word, string(os.PathSeparator)) {
		matches = append(matches, e.findCommands(word)...)
	}

	// Always search for files/directories
	matches = append(matches, e.findFiles(word, cwd)...)

	// Deduplicate matches
	uniqueMatches := make([]string, 0)
	seen := make(map[string]bool)
	for _, m := range matches {
		if !seen[m] {
			uniqueMatches = append(uniqueMatches, m)
			seen[m] = true
		}
	}

	return uniqueMatches, word
}

func isSeparator(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == ';' || c == '&' || c == '|' || c == '>' || c == '<'
}

func (e *CompletionEngine) findCommands(prefix string) []string {
	var matches []string
	path := os.Getenv("PATH")
	dirs := filepath.SplitList(path)

	for _, dir := range dirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			name := file.Name()
			if strings.HasPrefix(name, prefix) {
				// Check if executable
				if info, err := file.Info(); err == nil && info.Mode().IsRegular() && info.Mode()&0111 != 0 {
					matches = append(matches, name)
				}
			}
		}
	}
	return matches
}

func (e *CompletionEngine) findFiles(prefix string, cwd string) []string {
	var matches []string

	dir := cwd
	filePrefix := prefix

	// If prefix contains a path separator, split it
	if strings.Contains(prefix, string(os.PathSeparator)) {
		dir = filepath.Join(cwd, filepath.Dir(prefix))
		filePrefix = filepath.Base(prefix)
		// If prefix ends with separator, base is "." and we want everything in that dir
		if strings.HasSuffix(prefix, string(os.PathSeparator)) {
			dir = filepath.Join(cwd, prefix)
			filePrefix = ""
		}
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, file := range files {
		name := file.Name()
		if strings.HasPrefix(name, filePrefix) {
			fullPath := name
			if strings.Contains(prefix, string(os.PathSeparator)) {
				fullPath = filepath.Join(filepath.Dir(prefix), name)
			}

			if file.IsDir() {
				fullPath += string(os.PathSeparator)
			}
			matches = append(matches, fullPath)
		}
	}
	return matches
}
