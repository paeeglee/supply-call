// Package tray runs the Windows notification area icon and its menu.
package tray

import (
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"fyne.io/systray"
	"github.com/ncruces/zenity"

	"sc2overlay/internal/build"
	"sc2overlay/internal/sim"
)

// Options configures the tray.
type Options struct {
	Ctx     context.Context // cancelled on exit; closes an open folder dialog
	Icon    []byte          // .ico bytes
	Name    string          // program name, shown in the tooltip
	Version string
	OnQuit  func()   // called once when "Sair" is clicked
	Sim     *sim.Sim // adds the "Simulação" submenu when set (-sim)

	Entries  func() []build.Entry // builds in the current folder
	Selected func() string        // file name of the build in use
	OnSelect func(e build.Entry)  // a valid build was picked in the menu
	Folder   func() string        // current builds folder
	OnFolder func(dir string)     // a new builds folder was chosen
}

// buildMenu is the "Build order" submenu, rebuilt by Refresh.
var buildMenu struct {
	mu     sync.Mutex
	opts   Options
	parent *systray.MenuItem
	items  []*systray.MenuItem
	stop   chan struct{} // ends the click goroutines of the current items
}

// Start runs the tray on its own locked OS thread: on Windows the tray window
// and its message loop must live on the same thread, so systray.Run is used
// instead of RunWithExternalLoop. The returned channel closes after the tray
// has exited.
func Start(opts Options) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		systray.Run(func() { onReady(opts) }, func() { close(done) })
	}()
	return done
}

// FolderTitle is the folder picker's window title.
const FolderTitle = "Pasta das builds"

// dialogs counts open folder pickers.
var dialogs sync.WaitGroup

// Quit removes the icon and ends the tray loop. Cancel Options.Ctx first so
// an open folder picker is closed; Quit waits up to 2 s for it.
func Quit() {
	systray.Quit()
	done := make(chan struct{})
	go func() {
		dialogs.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
}

func onReady(opts Options) {
	systray.SetIcon(opts.Icon)
	systray.SetTooltip(opts.Name + " " + opts.Version)

	builds := systray.AddMenuItem("Build order", "Escolher a build")
	buildMenu.mu.Lock()
	buildMenu.opts, buildMenu.parent = opts, builds
	buildMenu.mu.Unlock()
	Refresh()
	folder := systray.AddMenuItem("Escolher pasta das builds…", "")
	go func() {
		for range folder.ClickedCh {
			chooseFolder(opts)
		}
	}()
	if opts.Sim != nil {
		addSimMenu(opts.Sim)
	}
	systray.AddSeparator()
	ver := systray.AddMenuItem("Versão "+opts.Version, "")
	ver.Disable()
	quit := systray.AddMenuItem("Sair", "Fechar o overlay")

	go func() {
		<-quit.ClickedCh
		opts.OnQuit()
	}()
}

// ChooseFolder shows the native folder picker and returns the chosen folder.
// ok is false when the user cancels. It runs inside the helper process
// started by chooseFolder.
func ChooseFolder(current string) (dir string, ok bool, err error) {
	dir, err = zenity.SelectFile(zenity.Directory(), zenity.Filename(current), zenity.Title(FolderTitle))
	if errors.Is(err, zenity.ErrCanceled) {
		return "", false, nil
	}
	return dir, err == nil, err
}

// chooseFolder runs the picker in a child process ("-choose-folder"). The
// native dialog loads shell extensions into its process, and some of them
// can freeze process teardown; keeping them out of the overlay process
// means quitting never hangs. Cancelling Options.Ctx kills the helper.
func chooseFolder(opts Options) {
	dialogs.Add(1)
	defer dialogs.Done()
	exe, err := os.Executable()
	if err != nil {
		log.Printf("diálogo de pasta: %v", err)
		return
	}
	out, err := exec.CommandContext(opts.Ctx, exe, "-choose-folder", opts.Folder()).Output()
	dir := strings.TrimSpace(string(out))
	switch {
	case opts.Ctx.Err() != nil:
		return
	case err != nil:
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 { // 1 = cancelled
			log.Printf("diálogo de pasta: %v", err)
		}
		return
	case dir == "":
		return
	}
	opts.OnFolder(dir)
	Refresh()
}

// Refresh rebuilds the "Build order" submenu from Options.Entries and marks
// Options.Selected. Safe to call from any goroutine, before or after the
// tray is ready.
func Refresh() {
	m := &buildMenu
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.parent == nil {
		return
	}
	if m.stop != nil {
		close(m.stop)
	}
	for _, it := range m.items {
		it.Remove()
	}
	m.items, m.stop = nil, make(chan struct{})

	entries := m.opts.Entries()
	if len(entries) == 0 {
		empty := m.parent.AddSubMenuItem("Nenhuma build encontrada", "")
		empty.Disable()
		m.items = append(m.items, empty)
		return
	}
	selected := m.opts.Selected()
	for _, e := range entries {
		it := m.parent.AddSubMenuItemCheckbox(e.Label(), e.File, e.Err == nil && e.File == selected)
		m.items = append(m.items, it)
		if e.Err != nil {
			it.Disable()
			continue
		}
		go func(e build.Entry, it *systray.MenuItem, stop chan struct{}) {
			for {
				select {
				case <-stop:
					return
				case <-it.ClickedCh:
					selectItem(it)
					m.opts.OnSelect(e)
				}
			}
		}(e, it, m.stop)
	}
}

// selectItem checks it and unchecks the other builds.
func selectItem(it *systray.MenuItem) {
	buildMenu.mu.Lock()
	defer buildMenu.mu.Unlock()
	for _, other := range buildMenu.items {
		if other == it {
			other.Check()
		} else {
			other.Uncheck()
		}
	}
}

// addSimMenu adds pause, speed and new-match controls for the simulator.
func addSimMenu(s *sim.Sim) {
	menu := systray.AddMenuItem("Simulação", "Controles do modo -sim")
	pause := menu.AddSubMenuItemCheckbox("Pausar", "", false)
	speeds := map[float64]*systray.MenuItem{}
	for _, v := range []float64{1, 2, 4} {
		speeds[v] = menu.AddSubMenuItemCheckbox(map[float64]string{1: "Velocidade 1x", 2: "Velocidade 2x", 4: "Velocidade 4x"}[v], "", s.Speed() == v)
	}
	newMatch := menu.AddSubMenuItem("Nova partida", "")

	go func() {
		for range pause.ClickedCh {
			if s.Paused() {
				s.Resume()
				pause.Uncheck()
			} else {
				s.Pause()
				pause.Check()
			}
		}
	}()
	for v, item := range speeds {
		go func() {
			for range item.ClickedCh {
				s.SetSpeed(v)
				for w, other := range speeds {
					if w == v {
						other.Check()
					} else {
						other.Uncheck()
					}
				}
			}
		}()
	}
	go func() {
		for range newMatch.ClickedCh {
			s.NewMatch()
		}
	}()
}
