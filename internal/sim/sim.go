// Package sim serves a fake SC2 Client API for testing without the game.
package sim

import (
	"context"
	"encoding/json"
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"sc2overlay/internal/api"
)

const phaseLength = 3 * time.Second // menu and loading screens

type phase int

const (
	menu phase = iota
	loading
	game
)

var (
	menuScreens    = []string{"ScreenBackgroundSC2/ScreenBackgroundSC2", "ScreenNavigationSC2/ScreenNavigationSC2", "ScreenHome/ScreenHome"}
	loadingScreens = []string{"ScreenLoading/ScreenLoading"}
	players        = []api.Player{
		{ID: 1, Name: "Jogador", Type: "user", Race: "Terr", Result: "Undecided"},
		{ID: 2, Name: "A.I. 1 (Very Easy)", Type: "computer", Race: "Zerg", Result: "Undecided"},
	}
)

// Sim cycles menu (3 s) -> loading (3 s) -> game (until length) -> menu.
// It is safe for concurrent use.
type Sim struct {
	mu       sync.Mutex
	clock    func() time.Time
	length   float64 // game seconds
	speed    float64
	paused   bool
	phase    phase
	inPhase  time.Duration // real time spent in menu/loading
	gameTime float64
	last     time.Time
}

// New creates a simulator. speed multiplies game time; length ends the match.
func New(now func() time.Time, speed float64, length time.Duration) *Sim {
	return &Sim{clock: now, speed: speed, length: length.Seconds(), last: now()}
}

// advance moves the simulation to the current real time. Caller holds mu.
func (s *Sim) advance() {
	t := s.clock()
	dt := t.Sub(s.last)
	s.last = t
	if s.paused || dt <= 0 {
		return
	}
	switch s.phase {
	case menu, loading:
		s.inPhase += dt
		if s.inPhase >= phaseLength {
			s.inPhase = 0
			s.phase++
		}
	case game:
		s.gameTime += dt.Seconds() * s.speed
		if s.gameTime >= s.length {
			s.phase, s.gameTime, s.inPhase = menu, 0, 0
		}
	}
}

func (s *Sim) locked(f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.advance()
	f()
}

// Pause freezes the simulation.
func (s *Sim) Pause() { s.locked(func() { s.paused = true }) }

// Resume continues after Pause.
func (s *Sim) Resume() { s.locked(func() { s.paused = false }) }

// Paused reports whether the simulation is paused.
func (s *Sim) Paused() (p bool) {
	s.locked(func() { p = s.paused })
	return p
}

// SetSpeed changes the game speed multiplier.
func (s *Sim) SetSpeed(speed float64) { s.locked(func() { s.speed = speed }) }

// Speed returns the game speed multiplier.
func (s *Sim) Speed() (v float64) {
	s.locked(func() { v = s.speed })
	return v
}

// NewMatch jumps to the loading screen of a new match starting at 0:00.
func (s *Sim) NewMatch() {
	s.locked(func() { s.phase, s.gameTime, s.inPhase = loading, 0, 0 })
}

// Handler serves /game and /ui in the SC2 Client API format.
func (s *Sim) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		g := api.Game{Players: []api.Player{}}
		s.locked(func() {
			if s.phase == game {
				g.DisplayTime = math.Floor(s.gameTime) // whole seconds, like the real API
				g.Players = players
			}
		})
		writeJSON(w, g)
	})
	mux.HandleFunc("/ui", func(w http.ResponseWriter, r *http.Request) {
		ui := api.UI{ActiveScreens: []string{}}
		s.locked(func() {
			switch s.phase {
			case menu:
				ui.ActiveScreens = menuScreens
			case loading:
				ui.ActiveScreens = loadingScreens
			}
		})
		writeJSON(w, ui)
	})
	return mux
}

// Start listens on addr (e.g. "localhost:6119") and serves until ctx is
// cancelled. It fails immediately when the port is taken.
func (s *Sim) Start(ctx context.Context, addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: s.Handler()}
	go srv.Serve(l)
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	return nil
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
