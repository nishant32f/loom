package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// CreateSession creates a new detached tmux session
func CreateSession(name string) error {
	cmd := exec.Command("tmux", "new-session", "-d", "-s", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}
	// Configure session
	run("tmux", "set", "-t", name, "remain-on-exit", "off")
	run("tmux", "set", "-t", name, "mouse", "on")
	run("tmux", "set", "-t", name, "pane-border-style", "fg=colour238")
	run("tmux", "set", "-t", name, "status", "off")
	run("tmux", "set", "-t", name, "base-index", "0")
	run("tmux", "set", "-t", name, "pane-base-index", "0")
	// Renumber the existing window to 0 if it started at 1
	run("tmux", "move-window", "-s", name+":1", "-t", name+":0")
	return nil
}

// SplitForSidebar splits window 0 with a left pane of the given width for the sidebar.
// After this call, :0.0 is left (sidebar), :0.1 is right (terminal).
func SplitForSidebar(session string, width int) error {
	target := session + ":0"
	// Split horizontally, new pane on the left (-b), with fixed column width
	cmd := exec.Command("tmux", "split-window", "-h", "-b", "-l", fmt.Sprintf("%d", width), "-t", target)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to split for sidebar: %w", err)
	}
	// Ensure the sidebar pane stays at the requested width
	sidebarPane := session + ":0.0"
	exec.Command("tmux", "resize-pane", "-t", sidebarPane, "-x", fmt.Sprintf("%d", width)).Run()
	return nil
}

// CreateWindow creates a new tmux window and returns its index
func CreateWindow(session string, name string) (int, error) {
	// Create window and print its index
	cmd := exec.Command("tmux", "new-window", "-d", "-t", session, "-n", name, "-P", "-F", "#{window_index}")
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to create window: %w", err)
	}
	idx, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse window index: %w", err)
	}
	return idx, nil
}

// SwapPane swaps two panes. source and target are like "session:0.1" and "session:2.0"
func SwapPane(session, source, target string) error {
	cmd := exec.Command("tmux", "swap-pane", "-s", source, "-t", target)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to swap pane: %w", err)
	}
	return nil
}

// SelectPane focuses a pane (e.g. "session:0.1")
func SelectPane(session, pane string) error {
	cmd := exec.Command("tmux", "select-pane", "-t", pane)
	return cmd.Run()
}

// RunInPane sends a command string to a pane
func RunInPane(session, pane, command string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", pane, command, "Enter")
	return cmd.Run()
}

// KillWindow kills a tmux window by index
func KillWindow(session string, windowIdx int) error {
	target := fmt.Sprintf("%s:%d", session, windowIdx)
	cmd := exec.Command("tmux", "kill-window", "-t", target)
	return cmd.Run()
}

// KillSession kills the entire tmux session
func KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	return cmd.Run()
}

// RenameWindow renames a tmux window
func RenameWindow(session string, windowIdx int, name string) error {
	target := fmt.Sprintf("%s:%d", session, windowIdx)
	cmd := exec.Command("tmux", "rename-window", "-t", target, name)
	return cmd.Run()
}

// Attach exec's into tmux attach-session (replaces the current process)
func Attach(session string) {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		fmt.Fprintf(os.Stderr, "tmux not found: %v\n", err)
		os.Exit(1)
	}
	syscall.Exec(tmuxPath, []string{"tmux", "attach-session", "-t", session}, os.Environ())
}

// SwitchClient switches the current tmux client to the given session
func SwitchClient(session string) error {
	cmd := exec.Command("tmux", "switch-client", "-t", session)
	return cmd.Run()
}

// IsTmuxAvailable checks if tmux is installed
func IsTmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// IsInsideTmux returns true if we're already inside a tmux session
func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

// SelectWindow selects a window (used internally)
func SelectWindow(session string, windowIdx int) error {
	target := fmt.Sprintf("%s:%d", session, windowIdx)
	cmd := exec.Command("tmux", "select-window", "-t", target)
	return cmd.Run()
}

// run is a helper that runs a command silently
func run(name string, args ...string) {
	exec.Command(name, args...).Run()
}
