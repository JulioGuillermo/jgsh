package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
)

func main() {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	cmd := exec.Command(shell)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		fmt.Printf("Error starting PTY: %v\n", err)
		return
	}
	defer ptmx.Close()
	defer cmd.Process.Kill()

	_ = pty.Setsize(ptmx, &pty.Winsize{Rows: 24, Cols: 80})

	fmt.Println("Shell started, waiting for initial prompt...")
	time.Sleep(500 * time.Millisecond)

	// Read initial output
	readAll(ptmx)

	fmt.Println("\n--- Sending 'ls -la' ---")
	io.WriteString(ptmx, "ls -la\n")
	time.Sleep(500 * time.Millisecond)
	output := readAll(ptmx)
	fmt.Printf("Output after 'ls -la' (len=%d):\n%s\n", len(output), output)

	fmt.Println("\n--- Sending 'echo hello' ---")
	io.WriteString(ptmx, "echo hello\n")
	time.Sleep(500 * time.Millisecond)
	output = readAll(ptmx)
	fmt.Printf("Output after 'echo hello' (len=%d):\n%s\n", len(output), output)

	fmt.Println("\n--- Test complete ---")
}

func readAll(r io.Reader) string {
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	return string(buf[:n])
}
