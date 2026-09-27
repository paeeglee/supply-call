package build

import "testing"

func timelineOf(t *testing.T, src string) *Timeline {
	t.Helper()
	b, err := Parse([]byte(src), "t.yml")
	if err != nil {
		t.Fatal(err)
	}
	return NewTimeline(b)
}

const simple = `
race: Terran
steps:
  - {time: "0:35", action: Depot}
  - {time: "0:48", action: Refinaria}
  - {time: "0:48", action: Barracks}
  - {time: "1:06", action: Factory}
`

func TestTimelineGroupsSameTime(t *testing.T) {
	tl := timelineOf(t, simple)
	if len(tl.Groups) != 3 {
		t.Fatalf("groups = %d, want 3", len(tl.Groups))
	}
	g := tl.Groups[1]
	if g.Time != 48 || len(g.Steps) != 2 || g.Steps[0].Action != "Refinaria" || g.Steps[1].Action != "Barracks" {
		t.Errorf("group 1 = %+v", g)
	}
}

func states(s Status) []State { return s.States }

func eq(a, b []State) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTimelineAt(t *testing.T) {
	tl := timelineOf(t, simple)
	cases := []struct {
		name      string
		t         float64
		want      []State
		current   int
		countdown int
		done      bool
	}{
		{"before start (t<0)", -2, []State{Current, Future, Future}, 0, 37, false},
		{"start", 0, []State{Current, Future, Future}, 0, 35, false},
		{"countdown rounds up", 29.2, []State{Current, Future, Future}, 0, 6, false},
		{"warning at exactly 5s", 30, []State{Warning, Future, Future}, 0, 5, false},
		{"warning at 0.5s", 34.5, []State{Warning, Future, Future}, 0, 1, false},
		{"now at its time", 35, []State{Now, Future, Future}, 0, 0, false},
		{"still now at 2.9s", 37.9, []State{Now, Future, Future}, 0, 0, false},
		{"passed at exactly +3s", 38, []State{Passed, Current, Future}, 1, 10, false},
		{"grouped now", 48, []State{Passed, Now, Future}, 1, 0, false},
		{"last now", 67, []State{Passed, Passed, Now}, 2, 0, false},
		{"done at last+3", 69, []State{Passed, Passed, Passed}, -1, 0, true},
		{"done long after", 600, []State{Passed, Passed, Passed}, -1, 0, true},
	}
	for _, c := range cases {
		s := tl.At(c.t, 5)
		if !eq(states(s), c.want) || s.Current != c.current || s.Countdown != c.countdown || s.Done != c.done {
			t.Errorf("%s: At(%v) = %+v, want states %v current %d countdown %d done %v",
				c.name, c.t, s, c.want, c.current, c.countdown, c.done)
		}
	}
}

func TestTimelineCloseGroupsLaterWins(t *testing.T) {
	tl := timelineOf(t, `
race: Terran
steps:
  - {time: "3:52", action: Starport}
  - {time: "3:54", action: Refinaria}
`)
	// 232 = 3:52, 234 = 3:54. At 234 the first is still inside its 3 s window,
	// but the later group has reached its time, so the first becomes passed.
	s := tl.At(234, 5)
	if !eq(s.States, []State{Passed, Now}) || s.Current != 1 {
		t.Errorf("At(234) = %+v, want [Passed Now] current 1", s)
	}
	s = tl.At(233, 5)
	if !eq(s.States, []State{Now, Warning}) {
		t.Errorf("At(233) = %+v, want [Now Warning]", s)
	}
}

func TestTimelineWarningSecondsConfigurable(t *testing.T) {
	tl := timelineOf(t, simple)
	if s := tl.At(25, 10); s.States[0] != Warning {
		t.Errorf("warn=10 at t=25 (10s left) = %v, want Warning", s.States[0])
	}
	if s := tl.At(25, 5); s.States[0] != Current {
		t.Errorf("warn=5 at t=25 (10s left) = %v, want Current", s.States[0])
	}
}

func TestTimelineSwitchBuildMidMatchUsesCurrentTime(t *testing.T) {
	// STEP-06: a new build evaluated at the same game time, no reset.
	other := timelineOf(t, `
race: Protoss
steps:
  - {time: "0:20", action: Pylon}
  - {time: "1:30", action: Gateway}
`)
	s := other.At(80, 5)
	if !eq(s.States, []State{Passed, Current}) || s.Countdown != 10 {
		t.Errorf("At(80) = %+v, want [Passed Current] countdown 10", s)
	}
}
