// Package clock keeps a smooth game clock between SC2 API readings.
//
// The real API reports displayTime in whole seconds and answers in 200-300 ms,
// so two consecutive polls often return the same value while the game runs.
// The clock therefore interpolates from the moment a value changes, measures
// the game speed over the recent changes, and only treats "no change" as a
// pause when a fresh reading confirms it long after the last change.
package clock

import (
	"math"
	"sync"
	"time"
)

const (
	// lead compensates the average delay between a displayTime change in
	// the game and the poll that sees it (half the 500 ms interval plus the
	// API latency).
	lead = 450 * time.Millisecond
	// maxExtrap: interpolate at most this long after the last change.
	maxExtrap = 3 * time.Second
	// pauseAfter: a reading with an unchanged value this long after the last
	// change means the game is paused.
	pauseAfter = 1600 * time.Millisecond
	// rateWindow and minRateSpan: game speed is measured over the changes in
	// the last 10 s once they span at least 2 s; 1x before that.
	rateWindow  = 10 * time.Second
	minRateSpan = 2 * time.Second
	minRate     = 0.25
	maxRate     = 8.0
	// newMatchDrop: a drop larger than this (seconds) is a new match.
	newMatchDrop = 2.0
)

type change struct {
	at time.Time
	d  float64
}

// Clock interpolates the game time between readings. It is safe for
// concurrent use.
type Clock struct {
	mu      sync.Mutex
	now     func() time.Time
	has     bool
	paused  bool
	changes []change // recent value changes, oldest first; last = anchor
	hold    float64  // shown value is never below this while running
}

// New creates a clock using now as the real-time source.
func New(now func() time.Time) *Clock {
	return &Clock{now: now}
}

// Observe records a displayTime reading from the API.
func (c *Clock) Observe(d float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	at := c.now()
	if !c.has {
		c.restart(d, at)
		return
	}
	shown := c.value(at)
	anchor := c.changes[len(c.changes)-1]
	switch {
	case d == anchor.d:
		if at.Sub(anchor.at) > pauseAfter {
			c.paused = true
		}
	case d < anchor.d-newMatchDrop:
		c.restart(d, at)
	case c.paused:
		// Resumed: the pause must not count in the measured speed.
		c.restart(d, at)
	default:
		c.changes = append(c.changes, change{at, d})
		for len(c.changes) > 2 && at.Sub(c.changes[0].at) > rateWindow {
			c.changes = c.changes[1:]
		}
		c.hold = shown
	}
}

// Reset forgets all readings; Now returns 0 until the next reading.
func (c *Clock) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.has, c.paused, c.changes, c.hold = false, false, nil, 0
}

// Now returns the game time to display, in seconds.
func (c *Clock) Now() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.has {
		return 0
	}
	return c.value(c.now())
}

func (c *Clock) restart(d float64, at time.Time) {
	c.has, c.paused = true, false
	c.changes = []change{{at, d}}
	c.hold = 0
}

// rate is the game speed measured over the recent changes.
func (c *Clock) rate() float64 {
	first, last := c.changes[0], c.changes[len(c.changes)-1]
	span := last.at.Sub(first.at)
	if span < minRateSpan {
		return 1
	}
	return math.Min(maxRate, math.Max(minRate, (last.d-first.d)/span.Seconds()))
}

func (c *Clock) value(at time.Time) float64 {
	anchor := c.changes[len(c.changes)-1]
	if c.paused {
		return anchor.d
	}
	el := min(max(at.Sub(anchor.at), 0), maxExtrap)
	return math.Max(c.hold, anchor.d+c.rate()*(el+lead).Seconds())
}
