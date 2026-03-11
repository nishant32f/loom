package model

import (
	"encoding/json"
	"fmt"
	"os"
)

// StateFile holds the persistent state for a loom session
type StateFile struct {
	Session   string      `json:"session"`
	ActiveTab int         `json:"active_tab"`
	Tabs      []TabState  `json:"tabs"`
	Groups    []GroupState `json:"groups"`
}

// TabState is the JSON-serializable form of a Tab
type TabState struct {
	Name          string `json:"name"`
	Group         string `json:"group"`
	HoldingWindow int    `json:"holding_window"`
}

// GroupState is the JSON-serializable form of a Group
type GroupState struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	Collapsed bool   `json:"collapsed"`
}

// StatePath returns the path for a session's state file
func StatePath(session string) string {
	return fmt.Sprintf("/tmp/loom_%s.json", session)
}

// SaveState writes the current state to disk
func SaveState(session string, groups []*Group, activeGroup, activeTab int) error {
	state := StateFile{
		Session:   session,
		ActiveTab: activeTab,
		Tabs:      make([]TabState, 0),
		Groups:    make([]GroupState, 0),
	}

	for _, g := range groups {
		state.Groups = append(state.Groups, GroupState{
			Name:      g.Name,
			Color:     string(g.Color),
			Collapsed: g.Collapsed,
		})
		for _, t := range g.Tabs {
			state.Tabs = append(state.Tabs, TabState{
				Name:          t.Name,
				Group:         t.GroupName,
				HoldingWindow: t.HoldingWindow,
			})
		}
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(StatePath(session), data, 0644)
}

// LoadState reads state from disk
func LoadState(session string) (*StateFile, error) {
	data, err := os.ReadFile(StatePath(session))
	if err != nil {
		return nil, err
	}
	var state StateFile
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// RemoveState deletes the state file
func RemoveState(session string) {
	os.Remove(StatePath(session))
}
