package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nishant32f/loom/config"
	"github.com/nishant32f/loom/model"
	"github.com/nishant32f/loom/tmux"
)

func main() {
	if !tmux.IsTmuxAvailable() {
		fmt.Println("Error: tmux is required but not found. Install it with:")
		fmt.Println("  brew install tmux     (macOS)")
		fmt.Println("  apt install tmux      (Linux)")
		os.Exit(1)
	}

	// Handle subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--sidebar":
			// Internal: run as sidebar TUI inside tmux pane
			runSidebar()
			return
		case "list":
			listSessions()
			return
		case "new":
			launchNew()
			return
		case "restore":
			if len(os.Args) < 3 {
				fmt.Println("Usage: loom restore <name>")
				os.Exit(1)
			}
			launchWithConfig(os.Args[2])
			return
		case "help", "--help", "-h":
			printHelp()
			return
		case "version", "--version", "-v":
			fmt.Println("Loom v0.2.0")
			return
		}
	}

	// Default: launch with saved config or default
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	launchSession(cfg)
}

// runSidebar runs the sidebar-only TUI (called inside the left tmux pane)
func runSidebar() {
	var sessionName string
	for i, arg := range os.Args {
		if arg == "--session" && i+1 < len(os.Args) {
			sessionName = os.Args[i+1]
			break
		}
	}
	if sessionName == "" {
		fmt.Fprintln(os.Stderr, "Error: --session <name> required with --sidebar")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	app, err := model.NewApp(cfg, sessionName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running sidebar: %v\n", err)
		os.Exit(1)
	}

	// When sidebar quits, clean up
	app.Cleanup()
}

// launchSession creates a tmux session with sidebar + terminal layout and attaches
func launchSession(cfg *config.Config) {
	sessionName := fmt.Sprintf("loom_%d", time.Now().UnixNano())

	// 1. Create the tmux session (detached)
	if err := tmux.CreateSession(sessionName); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// 2. Split window 0: left pane (sidebar, 24 cols) + right pane (terminal)
	if err := tmux.SplitForSidebar(sessionName, sidebarPaneWidth); err != nil {
		fmt.Printf("Error splitting: %v\n", err)
		tmux.KillSession(sessionName)
		os.Exit(1)
	}

	// 3. Get our executable path and start the sidebar TUI in the left pane (:0.0)
	exe, err := os.Executable()
	if err != nil {
		fmt.Printf("Error finding executable: %v\n", err)
		tmux.KillSession(sessionName)
		os.Exit(1)
	}

	sidebarCmd := fmt.Sprintf("%s --sidebar --session %s", exe, sessionName)
	if err := tmux.RunInPane(sessionName, sessionName+":0.0", sidebarCmd); err != nil {
		fmt.Printf("Error starting sidebar: %v\n", err)
		tmux.KillSession(sessionName)
		os.Exit(1)
	}

	// 4. Focus the right pane (the terminal)
	tmux.SelectPane(sessionName, sessionName+":0.1")

	// 5. Attach or switch
	if tmux.IsInsideTmux() {
		if err := tmux.SwitchClient(sessionName); err != nil {
			fmt.Printf("Error switching client: %v\n", err)
			os.Exit(1)
		}
	} else {
		tmux.Attach(sessionName)
	}
}

const sidebarPaneWidth = 26 // a bit wider than 24 to account for pane border

func launchNew() {
	cfg := &config.Config{
		Theme: "catppuccin",
		Sessions: []config.SessionConfig{
			{
				Name:  "default",
				Group: "general",
				Color: "#89b4fa",
				Tabs: []config.TabConfig{
					{Name: "shell"},
				},
			},
		},
	}
	launchSession(cfg)
}

func launchWithConfig(name string) {
	cfg, err := config.LoadNamed(name)
	if err != nil {
		fmt.Printf("Error loading session '%s': %v\n", name, err)
		fmt.Println("Available sessions:")
		listSessions()
		os.Exit(1)
	}
	launchSession(cfg)
}

func listSessions() {
	names, err := config.ListSaved()
	if err != nil {
		fmt.Printf("Error listing sessions: %v\n", err)
		os.Exit(1)
	}

	if len(names) == 0 {
		fmt.Println("No saved sessions.")
		return
	}

	fmt.Println("Saved sessions:")
	for _, name := range names {
		fmt.Printf("  • %s\n", name)
	}
}

func printHelp() {
	fmt.Println(`Loom - A beautiful TUI terminal tab manager (native tmux)

Usage:
  loom                    Launch with default/last session
  loom new                Start a fresh session
  loom restore <name>     Restore a saved session
  loom list               List saved sessions
  loom version            Show version
  loom help               Show this help

Inside Loom (sidebar):
  ↑/↓ or j/k      Navigate tabs
  Enter            Switch to tab (focus terminal)
  n                New tab
  d/x              Close tab
  g                New group
  r / F2           Rename tab
  Tab              Toggle group collapse
  s                Save session
  q / Ctrl+C       Quit

Navigation:
  Ctrl+B →         Focus terminal pane
  Ctrl+B ←         Focus sidebar pane`)
}
