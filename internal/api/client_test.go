package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func serve(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return NewClient(srv.URL)
}

func TestClientGame(t *testing.T) {
	c := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/game" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(`{"isReplay":false,"displayTime":123.5,"players":[
			{"id":1,"name":"Mateus","type":"user","race":"Terr","result":"Undecided"},
			{"id":2,"name":"A.I. 1 (Very Easy)","type":"computer","race":"Zerg","result":"Undecided"}],
			"extra":{"x":1}}`))
	})
	g, err := c.Game(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if g.IsReplay || g.DisplayTime != 123.5 || len(g.Players) != 2 {
		t.Fatalf("game = %+v", g)
	}
	p := g.Players[1]
	if p.ID != 2 || p.Name != "A.I. 1 (Very Easy)" || p.Type != "computer" || p.Race != "Zerg" || p.Result != "Undecided" {
		t.Errorf("player = %+v", p)
	}
}

func TestClientUI(t *testing.T) {
	c := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ui" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(`{"activeScreens":["ScreenLoading/ScreenLoading"]}`))
	})
	ui, err := c.UI(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ui.ActiveScreens) != 1 || ui.ActiveScreens[0] != "ScreenLoading/ScreenLoading" {
		t.Errorf("ui = %+v", ui)
	}
}

func TestClientHTTPError(t *testing.T) {
	c := serve(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	if _, err := c.Game(context.Background()); err == nil {
		t.Error("want error on 500")
	}
}

func TestClientInvalidJSON(t *testing.T) {
	c := serve(t, func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"displayTime":`)) })
	if _, err := c.Game(context.Background()); err == nil {
		t.Error("want error on invalid JSON")
	}
}

func TestClientTimeout(t *testing.T) {
	c := serve(t, func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() })
	c.HTTP.Timeout = 50 * time.Millisecond
	start := time.Now()
	if _, err := c.UI(context.Background()); err == nil {
		t.Error("want timeout error")
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Errorf("took %v", d)
	}
}

func TestClientDefaultTimeout(t *testing.T) {
	if got := NewClient(DefaultURL).HTTP.Timeout; got != time.Second {
		t.Errorf("timeout = %v, want 1s", got)
	}
}

func TestClientServerDown(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	if _, err := NewClient(url).Game(context.Background()); err == nil {
		t.Error("want error when server is down")
	}
}
