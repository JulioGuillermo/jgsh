package ports

// Highlighter defines the interface for syntax highlighting.
type Highlighter interface {
	Highlight(text string) string
}
