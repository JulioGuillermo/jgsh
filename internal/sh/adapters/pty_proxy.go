package adapters

import (
	"github.com/julioguillermo/jgsh/internal/sh/ports"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// PTYProxy implements the ShellManager interface using creack/pty.
type PTYProxy struct {
	shellPath string
	command   *exec.Cmd
	ptyFile   *os.File
}

// NewPTYProxy creates a new instance of PTYProxy.
func NewPTYProxy(shellPath string) ports.ShellManager {
	return &PTYProxy{
		shellPath: shellPath,
	}
}

// Start launches a clean bash session without loading user RCs to ensure a predictable prompt.
func (p *PTYProxy) Start() error {
	// Using bash --norc is the most reliable way to get a clean, simple prompt
	// that we can easily strip in the TUI.
	p.command = exec.Command("/bin/bash", "--norc")

	p.command.Env = append(os.Environ(),
		"PS1=JGSH> ",
		"TERM=xterm-256color",
	)

	var err error
	p.ptyFile, err = pty.Start(p.command)
	if err != nil {
		return err
	}

	// Set a default size for the PTY
	_ = pty.Setsize(p.ptyFile, &pty.Winsize{
		Rows: 24,
		Cols: 80,
	})

	return nil
}

// Write writes data to the PTY input.
func (p *PTYProxy) Write(b []byte) (n int, err error) {
	if p.ptyFile == nil {
		return 0, io.ErrClosedPipe
	}
	return p.ptyFile.Write(b)
}

// Read reads data from the PTY output.
func (p *PTYProxy) Read(b []byte) (n int, err error) {
	if p.ptyFile == nil {
		return 0, io.ErrClosedPipe
	}
	return p.ptyFile.Read(b)
}

// GetReader returns the PTY file for reading.
func (p *PTYProxy) GetReader() io.Reader {
	return p.ptyFile
}

// GetWriter returns the PTY file for writing.
func (p *PTYProxy) GetWriter() io.Writer {
	return p.ptyFile
}

// GetPID returns the process ID of the shell.
func (p *PTYProxy) GetPID() int {
	if p.command != nil && p.command.Process != nil {
		return p.command.Process.Pid
	}
	return 0
}

// SetSize updates the PTY's internal dimensions.
func (p *PTYProxy) SetSize(rows, cols int) error {
	if p.ptyFile == nil {
		return io.ErrClosedPipe
	}
	return pty.Setsize(p.ptyFile, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}

// Stop terminates the shell session.
func (p *PTYProxy) Stop() error {
	if p.ptyFile != nil {
		p.ptyFile.Close()
	}
	if p.command != nil && p.command.Process != nil {
		return p.command.Process.Kill()
	}
	return nil
}
