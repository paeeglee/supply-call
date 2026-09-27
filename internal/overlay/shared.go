package overlay

import (
	"sync"

	"sc2overlay/internal/api"
	"sc2overlay/internal/build"
	"sc2overlay/internal/clock"
)

// Status is what the overlay header shows besides the build.
type Status int

const (
	Waiting  Status = iota // "Aguardando partida…"
	InMatch                // clock running
	PortBusy               // -sim could not bind 6119
)

func (s Status) String() string {
	return [...]string{"Waiting", "InMatch", "PortBusy"}[s]
}

// Shared is the state exchanged between the poller, tray, hotkeys and the
// Ebitengine loop. All methods are safe for concurrent use.
type Shared struct {
	clock *clock.Clock

	mu           sync.Mutex
	build        *build.Build
	timeline     *build.Timeline
	status       Status
	visible      bool
	clickThrough bool

	quitOnce sync.Once
	quit     chan struct{}
}

// Snapshot is a consistent copy of Shared for one frame.
type Snapshot struct {
	Build        *build.Build
	Timeline     *build.Timeline
	Status       Status
	Time         float64 // game seconds
	Visible      bool
	ClickThrough bool
}

// Passthrough reports whether mouse clicks should go to the window below.
// A hidden overlay always lets clicks through.
func (s Snapshot) Passthrough() bool { return !s.Visible || s.ClickThrough }

// NewShared creates the shared state: visible, click-through on, waiting.
func NewShared(c *clock.Clock) *Shared {
	return &Shared{clock: c, visible: true, clickThrough: true, quit: make(chan struct{})}
}

// SetBuild switches the build (nil = none). The clock is not touched, so the
// new build applies at the current game time.
func (s *Shared) SetBuild(b *build.Build) {
	var tl *build.Timeline
	if b != nil {
		tl = build.NewTimeline(b)
	}
	s.mu.Lock()
	s.build, s.timeline = b, tl
	s.mu.Unlock()
}

// SetStatus sets the header status.
func (s *Shared) SetStatus(st Status) {
	s.mu.Lock()
	s.status = st
	s.mu.Unlock()
}

// HandleEvent applies a poller event to the clock and status.
func (s *Shared) HandleEvent(ev api.Event) {
	switch ev.Kind {
	case api.MatchStart:
		s.clock.Reset()
		s.clock.Observe(ev.Game.DisplayTime)
		s.SetStatus(InMatch)
	case api.Reading:
		s.clock.Observe(ev.Game.DisplayTime)
	case api.MatchEnd:
		s.clock.Reset()
		s.setWaiting()
	}
	if ev.Conn == api.WentOffline {
		s.setWaiting()
	}
}

// setWaiting switches to Waiting unless a port error must stay visible.
func (s *Shared) setWaiting() {
	s.mu.Lock()
	if s.status != PortBusy {
		s.status = Waiting
	}
	s.mu.Unlock()
}

// ToggleVisible shows or hides the overlay.
func (s *Shared) ToggleVisible() {
	s.mu.Lock()
	s.visible = !s.visible
	s.mu.Unlock()
}

// ToggleClickThrough switches between click-through and move mode.
func (s *Shared) ToggleClickThrough() {
	s.mu.Lock()
	s.clickThrough = !s.clickThrough
	s.mu.Unlock()
}

// RequestQuit asks the Ebitengine loop to end. Safe to call more than once.
func (s *Shared) RequestQuit() { s.quitOnce.Do(func() { close(s.quit) }) }

// Quit is closed after RequestQuit.
func (s *Shared) Quit() <-chan struct{} { return s.quit }

// Snapshot returns the current state.
func (s *Shared) Snapshot() Snapshot {
	t := s.clock.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	return Snapshot{
		Build: s.build, Timeline: s.timeline, Status: s.status, Time: t,
		Visible: s.visible, ClickThrough: s.clickThrough,
	}
}
