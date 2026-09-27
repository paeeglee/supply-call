// Package api reads the official SC2 Client API (http://localhost:6119).
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DefaultURL is where the SC2 client serves its API while the game is open.
const DefaultURL = "http://localhost:6119"

// Game is the /game response.
type Game struct {
	IsReplay    bool     `json:"isReplay"`
	DisplayTime float64  `json:"displayTime"`
	Players     []Player `json:"players"`
}

// Player is one entry of Game.Players.
type Player struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Race   string `json:"race"`
	Result string `json:"result"`
}

// UI is the /ui response. An empty ActiveScreens means the game view is shown.
type UI struct {
	ActiveScreens []string `json:"activeScreens"`
}

// Client calls the SC2 Client API.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// NewClient creates a client with a 1 s timeout per request (the real API
// takes 200-300 ms per request during a match).
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL, HTTP: &http.Client{Timeout: time.Second}}
}

// Game fetches /game.
func (c *Client) Game(ctx context.Context) (Game, error) {
	var g Game
	return g, c.get(ctx, "/game", &g)
}

// UI fetches /ui.
func (c *Client) UI(ctx context.Context) (UI, error) {
	var u UI
	return u, c.get(ctx, "/ui", &u)
}

func (c *Client) get(ctx context.Context, path string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", path, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}
