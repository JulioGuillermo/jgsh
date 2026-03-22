package components

import (
	"strings"
	"unicode/utf8"

	"github.com/julioguillermo/jgsh/internal/syntax/ports"
	"github.com/julioguillermo/jgsh/internal/ui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputField represents a styled command input field.
type InputField struct {
	textInput   textinput.Model
	highlighter ports.Highlighter
	width       int
}

// NewInputField creates a new InputField instance.
func NewInputField(highlighter ports.Highlighter) *InputField {
	ti := textinput.New()
	ti.Placeholder = "Enter command..."
	ti.Prompt = "" // Use our own prompt/header
	ti.Focus()
	// Set a visible cursor style by default using reverse for maximum visibility
	ti.CursorStyle = lipgloss.NewStyle().Reverse(true)
	return &InputField{
		textInput:   ti,
		highlighter: highlighter,
		width:       80,
	}
}

// Init initializes the input field.
func (i *InputField) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles input events.
func (i *InputField) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Ensure focus is maintained
	if !i.textInput.Focused() {
		i.textInput.Focus()
	}
	var cmd tea.Cmd
	i.textInput, cmd = i.textInput.Update(msg)
	return i, cmd
}

// SetWidth updates the width of the input field.
func (i *InputField) SetWidth(w int) {
	i.width = w
}

// View renders the input field with a modern boxed design.
func (i *InputField) View() string {
	value := i.textInput.Value()
	titleText := styles.InputPromptStyle.Render(" COMMAND ")
	titleWidth := lipgloss.Width(titleText)

	borderCol := lipgloss.Color("62")
	dashCount := i.width - 3 - titleWidth
	if dashCount < 0 {
		dashCount = 0
	}

	topBorder := lipgloss.NewStyle().Foreground(borderCol).Render("╭─") +
		titleText +
		lipgloss.NewStyle().Foreground(borderCol).Render(strings.Repeat("─", dashCount)+"╮")

	// Render content with cursor and highlighting
	var inputContent string
	if value == "" {
		inputContent = i.textInput.View()
	} else {
		// Highlight the value
		highlighted := i.highlighter.Highlight(value)

		// Get cursor position and handle it manually to keep syntax highlighting
		if !i.textInput.Focused() {
			inputContent = highlighted
		} else {
			// Get position from the textinput model
			pos := i.textInput.Position()
			inputContent = i.renderWithCursor(highlighted, pos)
		}
	}

	// Render the body
	body := styles.InputBoxStyle.
		BorderTop(false).
		Width(i.width - 2).
		Render(inputContent)

	return topBorder + "\n" + body
}

// renderWithCursor inserts the cursor into a highlighted (ANSI) string.
func (i *InputField) renderWithCursor(highlighted string, pos int) string {
	// Simple ANSI-aware character counting
	var out strings.Builder
	var charCount int
	var inAnsi bool
	var cursorInserted bool

	// Special case: cursor at the end
	textLen := utf8.RuneCountInString(i.textInput.Value())

	runes := []rune(highlighted)
	for j := 0; j < len(runes); j++ {
		r := runes[j]

		// Track ANSI escape sequences
		if r == '\x1b' {
			inAnsi = true
		}

		if !inAnsi {
			if charCount == pos {
				// Insert styled cursor. We wrap the next character if it exists.
				// Since we are in a highlighted string, we might want to keep the color
				// but change the background.
				cursorStyle := i.textInput.CursorStyle
				if cursorStyle.GetBackground() == lipgloss.Color("") {
					cursorStyle = cursorStyle.Reverse(true)
				}
				out.WriteString(cursorStyle.Render(string(r)))
				cursorInserted = true
				charCount++
				continue
			}
			charCount++
		}

		out.WriteRune(r)

		if inAnsi && r == 'm' {
			inAnsi = false
		}
	}

	// If cursor was at the end, append it
	if !cursorInserted && pos >= textLen {
		cursorStyle := i.textInput.CursorStyle
		if cursorStyle.GetBackground() == lipgloss.Color("") {
			cursorStyle = cursorStyle.Reverse(true)
		}
		out.WriteString(cursorStyle.Render(" "))
	}

	return out.String()
}

// Value returns the current text in the input field.
func (i *InputField) Value() string {
	return i.textInput.Value()
}

// Reset clears the input field.
func (i *InputField) Reset() {
	i.textInput.Reset()
}
