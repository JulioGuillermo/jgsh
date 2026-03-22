package logic

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ConfigManager handles user-defined configurations.
type ConfigManager struct {
	fullscreenFile string
}

// NewConfigManager creates a new ConfigManager instance.
func NewConfigManager() *ConfigManager {
	home, _ := os.UserHomeDir()
	return &ConfigManager{
		fullscreenFile: filepath.Join(home, ".jgsh_fullscreen"),
	}
}

// LoadFullscreenCommands reads the list of commands that require full-screen.
func (c *ConfigManager) LoadFullscreenCommands() []string {
	var commands []string

	// Ensure the file exists, if not create an empty one with some examples
	if _, err := os.Stat(c.fullscreenFile); os.IsNotExist(err) {
		examples := "# Add one command per line that requires full-screen handover\n# Example: htop\n# Example: custom-tui-app\n"
		_ = os.WriteFile(c.fullscreenFile, []byte(examples), 0644)
		return commands
	}

	file, err := os.Open(c.fullscreenFile)
	if err != nil {
		return commands
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			commands = append(commands, line)
		}
	}

	return commands
}
