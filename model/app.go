package model

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nishant32f/loom/config"
	"github.com/nishant32f/loom/keybinds"
	"github.com/nishant32f/loom/theme"
	"github.com/nishant32f/loom/tmux"
)

// Focus represents which panel is focused
type Focus int

const (
	FocusSidebar Focus = iota
	FocusTerminal
)

const sidebarWidth = 22

// App is the main application model
type App struct {
	Groups       []*Group
	ActiveGroup  int
	ActiveTab    int
	Focus        Focus
	Width        int
	Height       int
	Keys         keybinds.KeyMap
	TmuxSession  *tmux.Session
	TermContent  string
	Renaming     bool
	RenameInput  string
	StatusMsg    string
	StatusExpiry time.Time
	LastClick    time.Time // for double-click detection
}

// tickMsg is sent periodically to refresh the terminal
type tickMsg time.Time

// statusClearMsg clears the status message
type statusClearMsg struct{}

// NewApp creates a new App model
func NewApp(cfg *config.Config) (*App, error) {
	// Create tmux session
	session, err := tmux.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create tmux session: %w", err)
	}

	app := &App{
		Groups:      make([]*Group, 0),
		ActiveGroup: 0,
		ActiveTab:   0,
		Focus:       FocusSidebar,
		Keys:        keybinds.DefaultKeyMap(),
		TmuxSession: session,
	}

	// Build groups from config
	groupMap := make(map[string]*Group)
	groupIdx := 0

	for _, sc := range cfg.Sessions {
		group, exists := groupMap[sc.Group]
		if !exists {
			color := theme.GetGroupColor(groupIdx)
			if sc.Color != "" {
				color = lipgloss.Color(sc.Color)
			}
			group = &Group{
				Name:  sc.Group,
				Color: color,
			}
			groupMap[sc.Group] = group
			app.Groups = append(app.Groups, group)
			groupIdx++
		}

		for _, tc := range sc.Tabs {
			win, err := session.NewWindow(tc.Name, tc.Cmd, tc.Cwd)
			if err != nil {
				// If first window fails, use existing window 0
				win = &tmux.Window{
					ID:   session.Name + ":0",
					Name: tc.Name,
				}
				if tc.Cmd != "" {
					tmux.SendKeys(win.ID, tc.Cmd)
				}
			}

			tab := &Tab{
				Name:    tc.Name,
				TmuxID:  win.ID,
				Command: tc.Cmd,
				Cwd:     tc.Cwd,
			}
			group.AddTab(tab)
		}
	}

	// If no groups were created, create a default one
	if len(app.Groups) == 0 {
		group := &Group{
			Name:  "general",
			Color: theme.Blue,
		}
		tab := &Tab{
			Name:   "shell",
			TmuxID: session.Name + ":0",
		}
		group.AddTab(tab)
		app.Groups = append(app.Groups, group)
	}

	return app, nil
}

// Init implements tea.Model
func (a App) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		tea.SetWindowTitle("Loom"),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// ActiveTabRef returns the currently active tab
func (a *App) ActiveTabRef() *Tab {
	if a.ActiveGroup >= len(a.Groups) {
		return nil
	}
	group := a.Groups[a.ActiveGroup]
	if a.ActiveTab >= len(group.Tabs) {
		return nil
	}
	return group.Tabs[a.ActiveTab]
}

// TotalVisibleTabs counts all visible tabs across groups
func (a *App) TotalVisibleTabs() int {
	count := 0
	for _, g := range a.Groups {
		if !g.Collapsed {
			count += len(g.Tabs)
		}
	}
	return count
}

