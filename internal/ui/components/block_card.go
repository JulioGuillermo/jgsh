package components

import (
	"strings"
	"time"

	"github.com/julioguillermo/jgsh/internal/syntax/ports"
	"github.com/julioguillermo/jgsh/internal/ui/styles"

	"github.com/charmbracelet/lipgloss"
)

// BlockCard handles the rendering of a single command/output block.
type BlockCard struct {
	Highlighter ports.Highlighter
}

// NewBlockCard creates a new BlockCard instance.
func NewBlockCard(highlighter ports.Highlighter) *BlockCard {
	return &BlockCard{
		Highlighter: highlighter,
	}
}

// Render returns the string representation of a block card.
func (b *BlockCard) Render(title string, cmd string, output string, width int, duration time.Duration, isRunning bool) string {
	cmdPart := b.Highlighter.Highlight(cmd)
	topBorder := Title(title, cmdPart, width, duration, isRunning)

	// Build the content: full command + output
	var content strings.Builder

	// If the command is multi-line, we show it clearly at the top
	if strings.Contains(cmd, "\n") {
		// Highlighted multi-line command
		content.WriteString(cmdPart)
		content.WriteString("\n")
		// Separator
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("┄", width-4)))
		content.WriteString("\n")
	}

	content.WriteString(output)

	body := styles.BaseBlockStyle.
		BorderTop(false).
		Width(width).
		Render(content.String())

	return topBorder + "\n" + body
}
