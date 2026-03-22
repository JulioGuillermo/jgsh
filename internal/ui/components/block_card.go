package components

import (
	"github.com/julioguillermo/jgsh/internal/sh/logic"
	"github.com/julioguillermo/jgsh/internal/syntax/ports"
	"github.com/julioguillermo/jgsh/internal/ui/styles"
	"time"
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
	topBorder := Title(cmdPart, width, duration, isRunning)

	// Ensure content is clean
	outPart := ""
	if output != "" {
		outPart = logic.StripAnsi(output)
	}

	body := styles.BaseBlockStyle.
		BorderTop(false).
		Width(width).
		Render(outPart)

	return topBorder + "\n" + body
}
