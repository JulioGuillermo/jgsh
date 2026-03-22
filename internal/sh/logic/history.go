package logic

import (
	"bufio"
	"os"
	"path/filepath"
)

// HistoryManager handles persistent command history.
type HistoryManager struct {
	filePath string
}

// NewHistoryManager creates a new HistoryManager instance.
func NewHistoryManager() (*HistoryManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".jgsh_history")
	return &HistoryManager{filePath: path}, nil
}

// Load reads the history from the file.
func (h *HistoryManager) Load() ([]string, error) {
	file, err := os.Open(h.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var history []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			history = append(history, line)
		}
	}
	return history, scanner.Err()
}

// Append adds a command to the history file.
func (h *HistoryManager) Append(cmd string) error {
	file, err := os.OpenFile(h.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(cmd + "\n")
	return err
}
