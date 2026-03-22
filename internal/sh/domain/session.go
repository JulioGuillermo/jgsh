package domain

import "time"

// Session tracks the shell session state.
type Session struct {
	PID       int
	StartTime time.Time
	Cwd       string
	Shell     string
}
