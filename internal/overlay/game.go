// Package overlay draws the always-on-top build order window.
package overlay

import (
	"bytes"
	"image"
	"image/color"
	_ "image/png"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"

	"sc2overlay/internal/build"
)

// Width is the window width in device-independent pixels.
const Width = 320

// screenMargin is kept free below the window when it grows (OVL-06).
const screenMargin = 80

var palette = map[Color]color.RGBA{
	ColorText:    {235, 235, 235, 255},
	ColorDim:     {130, 130, 130, 255},
	ColorCurrent: {255, 204, 0, 255},
	ColorAlert:   {255, 120, 0, 255},
}

var (
	colorDone   = color.RGBA{110, 220, 110, 255}
	colorBorder = color.RGBA{255, 204, 0, 255}
)

// Options are the display settings from the config.
type Options struct {
	WarningSeconds int
	Opacity        float64 // panel alpha, 0..1
	Sound          bool    // beep when a step enters the warning window
}

// Game is the ebiten.Game for the overlay window.
type Game struct {
	shared  *Shared
	opts    Options
	start   time.Time
	regular *text.GoTextFaceSource
	bold    *text.GoTextFaceSource

	view        View
	visible     bool
	passthrough bool
	height      int

	manualScroll *float64
	lastCurrent  int

	beep        *beeper
	beepedGroup int // group index that already beeped

	dragging     bool
	dragX, dragY int
	posX, posY   int
	OnDragEnd    func(x, y int) // called on the Ebitengine goroutine
}

// NewGame creates the overlay for shared state s.
func NewGame(s *Shared, opts Options) (*Game, error) {
	regular, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		return nil, err
	}
	bold, err := text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if err != nil {
		return nil, err
	}
	if opts.Opacity <= 0 || opts.Opacity > 1 {
		opts.Opacity = 0.7
	}
	g := &Game{shared: s, opts: opts, start: time.Now(), regular: regular, bold: bold, lastCurrent: -1, beepedGroup: -1}
	if opts.Sound {
		g.beep = newBeeper()
	}
	return g, nil
}

// Run configures the window and blocks on the main thread until quit.
func Run(g *Game, icon []byte, x, y int) error {
	const title = "SC2 Build Overlay"
	ebiten.SetWindowTitle(title)
	go hideFromTaskbar(title)
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetWindowSize(Width, 120)
	ebiten.SetWindowPosition(x, y)
	g.posX, g.posY = x, y // valid even if the loop ends before the first frame
	ebiten.SetWindowMousePassthrough(true)
	g.passthrough = true
	if img, _, err := image.Decode(bytes.NewReader(icon)); err == nil {
		ebiten.SetWindowIcon([]image.Image{img})
	}
	return ebiten.RunGameWithOptions(g, &ebiten.RunGameOptions{ScreenTransparent: true})
}

// Update applies shared state to the window, and handles drag and scroll.
func (g *Game) Update() error {
	select {
	case <-g.shared.Quit():
		return ebiten.Termination
	default:
	}
	g.posX, g.posY = ebiten.WindowPosition()
	snap := g.shared.Snapshot()
	g.visible = snap.Visible
	if pt := snap.Passthrough(); pt != g.passthrough {
		g.passthrough = pt
		ebiten.SetWindowMousePassthrough(pt)
		g.dragging = false
	}

	g.updateBeep(snap)

	_, monH := ebiten.Monitor().Size()
	in := ViewInput{Snap: snap, Warn: g.opts.WarningSeconds, Wall: time.Since(g.start), MaxHeight: monH - screenMargin}
	v := BuildView(in)
	if v.Current != g.lastCurrent {
		g.manualScroll = nil // OVL-11: auto-scroll resumes when the group changes
		g.lastCurrent = v.Current
	}
	if !g.passthrough {
		if _, wy := ebiten.Wheel(); wy != 0 {
			s := v.Scroll - wy*StepRow
			g.manualScroll = &s
		}
		g.updateDrag()
	}
	in.ManualScroll = g.manualScroll
	g.view = BuildView(in)
	if g.manualScroll != nil {
		s := g.view.Scroll // keep the clamped value
		g.manualScroll = &s
	}

	if g.view.Height != g.height {
		g.height = g.view.Height
		ebiten.SetWindowSize(Width, g.height)
	}
	return nil
}

// updateBeep beeps once per group when it enters the warning window.
func (g *Game) updateBeep(snap Snapshot) {
	if g.beep == nil || snap.Timeline == nil || snap.Status != InMatch {
		return
	}
	st := snap.Timeline.At(snap.Time, g.opts.WarningSeconds)
	for i, state := range st.States {
		if state == build.Warning && i != g.beepedGroup {
			g.beepedGroup = i
			g.beep.play()
		}
	}
}

func (g *Game) updateDrag() {
	cx, cy := ebiten.CursorPosition()
	switch {
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		g.dragging, g.dragX, g.dragY = true, cx, cy
	case g.dragging && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft):
		if cx != g.dragX || cy != g.dragY {
			wx, wy := ebiten.WindowPosition()
			ebiten.SetWindowPosition(wx+cx-g.dragX, wy+cy-g.dragY)
		}
	case g.dragging:
		g.dragging = false
		if g.OnDragEnd != nil {
			g.OnDragEnd(ebiten.WindowPosition())
		}
	}
}

