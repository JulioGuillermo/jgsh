package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/julioguillermo/jgsh/internal/ui/styles"
)

func Title(content string, width int, duration time.Duration, isRunning bool) (string, bool) {
	// Meta info (duration/status)
	meta := ""
	if isRunning {
		meta = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(" ● RUNNING ")
	} else if duration > 0 {
		meta = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Render(fmt.Sprintf(" %s ", duration.Round(time.Millisecond)))
	}
	metaWidth := lipgloss.Width(meta)

	// If content is multi-line, take only the first line and add ellipsis
	displayContent := content
	show := false
	if strings.Contains(content, "\n") {
		displayContent = strings.Split(content, "\n")[0] + "..."
		show = true
	}

	if len(displayContent) > width-metaWidth-15 {
		displayContent = displayContent[:width-metaWidth-15] + "..."
		show = true
	}

	// Clean up ANSI for width calculation
	titleText := fmt.Sprintf(" %s ", displayContent)
	titleWidth := lipgloss.Width(titleText)

	// Math: ╭─[ + title + ] + ─...─ + meta + ╮
	dashCount := max(width-3-titleWidth-metaWidth, 0)

	topBorder := styles.BaseBlockBorderStyle.Render("╭─[") +
		titleText +
		styles.BaseBlockBorderStyle.Render("]"+strings.Repeat("─", dashCount)) +
		meta +
		styles.BaseBlockBorderStyle.Render("╮")

	return topBorder, show
}
