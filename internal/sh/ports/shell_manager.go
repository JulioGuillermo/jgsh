package ports

import "io"

// ShellManager defines the interface for managing a PTY shell session.
type ShellManager interface {
	Start() error
	Write(p []byte) (n int, err error)
	Read(p []byte) (n int, err error)
	GetReader() io.Reader
	GetWriter() io.Writer
	GetPID() int
	SetSize(rows, cols int) error
	Stop() error
}
