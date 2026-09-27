package api

import "testing"

func TestFindOpponent(t *testing.T) {
	ps := []Player{
		{ID: 1, Name: "Mateus", Race: "Terr"},
		{ID: 2, Name: "Rival", Race: "Zerg"},
	}
	me, opp, ok := FindOpponent(ps, "mateus")
	if !ok || me.ID != 1 || opp.ID != 2 || opp.Race != "Zerg" {
		t.Errorf("got %+v %+v %v", me, opp, ok)
	}
	if _, _, ok := FindOpponent(ps, "Outro"); ok {
		t.Error("unknown name must not match")
	}
	if _, _, ok := FindOpponent(ps, ""); ok {
		t.Error("empty name must not match")
	}
	if _, _, ok := FindOpponent(ps[:1], "Mateus"); ok {
		t.Error("no opponent must not match")
	}
}
