package logic

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitStatus represents the detailed status of a git repository.
type GitStatus struct {
	Branch     string
	Staged     int
	Modified   int
	Untracked  int
	Ahead      int
	Behind     int
	Insertions int
	Deletions  int
}

// GetShellCWD attempts to find the current working directory of the given PID.
func GetShellCWD(pid int) string {
	if pid <= 0 {
		return ""
	}
	// On Linux, we can read /proc/[pid]/cwd
	// We use fmt.Sprint instead of string(rune(pid)) because string(rune) converts to char
	cwd, err := os.Readlink(filepath.Join("/proc", fmt.Sprint(pid), "cwd"))
	if err == nil {
		return cwd
	}
	// Fallback to current process dir if /proc fails
	d, _ := os.Getwd()
	return d
}

// GetGitInfo returns the detailed git status.
func GetGitInfo(cwd string) GitStatus {
	var status GitStatus

	// Get branch name
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return status
	}
	status.Branch = strings.TrimSpace(string(out))

	// Get status counts
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = cwd
	out, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if len(line) < 2 {
				continue
			}
			st := line[0]
			mo := line[1]

			if st == 'M' || st == 'A' || st == 'D' || st == 'R' || st == 'C' {
				status.Staged++
			}
			if mo == 'M' || mo == 'D' {
				status.Modified++
			}
			if st == '?' {
				status.Untracked++
			}
		}
	}

	// Get ahead/behind
	cmd = exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
	cmd.Dir = cwd
	out, err = cmd.Output()
	if err == nil {
		parts := strings.Fields(string(out))
		if len(parts) >= 2 {
			fmt.Sscanf(parts[0], "%d", &status.Ahead)
			fmt.Sscanf(parts[1], "%d", &status.Behind)
		}
	}

	// Get insertions/deletions
	cmd = exec.Command("git", "diff", "--shortstat")
	cmd.Dir = cwd
	out, err = cmd.Output()
	if err == nil {
		str := string(out)
		// Example: 1 file changed, 5 insertions(+), 3 deletions(-)
		parts := strings.Split(str, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.Contains(part, "insertion") {
				fmt.Sscanf(part, "%d", &status.Insertions)
			}
			if strings.Contains(part, "deletion") {
				fmt.Sscanf(part, "%d", &status.Deletions)
			}
		}
	}

	return status
}

// GetVenvInfo returns the name of the active virtual environment if any.
func GetVenvInfo() string {
	venv := os.Getenv("VIRTUAL_ENV")
	if venv != "" {
		return filepath.Base(venv)
	}
	conda := os.Getenv("CONDA_DEFAULT_ENV")
	if conda != "" {
		return conda
	}
	return ""
}

// IsTUIBinary checks if a binary is likely a TUI application by inspecting its linked libraries.
func IsTUIBinary(path string) bool {
	if path == "" {
		return false
	}

	// Use ldd to check for ncurses, tinfo, or slang dependencies
	// This is very reliable for detecting TUI apps on Linux.
	cmd := exec.Command("ldd", path)
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	s := strings.ToLower(string(out))
	tuiLibs := []string{"ncurses", "tinfo", "slang", "readline", "libncurses", "libtinfo"}
	for _, lib := range tuiLibs {
		if strings.Contains(s, lib) {
			return true
		}
	}

	return false
}

// GetProjectInfo detects the project type based on files in the directory.
func GetProjectInfo(cwd string) string {
	// Common project indicators
	indicators := map[string]string{
		"go.mod":           "Go",
		"package.json":     "Node.js",
		"requirements.txt": "Python",
		"pyproject.toml":   "Python",
		"Cargo.toml":       "Rust",
		"Makefile":         "Make",
		"build.gradle":     "Java/Gradle",
		"pom.xml":          "Java/Maven",
		"composer.json":    "PHP",
		"Gemfile":          "Ruby",
	}

	for file, lang := range indicators {
		if _, err := os.Stat(filepath.Join(cwd, file)); err == nil {
			return lang
		}
	}

	return ""
}
