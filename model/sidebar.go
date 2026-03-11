package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nishant32f/loom/theme"
)

// SidebarItem represents a clickable item in the sidebar
type SidebarItem struct {
	Type     string // "group_header", "tab", "button"
	GroupIdx int
	TabIdx   int
	Label    string
}

// RenderSidebar renders the sidebar with groups and tabs
func RenderSidebar(groups []*Group, activeGroupIdx, activeTabIdx int, sidebarWidth, height int, focused bool, renaming bool, renameInput string) string {
	var lines []string

	// Logo
	logo := theme.Logo()
	lines = append(lines, logo)
	lines = append(lines, "")

	for gi, group := range groups {
		// Group header
		arrow := "▼"
		if group.Collapsed {
			arrow = "▶"
		}

		groupColor := group.Color
		headerStyle := theme.GroupHeaderStyle.Foreground(groupColor)
		header := headerStyle.Render(fmt.Sprintf("%s %s", arrow, strings.ToUpper(group.Name)))
		lines = append(lines, header)

		if !group.Collapsed {
			for ti, tab := range group.Tabs {
				isActive := gi == activeGroupIdx && ti == activeTabIdx

				// Build the tree connector
				connector := " ┣━"
				if ti == len(group.Tabs)-1 {
					connector = " ┗━"
				}

				connectorStyle := lipgloss.NewStyle().Foreground(groupColor).Background(theme.Mantle)

				if isActive {
					if renaming {
						// Show rename input
						label := connectorStyle.Render(connector) +
							theme.RenameInputStyle.Render(" "+renameInput+"█")
						lines = append(lines, label)
					} else {
						indicator := theme.TabIndicator.Render("●")
						label := connectorStyle.Render(connector) +
							theme.ActiveTabStyle.Render(" "+indicator+" "+tab.Name)
						lines = append(lines, label)
					}
				} else {
					label := connectorStyle.Render(connector) +
						theme.TabStyle.Render("  "+tab.Name)
					lines = append(lines, label)
				}
			}
		}
		lines = append(lines, "")
	}

	// Fill remaining space
	contentHeight := len(lines)
	for i := contentHeight; i < height-3; i++ {
		lines = append(lines, "")
	}

	// Bottom buttons
	divider := theme.DividerStyle.Render(strings.Repeat("─", sidebarWidth-2))
	lines = append(lines, divider)

	buttons := theme.ButtonStyle.Render("[+] tab") + " " + theme.ButtonStyle.Render("[g] grp")
	lines = append(lines, buttons)

	// Help text
	help := theme.HelpStyle.Render("r:rename n:new ↑↓:nav")
	lines = append(lines, help)

	content := strings.Join(lines, "\n")

	borderColor := theme.Surface2
	if focused {
		borderColor = theme.Lavender
	}

	sidebarContainer := lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(height).
		Background(theme.Mantle).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderRightForeground(borderColor)

	return sidebarContainer.Render(content)
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
			Label:    group.Name,
		})

		if !group.Collapsed {
			for ti, tab := range group.Tabs {
				items = append(items, SidebarItem{
					Type:     "tab",
					GroupIdx: gi,
					TabIdx:   ti,
					Label:    tab.Name,
				})
			}
		}
		items = append(items, SidebarItem{Type: "spacer"})
	}

	return items
}