// Update implements tea.Model
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.Width = msg.Width
		a.Height = msg.Height
		return a, nil

	case tickMsg:
		// Refresh terminal content
		tab := a.ActiveTabRef()
		if tab != nil {
			termWidth := a.Width - sidebarWidth - 4
			termHeight := a.Height - 4
			if termWidth > 10 && termHeight > 5 {
				content, err := tmux.CapturePaneContent(tab.TmuxID, termWidth, termHeight)
				if err == nil {
					a.TermContent = content
				}
			}
		}
		return a, tickCmd()

	case statusClearMsg:
		a.StatusMsg = ""
		return a, nil

	case tea.MouseMsg:
		return a.handleMouse(msg)

	case tea.KeyMsg:
		if a.Renaming {
			return a.handleRenameKey(msg)
		}

		if a.Focus == FocusTerminal {
			return a.handleTerminalKey(msg)
		}

		return a.handleSidebarKey(msg)
	}

	return a, nil
}

func (a App) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionRelease {
			return a, nil
		}
		x := msg.X
		if x < sidebarWidth {
			y := msg.Y
			sidebarHeight := a.Height - 1

			// Check if click is on bottom buttons (last 3 rows of sidebar)
			if y >= sidebarHeight-3 && y < sidebarHeight-1 {
				// Button row: [+] tab  [g] grp
				if x < 10 {
					return a.addNewTab()
				} else {
					return a.addNewGroup()
				}
			}

			// Click in sidebar tab/group area
			items := GetSidebarItems(a.Groups)
			if y >= 0 && y < len(items) {
				item := items[y]
				switch item.Type {
				case "group_header":
					a.Groups[item.GroupIdx].Collapsed = !a.Groups[item.GroupIdx].Collapsed
				case "tab":
					// Double-click detection for rename
					now := time.Now()
					if item.GroupIdx == a.ActiveGroup && item.TabIdx == a.ActiveTab &&
						now.Sub(a.LastClick) < 400*time.Millisecond {
						a.Renaming = true
						tab := a.ActiveTabRef()
						if tab != nil {
							a.RenameInput = tab.Name
						}
					} else {
						a.ActiveGroup = item.GroupIdx
						a.ActiveTab = item.TabIdx
					}
					a.LastClick = now
				}
			}
			a.Focus = FocusSidebar
		} else {
			// Click in terminal area
			a.Focus = FocusTerminal
		}

	case tea.MouseButtonWheelUp:
		if msg.X < sidebarWidth {
			a.moveToPrevTab()
		}
	case tea.MouseButtonWheelDown:
		if msg.X < sidebarWidth {
			a.moveToNextTab()
		}
	}

	return a, nil
}

func (a App) handleRenameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		tab := a.ActiveTabRef()
		if tab != nil && a.RenameInput != "" {
			tab.Name = a.RenameInput
			tmux.RenameWindow(tab.TmuxID, a.RenameInput)
		}
		a.Renaming = false
		a.RenameInput = ""
	case "esc":
		a.Renaming = false
		a.RenameInput = ""
	case "backspace":
		if len(a.RenameInput) > 0 {
			a.RenameInput = a.RenameInput[:len(a.RenameInput)-1]
		}
	default:
		if len(msg.String()) == 1 {
			a.RenameInput += msg.String()
		}
	}
	return a, nil
}

func (a App) handleTerminalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	tab := a.ActiveTabRef()

	switch msg.String() {
	case "esc":
		a.Focus = FocusSidebar
		return a, nil
	case "ctrl+c":
		return a, tea.Quit
	case "ctrl+\\":
		// Split pane
		if tab != nil {
			tmux.SplitWindow(tab.TmuxID, true, "")
		}
		return a, nil
	case "alt+n":
		return a.addNewTab()
	case "alt+w":
		return a.closeCurrentTab()
	case "alt+s":
		return a.saveSession()
	case "enter":
		if tab != nil {
			tmux.SendSpecialKey(tab.TmuxID, "Enter")
		}
	case "backspace":
		if tab != nil {
			tmux.SendSpecialKey(tab.TmuxID, "BSpace")
		}
	case "tab":
		if tab != nil {
			tmux.SendSpecialKey(tab.TmuxID, "Tab")
		}
	case "up":
		if tab != nil {
			tmux.SendSpecialKey(tab.TmuxID, "Up")
		}
	case "down":
		if tab != nil {
			tmux.SendSpecialKey(tab.TmuxID, "Down")
		}
	case "left":
		if tab != nil {
			tmux.SendSpecialKey(tab.TmuxID, "Left")
		}
	case "right":
		if tab != nil {
			tmux.SendSpecialKey(tab.TmuxID, "Right")
		}
	default:
		// Send keystrokes to tmux
		if tab != nil && len(msg.String()) == 1 {
			tmux.SendRawKeys(tab.TmuxID, msg.String())
		}
	}

	return a, nil
}

