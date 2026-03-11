package model

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nishant32f/loom/theme"
)

// RenderTerminal renders the terminal viewport with captured tmux content
func RenderTerminal(content string, tabName string, width, height int, focused bool) string {
	// Title bar
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Text).
		Background(theme.Surface0).
		Width(width).
		Padding(0, 1)

	title := titleStyle.Render("  " + tabName)

	// Terminal content area
	lines := strings.Split(content, "\n")

	// Pad or trim to fit height
	termHeight := height - 2 // account for title and bottom bar
	for len(lines) < termHeight {
		lines = append(lines, "")
	}
	if len(lines) > termHeight {
		lines = lines[:termHeight]
	}

	// Pad each line to width
	for i, line := range lines {
		lineLen := lipgloss.Width(line)
		if lineLen < width {
			lines[i] = line + strings.Repeat(" ", width-lineLen)
		}
	}

	termContent := strings.Join(lines, "\n")

	contentStyle := lipgloss.NewStyle().
		Background(theme.Base).
		Foreground(theme.Text).
		Width(width)

	// Status bar
	statusText := "Press Esc for sidebar"
	if !focused {
		statusText = "Press Enter to focus"
	}
	statusBar := theme.StatusBarStyle.
		Width(width).
		Render(statusText)

	borderColor := theme.Surface2
	if focused {
		borderColor = theme.Lavender
	}

	container := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor)

	inner := lipgloss.JoinVertical(lipgloss.Left,
		title,
		contentStyle.Render(termContent),
		statusBar,
	)

	return container.Render(inner)
}
