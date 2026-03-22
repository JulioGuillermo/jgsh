package adapters

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/julioguillermo/jgsh/internal/sh/logic"
	"github.com/julioguillermo/jgsh/internal/sh/ports"
	syntaxports "github.com/julioguillermo/jgsh/internal/syntax/ports"
	"github.com/julioguillermo/jgsh/internal/ui/components"
	"github.com/julioguillermo/jgsh/internal/ui/domain"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ShellMsg represents a chunk of data read from the shell.
type ShellMsg string

// TickMsg is sent every second to update the clock.
type TickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// BubbleTeaApp implements the TUI using Bubble Tea.
type BubbleTeaApp struct {
	shellManager ports.ShellManager
	inputField   *components.InputField
	blockCard    *components.BlockCard
	header       *components.Header
	statusBar    *components.StatusBar
	viewport     viewport.Model
	highlighter  syntaxports.Highlighter
	blocks       []domain.Block
	currentBlock *domain.Block
	history      *logic.HistoryManager
	historyCmds  []string
	historyIndex int
	draftCommand string
	ready        bool
	width        int
	height       int
}

// NewBubbleTeaApp creates a new BubbleTeaApp instance.
func NewBubbleTeaApp(shellManager ports.ShellManager, inputField *components.InputField, highlighter syntaxports.Highlighter, history *logic.HistoryManager) *BubbleTeaApp {
	// Load history only for navigation
	histCmds, _ := history.Load()

	return &BubbleTeaApp{
		shellManager: shellManager,
		inputField:   inputField,
		blockCard:    components.NewBlockCard(highlighter),
		header:       &components.Header{Title: "🚀 PROJECT GLOCK"},
		statusBar:    &components.StatusBar{},
		viewport:     viewport.New(0, 0),
		highlighter:  highlighter,
		blocks:       make([]domain.Block, 0),
		history:      history,
		historyCmds:  histCmds,
		historyIndex: -1,
		currentBlock: &domain.Block{
			Command: "",
			Output:  "",
		},
	}
}

// listenToShell is a tea.Cmd that reads from the shell and sends ShellMsg.
func listenToShell(reader io.Reader) tea.Cmd {
	return func() tea.Msg {
		buf := make([]byte, 4096)
		n, err := reader.Read(buf)
		if err != nil {
			return nil
		}
		return ShellMsg(string(buf[:n]))
	}
}

// Init initializes the Bubble Tea application.
func (m *BubbleTeaApp) Init() tea.Cmd {
	return tea.Batch(
		listenToShell(m.shellManager.GetReader()),
		m.inputField.Init(),
		tick(),
	)
}

