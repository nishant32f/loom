package theme

import "github.com/charmbracelet/lipgloss"

// Catppuccin Mocha palette
var (
	Rosewater = lipgloss.Color("#f5e0dc")
	Flamingo  = lipgloss.Color("#f2cdcd")
	Pink      = lipgloss.Color("#f5c2e7")
	Mauve     = lipgloss.Color("#cba6f7")
	Red       = lipgloss.Color("#f38ba8")
	Maroon    = lipgloss.Color("#eba0ac")
	Peach     = lipgloss.Color("#fab387")
	Yellow    = lipgloss.Color("#f9e2af")
	Green     = lipgloss.Color("#a6e3a1")
	Teal      = lipgloss.Color("#94e2d5")
	Sky       = lipgloss.Color("#89dceb")
	Sapphire  = lipgloss.Color("#74c7ec")
	Blue      = lipgloss.Color("#89b4fa")
	Lavender  = lipgloss.Color("#b4befe")
	Text      = lipgloss.Color("#cdd6f4")
	Subtext1  = lipgloss.Color("#bac2de")
	Subtext0  = lipgloss.Color("#a6adc8")
	Overlay2  = lipgloss.Color("#9399b2")
	Overlay1  = lipgloss.Color("#7f849c")
	Overlay0  = lipgloss.Color("#6c7086")
	Surface2  = lipgloss.Color("#585b70")
	Surface1  = lipgloss.Color("#45475a")
	Surface0  = lipgloss.Color("#313244")
	Base      = lipgloss.Color("#1e1e2e")
	Mantle    = lipgloss.Color("#181825")
	Crust     = lipgloss.Color("#11111b")
)

// GroupColors are cycled through for new groups
var GroupColors = []lipgloss.Color{
	Red, Green, Blue, Mauve, Peach, Pink, Teal, Yellow, Sapphire, Flamingo,
}

// Styles
var (
	SidebarStyle = lipgloss.NewStyle().
			Background(Mantle).
			Padding(1, 1)

	SidebarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(Lavender).
				Background(Mantle).
				Padding(0, 1).
				MarginBottom(1)

	GroupHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Background(Mantle).
			Padding(0, 1)

	TabStyle = lipgloss.NewStyle().
			Background(Mantle).
			Foreground(Subtext1).
			Padding(0, 1)

	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Background(Surface0).
			Foreground(Text).
			Padding(0, 1)

	TabIndicator = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	TerminalBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Surface2)

	TerminalActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Lavender)

	StatusBarStyle = lipgloss.NewStyle().
			Background(Surface0).
			Foreground(Subtext0).
			Padding(0, 1)

	ButtonStyle = lipgloss.NewStyle().
			Background(Surface1).
			Foreground(Text).
			Padding(0, 1).
			MarginTop(1)

	ButtonActiveStyle = lipgloss.NewStyle().
				Background(Mauve).
				Foreground(Crust).
				Padding(0, 1).
				MarginTop(1).
				Bold(true)

	DividerStyle = lipgloss.NewStyle().
			Foreground(Surface2)

	RenameInputStyle = lipgloss.NewStyle().
				Background(Surface0).
				Foreground(Yellow).
				Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Overlay0).
			Background(Mantle).
			Padding(0, 1)

	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Mauve).
			Background(Mantle)
)

// Logo returns the Loom ASCII logo
func Logo() string {
	return LogoStyle.Render("◈ Loom")
}

// GetGroupColor returns a color for a group index
func GetGroupColor(index int) lipgloss.Color {
	return GroupColors[index%len(GroupColors)]
}
