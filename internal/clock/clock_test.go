package clock

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
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

func newTest() (*Clock, *fakeTime) {
	ft := &fakeTime{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	return New(ft.Now), ft
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

const ms = time.Millisecond

// feed simulates the real API: integer displayTime polled every 500 ms
// while the game runs at `rate` from game time `from` for `dur`.
func feed(c *Clock, ft *fakeTime, from float64, rate float64, dur time.Duration) {
	for el := time.Duration(0); el <= dur; el += 500 * ms {
		c.Observe(math.Floor(from + rate*el.Seconds()))
		if el < dur {
			ft.Advance(500 * ms)
		}
	}
}

func TestClockZeroBeforeReadings(t *testing.T) {
	c, _ := newTest()
	if got := c.Now(); got != 0 {
		t.Errorf("Now() = %v, want 0", got)
	}
}

func TestClockChangeShowsReadingPlusLead(t *testing.T) {
	// CLK-01: a new value is shown as D + rate*0.45 (rate 1 without history).
	c, _ := newTest()
	c.Observe(42)
	if got := c.Now(); !near(got, 42.45) {
		t.Errorf("Now() = %v, want 42.45", got)
	}
}

func TestClockInterpolatesBetweenIntegerReadings(t *testing.T) {
	// CLK-02 at 1x: readings of whole seconds do not freeze the clock.
	c, ft := newTest()
	feed(c, ft, 10, 1, 5*time.Second) // last reading floor(15) = 15, changed at +5s
	prev := c.Now()
	for i := 0; i < 8; i++ { // 100 ms steps without readings
		ft.Advance(100 * ms)
		now := c.Now()
		if now <= prev {
			t.Fatalf("clock stalled or went back: %v -> %v", prev, now)
		}
		prev = now
	}
	if !near(prev, 15+0.45+0.8) {
		t.Errorf("Now() = %v, want %v", prev, 15+0.45+0.8)
	}
}

func TestClockSameValueWithinOneSecondIsNotPause(t *testing.T) {
	// Two equal readings 0.5 s apart are normal with integer seconds.
	c, ft := newTest()
	feed(c, ft, 10, 1, 3*time.Second) // ends right after a change to 13
	ft.Advance(500 * ms)
	c.Observe(13) // same value, 0.5 s later
	if got := c.Now(); !near(got, 13+0.45+0.5) {
		t.Errorf("Now() = %v, want %v (running)", got, 13+0.45+0.5)
	}
}

func TestClockRateMeasuredAt4x(t *testing.T) {
	// CLK-02 with the simulator at 4x.
	c, ft := newTest()
	feed(c, ft, 0, 4, 6*time.Second) // last change: 24 at +6s
	ft.Advance(250 * ms)
	want := 24 + 4*(0.45+0.25)
	if got := c.Now(); math.Abs(got-want) > 0.05 {
		t.Errorf("Now() = %v, want ~%v", got, want)
	}
}

func TestClockRateDefaultsTo1WithShortHistory(t *testing.T) {
	c, ft := newTest()
	c.Observe(5)
	ft.Advance(500 * ms)
	c.Observe(10) // 5 units in 0.5 s, but less than 2 s of history
	ft.Advance(200 * ms)
	if got := c.Now(); !near(got, 10+0.45+0.2) {
		t.Errorf("Now() = %v, want %v", got, 10+0.45+0.2)
	}
}

func TestClockRateClampedTo8x(t *testing.T) {
	c, ft := newTest()
	feed(c, ft, 0, 20, 3*time.Second) // 60 at +3s
	ft.Advance(100 * ms)
	if got := c.Now(); !near(got, 60+8*(0.45+0.1)) {
		t.Errorf("Now() = %v, want %v", got, 60+8*(0.45+0.1))
	}
}

func TestClockRateUsesOnlyLast10Seconds(t *testing.T) {
	// CLK-02: speed is measured over the changes of the last 10 s, so a
	// switch from 1x to 4x is followed once the old readings leave the window.
	c, ft := newTest()
	feed(c, ft, 0, 1, 20*time.Second)  // 0..20 at 1x
	feed(c, ft, 20, 4, 12*time.Second) // 20..68 at 4x
	ft.Advance(250 * ms)
	want := 68 + 4*(0.45+0.25)
	if got := c.Now(); math.Abs(got-want) > 0.05 {
		t.Errorf("Now() = %v, want ~%v (4x from the last 10 s)", got, want)
	}
}

func TestClockRateFloor(t *testing.T) {
	// CLK-02: a measured speed below 0.25x is clamped to 0.25x. The game runs
	// at 0.125x for 16 s (fractional readings, so no pause is detected); long
	// enough for the interpolation to pass the value held from the first
	// seconds, when the speed was still assumed to be 1x.
	c, ft := newTest()
	for i := 0; i <= 32; i++ {
		c.Observe(0.0625 * float64(i))
		if i < 32 {
			ft.Advance(500 * ms)
		}
	}
	ft.Advance(time.Second)
	want := 2 + 0.25*(0.45+1.0) // without the floor: 2 + 0.125*1.45
	if got := c.Now(); !near(got, want) {
		t.Errorf("Now() = %v, want %v (0.25x floor)", got, want)
	}
}

func TestClockMissingReadingsDoNotPauseButCapExtrapolation(t *testing.T) {
	// CLK-02 / CLK-03: API silent (slow or failing) -> keep running up to
	// 3 s after the last change, then hold; never snap back.
	c, ft := newTest()
	feed(c, ft, 10, 1, 4*time.Second) // 14 at +4s
	ft.Advance(1200 * ms)
	if got := c.Now(); !near(got, 14+0.45+1.2) {
		t.Errorf("Now() = %v, want %v (still running)", got, 14+0.45+1.2)
	}
	ft.Advance(5 * time.Second)
	if got := c.Now(); !near(got, 14+0.45+3) {
		t.Errorf("Now() = %v, want %v (capped)", got, 14+0.45+3)
	}
}

func TestClockPauseFreezesAtReading(t *testing.T) {
	// CLK-03: same value confirmed more than 1.6 s after the last change.
	c, ft := newTest()
	feed(c, ft, 10, 1, 4*time.Second) // 14 at +4s
	for i := 0; i < 3; i++ {          // readings at +0.5, +1.0, +1.5: still running
		ft.Advance(500 * ms)
		c.Observe(14)
	}
	if got := c.Now(); got <= 14.5 {
		t.Errorf("at +1.5 s Now() = %v, want still running", got)
	}
	ft.Advance(200 * ms)
	c.Observe(14) // +1.7 s
	if got := c.Now(); !near(got, 14) {
		t.Errorf("paused Now() = %v, want exactly 14", got)
	}
	ft.Advance(20 * time.Second)
	c.Observe(14)
	if got := c.Now(); !near(got, 14) {
		t.Errorf("during pause Now() = %v, want 14", got)
	}
	// Resume: the pause does not distort the rate.
	ft.Advance(500 * ms)
	c.Observe(15)
	ft.Advance(100 * ms)
	if got := c.Now(); !near(got, 15+0.45+0.1) {
		t.Errorf("after resume Now() = %v, want %v", got, 15+0.45+0.1)
	}
}

func TestClockSmallBackwardReadingHolds(t *testing.T) {
	// CLK-01: a change whose target is below the shown value (less than 2 s)
	// keeps the shown value until the interpolation catches up.
	c, ft := newTest()
	feed(c, ft, 10, 1, 4*time.Second) // 14 at +4s
	ft.Advance(1400 * ms)             // shown 14 + 0.45 + 1.4 = 15.85
	c.Observe(15)                     // target 15.45
	if got := c.Now(); !near(got, 15.85) {
		t.Errorf("Now() = %v, want 15.85 (held)", got)
	}
	ft.Advance(500 * ms) // interpolation passes the held value
	rate := 5 / 5.4      // changes 10@0 s ... 15@5.4 s
	if want := 15 + rate*0.95; !near(c.Now(), want) {
		t.Errorf("Now() = %v, want %v", c.Now(), want)
	}
}

func TestClockBigDropIsNewMatch(t *testing.T) {
	// CLK-04: drop of more than 2 s -> new value shown.
	c, ft := newTest()
	feed(c, ft, 300, 1, 3*time.Second)
	ft.Advance(500 * ms)
	c.Observe(0)
	if got := c.Now(); !near(got, 0.45) {
		t.Errorf("Now() = %v, want 0.45", got)
	}
}

func TestClockReset(t *testing.T) {
	c, ft := newTest()
	feed(c, ft, 100, 1, 2*time.Second)
	c.Reset()
	if got := c.Now(); got != 0 {
		t.Errorf("after Reset Now() = %v, want 0", got)
	}
}

// TestClockRealMatchRecording replays /game readings recorded from the real
// game (integer displayTime, ~1.1 s between polls, one 29 s pause).
func TestClockRealMatchRecording(t *testing.T) {
	f, err := os.Open("testdata/real_match.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	type rd struct{ at, d float64 }
	var rs []rd
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.Fields(line)
		at, _ := strconv.ParseFloat(parts[0], 64)
		d, _ := strconv.ParseFloat(parts[1], 64)
		rs = append(rs, rd{at, d})
	}

	c, ft := newTest()
	start := ft.Now()
	prevShown := 0.0
	stalledSince := -1.0
	for i, r := range rs {
		// Walk real time in 50 ms frames up to this reading.
		for {
			el := ft.Now().Sub(start).Seconds()
			if el >= r.at {
				break
			}
			ft.Advance(50 * ms)
			shown := c.Now()
			last := rs[max(0, i-1)]
			paused := last.d == 113 || last.d == 0
			if shown < prevShown-1e-9 && !paused {
				t.Fatalf("t=%.2f: clock went back %.2f -> %.2f", el, prevShown, shown)
			}
			// While readings keep arriving the clock must keep moving.
			if shown == prevShown && !paused && last.d > 1 && el-last.at < 1.5 {
				if stalledSince < 0 {
					stalledSince = el
				} else if el-stalledSince > 1.2 {
					t.Fatalf("t=%.2f: clock stalled at %.2f for %.1f s while the game ran", el, shown, el-stalledSince)
				}
			} else {
				stalledSince = -1
			}
			if !paused && (shown < last.d || shown > last.d+4) {
				t.Fatalf("t=%.2f: shown %.2f too far from last reading %.0f", el, shown, last.d)
			}
			prevShown = shown
		}
		c.Observe(r.d)
	}
	// Inside the pause (from 3 s after it starts) the clock shows exactly 113.
	c2, ft2 := newTest()
	start2 := ft2.Now()
	pauseStart := -1.0
	for _, r := range rs {
		ft2.Advance(time.Duration((r.at-ft2.Now().Sub(start2).Seconds())*1e9) * time.Nanosecond)
		c2.Observe(r.d)
		if r.d == 113 && pauseStart < 0 {
			pauseStart = r.at
		}
		if r.d == 113 && r.at > pauseStart+3 {
			if got := c2.Now(); got != 113 {
				t.Fatalf("t=%.2f: paused at 113, clock shows %.2f", r.at, got)
			}
		}
	}
	if pauseStart < 0 {
		t.Fatal("recording has no pause at 113")
	}
}

func TestClockConcurrentAccess(t *testing.T) {
	c, ft := newTest()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			ft.Advance(ms)
			c.Observe(float64(i / 10))
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = c.Now()
		}
	}()
	wg.Wait()
}
