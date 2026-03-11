package model

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nishant32f/loom/config"
	"github.com/nishant32f/loom/theme"
	"github.com/nishant32f/loom/tmux"
)

// App is the sidebar-only Bubble Tea model (runs inside the left tmux pane)
type App struct {
	Groups      []*Group
	ActiveGroup int
	ActiveTab   int
	Width       int
	Height      int
	Session     string // tmux session name
	Renaming    bool
	RenameInput string
	StatusMsg   string
	LastClick   time.Time
}

// statusClearMsg clears the status message
type statusClearMsg struct{}

// NewApp creates a new sidebar App from config
func NewApp(cfg *config.Config, session string) (*App, error) {
	app := &App{
		Groups:  make([]*Group, 0),
		Session: session,
	}

	groupMap := make(map[string]*Group)
	groupIdx := 0
	firstTab := true

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
				Tabs:  make([]*Tab, 0),
			}
			groupMap[sc.Group] = group
			app.Groups = append(app.Groups, group)
			groupIdx++
		}

		for _, tc := range sc.Tabs {
			if firstTab {
				tab := &Tab{
					Name:          tc.Name,
					HoldingWindow: -1,
					Command:       tc.Cmd,
					Cwd:           tc.Cwd,
					GroupName:     group.Name,
				}
				group.Tabs = append(group.Tabs, tab)
				if tc.Cmd != "" {
					tmux.RunInPane(session, tmux.TerminalPane(session), tc.Cmd)
				}
				firstTab = false
			} else {
				winIdx, err := tmux.CreateWindow(session, tc.Name)
				if err != nil {
					continue
				}
				tab := &Tab{
					Name:          tc.Name,
					HoldingWindow: winIdx,
					Command:       tc.Cmd,
					Cwd:           tc.Cwd,
					GroupName:     group.Name,
				}
				group.Tabs = append(group.Tabs, tab)
				if tc.Cmd != "" {
					tmux.RunInPane(session, tmux.WindowTarget(session, winIdx), tc.Cmd)
				}
			}
		}
	}

	if len(app.Groups) == 0 {
		app.Groups = append(app.Groups, &Group{
			Name:  "general",
			Color: theme.Blue,
			Tabs:  []*Tab{{Name: "shell", HoldingWindow: -1, GroupName: "general"}},
		})
	}

	tmux.SelectWindow(session, 0)
	app.saveState()
	return app, nil
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return tea.SetWindowTitle("Loom")
}

// ActiveTabRef returns the currently active tab
func (a *App) ActiveTabRef() *Tab {
	if a.ActiveGroup < 0 || a.ActiveGroup >= len(a.Groups) {
		return nil
	}
	group := a.Groups[a.ActiveGroup]
	if a.ActiveTab < 0 || a.ActiveTab >= len(group.Tabs) {
		return nil
	}
	return group.Tabs[a.ActiveTab]
}

// TotalTabs counts all tabs across all groups
func (a *App) TotalTabs() int {
	count := 0
	for _, g := range a.Groups {
		count += len(g.Tabs)
	}
	return count
}

// findActiveTab returns the tab whose pane is currently in the terminal
func (a *App) findActiveTab() *Tab {
	for _, g := range a.Groups {
		for _, t := range g.Tabs {
			if t.IsActive() {
				return t
			}
		}
	}
	return nil
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.Width = msg.Width
		a.Height = msg.Height
		return a, nil

	case statusClearMsg:
		a.StatusMsg = ""
		return a, nil

	case tea.MouseMsg:
		return a.handleMouse(msg)

	case tea.KeyMsg:
		if a.Renaming {
			return a.handleRenameKey(msg)
		}
		return a.handleSidebarKey(msg)
	}

	return a, nil
}

