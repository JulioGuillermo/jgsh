package components

import (
	"github.com/julioguillermo/jgsh/internal/ui/styles"
)

// Header renders the app title.
type Header struct {
	Title string
}

// Render returns the header as a string.
func (h *Header) Render() string {
	return styles.HeaderStyle.Render(h.Title)
}
