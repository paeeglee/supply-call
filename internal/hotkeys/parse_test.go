package hotkeys

import (
	"slices"
	"testing"

	"golang.design/x/hotkey"
)

func TestParseValid(t *testing.T) {
	cases := []struct {
		in   string
		mods []hotkey.Modifier
		key  hotkey.Key
	}{
		{"Ctrl+Alt+H", []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModAlt}, hotkey.KeyH},
		{"ctrl+alt+m", []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModAlt}, hotkey.KeyM},
		{" Shift + Win + 7 ", []hotkey.Modifier{hotkey.ModShift, hotkey.ModWin}, hotkey.Key7},
		{"Control+F9", []hotkey.Modifier{hotkey.ModCtrl}, hotkey.KeyF9},
		{"F11", nil, hotkey.KeyF11},
		{"Alt+F1", []hotkey.Modifier{hotkey.ModAlt}, hotkey.KeyF1},
		{"Ctrl+F24", []hotkey.Modifier{hotkey.ModCtrl}, hotkey.Key(0x87)}, // VK_F24
		{"Ctrl+Z", []hotkey.Modifier{hotkey.ModCtrl}, hotkey.KeyZ},
		{"Ctrl+0", []hotkey.Modifier{hotkey.ModCtrl}, hotkey.Key0},
	}
	for _, c := range cases {
		mods, key, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q): %v", c.in, err)
			continue
		}
		if !slices.Equal(mods, c.mods) || key != c.key {
			t.Errorf("Parse(%q) = %v %#x, want %v %#x", c.in, mods, key, c.mods, c.key)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, in := range []string{
		"",            // empty
		"Ctrl+Alt",    // no key
		"Ctrl+",       // empty key
		"Hyper+H",     // unknown modifier
		"Ctrl+Ctrl+H", // repeated modifier
		"Ctrl+H+J",    // two keys
		"Ctrl+F25",    // out of range
		"Ctrl+F0",     // out of range
		"Ctrl+Esc",    // unsupported key
		"Ctrl+ç",      // unsupported key
	} {
		if _, _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) accepted, want error", in)
		}
	}
}
