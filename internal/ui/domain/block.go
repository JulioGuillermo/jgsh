package domain

import "time"

// Block represents a single command execution and its output.
type Block struct {
	ID        string
	Command   string
	Output    string
	StartTime time.Time
	Duration  time.Duration
	ExitCode  int
	Finished  bool
	IsRunning bool
}
