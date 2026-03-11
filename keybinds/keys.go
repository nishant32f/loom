package keybinds

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for Loom
type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	NewTab     key.Binding
	CloseTab   key.Binding
	RenameTab  key.Binding
	SplitPane  key.Binding
	ToggleGroup key.Binding
	NewGroup   key.Binding
	SaveSession key.Binding
	FocusTerm  key.Binding
	BackToSide key.Binding
	Quit       key.Binding
	Confirm    key.Binding
	Cancel     key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		NewTab: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new tab"),
		),
		CloseTab: key.NewBinding(
			key.WithKeys("d", "x"),
			key.WithHelp("d/x", "close tab"),
		),
		RenameTab: key.NewBinding(
			key.WithKeys("r", "f2"),
			key.WithHelp("r/F2", "rename"),
		),
		SplitPane: key.NewBinding(
			key.WithKeys("ctrl+\\"),
			key.WithHelp("ctrl+\\", "split"),
		),
		ToggleGroup: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "toggle group"),
		),
		NewGroup: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "new group"),
		),
		SaveSession: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "save"),
		),
		FocusTerm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "focus terminal"),
		),
		BackToSide: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "sidebar"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// ShortHelp returns a short help string
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.NewTab, k.RenameTab, k.FocusTerm, k.Quit}
}

// FullHelp returns the full help string
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.NewTab, k.CloseTab},
		{k.RenameTab, k.SplitPane, k.ToggleGroup, k.NewGroup},
		{k.SaveSession, k.FocusTerm, k.BackToSide, k.Quit},
	}
}