// LastPosition is the window position seen on the last frame; it stays
// valid after the Ebitengine loop has ended.
func (g *Game) LastPosition() (x, y int) { return g.posX, g.posY }

// Draw paints the current view. A hidden overlay draws nothing.
func (g *Game) Draw(screen *ebiten.Image) {
	if !g.visible {
		return
	}
	v := g.view
	w, h := float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy())
	vector.FillRect(screen, 0, 0, w, h, color.RGBA{0, 0, 0, uint8(g.opts.Opacity * 255)}, false)

	// Header: clock, build name, status lines, reminders.
	g.text(screen, v.Clock, Pad, Pad, 20, palette[ColorText], true)
	if v.Title != "" {
		face := &text.GoTextFace{Source: g.regular, Size: 13}
		title := truncate(v.Title, face, Width-2*Pad-80)
		tw, _ := text.Measure(title, face, 0)
		g.text(screen, title, float64(Width-Pad)-tw, Pad+5, 13, palette[ColorDim], false)
	}
	y := float64(Pad + ClockRow)
	for _, s := range v.Status {
		c := palette[ColorText]
		if s == "Build concluída" {
			c = colorDone
		}
		g.text(screen, s, Pad, y, StepSize, c, false)
		y += StatusRow
	}
	if v.Reminders != "" {
		c := palette[ColorCurrent]
		if v.ReminderFlash && (time.Since(g.start)/blinkStep)%2 == 0 {
			c = palette[ColorAlert]
		}
		g.text(screen, v.Reminders, Pad, y, StepSize, c, true)
	}
	sepY := float32(v.ListTop - SepGap/2)
	vector.StrokeLine(screen, Pad, sepY, Width-Pad, sepY, 1, palette[ColorDim], false)

	// Build list, clipped to its area and scrolled.
	list := screen.SubImage(image.Rect(0, v.ListTop, Width, v.Height-Pad/2)).(*ebiten.Image)
	for _, l := range v.Lines {
		ly := float64(v.ListTop+l.Y) - v.Scroll
		if ly+float64(l.Height()) < float64(v.ListTop) || ly > float64(v.Height) {
			continue
		}
		c := palette[l.Color]
		textX := float64(Pad + 52)
		if l.Note {
			g.text(list, l.Text, textX, ly, l.Size, c, false)
			continue
		}
		if l.Marker {
			vector.FillPath(list, triangle(Pad, float32(ly)+4), &vector.FillOptions{}, &vector.DrawPathOptions{ColorScale: scale(c)})
		}
		g.text(list, l.Time, Pad+12, ly, 13, c, false)
		face := &text.GoTextFace{Source: g.regular, Size: l.Size}
		if l.Marker {
			face.Source = g.bold
		}
		s := truncate(l.Text, face, Width-int(textX)-Pad)
		g.draw(list, s, face, textX, ly, c)
		if l.Strike {
			tw, _ := text.Measure(s, face, 0)
			my := float32(ly + l.Size*0.6)
			vector.StrokeLine(list, Pad+12, my, float32(textX+tw), my, 1, c, false)
		}
	}

	if v.MoveMode {
		vector.StrokeRect(screen, 1, 1, w-2, h-2, 2, colorBorder, false)
		face := &text.GoTextFace{Source: g.regular, Size: NoteSize}
		tw, _ := text.Measure("modo mover", face, 0)
		g.draw(screen, "modo mover", face, float64(Width-Pad)-tw, float64(Pad+ClockRow)-4, colorBorder)
	}
}

func (g *Game) text(dst *ebiten.Image, s string, x, y, size float64, c color.Color, bold bool) {
	src := g.regular
	if bold {
		src = g.bold
	}
	g.draw(dst, s, &text.GoTextFace{Source: src, Size: size}, x, y, c)
}

func (g *Game) draw(dst *ebiten.Image, s string, face text.Face, x, y float64, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, face, op)
}

// truncate shortens s with "…" to fit maxWidth pixels.
func truncate(s string, face text.Face, maxWidth int) string {
	if w, _ := text.Measure(s, face, 0); w <= float64(maxWidth) {
		return s
	}
	r := []rune(s)
	for len(r) > 0 {
		r = r[:len(r)-1]
		if w, _ := text.Measure(string(r)+"…", face, 0); w <= float64(maxWidth) {
			break
		}
	}
	return string(r) + "…"
}

func triangle(x, y float32) *vector.Path {
	var p vector.Path
	p.MoveTo(x, y)
	p.LineTo(x+7, y+5)
	p.LineTo(x, y+10)
	p.Close()
	return &p
}

func scale(c color.RGBA) ebiten.ColorScale {
	var cs ebiten.ColorScale
	cs.ScaleWithColor(c)
	return cs
}

// Layout keeps one logical pixel per device-independent pixel.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
