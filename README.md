# 🚀 JGSH: The Modern Block Terminal

JGSH (project name Glock) is a modern TUI-based terminal wrapper written in Go. It intercepts shell sessions to organize command output into interactive, isolated "Blocks," providing IDE-grade syntax highlighting and a refined multi-line input experience.

## ✨ Key Features

- **📦 Block-Based Output**: Every command and its output are grouped into distinct, visually isolated cards.
- **🎨 IDE-Grade Syntax Highlighting**: Real-time Bash syntax highlighting powered by **Tree-sitter**.
- **⌨️ Multi-line Command Support**: Press `Shift+Enter` to insert newlines and write complex scripts directly in the prompt.
- **🔍 Smart Autocompletion**:
    - Context-aware suggestions (files, history, commands).
    - Navigate suggestions with `Up`/`Down` arrows.
    - Select with `Enter` or `Tab`.
- **📜 Persistent History**: searchable history across sessions with `Ctrl+R`.
- **⚡ Raw Mode Passthrough**: Automatic detection of full-screen TUI apps (vim, htop, etc.) for seamless native interaction.
- **🛡️ Secure Input**: Intelligent password prompt detection with automatic input masking.
- **🌀 Modern UI**: Built with the **Charm Bracelet** stack (Bubble Tea, Lipgloss, Bubbles).

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.25 or higher.
- A terminal with ANSI color support.

### Building from Source

```bash
# Clone the repository
git clone https://github.com/julioguillermo/jgsh.git
cd jgsh

# Build the binary
go build -o jgsh cmd/glock/main.go

# Run JGSH
./jgsh
```

## ⌨️ Key Bindings

| Key | Action |
|-----|--------|
| `Enter` | Execute command / Select autocomplete item |
| `Shift+Enter` | Insert newline (multi-line mode) |
| `Tab` | Open autocomplete / Next suggestion |
| `Shift+Tab` | Previous suggestion |
| `Up` / `Down` | Navigate history / Navigate autocomplete items |
| `Ctrl+R` | Open history search |
| `Ctrl+C` | Cancel current input / Interrupt running process |
| `Esc` | Close autocomplete / history search |

## 🏗️ Architecture

JGSH follows a **Modular Hexagonal Architecture** (Ports and Adapters) to ensure high decoupling and testability:

- **`internal/sh`**: PTY management and shell interaction using `creack/pty`.
- **`internal/ui`**: TUI implementation using `charmbracelet/bubbletea`.
- **`internal/syntax`**: Syntax highlighting logic using `tree-sitter`.
- **`internal/completion`**: Autocompletion engine and history management.

## 🛠️ Tech Stack

- **UI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Parser**: [Tree-sitter](https://github.com/tree-sitter/go-tree-sitter)
- **Terminal Control**: [PTY](https://github.com/creack/pty)

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
