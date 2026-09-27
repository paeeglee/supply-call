// Command Supply Call is a StarCraft II build order overlay driven only by
// the official SC2 Client API (http://localhost:6119).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"golang.org/x/sys/windows"

	"sc2overlay/assets"
	"sc2overlay/internal/api"
	"sc2overlay/internal/build"
	"sc2overlay/internal/clock"
	"sc2overlay/internal/config"
	"sc2overlay/internal/hotkeys"
	"sc2overlay/internal/overlay"
	"sc2overlay/internal/sim"
	"sc2overlay/internal/tray"
)

// appName is the program name shown to the user.
const appName = "Supply Call"

// version is set at build time with -ldflags "-X main.version=x.y.z".
var version = "dev"

const pollInterval = 500 * time.Millisecond

// settings guards the config shared by the tray, the window and the poller,
// plus the build in use, which an automatic suggestion may change without
// saving it.
type settings struct {
	dir    string
	mu     sync.Mutex
	cfg    config.Config
	inUse  string // file name of the build shown
	manual bool   // a build was picked in the menu since the last match ended
}

func (s *settings) use(file string, manual bool) {
	s.mu.Lock()
	s.inUse = file
	s.manual = s.manual || manual
	s.mu.Unlock()
}

func (s *settings) current() (file string, manual bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.inUse, s.manual
}

func (s *settings) matchEnded() {
	s.mu.Lock()
	s.manual = false
	s.mu.Unlock()
}

func (s *settings) get() config.Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cfg
}

// update changes the config and saves it.
func (s *settings) update(f func(*config.Config)) {
	s.mu.Lock()
	f(&s.cfg)
	c := s.cfg
	s.mu.Unlock()
	if err := config.Save(s.dir, c); err != nil {
		log.Printf("salvar config: %v", err)
	}
}

