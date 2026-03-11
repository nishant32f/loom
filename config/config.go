package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// TabConfig represents a single tab in a session
type TabConfig struct {
	Name string `yaml:"name"`
	Cmd  string `yaml:"cmd,omitempty"`
	Cwd  string `yaml:"cwd,omitempty"`
}

// SessionConfig represents a group of tabs
type SessionConfig struct {
	Name  string      `yaml:"name"`
	Group string      `yaml:"group"`
	Color string      `yaml:"color,omitempty"`
	Tabs  []TabConfig `yaml:"tabs"`
}

// Config is the top-level configuration
type Config struct {
	Theme    string          `yaml:"theme"`
	Sessions []SessionConfig `yaml:"sessions"`
}

// ConfigDir returns the configuration directory path
func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "loom")
}

// ConfigPath returns the full path to the sessions config file
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "sessions.yaml")
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	return os.MkdirAll(ConfigDir(), 0755)
}

// Load reads the config from disk
func Load() (*Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to disk
func Save(cfg *Config) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigPath(), data, 0644)
}

// DefaultConfig returns a starter config
func DefaultConfig() *Config {
	return &Config{
		Theme: "catppuccin",
		Sessions: []SessionConfig{
			{
				Name:  "default",
				Group: "general",
				Color: "#89b4fa",
				Tabs: []TabConfig{
					{Name: "shell", Cmd: ""},
				},
			},
		},
	}
}

// LoadNamed loads a specific named session file
func LoadNamed(name string) (*Config, error) {
	path := filepath.Join(ConfigDir(), name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveNamed saves config with a specific name
func SaveNamed(name string, cfg *Config) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	path := filepath.Join(ConfigDir(), name+".yaml")
	return os.WriteFile(path, data, 0644)
}

// ListSaved returns names of all saved session files
func ListSaved() ([]string, error) {
	dir := ConfigDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".yaml" {
			name := e.Name()[:len(e.Name())-5]
			names = append(names, name)
		}
	}
	return names, nil
}
