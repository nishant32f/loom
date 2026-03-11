package model

import "github.com/charmbracelet/lipgloss"

// Tab represents a single terminal tab
type Tab struct {
	Name          string
	HoldingWindow int // -1 = active (pane is at :0.1), N = stored in window N
	Command       string
	Cwd           string
	GroupName     string
}

// Group represents a collapsible group of tabs
type Group struct {
	Name      string
	Color     lipgloss.Color
	Tabs      []*Tab
	Collapsed bool
}

// VisibleTabs returns tabs only if the group is not collapsed
func (g *Group) VisibleTabs() []*Tab {
	if g.Collapsed {
		return nil
	}
	return g.Tabs
}

// AddTab adds a tab to the group
func (g *Group) AddTab(tab *Tab) {
	tab.GroupName = g.Name
	g.Tabs = append(g.Tabs, tab)
}

// RemoveTab removes a tab by index
func (g *Group) RemoveTab(index int) *Tab {
	if index < 0 || index >= len(g.Tabs) {
		return nil
	}
	tab := g.Tabs[index]
	g.Tabs = append(g.Tabs[:index], g.Tabs[index+1:]...)
	return tab
}
