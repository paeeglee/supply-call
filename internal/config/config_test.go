package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var example = []byte("name: Exemplo\nrace: Terran\nsteps: [{time: \"0:35\", action: Supply Depot}]\n")

func TestDirIsUnderAppData(t *testing.T) {
	d, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(d) != "SC2BuildOverlay" {
		t.Errorf("Dir() = %q, want .../SC2BuildOverlay", d)
	}
	if appdata := os.Getenv("APPDATA"); appdata != "" && filepath.Dir(d) != appdata {
		t.Errorf("Dir() = %q, want under %q", d, appdata)
	}
}

func TestFirstRunCreatesConfigAndExampleBuild(t *testing.T) {
	dir := t.TempDir()
	c, err := EnsureFirstRun(dir, "exemplo.yml", example)
	if err != nil {
		t.Fatal(err)
	}
	want := Config{
		BuildsFolder:   filepath.Join(dir, "builds"),
		SelectedBuild:  "exemplo.yml",
		Hotkeys:        Hotkeys{ToggleVisible: "Ctrl+Alt+H", ToggleClickThrough: "Ctrl+Alt+J"},
		Opacity:        0.7,
		WarningSeconds: 5,
		Sound:          false,
	}
	if c != want {
		t.Errorf("config = %+v\nwant     %+v", c, want)
	}
	got, err := os.ReadFile(filepath.Join(dir, "builds", "exemplo.yml"))
	if err != nil || string(got) != string(example) {
		t.Errorf("example build = %q, %v", got, err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "config.yml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"builds_folder:", "selected_build:", "hotkeys:", "toggle_visible:", "toggle_click_through:",
		"opacity:", "window_position:", "warning_seconds:", "sound:", "show_replays:", "player_name:"} {
		if !strings.Contains(string(raw), key) {
			t.Errorf("config.yml missing key %s:\n%s", key, raw)
		}
	}
}

func TestFirstRunKeepsExistingConfig(t *testing.T) {
	dir := t.TempDir()
	c := Defaults(dir)
	c.SelectedBuild = "mine.yml"
	c.Sound = true
	if err := Save(dir, c); err != nil {
		t.Fatal(err)
	}
	got, err := EnsureFirstRun(dir, "exemplo.yml", example)
	if err != nil {
		t.Fatal(err)
	}
	if got.SelectedBuild != "mine.yml" || !got.Sound {
		t.Errorf("existing config replaced: %+v", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "builds", "exemplo.yml")); !os.IsNotExist(err) {
		t.Errorf("example written on non-first run (err=%v)", err)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	c := Defaults(dir)
	c.BuildsFolder = `D:\Builds Ç`
	c.SelectedBuild = "pvz.yaml"
	c.Hotkeys.ToggleVisible = "Ctrl+Shift+F9"
	c.Opacity = 0.55
	c.WindowPosition = &Position{X: -1200, Y: 340}
	c.WarningSeconds = 8
	c.Sound = true
	c.ShowReplays = true
	c.PlayerName = "João"
	if err := Save(dir, c); err != nil {
		t.Fatal(err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.WindowPosition == nil || *got.WindowPosition != *c.WindowPosition {
		t.Fatalf("position = %v", got.WindowPosition)
	}
	got.WindowPosition, c.WindowPosition = nil, nil
	if got != c {
		t.Errorf("round trip = %+v\nwant         %+v", got, c)
	}
}

func TestLoadFillsDefaultsForMissingKeys(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.yml"), []byte("sound: true\nfuture_key: 1\n"), 0o644)
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !c.Sound || c.WarningSeconds != 5 || c.Opacity != 0.7 || c.Hotkeys.ToggleClickThrough != "Ctrl+Alt+J" ||
		c.BuildsFolder != filepath.Join(dir, "builds") {
		t.Errorf("config = %+v", c)
	}
}

func TestMalformedConfigUsesDefaultsAndIsNotOverwritten(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	broken := []byte("sound: [oops\n")
	os.WriteFile(path, broken, 0o644)
	c, err := EnsureFirstRun(dir, "exemplo.yml", example)
	if err == nil {
		t.Fatal("want error for malformed config")
	}
	if c != Defaults(dir) {
		t.Errorf("config = %+v, want defaults", c)
	}
	if got, _ := os.ReadFile(path); string(got) != string(broken) {
		t.Errorf("malformed file was overwritten: %q", got)
	}
}

func TestSaveLeavesNoTempFile(t *testing.T) {
	dir := t.TempDir()
	if err := Save(dir, Defaults(dir)); err != nil {
		t.Fatal(err)
	}
	if err := Save(dir, Defaults(dir)); err != nil { // replace existing
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 || entries[0].Name() != "config.yml" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("dir contents = %v, want only config.yml", names)
	}
}
