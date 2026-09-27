package build

import "testing"

var scv = []Reminder{{Text: "SCV!", EverySeconds: 12, Until: 420}}

func TestReminderStatus(t *testing.T) {
	cases := []struct {
		t         float64
		countdown int
		flash     bool
	}{
		{0, 12, false},    // no fire at t=0
		{5.5, 7, false},   // counting to first fire
		{12, 12, true},    // fires at 12
		{13.4, 11, true},  // still flashing (<1.5 s)
		{13.6, 11, false}, // flash over
		{24.2, 12, true},  // second fire
		{408, 12, true},   // fire at 408, next at 420 <= until
		{419, 1, false},   // counting to the last fire
		{420, -1, true},   // last fire exactly at until, no next
	}
	for _, c := range cases {
		got := ReminderStatus(scv, c.t)
		if len(got) != 1 {
			t.Errorf("t=%v: %d active, want 1", c.t, len(got))
			continue
		}
		r := got[0]
		if r.Text != "SCV!" || r.Countdown != c.countdown || r.Flash != c.flash {
			t.Errorf("t=%v: %+v, want countdown %d flash %v", c.t, r, c.countdown, c.flash)
		}
	}
}

func TestReminderGoneAfterUntil(t *testing.T) {
	if got := ReminderStatus(scv, 420.1); len(got) != 0 {
		t.Errorf("after until: %+v, want none", got)
	}
}

func TestReminderMultiple(t *testing.T) {
	rs := []Reminder{
		{Text: "SCV!", EverySeconds: 12, Until: 420},
		{Text: "Inject", EverySeconds: 29, Until: 60},
	}
	got := ReminderStatus(rs, 30)
	if len(got) != 2 || got[0].Text != "SCV!" || got[1].Text != "Inject" || got[1].Countdown != 28 || !got[1].Flash {
		t.Errorf("t=30: %+v", got)
	}
	got = ReminderStatus(rs, 61)
	if len(got) != 1 || got[0].Text != "SCV!" {
		t.Errorf("t=61: %+v, want only SCV!", got)
	}
}

func TestReminderIndependentOfBuildDone(t *testing.T) {
	// REM-04: reminders only depend on time, so they keep going after the
	// last step (5:40 in the example) until their own until (7:00).
	b, err := Load("testdata/terran_bio.yml")
	if err != nil {
		t.Fatal(err)
	}
	const t6 = 6 * 60
	if !NewTimeline(b).At(t6, 5).Done {
		t.Fatal("build should be done at 6:00")
	}
	if got := ReminderStatus(b.Reminders, t6); len(got) != 1 || got[0].Text != "SCV!" {
		t.Errorf("at 6:00: %+v, want SCV! active", got)
	}
}
