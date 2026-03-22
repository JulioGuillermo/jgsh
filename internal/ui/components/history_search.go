package components

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/julioguillermo/jgsh/internal/ui/styles"
)

// HistorySearch is a component for searching through command history.
type HistorySearch struct {
	textInput     textinput.Model
	history       []string
	filtered      []string
	selectedIndex int
	width         int
	active        bool
}

// NewHistorySearch creates a new HistorySearch instance.
func NewHistorySearch() *HistorySearch {
	ti := textinput.New()
	ti.Placeholder = "Search history..."
	ti.Prompt = "🔍 "
	return &HistorySearch{
		textInput: ti,
	}
}

// Activate enables the search with the given history and optional initial query.
func (h *HistorySearch) Activate(history []string, query string) {
	// Deduplicate history, keeping last occurrence
	seen := make(map[string]bool)
	var unique []string
	for i := len(history) - 1; i >= 0; i-- {
		cmd := strings.TrimSpace(history[i])
		if cmd != "" && !seen[cmd] {
			unique = append([]string{cmd}, unique...)
			seen[cmd] = true
		}
	}

	h.history = unique
	h.active = true
	h.textInput.Focus()
	h.textInput.SetValue(query)
	h.textInput.SetCursor(len(query))
	h.filter()
}

// Deactivate disables the search.
func (h *HistorySearch) Deactivate() {
	h.active = false
	h.textInput.Blur()
}

// IsActive returns true if search is currently active.
func (h *HistorySearch) IsActive() bool {
	return h.active
}

// Update handles input events for the search.
func (h *HistorySearch) Update(msg tea.Msg) (string, bool, tea.Cmd) {
	if !h.active {
		return "", false, nil
	}

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+r":
			h.Deactivate()
			return "", true, nil
		case "enter":
			if h.selectedIndex >= 0 && h.selectedIndex < len(h.filtered) {
				selected := h.filtered[h.selectedIndex]
				h.Deactivate()
				return selected, true, nil
			}
			h.Deactivate()
			return "", true, nil
		case "up":
			if h.selectedIndex > 0 {
				h.selectedIndex--
			}
			return "", false, nil
		case "down":
			if h.selectedIndex < len(h.filtered)-1 {
				h.selectedIndex++
			}
			return "", false, nil
		}
	}

	oldQuery := h.textInput.Value()
	h.textInput, cmd = h.textInput.Update(msg)
	newQuery := h.textInput.Value()

	if oldQuery != newQuery {
		h.filter()
	}

	return "", false, cmd
}

// filter updates the filtered list based on the search query.
func (h *HistorySearch) filter() {
	query := strings.ToLower(h.textInput.Value())
	if query == "" {
		h.filtered = h.history
	} else {
		words := strings.Fields(query)
		h.filtered = []string{}
		for _, cmd := range h.history {
			cmdLower := strings.ToLower(cmd)
			match := true
			lastIdx := 0
			for _, word := range words {
				idx := strings.Index(cmdLower[lastIdx:], word)
				if idx == -1 {
					match = false
					break
				}
				lastIdx += idx + len(word)
			}
			if match {
				h.filtered = append(h.filtered, cmd)
			}
		}
	}
	h.selectedIndex = len(h.filtered) - 1
	if h.selectedIndex < 0 {
		h.selectedIndex = 0
	}
}

// highlightMatch applies highlighting to matching parts of the string.
func (h *HistorySearch) highlightMatch(cmd string) string {
	query := h.textInput.Value()
	if query == "" {
		return cmd
	}

	words := strings.Fields(query)
	highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)

	// Escape words for regex
	var escapedWords []string
	for _, w := range words {
		escapedWords = append(escapedWords, regexp.QuoteMeta(w))
	}

	// Create a regex that matches any of the words
	re := regexp.MustCompile("(?i)" + strings.Join(escapedWords, "|"))

	return re.ReplaceAllStringFunc(cmd, func(match string) string {
		return highlightStyle.Render(match)
	})
}

// SetWidth updates the width.
func (h *HistorySearch) SetWidth(w int) {
	h.width = w
}

// View renders the search dialog.
func (h *HistorySearch) View() string {
	if !h.active {
		return ""
	}

	titleText := styles.InputPromptStyle.Render(" HISTORY SEARCH ")
	titleWidth := lipgloss.Width(titleText)

	borderCol := lipgloss.Color("62")
	dashCount := h.width - 3 - titleWidth
	if dashCount < 0 {
		dashCount = 0
	}

	topBorder := lipgloss.NewStyle().Foreground(borderCol).Render("╭─") +
		titleText +
		lipgloss.NewStyle().Foreground(borderCol).Render(strings.Repeat("─", dashCount)+"╮")

	style := lipgloss.NewStyle().
		BorderTop(false).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1, 0, 1). // Reduced bottom padding
		Width(h.width - 2)

	var b strings.Builder
	b.WriteString(h.textInput.View() + "\n")

	// Show results
	if len(h.filtered) > 0 {
		b.WriteString("\n")
		// Show up to 10 results
		pageSize := 10
		start := h.selectedIndex - pageSize/2
		if start < 0 {
			start = 0
		}
		end := start + pageSize
		if end > len(h.filtered) {
			end = len(h.filtered)
			start = end - pageSize
			if start < 0 {
				start = 0
			}
		}

		for i := start; i < end; i++ {
			line := h.filtered[i]
			highlightedLine := h.highlightMatch(line)
			if i == h.selectedIndex {
				b.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color("62")).
					Bold(true).
					Render("> "+highlightedLine) + "\n")
			} else {
				b.WriteString("  " + highlightedLine + "\n")
			}
		}
	} else {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" No matches found") + "\n")
	}

	return topBorder + style.Render(b.String())
}
