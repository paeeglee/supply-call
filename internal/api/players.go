package api

import "strings"

// FindOpponent finds the player named name (case insensitive) and the first
// other player. ok is false when either is missing.
func FindOpponent(players []Player, name string) (me, opp Player, ok bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return me, opp, false
	}
	foundMe, foundOpp := false, false
	for _, p := range players {
		switch {
		case !foundMe && strings.EqualFold(p.Name, name):
			me, foundMe = p, true
		case !foundOpp:
			opp, foundOpp = p, true
		}
	}
	return me, opp, foundMe && foundOpp
}
