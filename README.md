# Loom 🌈

A beautiful, colorful TUI terminal tab manager with a vertical sidebar — inspired by Arc browser. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [tmux](https://github.com/tmux/tmux).

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)
![License](https://img.shields.io/badge/License-MIT-purple?style=flat-square)

```
┌─ Loom ──────────┬──────────────────────────────────────┐
│                  │                                      │
│ ▼ WORK           │  $ npm run dev                       │
│ ┃                │  > Server ready on :3000             │
│ ┣━● dev-server   │  > Compiled in 120ms                 │
│ ┣━  logs         │                                      │
│ ┣━  k8s          │──────────────────────────────────────│
│ ┗━  ssh-prod     │  $ docker ps                         │
│                  │  CONTAINER  IMAGE   STATUS            │
│ ▼ PERSONAL       │  abc123     nginx   Up 2 hours        │
│ ┃                │                                      │
│ ┗━  notes        │                                      │
│                  │                                      │
│──────────────────│                                      │
│ [+] tab  [g] grp │                                      │
└──────────────────┴──────────────────────────────────────┘
```

## Features

- **Vertical tab sidebar** — Arc browser-style tab management in your terminal
- **Tab groups** — Organize tabs into collapsible, color-coded groups
- **Mouse support** — Click to switch tabs, scroll through the sidebar
- **tmux-powered** — Full terminal emulation with splits and persistence
- **Session save/restore** — Save your workspace layout to YAML config
- **Beautiful themes** — Catppuccin-inspired colors out of the box
- **Rename tabs** — Press F2 or double-click to rename any tab

## Prerequisites

- Go 1.21+
- tmux 3.0+

## Installation

```bash
go install github.com/nishant32f/loom@latest
```

Or build from source:

```bash
git clone https://github.com/nishant32f/loom.git
cd loom
go build -o loom .
```

## Usage

```bash
loom                    # Launch with default session
loom new                # Start a fresh session
loom restore <name>     # Restore a saved session
loom save <name>        # Save current layout
loom list               # List saved sessions
```

## Keybindings

| Action | Key | Mouse |
|--------|-----|-------|
| Switch tab | `↑`/`↓` | Click tab |
| New tab | `Ctrl+T` | Click `[+]` |
| Rename tab | `F2` | Double-click |
| Close tab | `Ctrl+W` | — |
| Split pane | `Ctrl+\` | — |
| Toggle group | `Tab` | Click ▼/▶ |
| New group | `Ctrl+G` | — |
| Save session | `Ctrl+S` | — |
| Focus terminal | `Enter` / Click terminal | — |
| Back to sidebar | `Esc` | Click sidebar |
| Quit | `Ctrl+C` | — |

## Configuration

Sessions are stored in `~/.config/loom/sessions.yaml`:

```yaml
theme: catppuccin

sessions:
  - name: "Backend"
    group: "work"
    color: "#f38ba8"
    tabs:
      - name: "dev-server"
        cmd: "npm run dev"
        cwd: "~/projects/api"
      - name: "logs"
        cmd: "tail -f /var/log/app.log"

  - name: "Infra"
    group: "work"
    color: "#a6e3a1"
    tabs:
      - name: "k8s"
        cmd: "k9s"
      - name: "ssh-prod"
        cmd: "ssh prod-server"

  - name: "scratch"
    group: "personal"
    color: "#cba6f7"
    tabs:
      - name: "notes"
        cmd: "nvim ~/notes"
```

## Architecture

```
Loom (Bubble Tea TUI)
  │
  ├── Sidebar ─── Tab list with groups, colors, renaming
  ├── Terminal ── tmux pane capture rendered in viewport
  └── Config ──── YAML session persistence
        │
        └── tmux (backend)
              ├── Sessions & windows for each tab
              ├── Pane splitting
              └── capture-pane for rendering
```

## License

MIT
