// Package hotkeys parses and registers the global hotkeys.
package hotkeys

import (
	"fmt"
	"strconv"
	"strings"

	"golang.design/x/hotkey"
)

var modifiers = map[string]hotkey.Modifier{
	"ctrl":    hotkey.ModCtrl,
	"control": hotkey.ModCtrl,
	"alt":     hotkey.ModAlt,
	"shift":   hotkey.ModShift,
	"win":     hotkey.ModWin,
}

// Parse reads "Mod+Mod+Key" (e.g. "Ctrl+Alt+H"). Modifiers: Ctrl, Alt,
// Shift, Win. Keys: A-Z, 0-9, F1-F24. Case and spaces are ignored.
func Parse(s string) ([]hotkey.Modifier, hotkey.Key, error) {
	parts := strings.Split(s, "+")
	var mods []hotkey.Modifier
	for _, p := range parts[:len(parts)-1] {
		m, ok := modifiers[strings.ToLower(strings.TrimSpace(p))]
		if !ok {
			return nil, 0, fmt.Errorf("atalho %q: modificador %q inválido (use Ctrl, Alt, Shift ou Win)", s, strings.TrimSpace(p))
		}
		for _, have := range mods {
			if have == m {
				return nil, 0, fmt.Errorf("atalho %q: modificador repetido", s)
			}
		}
		mods = append(mods, m)
	}
	key, err := parseKey(strings.ToUpper(strings.TrimSpace(parts[len(parts)-1])))
	if err != nil {
		return nil, 0, fmt.Errorf("atalho %q: %v", s, err)
	}
	return mods, key, nil
}

// parseKey maps an upper-case key name to its Windows virtual-key code.
func parseKey(k string) (hotkey.Key, error) {
	switch {
	case len(k) == 1 && k[0] >= 'A' && k[0] <= 'Z', len(k) == 1 && k[0] >= '0' && k[0] <= '9':
		return hotkey.Key(k[0]), nil // VK codes equal ASCII for A-Z and 0-9
	case len(k) >= 2 && k[0] == 'F':
		n, err := strconv.Atoi(k[1:])
		if err == nil && n >= 1 && n <= 24 {
			return hotkey.KeyF1 + hotkey.Key(n-1), nil
		}
	}
	if k == "" {
		return 0, fmt.Errorf("falta a tecla")
	}
	return 0, fmt.Errorf("tecla %q inválida (use A-Z, 0-9 ou F1-F24)", k)
}
