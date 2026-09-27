package build

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Entry is one build file found in the builds folder.
type Entry struct {
	File  string // base file name
	Name  string // display name
	Build *Build // nil when Err is set
	Err   error
}

// Label is the menu text: the display name, with " (erro)" for invalid files.
func (e Entry) Label() string {
	if e.Err != nil {
		return e.Name + " (erro)"
	}
	return e.Name
}

// Scan loads every .yml/.yaml file in dir, sorted by file name (case
// insensitive). A missing or unreadable folder yields no entries.
func Scan(dir string) []Entry {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []Entry
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if f.IsDir() || (ext != ".yml" && ext != ".yaml") {
			continue
		}
		e := Entry{File: f.Name(), Name: f.Name()}
		e.Build, e.Err = Load(filepath.Join(dir, f.Name()))
		if e.Build != nil {
			e.Name = e.Build.DisplayName()
		}
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].File) < strings.ToLower(out[j].File) })
	return out
}
