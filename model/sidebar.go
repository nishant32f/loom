package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nishant32f/loom/theme"
)

// Sidebar item type constants
const (
	itemSpacer      = "spacer"
	itemGroupHeader = "group_header"
	itemTab         = "tab"
)

// SidebarItem represents a clickable item in the sidebar
type SidebarItem struct {
	Type     string
	GroupIdx int
	TabIdx   int
}

// Cached styles (created once, not per render)
var (
	logoStyle      = lipgloss.NewStyle().Bold(true).Foreground(theme.Mauve)
	activeTabDot   = lipgloss.NewStyle().Foreground(theme.Green).Bold(true)
	activeTabName  = lipgloss.NewStyle().Foreground(theme.Text).Bold(true)
	inactiveTab    = lipgloss.NewStyle().Foreground(theme.Subtext1)
	renameStyle    = lipgloss.NewStyle().Foreground(theme.Yellow)
	dividerStyle   = lipgloss.NewStyle().Foreground(theme.Surface2)
	buttonStyle    = lipgloss.NewStyle().Foreground(theme.Text).Background(theme.Surface1)
	helpStyle      = lipgloss.NewStyle().Foreground(theme.Overlay0)
	navHintStyle   = lipgloss.NewStyle().Foreground(theme.Overlay1).Italic(true)
	statusStyle    = lipgloss.NewStyle().Foreground(theme.Subtext0)
	statusMsgStyle = lipgloss.NewStyle().Foreground(theme.Yellow)
)

// padLine pads a string to fill width with spaces
func padLine(s string, width int) string {
	visLen := lipgloss.Width(s)
	if visLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visLen)
}

// RenderSidebar renders the sidebar at full pane width
func RenderSidebar(groups []*Group, activeGroupIdx, activeTabIdx int, width, height int, renaming bool, renameInput string, statusMsg string) string {
	w := width - 1
	if w < 10 {
		w = 10
	}
	emptyLine := strings.Repeat(" ", w)
	var lines []string

	lines = append(lines, padLine(logoStyle.Render(" ◈ Loom"), w))
	lines = append(lines, emptyLine)

	for gi, group := range groups {
		arrow := "▼"
		if group.Collapsed {
			arrow = "▶"
		}

		groupStyle := lipgloss.NewStyle().Bold(true).Foreground(group.Color)
		lines = append(lines, padLine(groupStyle.Render(fmt.Sprintf(" %s %s", arrow, strings.ToUpper(group.Name))), w))

		if !group.Collapsed {
			connStyle := lipgloss.NewStyle().Foreground(group.Color)
			for ti, tab := range group.Tabs {
				isActive := gi == activeGroupIdx && ti == activeTabIdx

				connector := "┣━"
				if ti == len(group.Tabs)-1 {
					connector = "┗━"
				}
				prefix := connStyle.Render(" " + connector)

				var line string
				if isActive && renaming {
					line = prefix + renameStyle.Render(" "+renameInput+"█")
				} else if isActive {
					line = prefix + activeTabDot.Render(" ●") + activeTabName.Render(" "+tab.Name)
				} else {
					line = prefix + inactiveTab.Render("  "+tab.Name)
				}
				lines = append(lines, padLine(line, w))
			}
		}

		lines = append(lines, emptyLine)
	}

	// Fill remaining space
	bottomLines := 5
	for i := len(lines); i < height-bottomLines; i++ {
		lines = append(lines, emptyLine)
	}

	lines = append(lines, dividerStyle.Render(strings.Repeat("─", w)))
	lines = append(lines, padLine(" "+buttonStyle.Render(" [n] tab ")+" "+buttonStyle.Render(" [g] grp "), w))
	lines = append(lines, padLine(helpStyle.Render(" r:rename d:close ↑↓:nav"), w))
	lines = append(lines, padLine(navHintStyle.Render(" Ctrl+B ← to navigate here"), w))

	if statusMsg != "" {
		lines = append(lines, padLine(statusMsgStyle.Render(" "+statusMsg), w))
	} else {
		tabCount := 0
		for _, g := range groups {
			tabCount += len(g.Tabs)
		}
		lines = append(lines, padLine(statusStyle.Render(fmt.Sprintf(" %d groups │ %d tabs", len(groups), tabCount)), w))
	}

	container := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Background(theme.Mantle).
		Foreground(theme.Text)

	return container.Render(strings.Join(lines, "\n"))
}

// GetSidebarItems builds the list of clickable items for mouse interaction
func GetSidebarItems(groups []*Group) []SidebarItem {
	var items []SidebarItem

	items = append(items, SidebarItem{Type: itemSpacer}, SidebarItem{Type: itemSpacer})

	for gi, group := range groups {
		items = append(items, SidebarItem{Type: itemGroupHeader, GroupIdx: gi})

		if !group.Collapsed {
			for ti := range group.Tabs {
				items = append(items, SidebarItem{Type: itemTab, GroupIdx: gi, TabIdx: ti})
			}
		}
		items = append(items, SidebarItem{Type: itemSpacer})
	}

	return items
}
