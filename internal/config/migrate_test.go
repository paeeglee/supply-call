package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateMovesLegacyFolderAndFixesBuildsFolder(t *testing.T) {
	base := t.TempDir()
	legacy := filepath.Join(base, LegacyDirName)
	os.MkdirAll(filepath.Join(legacy, "builds"), 0o755)
	os.WriteFile(filepath.Join(legacy, "builds", "minha.yml"), []byte("race: Terran"), 0o644)
	c := Defaults(legacy)
	c.SelectedBuild = "minha.yml"
	c.PlayerName = "Mateus"
	if err := Save(legacy, c); err != nil {
		t.Fatal(err)
	}

	dir, err := Migrate(base)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(base, DirName)
	if dir != want {
		t.Fatalf("dir = %q, want %q", dir, want)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("legacy folder still exists (err=%v)", err)
	}
	if _, err := os.Stat(filepath.Join(want, "builds", "minha.yml")); err != nil {
		t.Errorf("build not moved: %v", err)
	}
	got, err := Load(want)
	if err != nil {
		t.Fatal(err)
	}
	if got.BuildsFolder != filepath.Join(want, "builds") || got.SelectedBuild != "minha.yml" || got.PlayerName != "Mateus" {
		t.Errorf("config after migration = %+v", got)
	}
}

func TestMigrateKeepsCustomBuildsFolder(t *testing.T) {
	base := t.TempDir()
	legacy := filepath.Join(base, LegacyDirName)
	custom := filepath.Join(base, "MinhasBuilds")
	c := Defaults(legacy)
	c.BuildsFolder = custom
	if err := Save(legacy, c); err != nil {
		t.Fatal(err)
	}
	dir, err := Migrate(base)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := Load(dir)
	if got.BuildsFolder != custom {
		t.Errorf("builds_folder = %q, want %q (outside the legacy folder)", got.BuildsFolder, custom)
	}
}

func TestMigrateNothingToDo(t *testing.T) {
	base := t.TempDir()
	dir, err := Migrate(base) // no legacy folder
	if err != nil || dir != filepath.Join(base, DirName) {
		t.Errorf("dir = %q, err = %v", dir, err)
	}
	// New folder already exists: the legacy one is left alone.
	os.MkdirAll(filepath.Join(base, LegacyDirName), 0o755)
	os.MkdirAll(filepath.Join(base, DirName), 0o755)
	dir, err = Migrate(base)
	if err != nil || dir != filepath.Join(base, DirName) {
		t.Errorf("dir = %q, err = %v", dir, err)
	}
	if _, err := os.Stat(filepath.Join(base, LegacyDirName)); err != nil {
		t.Errorf("legacy folder touched: %v", err)
	}
}

func TestMigrateFailureKeepsLegacyFolder(t *testing.T) {
	// The old version still running keeps overlay.log open, so Windows
	// refuses to rename the folder; this run keeps using the old folder.
	base := t.TempDir()
	legacy := filepath.Join(base, LegacyDirName)
	os.MkdirAll(legacy, 0o755)
	f, err := OpenLog(legacy)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dir, err := Migrate(base)
	if err == nil {
		t.Fatal("want error while a file in the legacy folder is open")
	}
	if dir != legacy {
		t.Errorf("dir = %q, want the legacy folder %q for this run", dir, legacy)
	}
}

func TestUpgradeLegacyHotkey(t *testing.T) {
	// Ctrl+Alt+M was the first default for move mode; it is taken by another
	// program on the author's PC, so it is upgraded to the current default.
	c := Config{Hotkeys: Hotkeys{ToggleVisible: "Ctrl+Alt+H", ToggleClickThrough: "Ctrl+Alt+M"}}
	if !UpgradeLegacyHotkey(&c) || c.Hotkeys.ToggleClickThrough != "Ctrl+Alt+J" || c.Hotkeys.ToggleVisible != "Ctrl+Alt+H" {
		t.Errorf("upgraded = %+v", c.Hotkeys)
	}
	for _, keep := range []string{"Ctrl+Alt+J", "Ctrl+Shift+F10", "ctrl+alt+k"} {
		c := Config{Hotkeys: Hotkeys{ToggleClickThrough: keep}}
		if UpgradeLegacyHotkey(&c) || c.Hotkeys.ToggleClickThrough != keep {
			t.Errorf("%q changed to %q", keep, c.Hotkeys.ToggleClickThrough)
		}
	}
	lower := Config{Hotkeys: Hotkeys{ToggleClickThrough: " ctrl+alt+m "}}
	if !UpgradeLegacyHotkey(&lower) || lower.Hotkeys.ToggleClickThrough != "Ctrl+Alt+J" {
		t.Errorf("case/space variant not upgraded: %q", lower.Hotkeys.ToggleClickThrough)
	}
}
