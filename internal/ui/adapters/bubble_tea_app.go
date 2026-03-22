package adapters

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
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
	shellManager       ports.ShellManager
	inputField         *components.InputField
	blockCard          *components.BlockCard
	historySearch      *components.HistorySearch
	completionSelector *components.CompletionSelector
	header             *components.Header
	statusBar          *components.StatusBar
	viewport           viewport.Model
	highlighter        syntaxports.Highlighter
	blocks             []domain.Block
	currentBlock       *domain.Block
	history            *logic.HistoryManager
	config             *logic.ConfigManager
	completionEngine   *logic.CompletionEngine
	historyCmds        []string
	customFullScreen   []string
	historyIndex       int
	draftCommand       string
	ready              bool
	width              int
	height             int
}

// NewBubbleTeaApp creates a new BubbleTeaApp instance.
func NewBubbleTeaApp(shellManager ports.ShellManager, inputField *components.InputField, highlighter syntaxports.Highlighter, history *logic.HistoryManager) *BubbleTeaApp {
	// Load history only for navigation
	histCmds, _ := history.Load()

	config := logic.NewConfigManager()
	customFS := config.LoadFullscreenCommands()

	return &BubbleTeaApp{
		shellManager:       shellManager,
		inputField:         inputField,
		blockCard:          components.NewBlockCard(highlighter),
		historySearch:      components.NewHistorySearch(),
		completionSelector: components.NewCompletionSelector(),
		header:             &components.Header{Title: "🚀 PROJECT GLOCK"},
		statusBar:          &components.StatusBar{},
		viewport:           viewport.New(0, 0),
		highlighter:        highlighter,
		blocks:             make([]domain.Block, 0),
		history:            history,
		config:             config,
		completionEngine:   logic.NewCompletionEngine(),
		historyCmds:        histCmds,
		customFullScreen:   customFS,
		historyIndex:       -1,
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

	// Handle history search if active
	if m.historySearch.IsActive() {
		selected, closed, searchCmd := m.historySearch.Update(msg)
		if searchCmd != nil {
			cmds = append(cmds, searchCmd)
		}
		if closed {
			if selected != "" {
				m.inputField.SetValue(selected)
			}
		}
		// If search is active, we mostly don't want other components to handle keys
		// unless it's a window size change
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			return m, tea.Batch(cmds...)
		}
	}

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
		// If a command is running, transfer ALL control to the PTY (Raw Passthrough)
		if m.currentBlock != nil && m.currentBlock.IsRunning {
			keyStr := msg.String()

			// Always allow Ctrl+C to send SIGINT
			if keyStr == "ctrl+c" {
				m.shellManager.Write([]byte("\x03"))
				return m, nil
			}

			// Handle special keys mapping to ANSI/ASCII
			switch keyStr {
			case "enter":
				m.shellManager.Write([]byte("\r"))
			case "backspace":
				m.shellManager.Write([]byte("\x7f"))
			case "tab":
				m.shellManager.Write([]byte("\t"))
			case "esc":
				m.shellManager.Write([]byte("\x1b"))
			case "up":
				m.shellManager.Write([]byte("\x1b[A"))
			case "down":
				m.shellManager.Write([]byte("\x1b[B"))
			case "right":
				m.shellManager.Write([]byte("\x1b[C"))
			case "left":
				m.shellManager.Write([]byte("\x1b[D"))
			case "space":
				m.shellManager.Write([]byte(" "))
			default:
				// For normal characters and other Ctrl keys
				if len(msg.Runes) > 0 {
					m.shellManager.Write([]byte(string(msg.Runes)))
				} else if strings.HasPrefix(keyStr, "ctrl+") && len(keyStr) == 6 {
					// Handle ctrl+a through ctrl+z
					char := keyStr[5]
					if char >= 'a' && char <= 'z' {
						m.shellManager.Write([]byte{char - 'a' + 1})
					}
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "tab":
			if m.inputField.IsCycling() {
				m.inputField.NextCompletion()
				m.completionSelector.SetIndex(m.inputField.CompletionIndex())
			} else {
				cwd := logic.GetShellCWD(m.shellManager.GetPID())
				val := m.inputField.Value()
				pos := m.inputField.Position()
				matches, word := m.completionEngine.Complete(val, pos, cwd)
				if len(matches) > 0 {
					m.inputField.SetCompletions(matches, word)
					m.completionSelector.Activate(matches, 0)
				}
			}
			return m, nil
		case "shift+tab":
			if m.inputField.IsCycling() {
				m.inputField.PrevCompletion()
				m.completionSelector.SetIndex(m.inputField.CompletionIndex())
			}
			return m, nil
		case "esc":
			if m.completionSelector.IsActive() {
				m.completionSelector.Deactivate()
				m.inputField.ResetCompletion()
				return m, nil
			}
		case "ctrl+r":
			m.historySearch.Activate(m.historyCmds, m.inputField.Value())
			return m, nil
		case "ctrl+c":
			// If a command is running, send SIGINT, otherwise clear input
			if m.currentBlock != nil && m.currentBlock.IsRunning {
				m.shellManager.Write([]byte("\x03"))
			} else {
				m.inputField.Reset()
				m.historyIndex = -1
				m.draftCommand = ""
				m.completionSelector.Deactivate()
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
			if m.completionSelector.IsActive() {
				m.completionSelector.Deactivate()
				m.inputField.ResetCompletion()
			}
			val := m.inputField.Value()

			// If a command is running, send the input to the process (e.g. password or confirmation)
			if m.currentBlock != nil && m.currentBlock.IsRunning {
				m.shellManager.Write([]byte(val + "\n"))
				m.inputField.Reset()
				return m, nil
			}

			if val == "" {
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

			// Check for full-screen/interactive apps that need native terminal control
			if m.isFullScreenApp(val) {
				m.history.Append(val)
				m.historyCmds = append(m.historyCmds, val)
				m.inputField.Reset()

				// Get current CWD to run the command in the same place
				cwd := logic.GetShellCWD(m.shellManager.GetPID())

				cmd := exec.Command("bash", "-c", val)
				cmd.Dir = cwd

				return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
					// Add block to history when finished
					// We'll create a synthetic block for the history log
					return nil
				})
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

		// Follow output
		m.viewport.GotoBottom()

		// Dynamic detection of full-screen apps
		if strings.Contains(msgStr, "\x1b[?1049h") || strings.Contains(msgStr, "\x1b[H\x1b[2J") {
			m.currentBlock.NeedsNative = true
		}

		// Detect password prompts to mask input
		if logic.IsPasswordPrompt(m.currentBlock.Output) {
			m.inputField.SetPasswordMode(true)
		} else {
			m.inputField.SetPasswordMode(false)
		}

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
	if m.currentBlock == nil || !m.currentBlock.IsRunning {
		_, inputCmd := m.inputField.Update(finalMsg)
		if inputCmd != nil {
			cmds = append(cmds, inputCmd)
		}
	}

	// Sync completion selector
	if !m.inputField.IsCycling() {
		m.completionSelector.Deactivate()
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
		out := m.currentBlock.Output
		if !m.currentBlock.IsRunning {
			out = logic.StripEcho(out, m.currentBlock.Command)
			out = logic.StripPrompt(out)
		}
		b.WriteString(m.blockCard.Render("EXEC", m.currentBlock.Command, out, m.viewport.Width-3, time.Since(m.currentBlock.StartTime), m.currentBlock.IsRunning))
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
	m.historySearch.SetWidth(m.width)
	m.completionSelector.SetWidth(m.width)
	m.statusBar.BlocksCount = len(m.blocks)
	m.statusBar.Width = m.width
	m.statusBar.CWD = logic.GetShellCWD(m.shellManager.GetPID())
	m.statusBar.Git = logic.GetGitInfo(m.statusBar.CWD)
	m.statusBar.Project = logic.GetProjectInfo(m.statusBar.CWD)
	m.statusBar.Venv = logic.GetVenvInfo()
	m.statusBar.Time = time.Now().Format("15:04:05")
	m.statusBar.Completions = m.inputField.Completions()
	m.statusBar.CompletionIndex = m.inputField.CompletionIndex()

	// Join the viewport and scrollbar horizontally
	viewArea := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.viewport.View(),
		m.renderScrollBar(),
	)

	bottomArea := "\n" + m.inputField.View()
	if m.historySearch.IsActive() {
		bottomArea = "\n" + m.historySearch.View()
	} else if m.currentBlock != nil && m.currentBlock.IsRunning {
		// Transfer control message
		runningStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("13")). // Magenta
			Bold(true).
			Width(m.width).
			Padding(0, 1)

		msg := "⚡ RAW CONTROL TRANSFERRED (Type directly...) [Ctrl+C to break]"
		if m.currentBlock.NeedsNative {
			msg = "⚠️ FULLSCREEN APP DETECTED - UI may break. [Prefix with ! to run natively next time]"
		}

		bottomArea = "\n" + runningStyle.Render(msg)
	} else {
		m.inputField.SetLabel(" COMMAND ")
	}

	if m.completionSelector.IsActive() {

		bottomArea += "\n" + m.completionSelector.View()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.header.Render(),
		viewArea,
		bottomArea,
		"\n"+m.statusBar.Render(),
	)
}

// isFullScreenApp checks if a command is likely a full-screen TUI.
func (m *BubbleTeaApp) isFullScreenApp(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return false
	}

	// Handle explicit native execution with ! prefix
	if strings.HasPrefix(cmd, "!") {
		return true
	}

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false
	}
	binaryName := parts[0]

	// 1. Check user custom list from ~/.jgsh_fullscreen
	for _, app := range m.customFullScreen {
		if binaryName == app {
			return true
		}
	}

	// Find the actual path of the binary
	path, err := exec.LookPath(binaryName)
	if err == nil {
		// Use deep library inspection for 99% accuracy
		if logic.IsTUIBinary(path) {
			return true
		}
	}

	binary := filepath.Base(binaryName)

	// 2. Exact matches for popular apps (mostly for aliases or scripts not easily checked via ldd)
	fullScreenApps := []string{

		"nvim", "vim", "vi", "nano", "emacs", "kak", "helix", "hx",
		"htop", "top", "btop", "nmon", "glances", "gtop",
		"tmux", "screen", "less", "more", "man", "pager",
		"fzf", "vifm", "ranger", "mc", "ncdu", "dua",
		"ssh", "mosh", "irssi", "weechat", "mutt", "neomutt",
		"tig", "lazygit", "lazydocker", "gh", "ipython", "bpython",
	}

	for _, app := range fullScreenApps {
		if binary == app {
			return true
		}
	}

	// 2. Suffix-based heuristics
	suffixes := []string{"view", "edit", "top", "mon", "tui", "gui"}
	for _, s := range suffixes {
		if strings.HasSuffix(binary, s) && binary != "mon" && binary != "top" {
			return true
		}
	}

	return false
}
