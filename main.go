package main

import (
	"fmt"
	"os"

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
		case "list":
			listSessions()
			return
		case "save":
			if len(os.Args) < 3 {
				fmt.Println("Usage: loom save <name>")
				os.Exit(1)
			}
			// Save is handled in the TUI via ctrl+s
			fmt.Println("Use ctrl+s inside Loom to save the current session.")
			return
		case "restore":
			if len(os.Args) < 3 {
				fmt.Println("Usage: loom restore <name>")
				os.Exit(1)
			}
			launchWithConfig(os.Args[2])
			return
		case "new":
			launchNew()
			return
		case "help", "--help", "-h":
			printHelp()
			return
		case "version", "--version", "-v":
			fmt.Println("Loom v0.1.0")
			return
		}
	}

	// Default: launch with saved config or default
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	launch(cfg)
}

func launch(cfg *config.Config) {
	app, err := model.NewApp(cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Cleanup()

	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running Loom: %v\n", err)
		os.Exit(1)
	}
}

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
	launch(cfg)
}

func launchWithConfig(name string) {
	cfg, err := config.LoadNamed(name)
	if err != nil {
		fmt.Printf("Error loading session '%s': %v\n", name, err)
		fmt.Println("Available sessions:")
		listSessions()
		os.Exit(1)
	}
	launch(cfg)
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
	fmt.Println(`Loom - A beautiful TUI terminal tab manager

Usage:
  loom                    Launch with default/last session
  loom new                Start a fresh session
  loom restore <name>     Restore a saved session
  loom list               List saved sessions
  loom version            Show version
  loom help               Show this help

Keybindings:
  ↑/↓ or j/k      Navigate tabs
  Enter            Focus terminal
  Esc              Back to sidebar
  Ctrl+T           New tab
  Ctrl+W           Close tab
  Ctrl+G           New group
  Ctrl+S           Save session
  Ctrl+\           Split pane
  F2               Rename tab
  Tab              Toggle group collapse
  q / Ctrl+C       Quit`)
}