func (a *App) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionRelease {
			return a, nil
		}
		y := msg.Y

		// Check bottom buttons area first (before building items list)
		if y >= a.Height-6 && y < a.Height-4 {
			if msg.X < 14 {
				return a.addNewTab()
			}
			return a.addNewGroup()
		}

		// Click in tab/group area
		items := GetSidebarItems(a.Groups)
		if y >= 0 && y < len(items) {
			item := items[y]
			switch item.Type {
			case itemGroupHeader:
				a.Groups[item.GroupIdx].Collapsed = !a.Groups[item.GroupIdx].Collapsed
				a.saveState()
			case itemTab:
				now := time.Now()
				if item.GroupIdx == a.ActiveGroup && item.TabIdx == a.ActiveTab &&
					now.Sub(a.LastClick) < 400*time.Millisecond {
					a.Renaming = true
					if tab := a.ActiveTabRef(); tab != nil {
						a.RenameInput = tab.Name
					}
				} else {
					a.switchToTab(item.GroupIdx, item.TabIdx)
				}
				a.LastClick = now
			}
		}

	case tea.MouseButtonWheelUp:
		a.moveToPrevTab()
	case tea.MouseButtonWheelDown:
		a.moveToNextTab()
	}

	return a, nil
}

func (a *App) handleRenameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		tab := a.ActiveTabRef()
		if tab != nil && a.RenameInput != "" {
			tab.Name = a.RenameInput
			if !tab.IsActive() {
				tmux.RenameWindow(a.Session, tab.HoldingWindow, a.RenameInput)
			}
			a.saveState()
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

func (a *App) handleSidebarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "up", "k":
		a.moveToPrevTab()
	case "down", "j":
		a.moveToNextTab()
	case "enter":
		a.switchToTab(a.ActiveGroup, a.ActiveTab)
	case "tab":
		if a.ActiveGroup < len(a.Groups) {
			a.Groups[a.ActiveGroup].Collapsed = !a.Groups[a.ActiveGroup].Collapsed
			a.saveState()
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
		if tab := a.ActiveTabRef(); tab != nil {
			a.Renaming = true
			a.RenameInput = tab.Name
		}
	}

	return a, nil
}

// switchToTab handles swapping panes for tab switching
func (a *App) switchToTab(groupIdx, tabIdx int) {
	if groupIdx < 0 || groupIdx >= len(a.Groups) {
		return
	}
	group := a.Groups[groupIdx]
	if tabIdx < 0 || tabIdx >= len(group.Tabs) {
		return
	}

	newTab := group.Tabs[tabIdx]

	// If the new tab is already active, just focus the terminal
	if newTab.IsActive() {
		a.ActiveGroup = groupIdx
		a.ActiveTab = tabIdx
		tmux.SelectPane(a.Session, tmux.TerminalPane(a.Session))
		return
	}

	currentTab := a.findActiveTab()
	holdingWin := newTab.HoldingWindow

	source := tmux.TerminalPane(a.Session)
	target := tmux.WindowTarget(a.Session, holdingWin)
	if err := tmux.SwapPane(a.Session, source, target); err != nil {
		a.StatusMsg = "Swap failed"
		return
	}

	if currentTab != nil {
		currentTab.HoldingWindow = holdingWin
	}
	newTab.HoldingWindow = -1

	a.ActiveGroup = groupIdx
	a.ActiveTab = tabIdx

	tmux.FocusTerminal(a.Session)
	a.saveState()
}

