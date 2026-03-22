package logic

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetShellCWD attempts to find the current working directory of the given PID.
func GetShellCWD(pid int) string {
	if pid <= 0 {
		return ""
	}
	// On Linux, we can read /proc/[pid]/cwd
	cwd, err := os.Readlink(filepath.Join("/proc", string(rune(pid)), "cwd"))
	if err == nil {
		return cwd
	}
	// Fallback to current process dir if /proc fails
	d, _ := os.Getwd()
	return d
}

// GetGitInfo returns the current git branch name.
func GetGitInfo(cwd string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
