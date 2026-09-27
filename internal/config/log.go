package config

import (
	"os"
	"path/filepath"
)

// maxLogSize: a log larger than this is emptied when the app starts.
const maxLogSize = 1 << 20

// OpenLog opens dir/overlay.log for appending, emptying it first when it is
// larger than 1 MB.
func OpenLog(dir string) (*os.File, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "overlay.log")
	flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if fi, err := os.Stat(path); err == nil && fi.Size() > maxLogSize {
		flags |= os.O_TRUNC
	}
	return os.OpenFile(path, flags, 0o644)
}
