package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/julioguillermo/jgsh/internal/ui/styles"
)

func Title(title string, width int, duration time.Duration, isRunning bool) string {
	titleText := fmt.Sprintf(" %s ", title)
	titleWidth := lipgloss.Width(titleText)

	// Meta info (duration/status)
	meta := ""
	if isRunning {
		meta = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(" ● RUNNING ")
	} else if duration > 0 {
		meta = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Render(fmt.Sprintf(" %s ", duration.Round(time.Millisecond)))
	}
	metaWidth := lipgloss.Width(meta)

	// Math: ╭─[ + title + ] + ─...─ + meta + ╮
	dashCount := max(width-3-titleWidth-metaWidth, 0)

	topBorder := styles.BaseBlockBorderStyle.Render("╭─[") +
		titleText +
		styles.BaseBlockBorderStyle.Render("]"+strings.Repeat("─", dashCount)) +
		meta +
		styles.BaseBlockBorderStyle.Render("╮")

	return topBorder
}
