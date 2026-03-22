package logic

import (
	"testing"
)

func TestStripAnsiRealShellOutput(t *testing.T) {
	// Simulate what ls -la might output (simplified)
	realOutput := "\x1b[0m\x1b[38;5;75m\x1b[48;5;234mtotal 12\x1b[0m\ndrwxr-xr-x  2 user user 4096 Jan 1 12:00 .\ndrwxr-xr-x  2 user user 4096 Jan 1 12:00 .."

	result := StripAnsi(realOutput)
	t.Logf("Input len: %d, Output len: %d", len(realOutput), len(result))
	t.Logf("Result: %q", result)

	if len(result) == 0 {
		t.Error("StripAnsi removed all content")
	}
}
