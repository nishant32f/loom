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

// padLine pads a string to fill sidebarWidth with spaces
func padLine(s string, width int) string {
	visLen := lipgloss.Width(s)
	if visLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visLen)
}

// RenderSidebar renders the sidebar with groups and tabs
func RenderSidebar(groups []*Group, activeGroupIdx, activeTabIdx int, sidebarWidth, height int, focused bool, renaming bool, renameInput string) string {
	w := sidebarWidth - 2 // content width (leave room for border)
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
	bottomLines := 3 // divider + buttons + help
	for i := contentLines; i < height-bottomLines; i++ {
		lines = append(lines, strings.Repeat(" ", w))
	}

	// Divider
	divStyle := lipgloss.NewStyle().Foreground(theme.Surface2)
	lines = append(lines, divStyle.Render(strings.Repeat("─", w)))

	// Buttons
	btnStyle := lipgloss.NewStyle().Foreground(theme.Text).Background(theme.Surface1)
	buttons := " " + btnStyle.Render(" [+] tab ") + " " + btnStyle.Render(" [g] grp ")
	lines = append(lines, padLine(buttons, w))

	// Help
	helpStyle := lipgloss.NewStyle().Foreground(theme.Overlay0)
	lines = append(lines, padLine(helpStyle.Render(" r:rename n:new ↑↓:nav"), w))

	content := strings.Join(lines, "\n")

	borderColor := theme.Surface2
	if focused {
		borderColor = theme.Lavender
	}

	container := lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(height).
		Background(theme.Mantle).
		Foreground(theme.Text).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderRightForeground(borderColor)

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
