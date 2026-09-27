package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunEmitsMatchEventsAndStopsOnCancel(t *testing.T) {
	var d atomic.Int64 // displayTime in tenths
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ui":
			json.NewEncoder(w).Encode(UI{ActiveScreens: []string{}})
		case "/game":
			json.NewEncoder(w).Encode(Game{DisplayTime: float64(d.Add(5)) / 10, Players: players})
		}
	}))
	defer srv.Close()

	var mu sync.Mutex
	var kinds []Kind
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		Run(ctx, NewClient(srv.URL), 5*time.Millisecond, func() bool { return false }, func(ev Event) {
			mu.Lock()
			kinds = append(kinds, ev.Kind)
			mu.Unlock()
		})
		close(done)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for {
		mu.Lock()
		n := len(kinds)
		mu.Unlock()
		if n >= 3 || time.Now().After(deadline) {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
	mu.Lock()
	defer mu.Unlock()
	if len(kinds) < 3 || kinds[0] != MatchStart || kinds[1] != Reading || kinds[2] != Reading {
		t.Errorf("kinds = %v, want MatchStart, Reading, Reading...", kinds)
	}
}

func TestRunReportsOfflineOnce(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()

	var mu sync.Mutex
	var events []Event
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	Run(ctx, NewClient(url), 5*time.Millisecond, func() bool { return false }, func(ev Event) {
		mu.Lock()
		events = append(events, ev)
		mu.Unlock()
	})
	offline := 0
	for _, ev := range events {
		if ev.Conn == WentOffline {
			offline++
		}
	}
	if len(events) < 2 || offline != 1 {
		t.Errorf("events = %d, WentOffline = %d, want >=2 polls and exactly 1 WentOffline", len(events), offline)
	}
}