func main() {
	showVersion := flag.Bool("version", false, "mostra a versão e sai")
	simMode := flag.Bool("sim", false, "simula a SC2 API em localhost:6119 (teste sem o jogo)")
	simSpeed := flag.Float64("sim-speed", 1, "velocidade da simulação (1, 2, 4)")
	simLength := flag.Duration("sim-length", 10*time.Minute, "duração de cada partida simulada")
	chooseFolder := flag.String("choose-folder", "", "uso interno: mostra o seletor de pasta e imprime a escolha")
	flag.Parse()

	if *chooseFolder != "" {
		runFolderPicker(*chooseFolder)
		return
	}

	if *showVersion {
		printConsole(appName + " " + version)
		return
	}

	base, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	// The app used to be "SC2 Build Overlay": move its folder on first run.
	dir, migrateErr := config.Migrate(base)
	if f, err := config.OpenLog(dir); err == nil {
		log.SetOutput(f)
		defer f.Close()
	}
	log.Printf("%s %s iniciado", appName, version)
	if migrateErr != nil {
		log.Printf("não foi possível mover %s para %s (%v); usando a pasta antiga nesta execução",
			config.LegacyDirName, config.DirName, migrateErr)
	}
	cfg, err := config.EnsureFirstRun(dir, assets.ExampleBuildName, assets.ExampleBuild)
	if err != nil {
		log.Printf("config: %v (usando padrões)", err)
	}
	st := &settings{dir: dir, cfg: cfg, inUse: cfg.SelectedBuild}

	shared := overlay.NewShared(clock.New(time.Now))
	loadSelected(shared, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var simulator *sim.Sim
	if *simMode {
		simulator = sim.New(time.Now, *simSpeed, *simLength)
		if err := simulator.Start(ctx, "localhost:6119"); err != nil {
			log.Printf("simulação: %v", err)
			shared.SetStatus(overlay.PortBusy)
		}
	}

	goSafe("poller", func() {
		api.Run(ctx, api.NewClient(api.DefaultURL), pollInterval,
			func() bool { return st.get().ShowReplays }, func(ev api.Event) {
				switch ev.Kind {
				case api.MatchStart:
					suggestBuild(st, shared, ev.Game.Players)
				case api.MatchEnd:
					st.matchEnded()
				}
				shared.HandleEvent(ev)
			})
	})

	for _, hk := range []struct {
		name, spec string
		fn         func()
	}{
		{"toggle_visible", cfg.Hotkeys.ToggleVisible, shared.ToggleVisible},
		{"toggle_click_through", cfg.Hotkeys.ToggleClickThrough, shared.ToggleClickThrough},
	} {
		if err := hotkeys.Listen(ctx, hk.spec, hk.fn); err != nil {
			log.Printf("atalho %s (%s) desativado: %v", hk.name, hk.spec, err)
		}
	}

	trayDone := tray.Start(tray.Options{
		Ctx: ctx, Icon: assets.IconICO, Name: appName, Version: version, OnQuit: shared.RequestQuit, Sim: simulator,
		Entries:  func() []build.Entry { return scanBuilds(st.get().BuildsFolder) },
		Selected: func() string { f, _ := st.current(); return f },
		OnSelect: func(e build.Entry) {
			shared.SetBuild(e.Build)
			st.use(e.File, true)
			st.update(func(c *config.Config) { c.SelectedBuild = e.File })
		},
		Folder: func() string { return st.get().BuildsFolder },
		OnFolder: func(dir string) {
			st.update(func(c *config.Config) { c.BuildsFolder = dir })
			st.use(st.get().SelectedBuild, false)
			loadSelected(shared, st.get())
		},
	})

	goSafe("watcher", func() { watchBuilds(ctx, st, shared) })

	g, err := overlay.NewGame(shared, overlay.Options{WarningSeconds: cfg.WarningSeconds, Opacity: cfg.Opacity, Sound: cfg.Sound})
	if err != nil {
		log.Fatal(err)
	}
	g.OnDragEnd = func(x, y int) { savePosition(st, x, y) }

	x, y := overlay.DefaultX, overlay.DefaultY
	if p := cfg.WindowPosition; p != nil {
		x, y = overlay.ClampPosition(p.X, p.Y, overlay.Monitors())
	}
	// Blocks on the main thread until "Sair".
	if err := overlay.Run(g, assets.IconPNG, x, y); err != nil {
		log.Print(err)
	}
	// Save only when the window moved: a malformed config is left untouched
	// until the user changes something (CFG-03).
	wx, wy := g.LastPosition()
	if overlay.PositionChanged(x, y, wx, wy) {
		savePosition(st, wx, wy)
	}

	cancel()
	tray.Quit()
	select {
	case <-trayDone:
	case <-time.After(2 * time.Second):
	}
}

// suggestBuild switches to the build matching the opponent's race when
// player_name is set and no build was picked in the menu since the last
// match. The choice is not saved to the config.
func suggestBuild(st *settings, shared *overlay.Shared, players []api.Player) {
	cfg := st.get()
	inUse, manual := st.current()
	if manual || cfg.PlayerName == "" {
		return
	}
	me, opp, ok := api.FindOpponent(players, cfg.PlayerName)
	if !ok {
		log.Printf("sugestão de build: jogador %q não encontrado na partida", cfg.PlayerName)
		return
	}
	entries := build.Scan(cfg.BuildsFolder)
	file, ok := build.Suggest(entries, me.Race, opp.Race)
	if !ok || file == inUse {
		return
	}
	for _, e := range entries {
		if e.File == file {
			shared.SetBuild(e.Build)
			st.use(file, false)
			log.Printf("sugestão de build: %s (%s vs %s)", file, me.Race, opp.Race)
			tray.Refresh()
			return
		}
	}
}

// watchBuilds rescans the builds folder every 2 s and, on any change,
// rebuilds the tray submenu and reloads the build in use.
func watchBuilds(ctx context.Context, st *settings, shared *overlay.Shared) {
	var w build.Watcher
	folder := ""
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for {
		dir := st.get().BuildsFolder
		if dir != folder { // the tray already refreshed after a folder change
			folder = dir
			w = build.Watcher{}
		}
		if w.Changed(dir) {
			log.Printf("pasta de builds alterada, recarregando")
			cfg := st.get()
			cfg.SelectedBuild, _ = st.current()
			loadSelected(shared, cfg)
			tray.Refresh()
		}
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}

func loadSelected(shared *overlay.Shared, cfg config.Config) {
	if cfg.SelectedBuild == "" {
		shared.SetBuild(nil)
		return
	}
	b, err := build.Load(filepath.Join(cfg.BuildsFolder, cfg.SelectedBuild))
	if err != nil {
		log.Printf("build: %v", err)
		shared.SetBuild(nil)
		return
	}
	shared.SetBuild(b)
}

// goSafe runs fn in a goroutine and logs a panic instead of crashing.
func goSafe(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("pânico em %s: %v\n%s", name, r, debug.Stack())
			}
		}()
		fn()
	}()
}

// printConsole writes a line to the console that started the app. A
// "-H windowsgui" build has no console of its own, so when stdout is not
// usable it attaches to the parent's console.
func printConsole(s string) {
	if _, err := fmt.Fprintln(os.Stdout, s); err == nil {
		return
	}
	const attachParentProcess = ^uintptr(0)
	windows.NewLazySystemDLL("kernel32.dll").NewProc("AttachConsole").Call(attachParentProcess)
	if con, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0); err == nil {
		fmt.Fprintln(con, s)
		con.Close()
	}
}

// runFolderPicker is the "-choose-folder" helper process: it prints the
// chosen folder and exits with 1 when the user cancels.
func runFolderPicker(current string) {
	dir, ok, err := tray.ChooseFolder(current)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if !ok {
		os.Exit(1)
	}
	fmt.Println(dir)
}

// scanBuilds lists the builds folder and logs why invalid files failed.
func scanBuilds(dir string) []build.Entry {
	entries := build.Scan(dir)
	for _, e := range entries {
		if e.Err != nil {
			log.Printf("build inválida: %v", e.Err)
		}
	}
	return entries
}

func savePosition(st *settings, x, y int) {
	st.update(func(c *config.Config) { c.WindowPosition = &config.Position{X: x, Y: y} })
}
