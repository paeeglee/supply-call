package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenLogAppends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "overlay.log")
	os.WriteFile(path, []byte("linha antiga\n"), 0o644)
	w, err := OpenLog(dir)
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("linha nova\n"))
	w.Close()
	got, _ := os.ReadFile(path)
	if string(got) != "linha antiga\nlinha nova\n" {
		t.Errorf("log = %q", got)
	}
}

func TestOpenLogTruncatesOverOneMB(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "overlay.log")
	os.WriteFile(path, []byte(strings.Repeat("x", 1<<20+1)), 0o644)
	w, err := OpenLog(dir)
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("novo\n"))
	w.Close()
	got, _ := os.ReadFile(path)
	if string(got) != "novo\n" {
		t.Errorf("log size = %d, want only the new line", len(got))
	}
}

func TestOpenLogCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "SC2BuildOverlay")
	w, err := OpenLog(dir)
	if err != nil {
		t.Fatal(err)
	}
	w.Close()
	if _, err := os.Stat(filepath.Join(dir, "overlay.log")); err != nil {
		t.Error(err)
	}
}
