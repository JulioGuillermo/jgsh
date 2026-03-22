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
		parentKind := ""
		if node.Parent() != nil {
			parentKind = node.Parent().Kind()
		}

		// For word nodes, check the parent kind to differentiate between command names and arguments
		if kind == "word" && parentKind != "" {
			if parentKind == "command_name" || parentKind == "simple_command" {
				kind = "command_name"
			}
		}
		return h.styleText(kind, parentKind, text[node.StartByte():node.EndByte()])
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
func (h *TSBashHighlighter) styleText(kind, parentKind, content string) string {
	switch kind {
	case "command_name", "program", "command", "simple_command":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Bold(true).Render(content)

	case "string", "raw_string", "string_content", "\"", "'":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff88")).Render(content)

	case "number":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffff")).Render(content)

	case "if", "then", "else", "elif", "fi", "case", "in", "esac", "for", "do", "done", "while", "until", "function", "time", "[[", "]]", "!":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#ff8855")).Bold(true).Render(content)

	case "$", "$(", "variable_name", "variable", "expansion", "simple_expansion", "command_substitution":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0088")).Render(content)

	case ")":
		if parentKind == "command_substitution" {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0088")).Render(content)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00ff")).Render(content)

	case "operator", "|", ">", ">>", "<", "<<", "&&", "||", "&", "redirect":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#0088ff")).Render(content)

	case "(", "[", "]", "{", "}", ";", "punctuation":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#ff00ff")).Render(content)

	case "option", "flag", "word", "argument", "word_content":
		if strings.HasPrefix(content, "--") {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#ffaa88")).Render(content)
		}
		if strings.HasPrefix(content, "-") {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#ffff88")).Render(content)
		}
		// Check if it is a number (for word nodes that should be numbers)
		isNumber := true
		if len(content) == 0 {
			isNumber = false
		}
		for _, r := range content {
			if (r < '0' || r > '9') && r != '.' {
				isNumber = false
				break
			}
		}
		if isNumber {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#00ffff")).Render(content)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Render(content)

	default:
		return lipgloss.NewStyle().Render(content)
	}
}