func (a App) handleSidebarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "up", "k":
		a.moveToPrevTab()
	case "down", "j":
		a.moveToNextTab()
	case "enter":
		a.Focus = FocusTerminal
	case "tab":
		// Toggle current group
		if a.ActiveGroup < len(a.Groups) {
			a.Groups[a.ActiveGroup].Collapsed = !a.Groups[a.ActiveGroup].Collapsed
		}
	case "n":
		return a.addNewTab()
	case "d", "x":
		return a.closeCurrentTab()
	case "g":
		return a.addNewGroup()
	case "s":
		return a.saveSession()
	case "r", "f2":
		tab := a.ActiveTabRef()
		if tab != nil {
			a.Renaming = true
			a.RenameInput = tab.Name
		}
	case "ctrl+\\":
		tab := a.ActiveTabRef()
		if tab != nil {
			tmux.SplitWindow(tab.TmuxID, true, "")
		}
	}

	return a, nil
}

func (a *App) moveToNextTab() {
	group := a.Groups[a.ActiveGroup]
	if group.Collapsed || a.ActiveTab >= len(group.Tabs)-1 {
		// Move to next group
		for gi := a.ActiveGroup + 1; gi < len(a.Groups); gi++ {
			if !a.Groups[gi].Collapsed && len(a.Groups[gi].Tabs) > 0 {
				a.ActiveGroup = gi
				a.ActiveTab = 0
				return
			}
		}
	} else {
		a.ActiveTab++
	}
}

func (a *App) moveToPrevTab() {
	if a.ActiveTab > 0 {
		a.ActiveTab--
	} else {
		// Move to previous group
		for gi := a.ActiveGroup - 1; gi >= 0; gi-- {
			if !a.Groups[gi].Collapsed && len(a.Groups[gi].Tabs) > 0 {
				a.ActiveGroup = gi
				a.ActiveTab = len(a.Groups[gi].Tabs) - 1
				return
			}
		}
	}
}

func (a App) addNewTab() (tea.Model, tea.Cmd) {
	if a.ActiveGroup >= len(a.Groups) {
		return a, nil
	}

	group := a.Groups[a.ActiveGroup]
	name := fmt.Sprintf("tab-%d", len(group.Tabs)+1)

	win, err := a.TmuxSession.NewWindow(name, "", "")
	if err != nil {
		a.StatusMsg = "Failed to create tab"
		return a, nil
	}

	tab := &Tab{
		Name:   name,
		TmuxID: win.ID,
	}
	group.AddTab(tab)
	a.ActiveTab = len(group.Tabs) - 1

	// Start renaming immediately
	a.Renaming = true
	a.RenameInput = name

	return a, nil
}

func (a App) closeCurrentTab() (tea.Model, tea.Cmd) {
	if a.ActiveGroup >= len(a.Groups) {
		return a, nil
	}

	group := a.Groups[a.ActiveGroup]
	if len(group.Tabs) == 0 {
		return a, nil
	}

	// Don't close the last tab
	if a.TotalVisibleTabs() <= 1 {
		a.StatusMsg = "Can't close last tab"
		return a, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return statusClearMsg{}
		})
	}

	tab := group.Tabs[a.ActiveTab]
	tmux.KillWindow(tab.TmuxID)
	group.RemoveTab(a.ActiveTab)

	// Adjust active tab
	if a.ActiveTab >= len(group.Tabs) {
		a.ActiveTab = len(group.Tabs) - 1
	}
	if len(group.Tabs) == 0 && a.ActiveGroup > 0 {
		a.ActiveGroup--
		a.ActiveTab = len(a.Groups[a.ActiveGroup].Tabs) - 1
	}

	return a, nil
}

