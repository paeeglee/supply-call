package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// realServer serves the /ui and /game responses recorded from the real game
// (SC2, 2026-09-26) for one screen: "menu", "loading" or "game".
func realServer(t *testing.T, screen *string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := map[string]string{"/ui": "_ui.json", "/game": "_game.json"}[r.URL.Path]
		data, err := os.ReadFile(filepath.Join("testdata", *screen+name))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(data)
	}))
	t.Cleanup(srv.Close)
	return NewClient(srv.URL)
}

func TestRealResponsesDecode(t *testing.T) {
	screen := "game"
	c := realServer(t, &screen)
	g, err := c.Game(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if g.IsReplay || g.DisplayTime != 60 || len(g.Players) != 2 {
		t.Fatalf("game = %+v", g)
	}
	me, opp, ok := FindOpponent(g.Players, "paeeglee")
	if !ok || me.Race != "Terr" || me.Type != "user" || opp.Type != "computer" || opp.Race != "random" || opp.Result != "Undecided" {
		t.Errorf("players = %+v / %+v", me, opp)
	}
}

func TestTrackerWithRealResponses(t *testing.T) {
	// API-02 / API-03 / API-04 with the real screens: players already appear
	// during loading, but the match only starts when activeScreens is empty.
	screen := "menu"
	c := realServer(t, &screen)
	var tr Tracker
	poll := func() Event {
		ui, err1 := c.UI(context.Background())
		g, err2 := c.Game(context.Background())
		return tr.Update(err1 == nil && err2 == nil, g, ui, false)
	}
	steps := []struct {
		screen string
		kind   Kind
	}{
		{"menu", None},
		{"loading", None},
		{"game", MatchStart},
		{"game", Reading},
		{"menu", MatchEnd},
	}
	for i, s := range steps {
		screen = s.screen
		if ev := poll(); ev.Kind != s.kind {
			t.Errorf("step %d (%s): kind %v, want %v", i, s.screen, ev.Kind, s.kind)
		}
	}
}
