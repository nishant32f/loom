package model

import "github.com/charmbracelet/lipgloss"

// Tab represents a single terminal tab
type Tab struct {
	Name          string
	HoldingWindow int // -1 = active (pane is at :0.1), N = stored in window N
	Command       string
	Cwd           string
	GroupName     string
	GitRepo       string // detected git repo name (empty if not a git repo)
	GitBranch     string // detected git branch
}

// IsActive returns true if the tab's pane is currently in the main terminal pane
func (t *Tab) IsActive() bool {
	return t.HoldingWindow == -1
}

// Group represents a collapsible group of tabs
type Group struct {
	Name      string
	Color     lipgloss.Color
	Tabs      []*Tab
	Collapsed bool
}