func (a App) addNewGroup() (tea.Model, tea.Cmd) {
	name := fmt.Sprintf("group-%d", len(a.Groups)+1)
	color := theme.GetGroupColor(len(a.Groups))

	group := &Group{
		Name:  name,
		Color: color,
	}

	// Create a tab in the new group
	tabName := "shell"
	win, err := a.TmuxSession.NewWindow(tabName, "", "")
	if err != nil {
		a.StatusMsg = "Failed to create group"
		return a, nil
	}

	tab := &Tab{
		Name:   tabName,
		TmuxID: win.ID,
	}
	group.AddTab(tab)
	a.Groups = append(a.Groups, group)
	a.ActiveGroup = len(a.Groups) - 1
	a.ActiveTab = 0

	return a, nil
}

func (a App) saveSession() (tea.Model, tea.Cmd) {
	cfg := &config.Config{
		Theme:    "catppuccin",
		Sessions: make([]config.SessionConfig, 0),
	}

	for _, group := range a.Groups {
		for _, tab := range group.Tabs {
			found := false
			for i, sc := range cfg.Sessions {
				if sc.Group == group.Name {
					cfg.Sessions[i].Tabs = append(cfg.Sessions[i].Tabs, config.TabConfig{
						Name: tab.Name,
						Cmd:  tab.Command,
						Cwd:  tab.Cwd,
					})
					found = true
					break
				}
			}
			if !found {
				cfg.Sessions = append(cfg.Sessions, config.SessionConfig{
					Name:  group.Name,
					Group: group.Name,
					Color: string(group.Color),
					Tabs: []config.TabConfig{
						{Name: tab.Name, Cmd: tab.Command, Cwd: tab.Cwd},
					},
				})
			}
		}
	}

	if err := config.Save(cfg); err != nil {
		a.StatusMsg = "Save failed!"
	} else {
		a.StatusMsg = "Session saved!"
	}

	return a, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

// View implements tea.Model
func (a App) View() string {
	if a.Width == 0 || a.Height == 0 {
		return "Loading..."
	}

	// Render sidebar
	sidebar := RenderSidebar(
		a.Groups,
		a.ActiveGroup,
		a.ActiveTab,
		sidebarWidth,
		a.Height-1, // leave room for status bar
		a.Focus == FocusSidebar,
		a.Renaming,
		a.RenameInput,
	)

	// Render terminal
	termWidth := a.Width - sidebarWidth - 4
	termHeight := a.Height - 3
	tabName := "shell"
	tab := a.ActiveTabRef()
	if tab != nil {
		tabName = tab.Name
	}

	terminal := RenderTerminal(
		a.TermContent,
		tabName,
		termWidth,
		termHeight,
		a.Focus == FocusTerminal,
	)

	// Join sidebar and terminal horizontally
	main := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, terminal)

	// Status bar at the bottom
	statusContent := a.StatusMsg
	if statusContent == "" {
		focusLabel := "SIDEBAR"
		if a.Focus == FocusTerminal {
			focusLabel = "TERMINAL"
		}
		statusContent = fmt.Sprintf(" %s │ %d groups │ %d tabs │ n:new  g:group  s:save  d:close",
			focusLabel, len(a.Groups), a.TotalVisibleTabs())
	}

	statusBar := lipgloss.NewStyle().
		Width(a.Width).
		Background(theme.Surface0).
		Foreground(theme.Subtext0).
		Render(statusContent)

	return lipgloss.JoinVertical(lipgloss.Left, main, statusBar)
}

// Cleanup kills the tmux session
func (a *App) Cleanup() {
	if a.TmuxSession != nil {
		a.TmuxSession.Kill()
	}
}
