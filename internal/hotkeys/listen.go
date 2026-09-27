package hotkeys

import (
	"context"

	"golang.design/x/hotkey"
)

// Listen registers the global hotkey described by spec (see Parse) and calls
// fn on every key press until ctx is cancelled. It returns an error when spec
// is invalid or Windows refuses the registration (e.g. another program
// already uses the combination).
func Listen(ctx context.Context, spec string, fn func()) error {
	mods, key, err := Parse(spec)
	if err != nil {
		return err
	}
	hk := hotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		return err
	}
	go func() {
		defer hk.Unregister()
		for {
			select {
			case <-ctx.Done():
				return
			case <-hk.Keydown():
				fn()
			}
		}
	}()
	return nil
}
