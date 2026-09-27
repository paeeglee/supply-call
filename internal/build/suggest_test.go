package build

import (
	"errors"
	"testing"
)

func entry(file, race, vs string) Entry {
	return Entry{File: file, Name: file, Build: &Build{Race: race, Vs: vs, File: file}}
}

func TestSuggest(t *testing.T) {
	entries := []Entry{
		entry("a_tvp.yml", "Terran", "Protoss"),
		entry("b_tvz.yml", "Terran", "Zerg"),
		entry("c_tvz_alt.yml", "Terran", "Zerg"),
		entry("d_pvz.yml", "Protoss", "Zerg"),
		entry("e_tvr.yml", "Terran", "Random"),
		entry("f_novs.yml", "Terran", ""),
		{File: "g_broken.yml", Name: "g_broken.yml", Err: errors.New("x")},
	}
	cases := []struct {
		me, opp string
		want    string
		ok      bool
	}{
		{"Terr", "Zerg", "b_tvz.yml", true}, // first match in file order
		{"Terr", "Prot", "a_tvp.yml", true}, // API abbreviations
		{"Prot", "Zerg", "d_pvz.yml", true},
		{"Terr", "random", "e_tvr.yml", true}, // opponent picked Random
		{"Terran", "Zerg", "b_tvz.yml", true}, // full names too
		{"Zerg", "Terr", "", false},           // AUTO-02: nothing matches
		{"Terr", "", "", false},               // unknown opponent race
	}
	for _, c := range cases {
		got, ok := Suggest(entries, c.me, c.opp)
		if got != c.want || ok != c.ok {
			t.Errorf("Suggest(%s vs %s) = %q,%v want %q,%v", c.me, c.opp, got, ok, c.want, c.ok)
		}
	}
}

func TestNormalizeAPIRace(t *testing.T) {
	cases := map[string]string{"Terr": "Terran", "Prot": "Protoss", "Zerg": "Zerg", "random": "Random", "Rand": "Random", "terran": "Terran", "": "", "Xel": ""}
	for in, want := range cases {
		if got := NormalizeAPIRace(in); got != want {
			t.Errorf("NormalizeAPIRace(%q) = %q, want %q", in, got, want)
		}
	}
}
