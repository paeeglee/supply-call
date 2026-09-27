package overlay

import (
	"image"
	"math"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/sys/windows"
)

// Default window position when none is saved or it is off-screen.
const (
	DefaultX = 20
	DefaultY = 200
)

// ClampPosition returns (x, y) when the top-left of the window (its clock)
// is on one of the monitors, and the default position otherwise. With no
// monitor information the position is kept.
func ClampPosition(x, y int, monitors []image.Rectangle) (int, int) {
	if len(monitors) == 0 {
		return x, y
	}
	grip := image.Pt(x+20, y+10) // a point inside the clock
	for _, m := range monitors {
		if grip.In(m) {
			return x, y
		}
	}
	return DefaultX, DefaultY
}

// PositionChanged reports whether the window moved during this run, i.e.
// whether its position must be saved on exit. Comparing with where the window
// opened (not with the config) keeps a malformed config.yml untouched when
// the user changed nothing (CFG-03).
func PositionChanged(openX, openY, x, y int) bool {
	return x != openX || y != openY
}

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procEnumDisplayMonitors = user32.NewProc("EnumDisplayMonitors")
)

// Monitors lists the monitor rectangles in device-independent pixels (the
// unit of ebiten.SetWindowPosition), using the primary monitor's scale.
// Call it on the main thread before ebiten.RunGame.
func Monitors() []image.Rectangle {
	var rects []image.Rectangle
	cb := windows.NewCallback(func(_ uintptr, _ uintptr, r *windows.Rect, _ uintptr) uintptr {
		rects = append(rects, image.Rect(int(r.Left), int(r.Top), int(r.Right), int(r.Bottom)))
		return 1
	})
	if ok, _, _ := procEnumDisplayMonitors.Call(0, 0, cb, uintptr(unsafe.Pointer(nil))); ok == 0 {
		return nil
	}
	scale := ebiten.Monitor().DeviceScaleFactor()
	if scale <= 0 {
		return rects
	}
	for i, r := range rects {
		rects[i] = image.Rect(
			int(math.Floor(float64(r.Min.X)/scale)), int(math.Floor(float64(r.Min.Y)/scale)),
			int(math.Ceil(float64(r.Max.X)/scale)), int(math.Ceil(float64(r.Max.Y)/scale)))
	}
	return rects
}
