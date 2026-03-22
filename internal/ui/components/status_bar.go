package components

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/julioguillermo/jgsh/internal/sh/logic"
	"strings"
)

// StatusBar renders the bottom info bar.
type StatusBar struct {
	BlocksCount     int
	Width           int
	CWD             string
	Git             logic.GitStatus
	Project         string
	Venv            string
	Time            string
	Completions     []string
	CompletionIndex int
}

// getLangIcon returns a nerd font icon for common languages.
func (s *StatusBar) getLangIcon(lang string) string {
	switch lang {
	case "Go":
		return "󰟓 "
	case "Node.js":
		return "󰎙 "
	case "Python":
		return "󰌠 "
	case "Rust":
		return "󱘗 "
	case "Java/Gradle", "Java/Maven":
		return "󰬷 "
	case "PHP":
		return "󰌟 "
	case "Ruby":
		return "󰴭 "
	case "Make":
		return "󱁤 "
	}
	return ""
}

// Render returns the status bar as a string.
func (s *StatusBar) Render() string {
	// Styles
	cwdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	gitStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	projectStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	venvStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// Git sub-styles
	stagedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	modifiedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	untrackedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	insertionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	deletionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	aheadStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	behindStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	cwdText := "📂 " + s.CWD
	gitPart := ""
	if s.Git.Branch != "" {
		branch := gitStyle.Render("  " + s.Git.Branch)
		var stats []string
		if s.Git.Staged > 0 {
			stats = append(stats, stagedStyle.Render(fmt.Sprintf("●%d", s.Git.Staged)))
		}
		if s.Git.Modified > 0 {
			stats = append(stats, modifiedStyle.Render(fmt.Sprintf("✚%d", s.Git.Modified)))
		}
		if s.Git.Untracked > 0 {
			stats = append(stats, untrackedStyle.Render(fmt.Sprintf("?%d", s.Git.Untracked)))
		}
		if s.Git.Insertions > 0 {
			stats = append(stats, insertionStyle.Render(fmt.Sprintf("+%d", s.Git.Insertions)))
		}
		if s.Git.Deletions > 0 {
			stats = append(stats, deletionStyle.Render(fmt.Sprintf("-%d", s.Git.Deletions)))
		}
		if s.Git.Ahead > 0 {
			stats = append(stats, aheadStyle.Render(fmt.Sprintf("↑%d", s.Git.Ahead)))
		}
		if s.Git.Behind > 0 {
			stats = append(stats, behindStyle.Render(fmt.Sprintf("↓%d", s.Git.Behind)))
		}

		statusStr := ""
		if len(stats) > 0 {
			statusStr = " [" + strings.Join(stats, " ") + "]"
		}
		gitPart = branch + statusStr
	}

	projectPart := ""
	if s.Project != "" {
		projectPart = projectStyle.Render("  " + s.getLangIcon(s.Project) + s.Project)
	}

	venvPart := ""
	if s.Venv != "" {
		venvPart = venvStyle.Render("  󱆍 " + s.Venv)
	}

	timePart := timeStyle.Render(" 󱑎 " + s.Time + " ")

	// Build left side
	leftSide := cwdStyle.Render(cwdText) + gitPart + projectPart + venvPart

	// Calculate widths
	leftWidth := lipgloss.Width(leftSide)
	rightWidth := lipgloss.Width(timePart)

	// If too wide, truncate the CWD
	if leftWidth+rightWidth > s.Width && s.Width > 20 {
		availableForLeft := s.Width - rightWidth - 5
		if availableForLeft > 10 {
			// Recalculate leftSide with truncated CWD
			cwdText = "📂 …" + s.CWD[len(s.CWD)-(availableForLeft-4):]
			leftSide = cwdStyle.Render(cwdText) + gitPart + projectPart + venvPart
			leftWidth = lipgloss.Width(leftSide)
		}
	}

	spaceCount := s.Width - leftWidth - rightWidth
	if spaceCount < 0 {
		spaceCount = 0
	}

	return leftSide + strings.Repeat(" ", spaceCount) + timePart
}
