package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nishant32f/loom/theme"
)

// SidebarItem represents a clickable item in the sidebar
type SidebarItem struct {
	Type     string // "group_header", "tab", "button", "spacer"
	GroupIdx int
	TabIdx   int
}

// padLine pads a string to fill width with spaces
func padLine(s string, width int) string {
	visLen := lipgloss.Width(s)
	if visLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visLen)
}

// RenderSidebar renders the sidebar at full pane width (no right border — tmux provides it)
func RenderSidebar(groups []*Group, activeGroupIdx, activeTabIdx int, width, height int, renaming bool, renameInput string, statusMsg string) string {
	w := width - 1 // small right margin
	if w < 10 {
		w = 10
	}
	var lines []string

	// Logo line
	logo := lipgloss.NewStyle().Bold(true).Foreground(theme.Mauve).Render(" ◈ Loom")
	lines = append(lines, padLine(logo, w))

	// Empty separator
	lines = append(lines, strings.Repeat(" ", w))

	for gi, group := range groups {
		// Group header
		arrow := "▼"
		if group.Collapsed {
			arrow = "▶"
		}

		groupStyle := lipgloss.NewStyle().Bold(true).Foreground(group.Color)
		header := groupStyle.Render(fmt.Sprintf(" %s %s", arrow, strings.ToUpper(group.Name)))
		lines = append(lines, padLine(header, w))

		if !group.Collapsed {
			for ti, tab := range group.Tabs {
				isActive := gi == activeGroupIdx && ti == activeTabIdx

				connector := "┣━"
				if ti == len(group.Tabs)-1 {
					connector = "┗━"
				}

				connStyle := lipgloss.NewStyle().Foreground(group.Color)
				prefix := connStyle.Render(" " + connector)

				if isActive {
					if renaming {
						inputStyle := lipgloss.NewStyle().Foreground(theme.Yellow)
						line := prefix + inputStyle.Render(" "+renameInput+"█")
						lines = append(lines, padLine(line, w))
					} else {
						dotStyle := lipgloss.NewStyle().Foreground(theme.Green).Bold(true)
						nameStyle := lipgloss.NewStyle().Foreground(theme.Text).Bold(true)
						line := prefix + dotStyle.Render(" ●") + nameStyle.Render(" "+tab.Name)
						lines = append(lines, padLine(line, w))
					}
				} else {
					nameStyle := lipgloss.NewStyle().Foreground(theme.Subtext1)
					line := prefix + nameStyle.Render("  "+tab.Name)
					lines = append(lines, padLine(line, w))
				}
			}
		}

		// Spacer after group
		lines = append(lines, strings.Repeat(" ", w))
	}

	// Fill remaining space
	contentLines := len(lines)
	bottomLines := 5 // divider + buttons + help + nav hint + status
	for i := contentLines; i < height-bottomLines; i++ {
		lines = append(lines, strings.Repeat(" ", w))
	}

	// Divider
	divStyle := lipgloss.NewStyle().Foreground(theme.Surface2)
	lines = append(lines, divStyle.Render(strings.Repeat("─", w)))

	// Buttons
	btnStyle := lipgloss.NewStyle().Foreground(theme.Text).Background(theme.Surface1)
	buttons := " " + btnStyle.Render(" [n] tab ") + " " + btnStyle.Render(" [g] grp ")
	lines = append(lines, padLine(buttons, w))

	// Help
	helpStyle := lipgloss.NewStyle().Foreground(theme.Overlay0)
	lines = append(lines, padLine(helpStyle.Render(" r:rename d:close ↑↓:nav"), w))

	// Navigation hint
	navStyle := lipgloss.NewStyle().Foreground(theme.Overlay1).Italic(true)
	lines = append(lines, padLine(navStyle.Render(" Ctrl+B ← to navigate here"), w))

	// Status line
	if statusMsg != "" {
		statStyle := lipgloss.NewStyle().Foreground(theme.Yellow)
		lines = append(lines, padLine(statStyle.Render(" "+statusMsg), w))
	} else {
		tabCount := 0
		for _, g := range groups {
			tabCount += len(g.Tabs)
		}
		statStyle := lipgloss.NewStyle().Foreground(theme.Subtext0)
		lines = append(lines, padLine(statStyle.Render(fmt.Sprintf(" %d groups │ %d tabs", len(groups), tabCount)), w))
	}

	content := strings.Join(lines, "\n")

	container := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Background(theme.Mantle).
		Foreground(theme.Text)

	return container.Render(content)
}

// GetSidebarItems builds the list of clickable items for mouse interaction
func GetSidebarItems(groups []*Group) []SidebarItem {
	var items []SidebarItem

	// Logo line + empty line
	items = append(items, SidebarItem{Type: "spacer"})
	items = append(items, SidebarItem{Type: "spacer"})

	for gi, group := range groups {
		items = append(items, SidebarItem{
			Type:     "group_header",
			GroupIdx: gi,
		})

		if !group.Collapsed {
			for ti := range group.Tabs {
				items = append(items, SidebarItem{
					Type:     "tab",
					GroupIdx: gi,
					TabIdx:   ti,
				})
			}
		}
		items = append(items, SidebarItem{Type: "spacer"})
	}

	return items
}