// Update handles messages and updates the application state.
func (m *BubbleTeaApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	finalMsg := msg

	switch msg := msg.(type) {
	case TickMsg:
		cmds = append(cmds, tick())

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Header + Input(3) + Status(1) + Spacing(3)
		m.viewport.Width = m.width
		m.viewport.Height = m.height - 10
		m.ready = true

	case tea.MouseMsg:
		if !m.ready {
			break
		}
		// Translate mouse clicks for the input field
		headerHeight := lipgloss.Height(m.header.Render())
		inputFieldY := headerHeight + m.viewport.Height + 3

		// Right click to copy block
		if msg.Type == tea.MouseRight {
			headerHeight := 3
			// Check if click is within viewport bounds
			if msg.Y >= headerHeight && msg.Y < headerHeight+m.viewport.Height {
				viewportY := msg.Y - headerHeight + m.viewport.YOffset
				block := m.findBlockByY(viewportY)
				if block != nil && block.Command != "" {
					content := fmt.Sprintf("$ %s\n%s", block.Command, block.Output)
					// Remove trailing newlines and carriage returns
					content = strings.TrimRight(content, "\n\r ")
					clipboard.WriteAll(content)
					// Visual feedback in status bar
					m.statusBar.Time = "📋 COPIED: " + block.Command
				}
			}
		}

		// Only translate and pass mouse clicks to the input field, not scroll events
		isClick := msg.Type == tea.MouseLeft || msg.Type == tea.MouseRelease
		if isClick && msg.Y >= inputFieldY-1 && msg.Y <= inputFieldY+1 {
			msg.X -= 2
			msg.Y = 0
			finalMsg = msg
		}
		// Pass to viewport for scrolling
		var vpCmd tea.Cmd
		m.viewport, vpCmd = m.viewport.Update(msg)
		if vpCmd != nil {
			cmds = append(cmds, vpCmd)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+shift+c":
			// If we had selection support, we'd copy it here.
			// For now, let's copy the current input if it's not empty.
			val := m.inputField.Value()
			if val != "" {
				clipboard.WriteAll(val)
			}
			return m, nil
		case "ctrl+c":
			// If a command is running, send SIGINT, otherwise clear input
			if m.currentBlock != nil && m.currentBlock.IsRunning {
				m.shellManager.Write([]byte("\x03"))
			} else {
				m.inputField.Reset()
				m.historyIndex = -1
				m.draftCommand = ""
			}
			return m, nil
		case "up":
			if len(m.historyCmds) > 0 && (m.currentBlock == nil || !m.currentBlock.IsRunning) {
				if m.historyIndex == -1 {
					m.draftCommand = m.inputField.Value()
					m.historyIndex = len(m.historyCmds) - 1
				} else if m.historyIndex > 0 {
					m.historyIndex--
				}
				m.inputField.SetValue(m.historyCmds[m.historyIndex])
			}
			return m, nil
		case "down":
			if m.currentBlock == nil || !m.currentBlock.IsRunning {
				if m.historyIndex != -1 {
					if m.historyIndex < len(m.historyCmds)-1 {
						m.historyIndex++
						m.inputField.SetValue(m.historyCmds[m.historyIndex])
					} else {
						m.historyIndex = -1
						m.inputField.SetValue(m.draftCommand)
					}
				}
			}
			return m, nil
		case "enter":
			val := m.inputField.Value()
			if val == "" || (m.currentBlock != nil && m.currentBlock.IsRunning) {
				return m, nil
			}

			// Reset history index
			m.historyIndex = -1
			m.draftCommand = ""

			// Save to persistent history and local navigation buffer
			m.history.Append(val)
			m.historyCmds = append(m.historyCmds, val)

			// Special command: exit
			if val == "exit" {
				return m, tea.Quit
			}

			// Send command
			m.shellManager.Write([]byte(val + "\n"))

			// Setup current block
			m.currentBlock.Command = val
			m.currentBlock.StartTime = time.Now()
			m.currentBlock.IsRunning = true

			m.inputField.Reset()
		}
	case ShellMsg:
		msgStr := string(msg)
		m.currentBlock.Output += msgStr

		if logic.DetectPrompt([]byte(m.currentBlock.Output)) {
			// Finish current block
			m.currentBlock.Duration = time.Since(m.currentBlock.StartTime)
			m.currentBlock.Finished = true
			m.currentBlock.IsRunning = false

			m.currentBlock.Output = logic.StripEcho(m.currentBlock.Output, m.currentBlock.Command)
			m.currentBlock.Output = logic.StripPrompt(m.currentBlock.Output)

			if m.currentBlock.Command != "" {
				m.blocks = append(m.blocks, *m.currentBlock)
			}

			// Start new fresh block
			m.currentBlock = &domain.Block{
				Command: "",
				Output:  "",
			}
			m.viewport.GotoBottom()
		}
		cmds = append(cmds, listenToShell(m.shellManager.GetReader()))
	}

	// Update components
	_, inputCmd := m.inputField.Update(finalMsg)
	if inputCmd != nil {
		cmds = append(cmds, inputCmd)
	}

	// Update viewport content
	m.viewport.SetContent(m.renderAllBlocks())

	return m, tea.Batch(cmds...)
}

