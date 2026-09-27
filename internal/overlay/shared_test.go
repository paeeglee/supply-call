package overlay

import (
	"sync"
	"testing"
	"time"

	"sc2overlay/internal/api"
	"sc2overlay/internal/build"
	"sc2overlay/internal/clock"
)

type fakeTime struct {
	mu sync.Mutex
	t  time.Time
}

func (f *fakeTime) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.t
}

func (f *fakeTime) Advance(d time.Duration) {
	f.mu.Lock()
	f.t = f.t.Add(d)
	f.mu.Unlock()
}

func approx(a, b float64) bool { return a-b < 1e-6 && b-a < 1e-6 }

func newShared() (*Shared, *fakeTime) {
	ft := &fakeTime{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	return NewShared(clock.New(ft.Now)), ft
}

func testBuild(t *testing.T) *build.Build {
	t.Helper()
	b, err := build.Parse([]byte("name: B\nrace: Terran\nsteps: [{time: \"0:35\", action: Depot}]"), "b.yml")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestSharedDefaults(t *testing.T) {
	s, _ := newShared()
	snap := s.Snapshot()
	if snap.Status != Waiting || !snap.Visible || !snap.ClickThrough || snap.Build != nil || snap.Timeline != nil {
		t.Errorf("defaults = %+v", snap)
	}
}

func TestSharedSetBuildBuildsTimeline(t *testing.T) {
	s, _ := newShared()
	b := testBuild(t)
	s.SetBuild(b)
	snap := s.Snapshot()
	if snap.Build != b || snap.Timeline == nil || len(snap.Timeline.Groups) != 1 {
		t.Errorf("snapshot = %+v", snap)
	}
	s.SetBuild(nil)
	if snap := s.Snapshot(); snap.Build != nil || snap.Timeline != nil {
		t.Errorf("after SetBuild(nil) = %+v", snap)
	}
}

func TestSharedMatchEventsDriveClockAndStatus(t *testing.T) {
	s, ft := newShared()
	// The clock shows a new reading plus its 0.45 s lead (see package clock).
	s.HandleEvent(api.Event{Kind: api.MatchStart, Game: api.Game{DisplayTime: 1}})
	if snap := s.Snapshot(); snap.Status != InMatch || !approx(snap.Time, 1.45) {
		t.Fatalf("after MatchStart = %+v", snap)
	}
	ft.Advance(time.Second)
	s.HandleEvent(api.Event{Kind: api.Reading, Game: api.Game{DisplayTime: 2}})
	ft.Advance(250 * time.Millisecond)
	if got := s.Snapshot().Time; !approx(got, 2.7) {
		t.Errorf("interpolated time = %v, want 2.7", got)
	}
	// STEP-06: switching build keeps the clock.
	s.SetBuild(testBuild(t))
	if got := s.Snapshot().Time; !approx(got, 2.7) {
		t.Errorf("time after SetBuild = %v, want 2.7", got)
	}
	// API-04: leaving the match resets to 0 and waits.
	s.HandleEvent(api.Event{Kind: api.MatchEnd})
	if snap := s.Snapshot(); snap.Status != Waiting || snap.Time != 0 {
		t.Errorf("after MatchEnd = %+v", snap)
	}
	// API-03: a new match starts from its own time.
	s.HandleEvent(api.Event{Kind: api.MatchStart, Game: api.Game{DisplayTime: 0}})
	if snap := s.Snapshot(); snap.Status != InMatch || !approx(snap.Time, 0.45) {
		t.Errorf("after second MatchStart = %+v", snap)
	}
}

func TestSharedPortBusyStatus(t *testing.T) {
	s, _ := newShared()
	s.SetStatus(PortBusy)
	s.HandleEvent(api.Event{Conn: api.WentOffline})
	if got := s.Snapshot().Status; got != PortBusy {
		t.Errorf("status = %v, want PortBusy kept", got)
	}
}

func TestSharedToggles(t *testing.T) {
	s, _ := newShared()
	s.ToggleClickThrough()
	// Clicks reach the overlay only when it is visible and click-through is off.
	if snap := s.Snapshot(); snap.ClickThrough || snap.Passthrough() {
		t.Errorf("after ToggleClickThrough = %+v passthrough=%v", snap, snap.Passthrough())
	}
	s.ToggleVisible()
	snap := s.Snapshot()
	if snap.Visible || !snap.Passthrough() {
		t.Errorf("hidden must be passthrough: %+v passthrough=%v", snap, snap.Passthrough())
	}
	s.ToggleVisible()
	if snap := s.Snapshot(); !snap.Visible || snap.Passthrough() {
		t.Errorf("visible again with click-through off: %+v passthrough=%v", snap, snap.Passthrough())
	}
}

func TestSharedRequestQuitIsIdempotent(t *testing.T) {
	s, _ := newShared()
	s.RequestQuit()
	s.RequestQuit()
	select {
	case <-s.Quit():
	default:
		t.Error("Quit channel not closed")
	}
}

func TestSharedConcurrentAccess(t *testing.T) {
	s, _ := newShared()
	b := testBuild(t)
	var wg sync.WaitGroup
	writers := []func(){
		func() { s.SetBuild(b) },
		func() { s.HandleEvent(api.Event{Kind: api.Reading, Game: api.Game{DisplayTime: 3}}) },
		func() { s.ToggleVisible() },
		func() { s.ToggleClickThrough() },
		func() { s.SetStatus(InMatch) },
	}
	for _, w := range writers {
		wg.Add(1)
		go func(w func()) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				w()
			}
		}(w)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = s.Snapshot()
		}
	}()
	wg.Wait()
}
