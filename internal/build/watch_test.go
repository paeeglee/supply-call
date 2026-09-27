package build

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherDetectsCreateModifyDelete(t *testing.T) {
	dir := t.TempDir()
	var w Watcher
	if w.Changed(dir) {
		t.Error("first scan is the baseline, want false")
	}
	if w.Changed(dir) {
		t.Error("nothing changed, want false")
	}

	path := filepath.Join(dir, "nova.yml")
	os.WriteFile(path, []byte("a"), 0o644)
	if !w.Changed(dir) {
		t.Error("created .yml, want true")
	}
	if w.Changed(dir) {
		t.Error("no change since last scan, want false")
	}

	// Same size, newer modification time.
	os.WriteFile(path, []byte("b"), 0o644)
	later := time.Now().Add(time.Minute)
	os.Chtimes(path, later, later)
	if !w.Changed(dir) {
		t.Error("modified .yml, want true")
	}

	os.WriteFile(filepath.Join(dir, "outra.yaml"), []byte("c"), 0o644)
	if !w.Changed(dir) {
		t.Error("created .yaml, want true")
	}

	os.Remove(path)
	if !w.Changed(dir) {
		t.Error("deleted .yml, want true")
	}
}

func TestWatcherIgnoresOtherFiles(t *testing.T) {
	dir := t.TempDir()
	var w Watcher
	w.Changed(dir)
	os.WriteFile(filepath.Join(dir, "notas.txt"), []byte("x"), 0o644)
	if w.Changed(dir) {
		t.Error(".txt must not count as a change")
	}
}
