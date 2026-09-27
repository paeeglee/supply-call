package overlay

import (
	"image"
	"testing"
)

func TestClampPosition(t *testing.T) {
	primary := image.Rect(0, 0, 1920, 1080)
	left := image.Rect(-1280, 0, 0, 1024) // second monitor on the left
	monitors := []image.Rectangle{primary, left}
	cases := []struct {
		name       string
		x, y       int
		wantX, wty int
	}{
		{"inside primary", 1500, 300, 1500, 300},
		{"inside left monitor", -1200, 100, -1200, 100},
		{"partly off the right edge but header visible", 1880, 10, 1880, 10},
		{"off every monitor", 5000, 300, DefaultX, DefaultY},
		{"above the screen", 100, -500, DefaultX, DefaultY},
		{"left monitor unplugged", -1200, 100, DefaultX, DefaultY},
	}
	for _, c := range cases {
		ms := monitors
		if c.name == "left monitor unplugged" {
			ms = []image.Rectangle{primary}
		}
		x, y := ClampPosition(c.x, c.y, ms)
		if x != c.wantX || y != c.wty {
			t.Errorf("%s: ClampPosition(%d,%d) = (%d,%d), want (%d,%d)", c.name, c.x, c.y, x, y, c.wantX, c.wty)
		}
	}
}

func TestClampPositionNoMonitorInfoKeepsPosition(t *testing.T) {
	// When the monitor list is unavailable, do not move the window.
	if x, y := ClampPosition(700, 400, nil); x != 700 || y != 400 {
		t.Errorf("got (%d,%d), want (700,400)", x, y)
	}
}

func TestPositionChanged(t *testing.T) {
	// CFG-03 / TRAY-08: save on exit only when the window moved in this run.
	// A malformed config opens at the default (20,200) with no saved
	// position; quitting without moving must not rewrite it.
	if PositionChanged(DefaultX, DefaultY, DefaultX, DefaultY) {
		t.Error("window did not move: want no save")
	}
	if !PositionChanged(DefaultX, DefaultY, 170, 230) {
		t.Error("window moved: want save")
	}
	if !PositionChanged(170, 230, 170, 231) {
		t.Error("moved by one pixel: want save")
	}
}
