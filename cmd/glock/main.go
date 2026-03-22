package main

import (
	"fmt"
	"os"

	shadapters "github.com/julioguillermo/jgsh/internal/sh/adapters"
	syntaxadapters "github.com/julioguillermo/jgsh/internal/syntax/adapters"
	uiadapters "github.com/julioguillermo/jgsh/internal/ui/adapters"
	"github.com/julioguillermo/jgsh/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Create the PTY Proxy (Shell Manager)
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = "/bin/bash"
	}

	shellProxy := shadapters.NewPTYProxy(shellPath)
	if err := shellProxy.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting shell: %v\n", err)
		os.Exit(1)
	}
	defer shellProxy.Stop()

	// Create the Syntax Highlighter
	highlighter, err := syntaxadapters.NewTSBashHighlighter()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating highlighter: %v\n", err)
		os.Exit(1)
	}

	// Create the Input Component
	inputField := components.NewInputField(highlighter)

	// Create the Bubble Tea TUI
	app := uiadapters.NewBubbleTeaApp(shellProxy, inputField, highlighter)
	p := tea.NewProgram(app, tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
