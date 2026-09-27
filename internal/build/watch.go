package build

import (
	"maps"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type fileStamp struct {
	mod  time.Time
	size int64
}

// Watcher detects created, modified and deleted build files by polling.
// The zero value is ready; it is not safe for concurrent use.
type Watcher struct {
	started bool
	last    map[string]fileStamp
}

// Changed scans dir and reports whether its .yml/.yaml files differ from the
// previous scan. The first scan only records the baseline.
func (w *Watcher) Changed(dir string) bool {
	now := map[string]fileStamp{}
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if f.IsDir() || (ext != ".yml" && ext != ".yaml") {
			continue
		}
		if info, err := f.Info(); err == nil {
			now[f.Name()] = fileStamp{info.ModTime(), info.Size()}
		}
	}
	changed := w.started && !maps.Equal(now, w.last)
	w.started, w.last = true, now
	return changed
}
