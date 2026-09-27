package overlay

import (
	"math"
	"strings"
	"time"

	"sc2overlay/internal/build"
)

// Layout, in device-independent pixels.
const (
	Pad       = 10
	ClockRow  = 28
	StatusRow = 20
	RemRow    = 20
	SepGap    = 8
	StepRow   = 20
	NoteRow   = 15
	StepSize  = 14
	NoteSize  = 11
	blinkStep = 250 * time.Millisecond
)

// Color is a semantic color; Draw maps it to RGB.
type Color int

const (
	ColorText    Color = iota // future steps
	ColorDim                  // passed steps, notes
	ColorCurrent              // current step (yellow)
	ColorAlert                // warning blink / now (orange)
)

// Line is one row of the build list. Y is relative to the list top.
type Line struct {
	Y      int
	Time   string // planned time, "m:ss"
	Text   string
	Size   float64
	Color  Color
	Strike bool
	Marker bool // highlighted group
	Note   bool
}

// Height of the row.
func (l Line) Height() int {
	if l.Note {
		return NoteRow
	}
	return StepRow
}

// ViewInput is everything BuildView needs for one frame.
type ViewInput struct {
	Snap         Snapshot
	Warn         int           // warning_seconds
	Wall         time.Duration // real time since start, for blinking
	MaxHeight    int           // window height cap (0 = none)
	ManualScroll *float64      // nil = automatic scrolling
}

// View is what Draw paints.
type View struct {
	Clock         string
	Title         string
	Status        []string // status lines under the clock
	Reminders     string   // "" = hidden
	ReminderFlash bool
	MoveMode      bool
	Lines         []Line
	ListTop       int     // y where the list starts
	ContentHeight int     // full list height
	Height        int     // window height
	Scroll        float64 // list offset in px
	Current       int     // index in Lines of the first highlighted step, -1 if none
}

// BuildView computes the frame for a snapshot. It is pure.
func BuildView(in ViewInput) View {
	s := in.Snap
	v := View{Clock: build.FormatTime(int(math.Floor(s.Time))), Current: -1, MoveMode: s.Visible && !s.ClickThrough}
	if s.Build != nil {
		v.Title = s.Build.DisplayName()
	}
	blinkOn := (in.Wall/blinkStep)%2 == 0

	var st build.Status
	if s.Timeline != nil {
		st = s.Timeline.At(s.Time, in.Warn)
	}

	switch s.Status {
	case PortBusy:
		v.Status = append(v.Status, "Porta 6119 ocupada (jogo aberto?)")
	case Waiting:
		v.Status = append(v.Status, "Aguardando partida…")
	}
	if s.Build == nil {
		v.Status = append(v.Status, "Nenhuma build selecionada")
	} else if st.Done {
		v.Status = append(v.Status, "Build concluída")
	}

	if s.Build != nil {
		var parts []string
		for _, r := range build.ReminderStatus(s.Build.Reminders, s.Time) {
			p := r.Text
			if r.Countdown >= 0 {
				p += " " + build.FormatTime(r.Countdown)
			}
			parts = append(parts, p)
			v.ReminderFlash = v.ReminderFlash || r.Flash
		}
		v.Reminders = strings.Join(parts, " · ")
	}

	v.ListTop = Pad + ClockRow + StatusRow*len(v.Status) + SepGap
	if v.Reminders != "" {
		v.ListTop += RemRow
	}

	if s.Timeline != nil {
		y := 0
		for gi, g := range s.Timeline.Groups {
			state := st.States[gi]
			for _, step := range g.Steps {
				l := Line{Y: y, Time: build.FormatTime(g.Time), Text: step.Action, Size: StepSize}
				switch state {
				case build.Passed:
					l.Color, l.Strike = ColorDim, true
				case build.Future:
					l.Color = ColorText
				case build.Now:
					l.Color, l.Marker = ColorAlert, true
					l.Text += " — AGORA"
				case build.Current, build.Warning:
					l.Marker = true
					l.Text += " em " + build.FormatTime(build.Countdown(float64(g.Time)-s.Time))
					l.Color = ColorCurrent
					if state == build.Warning && blinkOn {
						l.Color = ColorAlert
					}
				}
				if l.Marker && v.Current < 0 && gi == st.Current {
					v.Current = len(v.Lines)
				}
				v.Lines = append(v.Lines, l)
				y += StepRow
				if step.Note != "" {
					v.Lines = append(v.Lines, Line{Y: y, Text: step.Note, Size: NoteSize, Color: ColorDim, Note: true})
					y += NoteRow
				}
			}
		}
		v.ContentHeight = y
	}

	v.Height = v.ListTop + v.ContentHeight + Pad
	if in.MaxHeight > 0 && v.Height > in.MaxHeight {
		v.Height = in.MaxHeight
	}
	visible := float64(v.Height - v.ListTop - Pad)
	maxScroll := math.Max(0, float64(v.ContentHeight)-visible)
	switch {
	case in.ManualScroll != nil:
		v.Scroll = *in.ManualScroll
	case v.Current >= 0:
		v.Scroll = float64(v.Lines[v.Current].Y) - visible/4
	case st.Done:
		v.Scroll = maxScroll
	}
	v.Scroll = math.Min(maxScroll, math.Max(0, v.Scroll))
	return v
}
