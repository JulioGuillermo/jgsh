package components

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/julioguillermo/jgsh/internal/ui/styles"
	"strings"
)

// StatusBar renders the bottom info bar.
type StatusBar struct {
	BlocksCount int
	Width       int
	CWD         string
	GitBranch   string
}

// Render returns the status bar as a string.
func (s *StatusBar) Render() string {
	cwd := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(s.CWD)
	git := ""
	if s.GitBranch != "" {
		git = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("  " + s.GitBranch)
	}

	left := fmt.Sprintf(" %s%s ", cwd, git)
	right := fmt.Sprintf(" Blocks: %d | Ctrl+C to exit ", s.BlocksCount)

	// Spacing to push right part to the end
	space := s.Width - lipgloss.Width(left) - lipgloss.Width(right)
	if space < 0 {
		space = 0
	}

	content := left + strings.Repeat(" ", space) + right
	return styles.StatusStyle.Width(s.Width).Render(content)
}
