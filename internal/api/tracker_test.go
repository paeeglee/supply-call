package api

import "testing"

var (
	menuUI    = UI{ActiveScreens: []string{"ScreenBackgroundSC2/ScreenBackgroundSC2", "ScreenHome/ScreenHome"}}
	loadingUI = UI{ActiveScreens: []string{"ScreenLoading/ScreenLoading"}}
	gameUI    = UI{ActiveScreens: []string{}}
	players   = []Player{{ID: 1, Name: "Me", Race: "Terr"}, {ID: 2, Name: "AI", Race: "Zerg"}}
)

func inGame(d float64) Game { return Game{DisplayTime: d, Players: players} }

type step struct {
	ok      bool
	g       Game
	ui      UI
	replays bool
	kind    Kind
	conn    Conn
}

func run(t *testing.T, steps []step) {
	t.Helper()
	var tr Tracker
	for i, s := range steps {
		ev := tr.Update(s.ok, s.g, s.ui, s.replays)
		if ev.Kind != s.kind || ev.Conn != s.conn {
			t.Errorf("step %d: got kind %v conn %v, want kind %v conn %v", i, ev.Kind, ev.Conn, s.kind, s.conn)
		}
		if ev.Kind == MatchStart || ev.Kind == Reading {
			if ev.Game.DisplayTime != s.g.DisplayTime {
				t.Errorf("step %d: event time %v, want %v", i, ev.Game.DisplayTime, s.g.DisplayTime)
			}
		}
	}
}

func TestTrackerFullLifecycle(t *testing.T) {
	run(t, []step{
		{ok: false, kind: None, conn: WentOffline},                       // game closed at start
		{ok: false, kind: None, conn: NoChange},                          // still closed: no repeated log
		{ok: true, g: Game{}, ui: menuUI, kind: None, conn: WentOnline},  // menu
		{ok: true, g: Game{}, ui: loadingUI, kind: None, conn: NoChange}, // loading
		{ok: true, g: inGame(0.1), ui: gameUI, kind: MatchStart},         // API-02/03
		{ok: true, g: inGame(0.6), ui: gameUI, kind: Reading},            // readings
		{ok: true, g: inGame(1.1), ui: gameUI, kind: Reading},
		{ok: true, g: inGame(300), ui: menuUI, kind: MatchEnd},   // API-04 score screen
		{ok: true, g: inGame(300), ui: menuUI, kind: None},       // still out
		{ok: true, g: inGame(0.2), ui: gameUI, kind: MatchStart}, // next match
		{ok: false, kind: None, conn: NoChange},                  // 1st failure: tolerated
		{ok: false, kind: None, conn: NoChange},                  // 2nd failure: tolerated
		{ok: false, kind: MatchEnd, conn: WentOffline},           // 3rd: game closed mid-match
		{ok: false, kind: None, conn: NoChange},
	})
}

func TestTrackerDropIsNewMatch(t *testing.T) {
	// API-03: displayTime falls more than 2 s inside a match.
	run(t, []step{
		{ok: true, g: inGame(200), ui: gameUI, kind: MatchStart, conn: WentOnline},
		{ok: true, g: inGame(200.5), ui: gameUI, kind: Reading},
		{ok: true, g: inGame(199), ui: gameUI, kind: Reading}, // 1.5 s: jitter, same match
		{ok: true, g: inGame(3), ui: gameUI, kind: MatchStart},
	})
}

func TestTrackerNeedsPlayers(t *testing.T) {
	run(t, []step{
		{ok: true, g: Game{DisplayTime: 5}, ui: gameUI, kind: None, conn: WentOnline},
	})
}

func TestTrackerIgnoresReplaysByDefault(t *testing.T) {
	// API-06
	replay := Game{IsReplay: true, DisplayTime: 60, Players: players}
	run(t, []step{
		{ok: true, g: replay, ui: gameUI, kind: None, conn: WentOnline},
		{ok: true, g: replay, ui: gameUI, kind: None},
	})
}

func TestTrackerShowsReplaysWhenEnabled(t *testing.T) {
	// API-07
	replay := Game{IsReplay: true, DisplayTime: 60, Players: players}
	next := replay
	next.DisplayTime = 60.5
	run(t, []step{
		{ok: true, g: replay, ui: gameUI, replays: true, kind: MatchStart, conn: WentOnline},
		{ok: true, g: next, ui: gameUI, replays: true, kind: Reading},
	})
}

func TestTrackerToleratesTransientFailures(t *testing.T) {
	// API-05: one or two slow/failed polls in a match (seen on the real
	// API) keep the match going; the next reading continues it.
	run(t, []step{
		{ok: true, g: inGame(30), ui: gameUI, kind: MatchStart, conn: WentOnline},
		{ok: false, kind: None, conn: NoChange},
		{ok: false, kind: None, conn: NoChange},
		{ok: true, g: inGame(32), ui: gameUI, kind: Reading},
		{ok: false, kind: None, conn: NoChange}, // counter restarted
		{ok: false, kind: None, conn: NoChange},
		{ok: true, g: inGame(34), ui: gameUI, kind: Reading},
	})
}
