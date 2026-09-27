package overlay

import (
	"strings"
	"testing"
	"time"

	"sc2overlay/internal/build"
)

const exampleYAML = `
name: Terran Bio
race: Terran
steps:
  - {time: "0:35", action: Supply Depot}
  - {time: "0:48", action: Refinaria, note: "3 SCVs no gás quando terminar"}
  - {time: "1:06", action: Barrack}
  - {time: "1:06", action: Reaper}
  - {time: "1:55", action: Factory}
reminders:
  - {text: "SCV!", every_seconds: 12, until: "7:00"}
`

func parse(t *testing.T, src string) *build.Build {
	t.Helper()
	b, err := build.Parse([]byte(src), "x.yml")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func input(t *testing.T, b *build.Build, status Status, gameTime float64) ViewInput {
	snap := Snapshot{Status: status, Time: gameTime, Visible: true, ClickThrough: true}
	if b != nil {
		snap.Build, snap.Timeline = b, build.NewTimeline(b)
	}
	return ViewInput{Snap: snap, Warn: 5, MaxHeight: 1000}
}

// stepLines returns the non-note lines.
func stepLines(v View) []Line {
	var out []Line
	for _, l := range v.Lines {
		if !l.Note {
			out = append(out, l)
		}
	}
	return out
}

func TestViewHeaderClockAndName(t *testing.T) {
	v := BuildView(input(t, parse(t, exampleYAML), InMatch, 184.7))
	if v.Clock != "3:04" || v.Title != "Terran Bio" {
		t.Errorf("clock/title = %q/%q, want 3:04/Terran Bio", v.Clock, v.Title)
	}
}

func TestViewWaiting(t *testing.T) {
	// API-04 / API-05: out of match shows the message and the list at 0:00.
	v := BuildView(input(t, parse(t, exampleYAML), Waiting, 0))
	if v.Clock != "0:00" || len(v.Status) != 1 || v.Status[0] != "Aguardando partida…" {
		t.Errorf("clock %q status %q", v.Clock, v.Status)
	}
	l := stepLines(v)[0]
	if l.Text != "Supply Depot em 0:35" || l.Color != ColorCurrent {
		t.Errorf("first line = %+v", l)
	}
}

func TestViewPortBusy(t *testing.T) {
	v := BuildView(input(t, parse(t, exampleYAML), PortBusy, 0))
	if len(v.Status) == 0 || v.Status[0] != "Porta 6119 ocupada (jogo aberto?)" {
		t.Errorf("status = %q", v.Status)
	}
}

func TestViewNoBuild(t *testing.T) {
	// OVL-08: no build keeps the clock.
	v := BuildView(input(t, nil, InMatch, 65))
	if v.Clock != "1:05" || len(v.Lines) != 0 || len(v.Status) != 1 || v.Status[0] != "Nenhuma build selecionada" {
		t.Errorf("view = %+v", v)
	}
}

func TestViewStepStyles(t *testing.T) {
	// OVL-03 at t=50: Depot passed, Refinaria passed (48+3=51 -> still now), ...
	b := parse(t, exampleYAML)
	v := BuildView(input(t, b, InMatch, 40))
	lines := stepLines(v)
	want := []struct {
		time, text string
		color      Color
		strike     bool
		marker     bool
	}{
		{"0:35", "Supply Depot", ColorDim, true, false},
		{"0:48", "Refinaria em 0:08", ColorCurrent, false, true},
		{"1:06", "Barrack", ColorText, false, false},
		{"1:06", "Reaper", ColorText, false, false},
		{"1:55", "Factory", ColorText, false, false},
	}
	if len(lines) != len(want) {
		t.Fatalf("lines = %d, want %d", len(lines), len(want))
	}
	for i, w := range want {
		l := lines[i]
		if l.Time != w.time || l.Text != w.text || l.Color != w.color || l.Strike != w.strike || l.Marker != w.marker || l.Size != 14 {
			t.Errorf("line %d = %+v, want %+v", i, l, w)
		}
	}
}

func TestViewGroupedCurrentHighlightsAllSteps(t *testing.T) {
	v := BuildView(input(t, parse(t, exampleYAML), InMatch, 55))
	lines := stepLines(v)
	if lines[2].Text != "Barrack em 0:11" || lines[3].Text != "Reaper em 0:11" ||
		lines[2].Color != ColorCurrent || lines[3].Color != ColorCurrent {
		t.Errorf("grouped = %+v / %+v", lines[2], lines[3])
	}
}

func TestViewNowState(t *testing.T) {
	v := BuildView(input(t, parse(t, exampleYAML), InMatch, 36))
	l := stepLines(v)[0]
	if l.Text != "Supply Depot — AGORA" || l.Color != ColorAlert || l.Strike || !l.Marker {
		t.Errorf("now line = %+v", l)
	}
}

func TestViewWarningBlinks(t *testing.T) {
	// OVL-04: alternates every 250 ms between alert and current colors.
	in := input(t, parse(t, exampleYAML), InMatch, 31) // 4 s to 0:35
	in.Wall = 0
	a := stepLines(BuildView(in))[0]
	in.Wall = 250 * time.Millisecond
	b := stepLines(BuildView(in))[0]
	in.Wall = 500 * time.Millisecond
	c := stepLines(BuildView(in))[0]
	if a.Color != ColorAlert || b.Color != ColorCurrent || c.Color != ColorAlert {
		t.Errorf("blink colors = %v %v %v, want Alert Current Alert", a.Color, b.Color, c.Color)
	}
	if a.Text != "Supply Depot em 0:04" {
		t.Errorf("text = %q", a.Text)
	}
}

func TestViewNotes(t *testing.T) {
	// OVL-05: note under its step, 11 px.
	v := BuildView(input(t, parse(t, exampleYAML), InMatch, 0))
	var idx int
	for i, l := range v.Lines {
		if strings.HasPrefix(l.Text, "Refinaria") {
			idx = i
		}
	}
	n := v.Lines[idx+1]
	if !n.Note || n.Text != "3 SCVs no gás quando terminar" || n.Size != 11 || n.Y <= v.Lines[idx].Y {
		t.Errorf("note = %+v", n)
	}
}

func TestViewBuildDone(t *testing.T) {
	// OVL-02 / STEP-05 / REM-04
	v := BuildView(input(t, parse(t, exampleYAML), InMatch, 200))
	if len(v.Status) != 1 || v.Status[0] != "Build concluída" {
		t.Errorf("status = %q", v.Status)
	}
	for _, l := range stepLines(v) {
		if !l.Strike || l.Color != ColorDim {
			t.Errorf("line not passed: %+v", l)
		}
	}
	if len(stepLines(v)) != 5 {
		t.Errorf("list must keep every step, got %d", len(stepLines(v)))
	}
	if v.Reminders == "" {
		t.Error("reminders must continue after the build is done")
	}
}

func TestViewReminderLine(t *testing.T) {
	// REM-02 / REM-01 / REM-03
	in := input(t, parse(t, exampleYAML), InMatch, 5.5)
	if v := BuildView(in); v.Reminders != "SCV! 0:07" || v.ReminderFlash {
		t.Errorf("reminder = %q flash %v", v.Reminders, v.ReminderFlash)
	}
	in.Snap.Time = 12.2
	if v := BuildView(in); !v.ReminderFlash {
		t.Errorf("want flash right after firing")
	}
	in.Snap.Time = 421
	if v := BuildView(in); v.Reminders != "" {
		t.Errorf("after until reminder = %q, want empty", v.Reminders)
	}
}

func TestViewMultipleRemindersJoined(t *testing.T) {
	b := parse(t, exampleYAML)
	b.Reminders = append(b.Reminders, build.Reminder{Text: "Inject", EverySeconds: 29, Until: 60})
	v := BuildView(input(t, b, InMatch, 5.5))
	if v.Reminders != "SCV! 0:07 · Inject 0:24" {
		t.Errorf("reminders = %q", v.Reminders)
	}
}

func TestViewMoveMode(t *testing.T) {
	in := input(t, parse(t, exampleYAML), InMatch, 0)
	if BuildView(in).MoveMode {
		t.Error("click-through on: no move mode")
	}
	in.Snap.ClickThrough = false
	if !BuildView(in).MoveMode {
		t.Error("OVL-09: click-through off shows move mode")
	}
}

func TestViewHeightFitsBuild(t *testing.T) {
	// OVL-06: fits the whole build when the screen allows.
	v := BuildView(input(t, parse(t, exampleYAML), InMatch, 0))
	last := v.Lines[len(v.Lines)-1]
	if v.Height != v.ListTop+last.Y+last.Height()+Pad || v.Scroll != 0 {
		t.Errorf("height = %d (listTop %d, last %+v), scroll %v", v.Height, v.ListTop, last, v.Scroll)
	}
}

func longBuild(n int) string {
	var sb strings.Builder
	sb.WriteString("race: Zerg\nsteps:\n")
	for i := 0; i < n; i++ {
		sb.WriteString("  - {time: \"" + build.FormatTime(10+i*10) + "\", action: Step}\n")
	}
	return sb.String()
}

func TestViewHeightCappedAndAutoScroll(t *testing.T) {
	// OVL-06 / OVL-07: 60 steps do not fit in 400 px; the current group stays
	// in the first third of the visible list.
	in := input(t, parse(t, longBuild(60)), InMatch, 305) // current = step at 5:10 (index 30)
	in.MaxHeight = 400
	v := BuildView(in)
	if v.Height != 400 {
		t.Fatalf("height = %d, want 400", v.Height)
	}
	cur := stepLines(v)[30]
	if !cur.Marker {
		t.Fatalf("line 30 = %+v, want current", cur)
	}
	visible := float64(v.Height - v.ListTop - Pad)
	pos := float64(cur.Y) - v.Scroll
	if pos < 0 || pos > visible/3 {
		t.Errorf("current at %v of %v visible px (scroll %v), want in first third", pos, visible, v.Scroll)
	}
	if len(stepLines(v)) != 60 {
		t.Errorf("passed steps must stay in the list")
	}
}

func TestViewScrollClamped(t *testing.T) {
	in := input(t, parse(t, longBuild(60)), InMatch, 0)
	in.MaxHeight = 400
	if v := BuildView(in); v.Scroll != 0 {
		t.Errorf("start scroll = %v, want 0", v.Scroll)
	}
	in.Snap.Time = 10000 // done: show the end
	v := BuildView(in)
	maxScroll := float64(v.ContentHeight - (v.Height - v.ListTop - Pad))
	if v.Scroll != maxScroll {
		t.Errorf("done scroll = %v, want %v", v.Scroll, maxScroll)
	}
}

func TestViewManualScroll(t *testing.T) {
	// OVL-11: a manual offset overrides auto-scroll, clamped to the content.
	in := input(t, parse(t, longBuild(60)), InMatch, 305)
	in.MaxHeight = 400
	manual := 40.0
	in.ManualScroll = &manual
	if v := BuildView(in); v.Scroll != 40 {
		t.Errorf("scroll = %v, want 40", v.Scroll)
	}
	manual = -50
	if v := BuildView(in); v.Scroll != 0 {
		t.Errorf("scroll = %v, want clamped 0", v.Scroll)
	}
	manual = 1e6
	v := BuildView(in)
	if maxScroll := float64(v.ContentHeight - (v.Height - v.ListTop - Pad)); v.Scroll != maxScroll {
		t.Errorf("scroll = %v, want clamped %v", v.Scroll, maxScroll)
	}
}
