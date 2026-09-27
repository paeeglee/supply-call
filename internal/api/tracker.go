package api

// Kind is what happened to the match on an update.
type Kind int

const (
	None       Kind = iota // nothing to do (out of match, or offline)
	MatchStart             // a new match started: reset clock and build
	Reading                // a new displayTime reading inside the match
	MatchEnd               // left the match (or the API went away)
)

func (k Kind) String() string {
	return [...]string{"None", "MatchStart", "Reading", "MatchEnd"}[k]
}

// Conn reports API availability changes, for logging.
type Conn int

const (
	NoChange Conn = iota
	WentOnline
	WentOffline
)

func (c Conn) String() string {
	return [...]string{"NoChange", "WentOnline", "WentOffline"}[c]
}

// Event is the result of one poll.
type Event struct {
	Kind Kind
	Conn Conn
	Game Game // last /game response (valid for MatchStart and Reading)
}

type connState int

const (
	unknown connState = iota
	online
	offline
)

// newMatchDrop: a displayTime drop larger than this (seconds) is a new match.
const newMatchDrop = 2.0

// OfflineAfter consecutive failed polls mark the API as unavailable. The real
// API sometimes takes more than 1 s to answer during a match; a single slow
// poll must not end the match.
const OfflineAfter = 3

// Tracker turns API snapshots into match lifecycle events. The zero value is
// ready to use. It is not safe for concurrent use.
type Tracker struct {
	conn     connState
	inMatch  bool
	lastD    float64
	failures int // consecutive failed polls
}

// Update feeds one poll result. ok is false when either request failed.
func (t *Tracker) Update(ok bool, g Game, ui UI, showReplays bool) Event {
	var ev Event
	if !ok {
		t.failures++
		if t.failures < OfflineAfter && t.conn != unknown {
			return ev // transient failure: keep the current state
		}
		if t.conn != offline {
			ev.Conn = WentOffline
			t.conn = offline
		}
		if t.inMatch {
			ev.Kind = MatchEnd
			t.inMatch = false
		}
		return ev
	}
	t.failures = 0
	if t.conn != online {
		ev.Conn = WentOnline
		t.conn = online
	}

	playing := len(ui.ActiveScreens) == 0 && len(g.Players) > 0 && (!g.IsReplay || showReplays)
	switch {
	case playing && (!t.inMatch || g.DisplayTime < t.lastD-newMatchDrop):
		ev.Kind = MatchStart
	case playing:
		ev.Kind = Reading
	case t.inMatch:
		ev.Kind = MatchEnd
	}
	t.inMatch = playing
	if playing {
		t.lastD = g.DisplayTime
		ev.Game = g
	}
	return ev
}
