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
	activeTabDot   = lipgloss.NewStyle().Foreground(theme.Green).Bold(true).Background(theme.Surface0)
	activeTabName  = lipgloss.NewStyle().Foreground(theme.Text).Bold(true).Background(theme.Surface0)
	activeBranch   = lipgloss.NewStyle().Foreground(theme.Teal).Background(theme.Surface0)
	inactiveTab    = lipgloss.NewStyle().Foreground(theme.Subtext1)
	inactiveBranch = lipgloss.NewStyle().Foreground(theme.Overlay0)
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

// buildSidebarContent builds both the rendered lines and the click item map in a single pass
func buildSidebarContent(groups []*Group, activeGroupIdx, activeTabIdx int, w int, renaming bool, renameInput string) ([]string, []SidebarItem) {
	emptyLine := strings.Repeat(" ", w)
	var lines []string
	var items []SidebarItem

	// Logo + empty line
	lines = append(lines, padLine(logoStyle.Render(" ◈ Loom"), w))
	items = append(items, SidebarItem{Type: itemSpacer})
	lines = append(lines, emptyLine)
	items = append(items, SidebarItem{Type: itemSpacer})

	for gi, group := range groups {
		arrow := "▼"
		if group.Collapsed {
			arrow = "▶"
		}

		groupStyle := lipgloss.NewStyle().Bold(true).Foreground(group.Color)
		lines = append(lines, padLine(groupStyle.Render(fmt.Sprintf(" %s %s", arrow, strings.ToUpper(group.Name))), w))
		items = append(items, SidebarItem{Type: itemGroupHeader, GroupIdx: gi})

		if !group.Collapsed {
			connStyle := lipgloss.NewStyle().Foreground(group.Color)
			for ti, tab := range group.Tabs {
				isActive := gi == activeGroupIdx && ti == activeTabIdx

				connector := "┣━"
				if ti == len(group.Tabs)-1 {
					connector = "┗━"
				}
				prefix := connStyle.Render(" " + connector)

				// Build display name with git info
				displayName := tab.Name
				branchInfo := ""
				if tab.GitRepo != "" {
					displayName = tab.GitRepo
					if tab.GitBranch != "" {
						branchInfo = tab.GitBranch
					}
				}

				// Truncate display name if too long for sidebar
				maxNameLen := w - 6
				if len(displayName) > maxNameLen && maxNameLen > 3 {
					displayName = displayName[:maxNameLen-1] + "…"
				}

				var line string
				if isActive && renaming {
					line = prefix + renameStyle.Render(" "+renameInput+"█")
				} else if isActive {
					tabContent := activeTabDot.Render(" ●") + activeTabName.Render(" "+displayName)
					contentLen := lipgloss.Width(tabContent)
					remaining := w - lipgloss.Width(prefix) - contentLen
					if remaining > 0 {
						tabContent += activeTabName.Render(strings.Repeat(" ", remaining))
					}
					line = prefix + tabContent
				} else {
					line = prefix + inactiveTab.Render("  "+displayName)
				}
				lines = append(lines, padLine(line, w))
				items = append(items, SidebarItem{Type: itemTab, GroupIdx: gi, TabIdx: ti})

				// Branch on second line if present
				if branchInfo != "" {
					branchPrefix := "     "
					maxBranchLen := w - len(branchPrefix) - 2
					if len(branchInfo) > maxBranchLen && maxBranchLen > 3 {
						branchInfo = branchInfo[:maxBranchLen-1] + "…"
					}
					if isActive {
						branchLine := activeTabName.Render(branchPrefix) + activeBranch.Render("↳ "+branchInfo)
						bl := lipgloss.Width(branchLine)
						rem := w - bl
						if rem > 0 {
							branchLine += activeTabName.Render(strings.Repeat(" ", rem))
						}
						lines = append(lines, padLine(branchLine, w))
					} else {
						lines = append(lines, padLine(inactiveBranch.Render(branchPrefix+"↳ "+branchInfo), w))
					}
					// Clicking on branch line selects the same tab
					items = append(items, SidebarItem{Type: itemTab, GroupIdx: gi, TabIdx: ti})
				}
			}
		}

		lines = append(lines, emptyLine)
		items = append(items, SidebarItem{Type: itemSpacer})
	}

	return lines, items
}

// RenderSidebar renders the sidebar at full pane width
func RenderSidebar(groups []*Group, activeGroupIdx, activeTabIdx int, width, height int, renaming bool, renameInput string, statusMsg string) string {
	w := width - 1
	if w < 10 {
		w = 10
	}
	emptyLine := strings.Repeat(" ", w)

	lines, _ := buildSidebarContent(groups, activeGroupIdx, activeTabIdx, w, renaming, renameInput)

	// Fill remaining space
	bottomLines := 5
	for i := len(lines); i < height-bottomLines; i++ {
		lines = append(lines, emptyLine)
	}

	lines = append(lines, dividerStyle.Render(strings.Repeat("─", w)))
	lines = append(lines, padLine(" "+buttonStyle.Render(" [n] tab ")+" "+buttonStyle.Render(" [g] grp "), w))
	lines = append(lines, padLine(helpStyle.Render(" r:rename d:close ↑↓:nav"), w))
	lines = append(lines, padLine(navHintStyle.Render(" Ctrl+B ← → switch panes"), w))

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

// GetSidebarItems builds the list of clickable items for mouse interaction.
// Uses the same buildSidebarContent function as the renderer to stay in sync.
func GetSidebarItems(groups []*Group, activeGroupIdx, activeTabIdx int, width int, renaming bool, renameInput string) []SidebarItem {
	w := width - 1
	if w < 10 {
		w = 10
	}
	_, items := buildSidebarContent(groups, activeGroupIdx, activeTabIdx, w, renaming, renameInput)
	return items
}
