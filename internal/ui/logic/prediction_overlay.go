package logic

import (
	"github.com/charmbracelet/lipgloss"
)

// RenderGhostText takes current input and a predicted suggestion and overlays them.
func RenderGhostText(input, suggestion string) string {
	if suggestion == "" {
		return input
	}

	// Suggestion should contain the full predicted command.
	// Highlight only the parts that are NOT in the input.
	ghostStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Dim gray

	// Basic implementation: if suggestion starts with input, return input + the rest in gray.
	if len(suggestion) > len(input) && suggestion[:len(input)] == input {
		return input + ghostStyle.Render(suggestion[len(input):])
	}

	return input
}
