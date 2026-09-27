package build

import "strings"

// NormalizeAPIRace maps SC2 API races ("Terr", "Prot", "Zerg", "random")
// and full names to Terran, Protoss, Zerg or Random; unknown values to "".
func NormalizeAPIRace(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	for _, r := range Races {
		if strings.HasPrefix(strings.ToLower(r), s) || strings.HasPrefix(s, strings.ToLower(r)[:4]) {
			return r
		}
	}
	return ""
}

// Suggest returns the first valid build (in entries order) whose race is the
// player's race and whose vs is the opponent's race. Races may use the API
// abbreviations.
func Suggest(entries []Entry, myRace, oppRace string) (file string, ok bool) {
	me, opp := NormalizeAPIRace(myRace), NormalizeAPIRace(oppRace)
	if me == "" || opp == "" {
		return "", false
	}
	for _, e := range entries {
		if e.Err == nil && e.Build.Race == me && e.Build.Vs == opp {
			return e.File, true
		}
	}
	return "", false
}