func (a *App) moveToNextTab() {
	if len(a.Groups) == 0 {
		return
	}
	group := a.Groups[a.ActiveGroup]
	if group.Collapsed || a.ActiveTab >= len(group.Tabs)-1 {
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
	if len(a.Groups) == 0 {
		return
	}
	if a.ActiveTab > 0 {
		a.ActiveTab--
	} else {
		for gi := a.ActiveGroup - 1; gi >= 0; gi-- {
			if !a.Groups[gi].Collapsed && len(a.Groups[gi].Tabs) > 0 {
				a.ActiveGroup = gi
				a.ActiveTab = len(a.Groups[gi].Tabs) - 1
				return
			}
		}
	}
}

func (a *App) addNewTab() (tea.Model, tea.Cmd) {
	if a.ActiveGroup >= len(a.Groups) {
		return a, nil
	}

	group := a.Groups[a.ActiveGroup]
	name := fmt.Sprintf("tab-%d", len(group.Tabs)+1)

	winIdx, err := tmux.CreateWindow(a.Session, name)
	if err != nil {
		a.StatusMsg = "Failed to create tab"
		return a, nil
	}

	tab := &Tab{
		Name:          name,
		HoldingWindow: winIdx,
		GroupName:     group.Name,
	}
	group.Tabs = append(group.Tabs, tab)

	// switchToTab already saves state
	a.switchToTab(a.ActiveGroup, len(group.Tabs)-1)

	a.Renaming = true
	a.RenameInput = name
	return a, nil
}

func (a *App) closeCurrentTab() (tea.Model, tea.Cmd) {
	if a.ActiveGroup >= len(a.Groups) {
		return a, nil
	}

	group := a.Groups[a.ActiveGroup]
	if len(group.Tabs) == 0 {
		return a, nil
	}

	if a.TotalTabs() <= 1 {
		a.StatusMsg = "Can't close last tab"
		return a, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return statusClearMsg{}
		})
	}

	tab := group.Tabs[a.ActiveTab]

	if tab.IsActive() {
		// Find a replacement tab to swap in before killing
		found := false
		for gi, g := range a.Groups {
			for ti, t := range g.Tabs {
				if t != tab && !t.IsActive() {
					repTab := a.Groups[gi].Tabs[ti]
					holdingWin := repTab.HoldingWindow
					tmux.SwapPane(a.Session, tmux.TerminalPane(a.Session), tmux.WindowTarget(a.Session, holdingWin))
					repTab.HoldingWindow = -1
					tmux.KillWindow(a.Session, holdingWin)
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	} else {
		tmux.KillWindow(a.Session, tab.HoldingWindow)
	}

	// Remove from group
	group.Tabs = append(group.Tabs[:a.ActiveTab], group.Tabs[a.ActiveTab+1:]...)

	if a.ActiveTab >= len(group.Tabs) {
		a.ActiveTab = len(group.Tabs) - 1
	}
	if len(group.Tabs) == 0 && a.ActiveGroup > 0 {
		a.ActiveGroup--
		a.ActiveTab = len(a.Groups[a.ActiveGroup].Tabs) - 1
	}

	tmux.SelectWindow(a.Session, 0)
	a.saveState()
	return a, nil
}

func (a *App) addNewGroup() (tea.Model, tea.Cmd) {
	name := fmt.Sprintf("group-%d", len(a.Groups)+1)
	color := theme.GetGroupColor(len(a.Groups))

	winIdx, err := tmux.CreateWindow(a.Session, "shell")
	if err != nil {
		a.StatusMsg = "Failed to create group"
		return a, nil
	}

	group := &Group{
		Name:  name,
		Color: color,
		Tabs:  []*Tab{{Name: "shell", HoldingWindow: winIdx, GroupName: name}},
	}
	a.Groups = append(a.Groups, group)

	// switchToTab already saves state
	a.switchToTab(len(a.Groups)-1, 0)
	return a, nil
}

func (a *App) saveSession() (tea.Model, tea.Cmd) {
	cfg := &config.Config{
		Theme:    "catppuccin",
		Sessions: make([]config.SessionConfig, 0),
	}

	for _, group := range a.Groups {
		sc := config.SessionConfig{
			Name:  group.Name,
			Group: group.Name,
			Color: string(group.Color),
			Tabs:  make([]config.TabConfig, 0, len(group.Tabs)),
		}
		for _, tab := range group.Tabs {
			sc.Tabs = append(sc.Tabs, config.TabConfig{
				Name: tab.Name,
				Cmd:  tab.Command,
				Cwd:  tab.Cwd,
			})
		}
		cfg.Sessions = append(cfg.Sessions, sc)
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

func (a *App) saveState() {
	SaveState(a.Session, a.Groups, a.ActiveGroup, a.ActiveTab)
}

// View implements tea.Model
func (a *App) View() string {
	if a.Width == 0 || a.Height == 0 {
		return "Loading..."
	}

	return RenderSidebar(
		a.Groups,
		a.ActiveGroup,
		a.ActiveTab,
		a.Width,
		a.Height,
		a.Renaming,
		a.RenameInput,
		a.StatusMsg,
	)
}

// Cleanup removes state file and kills the tmux session
func (a *App) Cleanup() {
	RemoveState(a.Session)
	tmux.KillSession(a.Session)
}
