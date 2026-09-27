package overlay

import (
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procEnumWindows              = user32.NewProc("EnumWindows")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procGetWindowLongPtrW        = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW        = user32.NewProc("SetWindowLongPtrW")
	procShowWindowAsync          = user32.NewProc("ShowWindowAsync")
)

// hideFromTaskbar waits for this process's overlay window to appear and marks
// it as a tool window, which keeps it out of the taskbar and Alt+Tab
// (OVL-12). Ebitengine runs Update on a different thread from the window's,
// so this runs in its own goroutine and only uses calls that cannot block on
// the window thread for long (ShowWindowAsync).
func hideFromTaskbar(title string) {
	const (
		gwlExStyle       = ^uintptr(19) // GWL_EXSTYLE (-20)
		wsExToolWindow   = 0x00000080
		wsExAppWindow    = 0x00040000
		swHide           = 0
		swShowNoActivate = 4
	)
	self := windows.GetCurrentProcessId()
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(50 * time.Millisecond) {
		var hwnd uintptr
		cb := windows.NewCallback(func(h, _ uintptr) uintptr {
			var pid uint32
			procGetWindowThreadProcessId.Call(h, uintptr(unsafe.Pointer(&pid)))
			if pid != self {
				return 1
			}
			if v, _, _ := procIsWindowVisible.Call(h); v == 0 {
				return 1
			}
			buf := make([]uint16, 128)
			procGetWindowTextW.Call(h, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
			if windows.UTF16ToString(buf) == title {
				hwnd = h
				return 0
			}
			return 1
		})
		procEnumWindows.Call(cb, 0)
		if hwnd == 0 {
			continue
		}
		ex, _, _ := procGetWindowLongPtrW.Call(hwnd, gwlExStyle)
		// The taskbar only notices the change when the window is shown again.
		procShowWindowAsync.Call(hwnd, swHide)
		procSetWindowLongPtrW.Call(hwnd, gwlExStyle, ex&^wsExAppWindow|wsExToolWindow)
		procShowWindowAsync.Call(hwnd, swShowNoActivate)
		return
	}
}
