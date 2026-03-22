package adapters

import (
	"context"
	"strings"

	"github.com/julioguillermo/jgsh/internal/syntax/ports"

	"github.com/charmbracelet/lipgloss"
	sitter "github.com/tree-sitter/go-tree-sitter"
	bash "github.com/tree-sitter/tree-sitter-bash/bindings/go"
)

// TSBashHighlighter implements the Highlighter interface for Bash.
type TSBashHighlighter struct {
	parser *sitter.Parser
}

// NewTSBashHighlighter creates a new TSBashHighlighter instance.
func NewTSBashHighlighter() (ports.Highlighter, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(bash.Language()))
	return &TSBashHighlighter{
		parser: parser,
	}, nil
}

// Highlight highlights the given bash text using Tree-sitter nodes.
func (h *TSBashHighlighter) Highlight(text string) string {
	if text == "" {
		return ""
	}

	tree := h.parser.ParseCtx(context.Background(), []byte(text), nil)
	if tree == nil {
		return text
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	return h.renderNode(rootNode, text)
}

// renderNode recursively applies styles to nodes.
func (h *TSBashHighlighter) renderNode(node *sitter.Node, text string) string {
	if node.ChildCount() == 0 {
		kind := node.Kind()
		// For word nodes, check the parent kind to differentiate between command names and arguments
		if kind == "word" && node.Parent() != nil {
			parentKind := node.Parent().Kind()
			if parentKind == "command_name" || parentKind == "simple_command" {
				kind = "command_name"
			}
		}
		return h.styleText(kind, text[node.StartByte():node.EndByte()])
	}

	var result strings.Builder
	lastEnd := node.StartByte()

	for i := uint(0); i < node.ChildCount(); i++ {
		child := node.Child(i)

		// Fill in gaps (spaces, symbols not in children)
		if child.StartByte() > lastEnd {
			result.WriteString(text[lastEnd:child.StartByte()])
		}

		result.WriteString(h.renderNode(child, text))
		lastEnd = child.EndByte()
	}

	// Fill in remaining text
	if lastEnd < node.EndByte() {
		result.WriteString(text[lastEnd:node.EndByte()])
	}

	return result.String()
}

// styleText applies lipgloss styles based on node types.
func (h *TSBashHighlighter) styleText(kind, content string) string {
	// Reverted to your preferred colors but with improved mapping
	switch kind {
	case "command_name", "program", "command", "simple_command":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true).Render(content) // Yellow (Color 3)
	case "argument", "word", "word_content":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Render(content) // White (Color 15)
	case "string", "raw_string", "string_content", "concatenation":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render(content) // Green (Color 2)
	case "variable_name", "variable":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render(content) // Cyan (Color 6)
	case "option", "flag":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render(content) // Magenta (Color 5)
	case "operator", "|", ">", ">>", "&&", "||", ";", "(", ")", "redirect":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render(content) // Red (Color 1)
	default:
		// Default bold for anything else
		return lipgloss.NewStyle().Bold(true).Render(content)
	}
}
