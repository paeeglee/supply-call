// Package config loads and saves %APPDATA%\SC2BuildOverlay\config.yml.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const fileName = "config.yml"

// Config is the persisted user configuration. Keys are in English.
type Config struct {
	BuildsFolder   string    `yaml:"builds_folder"`
	SelectedBuild  string    `yaml:"selected_build"` // file name inside BuildsFolder
	Hotkeys        Hotkeys   `yaml:"hotkeys"`
	Opacity        float64   `yaml:"opacity"`
	WindowPosition *Position `yaml:"window_position"`
	WarningSeconds int       `yaml:"warning_seconds"`
	Sound          bool      `yaml:"sound"`
	ShowReplays    bool      `yaml:"show_replays"`
	PlayerName     string    `yaml:"player_name"`
}

// Hotkeys are written as "Ctrl+Alt+H".
type Hotkeys struct {
	ToggleVisible      string `yaml:"toggle_visible"`
	ToggleClickThrough string `yaml:"toggle_click_through"`
}

// Position is the window's top-left corner in screen coordinates.
type Position struct {
	X int `yaml:"x"`
	Y int `yaml:"y"`
}

// Dir returns %APPDATA%\SC2BuildOverlay.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "SC2BuildOverlay"), nil
}

// Defaults returns the default configuration for a config directory.
func Defaults(dir string) Config {
	return Config{
		BuildsFolder:   filepath.Join(dir, "builds"),
		Hotkeys:        Hotkeys{ToggleVisible: "Ctrl+Alt+H", ToggleClickThrough: "Ctrl+Alt+J"},
		Opacity:        0.7,
		WarningSeconds: 5,
	}
}

// Load reads dir/config.yml. Missing keys keep their default values.
func Load(dir string) (Config, error) {
	c := Defaults(dir)
	data, err := os.ReadFile(filepath.Join(dir, fileName))
	if err != nil {
		return c, err
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return Defaults(dir), fmt.Errorf("%s: %w", filepath.Join(dir, fileName), err)
	}
	return c, nil
}

// Save writes dir/config.yml atomically (temporary file + rename).
func Save(dir string, c Config) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, fileName+".*.tmp")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), filepath.Join(dir, fileName)); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return nil
}

// EnsureFirstRun loads the config. When it does not exist yet, it creates the
// builds folder with the example build, selects it and saves the config.
// A malformed config returns the defaults and an error, and the file is left
// untouched.
func EnsureFirstRun(dir, exampleName string, example []byte) (Config, error) {
	c, err := Load(dir)
	if !errors.Is(err, os.ErrNotExist) {
		return c, err
	}
	c = Defaults(dir)
	if err := os.MkdirAll(c.BuildsFolder, 0o755); err != nil {
		return c, err
	}
	if err := os.WriteFile(filepath.Join(c.BuildsFolder, exampleName), example, 0o644); err != nil {
		return c, err
	}
	c.SelectedBuild = exampleName
	return c, Save(dir, c)
}
