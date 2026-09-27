package sim

import (
	"context"
	"net"
	"net/http/httptest"
	"testing"
	"time"

	"sc2overlay/internal/api"
)

type fakeTime struct{ t time.Time }

func (f *fakeTime) now() time.Time      { return f.t }
func (f *fakeTime) adv(d time.Duration) { f.t = f.t.Add(d) }

func newSim(speed float64) (*Sim, *fakeTime) {
	ft := &fakeTime{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	return New(ft.now, speed, 10*time.Minute), ft
}

func snapshot(t *testing.T, s *Sim) (api.Game, api.UI) {
	t.Helper()
	srv := httptest.NewServer(s.Handler())
	defer srv.Close()
	c := api.NewClient(srv.URL)
	g, err := c.Game(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ui, err := c.UI(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return g, ui
}

func phaseOf(ui api.UI) string {
	switch {
	case len(ui.ActiveScreens) == 0:
		return "game"
	case ui.ActiveScreens[0] == "ScreenLoading/ScreenLoading":
		return "loading"
	default:
		return "menu"
	}
}

// toGame moves a fresh simulator from the menu into the match at 0:00.
func toGame(t *testing.T, s *Sim, ft *fakeTime) {
	t.Helper()
	ft.adv(3 * time.Second)
	snapshot(t, s)
	ft.adv(3 * time.Second)
	snapshot(t, s)
}

func TestSimCycle(t *testing.T) {
	// SIM-02: menu 3 s -> loading 3 s -> game until 10:00 -> menu.
	s, ft := newSim(1)
	check := func(wantPhase string, wantD float64) {
		t.Helper()
		g, ui := snapshot(t, s)
		if p := phaseOf(ui); p != wantPhase || g.DisplayTime != wantD {
			t.Fatalf("phase %s d=%v, want %s d=%v", p, g.DisplayTime, wantPhase, wantD)
		}
		if wantPhase == "game" && (len(g.Players) != 2 || g.IsReplay) {
			t.Fatalf("game players = %+v replay=%v", g.Players, g.IsReplay)
		}
		if wantPhase != "game" && len(g.Players) != 0 {
			t.Fatalf("players outside game = %+v", g.Players)
		}
	}
	check("menu", 0)
	ft.adv(3 * time.Second)
	check("loading", 0)
	ft.adv(3 * time.Second)
	check("game", 0)
	ft.adv(90 * time.Second)
	check("game", 90)
	ft.adv(510 * time.Second)
	check("menu", 0)
	ft.adv(3 * time.Second)
	check("loading", 0)
	ft.adv(3 * time.Second)
	check("game", 0) // repeats
}

func TestSimSpeed(t *testing.T) {
	// SIM-03
	for _, speed := range []float64{2, 4} {
		s, ft := newSim(speed)
		toGame(t, s, ft)
		ft.adv(10 * time.Second)
		if g, _ := snapshot(t, s); g.DisplayTime != 10*speed {
			t.Errorf("speed %v: d=%v, want %v", speed, g.DisplayTime, 10*speed)
		}
	}
}

func TestSimSetSpeedMidGame(t *testing.T) {
	s, ft := newSim(1)
	toGame(t, s, ft)
	ft.adv(10 * time.Second)
	s.SetSpeed(4)
	ft.adv(10 * time.Second)
	if g, _ := snapshot(t, s); g.DisplayTime != 50 {
		t.Errorf("d=%v, want 50", g.DisplayTime)
	}
}

func TestSimPause(t *testing.T) {
	// SIM-04
	s, ft := newSim(1)
	toGame(t, s, ft)
	ft.adv(20 * time.Second)
	s.Pause()
	ft.adv(30 * time.Second)
	if g, _ := snapshot(t, s); g.DisplayTime != 20 || !s.Paused() {
		t.Errorf("paused d=%v, want 20", g.DisplayTime)
	}
	s.Resume()
	ft.adv(5 * time.Second)
	if g, _ := snapshot(t, s); g.DisplayTime != 25 {
		t.Errorf("resumed d=%v, want 25", g.DisplayTime)
	}
}

func TestSimNewMatch(t *testing.T) {
	// SIM-05
	s, ft := newSim(1)
	toGame(t, s, ft)
	ft.adv(200 * time.Second)
	s.NewMatch()
	if _, ui := snapshot(t, s); phaseOf(ui) != "loading" {
		t.Fatalf("after NewMatch phase = %s, want loading", phaseOf(ui))
	}
	ft.adv(3 * time.Second)
	if g, ui := snapshot(t, s); phaseOf(ui) != "game" || g.DisplayTime != 0 {
		t.Errorf("phase %s d=%v, want game 0", phaseOf(ui), g.DisplayTime)
	}
}

func TestStartFailsWhenPortBusy(t *testing.T) {
	// SIM-06
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	s, _ := newSim(1)
	if err := s.Start(context.Background(), l.Addr().String()); err == nil {
		t.Error("want error on busy port")
	}
}

func TestStartServes(t *testing.T) {
	s, _ := newSim(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := l.Addr().String()
	l.Close()
	if err := s.Start(ctx, addr); err != nil {
		t.Fatal(err)
	}
	if _, err := api.NewClient("http://" + addr).UI(context.Background()); err != nil {
		t.Errorf("UI: %v", err)
	}
}
