package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// CompletionSelector renders a list of completions.
type CompletionSelector struct {
	completions   []string
	selectedIndex int
	width         int
	active        bool
}

// NewCompletionSelector creates a new CompletionSelector instance.
func NewCompletionSelector() *CompletionSelector {
	return &CompletionSelector{}
}

// Activate enables the selector with the given completions.
func (s *CompletionSelector) Activate(completions []string, index int) {
	s.completions = completions
	s.selectedIndex = index
	s.active = true
}

// Deactivate disables the selector.
func (s *CompletionSelector) Deactivate() {
	s.active = false
}

// IsActive returns true if the selector is active.
func (s *CompletionSelector) IsActive() bool {
	return s.active && len(s.completions) > 0
}

// SetIndex updates the selected index.
func (s *CompletionSelector) SetIndex(index int) {
	s.selectedIndex = index
}

// SetWidth updates the width.
func (s *CompletionSelector) SetWidth(w int) {
	s.width = w
}

// View renders the completion list.
func (s *CompletionSelector) View() string {
	if !s.IsActive() {
		return ""
	}

	titleText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("13")). // Magenta background
		Bold(true).
		Padding(0, 1).
		Render(" COMPLETIONS ")
	titleWidth := lipgloss.Width(titleText)

	borderCol := lipgloss.Color("13")
	dashCount := s.width - 3 - titleWidth
	if dashCount < 0 {
		dashCount = 0
	}

	topBorder := lipgloss.NewStyle().Foreground(borderCol).Render("╭─") +
		titleText +
		lipgloss.NewStyle().Foreground(borderCol).Render(strings.Repeat("─", dashCount)+"╮")

	style := lipgloss.NewStyle().
		BorderForeground(borderCol).
		Border(lipgloss.RoundedBorder(), false, true, true, true).
		Padding(0, 1).
		Width(s.width - 2)

	var b strings.Builder

	// Show up to 8 completions
	pageSize := 8
	start := s.selectedIndex - pageSize/2
	if start < 0 {
		start = 0
	}
	end := start + pageSize
	if end > len(s.completions) {
		end = len(s.completions)
		start = end - pageSize
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		comp := s.completions[i]
		if i == s.selectedIndex {
			b.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("13")).
				Bold(true).
				Render("> "+comp) + "\n")
		} else {
			b.WriteString("  " + comp + "\n")
		}
	}

	if len(s.completions) > end {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("  ... %d more", len(s.completions)-end)) + "\n")
	}

	return topBorder + "\n" + style.Render(b.String())
}
