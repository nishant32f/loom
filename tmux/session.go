package tmux

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Session manages tmux sessions for Loom
type Session struct {
	Name    string
	Windows []Window
}

// Window represents a tmux window (maps to a Loom tab)
type Window struct {
	ID   string
	Name string
}

// sessionPrefix is used to namespace loom tmux sessions
const sessionPrefix = "loom_"

// NewSession creates a new tmux session
func NewSession() (*Session, error) {
	name := fmt.Sprintf("%s%d", sessionPrefix, time.Now().UnixNano())

	cmd := exec.Command("tmux", "new-session", "-d", "-s", name, "-x", "200", "-y", "50")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create tmux session: %w", err)
	}

	s := &Session{
		Name: name,
		Windows: []Window{
			{ID: name + ":0", Name: "shell"},
		},
	}
	return s, nil
}

// NewWindow creates a new window in the session
func (s *Session) NewWindow(name string, command string, cwd string) (*Window, error) {
	args := []string{"new-window", "-t", s.Name, "-n", name}
	if cwd != "" {
		args = append(args, "-c", expandHome(cwd))
	}

	cmd := exec.Command("tmux", args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}

	// Get the window ID
	winID := fmt.Sprintf("%s:%d", s.Name, len(s.Windows))
	w := Window{ID: winID, Name: name}
	s.Windows = append(s.Windows, w)

	if command != "" {
		SendKeys(winID, command)
	}

	return &w, nil
}

// SendKeys sends keystrokes to a tmux pane
func SendKeys(target string, keys string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, keys, "Enter")
	return cmd.Run()
}

// CapturePaneContent captures the visible content of a tmux pane
func CapturePaneContent(target string, width, height int) (string, error) {
	// Resize the pane to match our viewport
	resizeCmd := exec.Command("tmux", "resize-window", "-t", target, "-x", fmt.Sprintf("%d", width), "-y", fmt.Sprintf("%d", height))
	_ = resizeCmd.Run()

	cmd := exec.Command("tmux", "capture-pane", "-t", target, "-p", "-e")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to capture pane: %w", err)
	}
	return string(out), nil
}

// SplitWindow splits the current window
func SplitWindow(target string, horizontal bool, command string) error {
	direction := "-v"
	if horizontal {
		direction = "-h"
	}

	args := []string{"split-window", direction, "-t", target}
	cmd := exec.Command("tmux", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to split window: %w", err)
	}

	if command != "" {
		SendKeys(target, command)
	}
	return nil
}

// KillWindow kills a tmux window
func KillWindow(target string) error {
	cmd := exec.Command("tmux", "kill-window", "-t", target)
	return cmd.Run()
}

// KillSession kills the entire tmux session
func (s *Session) Kill() error {
	cmd := exec.Command("tmux", "kill-session", "-t", s.Name)
	return cmd.Run()
}

// RenameWindow renames a tmux window
func RenameWindow(target string, newName string) error {
	cmd := exec.Command("tmux", "rename-window", "-t", target, newName)
	return cmd.Run()
}

// SelectWindow switches to a specific window
func SelectWindow(target string) error {
	cmd := exec.Command("tmux", "select-window", "-t", target)
	return cmd.Run()
}

// IsTmuxAvailable checks if tmux is installed
func IsTmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// expandHome expands ~ to the home directory
func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := exec.Command("sh", "-c", "echo $HOME").Output()
		return strings.TrimSpace(string(home)) + path[1:]
	}
	return path
}

// SendRawKeys sends raw key input to a pane (for terminal interaction)
func SendRawKeys(target string, keys string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, "-l", keys)
	return cmd.Run()
}

// SendSpecialKey sends a special key (like Enter, Backspace, etc.)
func SendSpecialKey(target string, key string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", target, key)
	return cmd.Run()
}
