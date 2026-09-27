package api

import (
	"context"
	"log"
	"time"
)

// Run polls /ui and /game (in parallel) every interval until ctx is cancelled, feeding a
// Tracker and passing every resulting event to sink. Availability changes
// are logged once per transition, never once per failed poll.
func Run(ctx context.Context, c *Client, interval time.Duration, showReplays func() bool, sink func(Event)) {
	var tr Tracker
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		// Both requests in parallel: each takes 200-300 ms on the real API.
		var g Game
		var errGame error
		done := make(chan struct{})
		go func() {
			g, errGame = c.Game(ctx)
			close(done)
		}()
		ui, errUI := c.UI(ctx)
		<-done
		if ctx.Err() != nil {
			return
		}
		ev := tr.Update(errUI == nil && errGame == nil, g, ui, showReplays())
		switch ev.Conn {
		case WentOnline:
			log.Print("SC2 API disponível")
		case WentOffline:
			log.Printf("SC2 API indisponível (%v)", firstErr(errUI, errGame))
		}
		sink(ev)

		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}

func firstErr(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}
