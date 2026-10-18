package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// Motion is how the window animates. Every animation is short, purposeful and
// ends in its resting state, after which nothing is redrawn. They run on
// Fyne's own animation API: the driver ticks them from its render loop, on the
// same thread that handles events and paints (internal/driver/glfw/loop.go),
// so a Tick may touch widgets like any event handler can, and the app starts
// no goroutines of its own (D2).
type Motion struct {
	// Enabled is false when animations are off (Windows' "Animation effects"
	// setting, or the zero value in tests): every end state is shown at once.
	Enabled bool
	// Run starts an animation; nil means (*fyne.Animation).Start. Tests replace
	// it to capture animations and tick them by hand.
	Run func(*fyne.Animation)

	Flash   time.Duration // the Copy pill and the readout brighten and return
	Startup time.Duration // the frame lights up, once per launch
	Fade    time.Duration // the selection highlight eases in and out

	skipStartup bool // set when the main window opens from the recovery window
}

// DefaultMotion is the real app's motion: on unless the system turns animations off.
func DefaultMotion() Motion {
	return Motion{
		Enabled: systemAnimations(),
		Flash:   200 * time.Millisecond,
		Startup: 300 * time.Millisecond,
		Fade:    110 * time.Millisecond,
	}
}

func (m *MainWindow) animate(d time.Duration, curve fyne.AnimationCurve, tick func(float32)) *fyne.Animation {
	a := fyne.NewAnimation(d, tick)
	a.Curve = curve
	if run := m.deps.Motion.Run; run != nil {
		run(a)
	} else {
		a.Start()
	}
	return a
}

func (m *MainWindow) stopAnimation(a *fyne.Animation) {
	if a != nil && m.deps.Motion.Run == nil {
		a.Stop()
	}
}

// fade is c at fraction f, as an alpha over the black canvas: a fill that fades in.
func fade(c color.Color, f float32) color.Color {
	if f >= 1 {
		return c
	}
	if f <= 0 {
		return color.Transparent
	}
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	n.A = uint8(f*255 + 0.5)
	return n
}

// brighten moves c a fraction f of the way to white.
func brighten(c color.Color, f float32) color.Color {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	up := func(v uint8) uint8 { return uint8(float32(v) + (255-float32(v))*f + 0.5) }
	return color.NRGBA{R: up(n.R), G: up(n.G), B: up(n.B), A: 0xFF}
}

// flashCopy is the confirmation of a real copy: the Copy pill and the status
// readout brighten, then ease back over Flash. Brighter fills only raise the
// contrast of the black text on them.
func (m *MainWindow) flashCopy() {
	mo := m.deps.Motion
	if !mo.Enabled || m.frame == nil {
		return
	}
	m.stopAnimation(m.flash)
	pill, readout := m.pills[0].fill, m.frame.readout
	pillBase, readoutBase := m.pills[0].c, color.Color(colRed)
	m.flash = m.animate(mo.Flash, fyne.AnimationEaseOut, func(p float32) {
		f := 0.5 * (1 - p)
		if p >= 1 {
			pill.FillColor, readout.FillColor = pillBase, readoutBase
		} else {
			pill.FillColor, readout.FillColor = brighten(pillBase, f), brighten(readoutBase, f)
		}
		pill.Refresh()
		readout.Refresh()
	})
}

// lightPart is a piece of the frame that the startup sequence lights up.
type lightPart struct {
	rect  *canvas.Rectangle
	base  color.Color
	label fyne.CanvasObject // shown once the part is fully lit, so black text never sits on a half-lit block
}

// startup lights the frame up, piece by piece, once. Everything is dark at the
// first frame and exact at the last. Input is never held: the search box is
// focused and the panel is live throughout.
func (m *MainWindow) startup() {
	mo := m.deps.Motion
	if !mo.Enabled || mo.skipStartup || m.frame == nil {
		return
	}
	f, p := m.frame, m.parts
	var parts []lightPart
	add := func(r *canvas.Rectangle, label fyne.CanvasObject) {
		parts = append(parts, lightPart{r, r.FillColor, label})
	}
	add(f.topCap, nil)
	add(f.topJoin, nil)
	add(f.topBar, nil)
	add(f.topEnd, nil)
	for i, b := range f.blocks {
		add(b, f.blockTexts[i])
	}
	add(f.botJoin, nil)
	add(f.botCap, nil)
	add(f.readout, f.status)
	add(f.hintNew, f.hintNewText)
	add(f.hintUndo, f.hintUndoText)
	add(f.botEnd, nil)
	for _, d := range p.divider {
		add(d, nil)
	}
	for _, part := range parts {
		part.rect.FillColor = color.Transparent
		part.rect.Refresh()
		if part.label != nil {
			part.label.Hide()
		}
	}
	const window = 0.4 // each piece takes this share of the whole, staggered across the rest
	step := (1 - window) / float32(len(parts)-1)
	m.animate(mo.Startup, fyne.AnimationLinear, func(t float32) {
		for i, part := range parts {
			q := min(max((t-step*float32(i))/window, 0), 1)
			if t >= 1 {
				q = 1
			}
			part.rect.FillColor = fade(part.base, q)
			part.rect.Refresh()
			if q >= 1 && part.label != nil && !part.label.Visible() {
				part.label.Show()
			}
		}
	})
}

// fadeRow eases a row's highlight toward selected (or away from it) over Fade.
// Each row has at most one fade running; a newer one replaces it, so holding
// Down never queues anything. The row may be reused for another snippet by the
// list before the fade ends, so every tick checks it still shows id.
func (m *MainWindow) fadeRow(refs *rowRefs, id string, selected bool) {
	mo := m.deps.Motion
	if !mo.Enabled {
		return
	}
	m.stopAnimation(refs.fade)
	refs.gen++
	gen := refs.gen
	from, to := float32(1), float32(0)
	if selected {
		from, to = 0, 1
	}
	refs.paint(from)
	refs.fade = m.animate(mo.Fade, fyne.AnimationEaseOut, func(t float32) {
		if refs.id != id || refs.gen != gen {
			return // the row moved on to another snippet, or a newer fade took over
		}
		if t >= 1 {
			refs.paint(to)
			return
		}
		refs.paint(from + (to-from)*t)
	})
}