// findBlockByY returns the block at the given viewport-relative Y coordinate.
func (m *BubbleTeaApp) findBlockByY(y int) *domain.Block {
	if m.viewport.Width <= 0 {
		return nil
	}
	currentY := 0
	for i, block := range m.blocks {
		h := lipgloss.Height(m.blockCard.Render(fmt.Sprintf("BLOCK %d", i), block.Command, block.Output, m.viewport.Width-3, block.Duration, false))
		if y >= currentY && y < currentY+h {
			return &m.blocks[i]
		}
		currentY += h + 1 // +1 for the newline between cards
	}

	// Check current block
	if m.currentBlock != nil && (m.currentBlock.Command != "" || m.currentBlock.Output != "") {
		out := m.currentBlock.Output
		h := lipgloss.Height(m.blockCard.Render("EXEC", m.currentBlock.Command, out, m.viewport.Width-3, time.Since(m.currentBlock.StartTime), true))
		if y >= currentY && y < currentY+h {
			return m.currentBlock
		}
	}

	return nil
}

// renderAllBlocks renders all blocks into a single string for the viewport.
func (m *BubbleTeaApp) renderAllBlocks() string {
	var b strings.Builder
	for i, block := range m.blocks {
		b.WriteString(m.blockCard.Render(fmt.Sprintf("BLOCK %d", i), block.Command, block.Output, m.viewport.Width-3, block.Duration, false))
		b.WriteString("\n")
	}

	if m.currentBlock != nil && m.currentBlock.Command != "" {
		out := logic.StripEcho(m.currentBlock.Output, m.currentBlock.Command)
		out = logic.StripPrompt(out)
		b.WriteString(m.blockCard.Render("EXEC", m.currentBlock.Command, out, m.viewport.Width-3, time.Since(m.currentBlock.StartTime), true))
	}
	return b.String()
}

// renderScrollBar creates a vertical scroll indicator.
func (m *BubbleTeaApp) renderScrollBar() string {
	height := m.viewport.Height
	if height <= 0 {
		return ""
	}

	scrollPct := m.viewport.ScrollPercent()
	barPos := int(float64(height-1) * scrollPct)

	var s strings.Builder
	for i := 0; i < height; i++ {
		if i == barPos {
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("┃"))
		} else {
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("│"))
		}
		if i < height-1 {
			s.WriteString("\n")
		}
	}
	return s.String()
}

// View renders the application UI.
func (m *BubbleTeaApp) View() string {
	if !m.ready {
		return "Initializing UI..."
	}

	m.inputField.SetWidth(m.width)
	m.statusBar.BlocksCount = len(m.blocks)
	m.statusBar.Width = m.width
	m.statusBar.CWD = logic.GetShellCWD(m.shellManager.GetPID())
	m.statusBar.Git = logic.GetGitInfo(m.statusBar.CWD)
	m.statusBar.Project = logic.GetProjectInfo(m.statusBar.CWD)
	m.statusBar.Venv = logic.GetVenvInfo()
	m.statusBar.Time = time.Now().Format("15:04:05")

	// Join the viewport and scrollbar horizontally
	viewArea := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.viewport.View(),
		m.renderScrollBar(),
	)

	bottomArea := "\n" + m.inputField.View()
	if m.currentBlock != nil && m.currentBlock.IsRunning {
		// Replace input with a "Command Running" message of the same height to avoid jumpy UI
		runningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Bold(true).
			Width(m.width).
			Padding(1, 2).
			Height(3) // Match input field height (topBorder + newline + body)
		bottomArea = "\n" + runningStyle.Render("⏳ COMMAND RUNNING... [Ctrl+C to stop]")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.header.Render(),
		viewArea,
		bottomArea,
		"\n"+m.statusBar.Render(),
	)
}
