package components

import (
	"strings"
	"unicode/utf8"

	"github.com/julioguillermo/jgsh/internal/syntax/ports"
	"github.com/julioguillermo/jgsh/internal/ui/styles"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputField represents a styled command input field.
type InputField struct {
	textArea    textarea.Model
	highlighter ports.Highlighter
	width       int
	label       string

	// Completion state
	completions     []string
	completionIndex int
	originalWord    string
	originalValue   string
	cursorBefore    int

	// Password mode
	isPassword bool
}

// NewInputField creates a new InputField instance.
func NewInputField(highlighter ports.Highlighter) *InputField {
	ta := textarea.New()
	ta.Placeholder = "Enter command..."
	ta.Prompt = "" // Use our own prompt/header
	ta.Focus()
	// Set a visible cursor style by default using reverse for maximum visibility
	ta.Cursor.Style = lipgloss.NewStyle().Reverse(true)
	ta.ShowLineNumbers = false
	ta.SetHeight(1) // Start with 1 line, will grow if needed

	// Remove backgrounds from focused and blurred styles
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.BlurredStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()

	return &InputField{
		textArea:    ta,
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
	i.textArea.SetWidth(w - 4) // Account for padding/borders
}

// Init initializes the input field.
func (i *InputField) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles input events.
func (i *InputField) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Ensure focus is maintained
	if !i.textArea.Focused() {
		i.textArea.Focus()
	}

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Any key except Tab/Shift+Tab resets completion
		if msg.String() != "tab" && msg.String() != "shift+tab" {
			i.ResetCompletion()
		}
	}

	i.textArea, cmd = i.textArea.Update(msg)

	// Adjust height based on content
	lines := strings.Split(i.textArea.Value(), "\n")
	newHeight := len(lines)
	if newHeight < 1 {
		newHeight = 1
	}
	if newHeight > 10 { // Cap height
		newHeight = 10
	}
	i.textArea.SetHeight(newHeight)

	return i, cmd
}

// Position returns the current cursor position.
func (i *InputField) Position() int {
	// textarea uses (row, col) internally for cursor, but we need absolute position for completion
	// Let's calculate it from the value and cursor position
	val := i.textArea.Value()
	cursorLine := i.textArea.Line()
	cursorCol := i.textArea.LineInfo().CharOffset

	lines := strings.Split(val, "\n")
	pos := 0
	for l := 0; l < cursorLine; l++ {
		pos += utf8.RuneCountInString(lines[l]) + 1 // +1 for newline
	}
	pos += cursorCol
	return pos
}

// Focus focuses the input field.
func (i *InputField) Focus() tea.Cmd {
	return i.textArea.Focus()
}

// Blur blurs the input field.
func (i *InputField) Blur() {
	i.textArea.Blur()
}

// View renders the input field with a modern boxed design.
func (i *InputField) View() string {
	value := i.textArea.Value()
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
		inputContent = i.textArea.View()
	} else {
		displayValue := value
		if i.isPassword {
			displayValue = strings.Repeat("*", utf8.RuneCountInString(value))
		}

		// Highlight the value
		var highlighted string
		if i.isPassword {
			highlighted = displayValue
		} else {
			highlighted = i.highlighter.Highlight(value)
		}

		// Get cursor position and handle it manually to keep syntax highlighting
		if !i.textArea.Focused() {
			inputContent = highlighted
		} else {
			// Check blinking state from the original View
			tiView := i.textArea.View()
			// If tiView contains the reverse escape code, the cursor is currently visible
			showCursor := strings.Contains(tiView, "\x1b[7m") || strings.Contains(tiView, "\x1b[27m")

			if showCursor {
				pos := i.Position()
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
	textLen := utf8.RuneCountInString(i.textArea.Value())

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
				cursorStyle := i.textArea.Cursor.Style
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
		cursorStyle := i.textArea.Cursor.Style
		if cursorStyle.GetBackground() == lipgloss.Color("") {
			cursorStyle = cursorStyle.Reverse(true)
		}
		out.WriteString(cursorStyle.Render(" "))
	}

	return out.String()
}

// Value returns the current text in the input field.
func (i *InputField) Value() string {
	return i.textArea.Value()
}

// SetValue updates the current text in the input field.
func (i *InputField) SetValue(s string) {
	i.textArea.SetValue(s)
	// Reset cursor to end of text
	lines := strings.Split(s, "\n")
	i.textArea.SetCursor(utf8.RuneCountInString(lines[len(lines)-1]))
}

// InsertNewline inserts a newline at the current cursor position.
func (i *InputField) InsertNewline() {
	// textarea handles newlines natively with Enter.
	// But the user specifically wants Shift+Enter to insert a newline.
	// If we are calling this from bubble_tea_app.go, we should insert it manually.
	i.textArea.InsertString("\n")
}

// SetPasswordMode toggles password masking.
func (i *InputField) SetPasswordMode(on bool) {
	i.isPassword = on
}

// Reset clears the input field.
func (i *InputField) Reset() {
	i.textArea.Reset()
	i.textArea.SetHeight(1)
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
	i.originalValue = i.textArea.Value()
	i.cursorBefore = i.Position()
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
	start := i.cursorBefore - utf8.RuneCountInString(i.originalWord)
	if start < 0 {
		start = 0
	}

	// Work with runes for safe indexing
	runes := []rune(i.originalValue)
	prefix := string(runes[:start])
	suffix := string(runes[i.cursorBefore:])

	newValue := prefix + comp + suffix
	i.textArea.SetValue(newValue)

	// Set cursor at the end of the completion
	newPos := start + utf8.RuneCountInString(comp)
	i.setAbsolutePosition(newPos)
}

func (i *InputField) setAbsolutePosition(pos int) {
	// 1. Go to the beginning of the entire input
	for i.textArea.Line() > 0 {
		i.textArea.CursorUp()
	}
	i.textArea.CursorStart()

	// 2. Count characters to find the target logical line and column
	val := i.textArea.Value()
	lines := strings.Split(val, "\n")
	current := 0
	targetRow := 0
	targetCol := 0
	for r, line := range lines {
		lineLen := utf8.RuneCountInString(line)
		if pos <= current+lineLen {
			targetRow = r
			targetCol = pos - current
			break
		}
		current += lineLen + 1
		if r == len(lines)-1 {
			targetRow = r
			targetCol = lineLen
		}
	}

	// 3. Move to targetRow
	for i.textArea.Line() < targetRow {
		i.textArea.CursorDown()
	}
	// 4. Move to targetCol
	i.textArea.SetCursor(targetCol)
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
