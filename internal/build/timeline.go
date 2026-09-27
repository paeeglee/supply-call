package build

// NowWindow is how long (seconds) a step stays in the "now" state after its time.
const NowWindow = 3

// State of a step group at a given game time.
type State int

const (
	Future  State = iota // after the current group
	Current              // next group to happen, with countdown
	Warning              // current group within the warning window
	Now                  // its time has come, for NowWindow seconds
	Passed               // done
)

func (s State) String() string {
	return [...]string{"Future", "Current", "Warning", "Now", "Passed"}[s]
}

// Group is a set of steps sharing the same time.
type Group struct {
	Time  int
	Steps []Step
}

// Timeline is a build's steps grouped by time.
type Timeline struct {
	Groups []Group
}

// Status is the state of every group at a given time.
type Status struct {
	States    []State
	Current   int  // index of the highlighted group (Current, Warning or Now); -1 when done
	Countdown int  // whole seconds until the Current/Warning group
	Done      bool // every group passed
}

// NewTimeline groups the (already sorted) steps of b by time.
func NewTimeline(b *Build) *Timeline {
	tl := &Timeline{}
	for _, s := range b.Steps {
		n := len(tl.Groups)
		if n > 0 && tl.Groups[n-1].Time == s.Time {
			tl.Groups[n-1].Steps = append(tl.Groups[n-1].Steps, s)
			continue
		}
		tl.Groups = append(tl.Groups, Group{Time: s.Time, Steps: []Step{s}})
	}
	return tl
}

// At computes the state of every group at game time t (seconds). A group
// whose countdown is at most warn seconds is in the Warning state.
func (tl *Timeline) At(t float64, warn int) Status {
	st := Status{States: make([]State, len(tl.Groups)), Current: -1}

	// Last group whose time has been reached.
	reached := -1
	for i, g := range tl.Groups {
		if float64(g.Time) <= t {
			reached = i
		}
	}
	for i := 0; i <= reached; i++ {
		st.States[i] = Passed
	}

	if reached >= 0 && t < float64(tl.Groups[reached].Time+NowWindow) {
		st.States[reached] = Now
		st.Current = reached
		if next := reached + 1; next < len(tl.Groups) {
			// A following group may already be inside its own warning window.
			if Countdown(float64(tl.Groups[next].Time)-t) <= warn {
				st.States[next] = Warning
			}
		}
		return st
	}

	next := reached + 1
	if next >= len(tl.Groups) {
		st.Done = true
		return st
	}
	st.Current = next
	st.Countdown = Countdown(float64(tl.Groups[next].Time) - t)
	st.States[next] = Current
	if st.Countdown <= warn {
		st.States[next] = Warning
	}
	return st
}
