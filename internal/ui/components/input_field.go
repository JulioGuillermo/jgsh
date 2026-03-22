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
	label       string

	// Completion state
	completions     []string
	completionIndex int
	originalWord    string
	originalValue   string
	cursorBefore    int
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
		label:       " COMMAND ",
	}
}

// SetLabel updates the label of the input field.
func (i *InputField) SetLabel(label string) {
	i.label = label
}

// SetWidth updates the width of the input field.
func (i *InputField) SetWidth(w int) {
	i.width = w
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Any key except Tab/Shift+Tab resets completion
		if msg.String() != "tab" && msg.String() != "shift+tab" {
			i.ResetCompletion()
		}
	}

	i.textInput, cmd = i.textInput.Update(msg)
	return i, cmd
}

// View renders the input field with a modern boxed design.
func (i *InputField) View() string {
	value := i.textInput.Value()
	titleText := styles.InputPromptStyle.Render(i.label)
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
			// Check blinking state from the original View
			tiView := i.textInput.View()
			// If tiView contains the reverse escape code, the cursor is currently visible
			showCursor := strings.Contains(tiView, "\x1b[7m") || strings.Contains(tiView, "\x1b[27m")

			if showCursor {
				pos := i.textInput.Position()
				inputContent = i.renderWithCursor(highlighted, pos)
			} else {
				inputContent = highlighted
			}
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

// SetValue updates the current text in the input field.
func (i *InputField) SetValue(s string) {
	i.textInput.SetValue(s)
	i.textInput.SetCursor(len(s))
}

// SetPasswordMode toggles password masking.
func (i *InputField) SetPasswordMode(on bool) {
	if on {
		i.textInput.EchoMode = textinput.EchoPassword
		i.textInput.EchoCharacter = '*'
	} else {
		i.textInput.EchoMode = textinput.EchoNormal
	}
}

// Reset clears the input field.
func (i *InputField) Reset() {
	i.textInput.Reset()
	i.ResetCompletion()
}

// ResetCompletion resets the completion state.
func (i *InputField) ResetCompletion() {
	i.completions = nil
	i.completionIndex = -1
	i.originalWord = ""
	i.originalValue = ""
	i.cursorBefore = -1
}

// SetCompletions sets new completions and starts cycling.
func (i *InputField) SetCompletions(completions []string, word string) {
	if len(completions) == 0 {
		return
	}
	i.completions = completions
	i.originalWord = word
	i.originalValue = i.textInput.Value()
	i.cursorBefore = i.textInput.Position()
	i.completionIndex = 0
	i.applyCompletion()
}

// NextCompletion cycles to the next completion.
func (i *InputField) NextCompletion() {
	if len(i.completions) == 0 {
		return
	}
	i.completionIndex = (i.completionIndex + 1) % len(i.completions)
	i.applyCompletion()
}

// PrevCompletion cycles to the previous completion.
func (i *InputField) PrevCompletion() {
	if len(i.completions) == 0 {
		return
	}
	i.completionIndex = (i.completionIndex - 1 + len(i.completions)) % len(i.completions)
	i.applyCompletion()
}

func (i *InputField) applyCompletion() {
	if i.completionIndex < 0 || i.completionIndex >= len(i.completions) {
		return
	}

	comp := i.completions[i.completionIndex]

	// Replace the word at the cursor with the completion
	start := i.cursorBefore - len(i.originalWord)
	if start < 0 {
		start = 0
	}

	prefix := i.originalValue[:start]
	suffix := i.originalValue[i.cursorBefore:]

	i.textInput.SetValue(prefix + comp + suffix)
	i.textInput.SetCursor(len(prefix + comp))
}

// IsCycling returns true if the input field is currently cycling completions.
func (i *InputField) IsCycling() bool {
	return len(i.completions) > 0
}

// Completions returns the current list of completions.
func (i *InputField) Completions() []string {
	return i.completions
}

// CompletionIndex returns the current completion index.
func (i *InputField) CompletionIndex() int {
	return i.completionIndex
}

// Position returns the current cursor position.
func (i *InputField) Position() int {
	return i.textInput.Position()
}
