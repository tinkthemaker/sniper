package ui

import (
	"image/color"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// recorder stands in for Fyne's animation runner: it keeps every animation
// started, so a test can tick it to any point, deterministically.
type recorder struct{ anims []*fyne.Animation }

func (r *recorder) run(a *fyne.Animation) { r.anims = append(r.anims, a) }

func (r *recorder) ofDuration(d time.Duration) []*fyne.Animation {
	var out []*fyne.Animation
	for _, a := range r.anims {
		if a.Duration == d {
			out = append(out, a)
		}
	}
	return out
}

// Distinct durations, so a test can tell the animations apart.
var testMotion = Motion{Enabled: true, Flash: 200 * time.Millisecond, Startup: 300 * time.Millisecond, Fade: 110 * time.Millisecond}

// withMotion makes the next harnesses animate through a recorder.
func withMotion(t *testing.T, m Motion) *recorder {
	t.Helper()
	rec := &recorder{}
	m.Run = rec.run
	harnessMotion = m
	t.Cleanup(func() { harnessMotion = Motion{} })
	return rec
}

func sameColor(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

func TestCopyFlashRunsOnARealCopyOnly(t *testing.T) {
	rec := withMotion(t, testMotion)
	h := newHarness(t, sampleLibrary())
	pill, readout := h.m.pills[0].fill, h.m.frame.readout
	if got := len(rec.ofDuration(testMotion.Flash)); got != 0 {
		t.Fatalf("%d flashes before any copy", got)
	}
	h.key(fyne.KeyReturn) // Enter
	flashes := rec.ofDuration(testMotion.Flash)
	if len(flashes) != 1 {
		t.Fatalf("Enter started %d flashes, want 1", len(flashes))
	}
	flashes[0].Tick(0)
	for name, c := range map[string][2]color.Color{"Copy pill": {pill.FillColor, colOrange}, "readout": {readout.FillColor, colRed}} {
		if luminance(c[0]) <= luminance(c[1]) {
			t.Errorf("%s did not brighten at the start of the flash: %v", name, c[0])
		}
		if contrast(colCanvas, c[0]) < 4.5 {
			t.Errorf("black on the brightened %s = %.2f:1, want at least 4.5:1", name, contrast(colCanvas, c[0]))
		}
	}
	flashes[0].Tick(0.5)
	if mid := pill.FillColor; luminance(mid) <= luminance(colOrange) {
		t.Errorf("mid-flash Copy pill = %v, want still brighter than orange", mid)
	}
	flashes[0].Tick(1)
	if !sameColor(pill.FillColor, colOrange) || !sameColor(readout.FillColor, colRed) {
		t.Errorf("flash did not return to the resting colors: %v, %v", pill.FillColor, readout.FillColor)
	}
	// The Copy button and a double-click copy for real too.
	test.Tap(h.m.copyBtn)
	h.doubleClick(h.idOfTitle("Sign-off", 0))
	if got := len(rec.ofDuration(testMotion.Flash)); got != 3 {
		t.Errorf("after Enter, the button and a double-click there were %d flashes, want 3", got)
	}
	// Nothing to copy: no flash.
	h.canvasShortcut(shortcutFind)
	h.typeText("nomatch")
	h.key(fyne.KeyReturn)
	test.Tap(h.m.copyBtn) // disabled
	if got := len(rec.ofDuration(testMotion.Flash)); got != 3 {
		t.Errorf("copy with nothing selected started a flash (%d in all)", got)
	}
}

func TestStartupLightsTheFrameOnceAndNeverBlocksInput(t *testing.T) {
	rec := withMotion(t, testMotion)
	h := newHarness(t, sampleLibrary())
	starts := rec.ofDuration(testMotion.Startup)
	if len(starts) != 1 {
		t.Fatalf("%d startup sequences, want 1", len(starts))
	}
	f := h.m.frame
	// First frame: dark blocks, and no black text on a dark block.
	if f.blocks[0].FillColor != color.Transparent || f.blockTexts[0].Visible() || f.hintNewText.Visible() {
		t.Errorf("the first frame is not dark: block %v, label visible %v", f.blocks[0].FillColor, f.blockTexts[0].Visible())
	}
	// Input works from the first frame.
	if h.focused() != fyne.Focusable(h.m.search) {
		t.Fatalf("focus is on %T, want the search box", h.focused())
	}
	h.typeText("greet")
	if len(h.m.Results()) != 2 {
		t.Errorf("typing during startup found %d results, want 2", len(h.m.Results()))
	}
	// Halfway, the first pieces are lit and the last are not yet.
	starts[0].Tick(0.5)
	if !sameColor(f.topCap.FillColor, colPeach) || h.m.parts.divider[4].FillColor != color.Transparent {
		t.Errorf("halfway: the first piece should be lit and the last still dark, got %v and %v", f.topCap.FillColor, h.m.parts.divider[4].FillColor)
	}
	starts[0].Tick(1)
	for i, want := range blockColors {
		if !sameColor(f.blocks[i].FillColor, want) || !f.blockTexts[i].Visible() {
			t.Errorf("block %d ends as %v (label visible %v), want %v shown", i, f.blocks[i].FillColor, f.blockTexts[i].Visible(), want)
		}
	}
	for i, c := range dividerColors {
		if !sameColor(h.m.parts.divider[i].FillColor, c) {
			t.Errorf("divider %d ends as %v, want %v", i, h.m.parts.divider[i].FillColor, c)
		}
	}
	if !sameColor(f.readout.FillColor, colRed) || !sameColor(f.topCap.FillColor, colPeach) || !f.hintNewText.Visible() || !f.status.Visible() {
		t.Error("the elbows and bottom bar do not end lit")
	}
	// Once per launch: more activity and a dialog start no more.
	h.key(fyne.KeyEscape)
	h.key(fyne.KeyReturn)
	h.shortcut(shortcutNew)
	h.dismissTop()
	if got := len(rec.ofDuration(testMotion.Startup)); got != 1 {
		t.Errorf("%d startup sequences after use, want 1", got)
	}
}

func TestStartupDoesNotRunWhenOpeningFromRecovery(t *testing.T) {
	rec := withMotion(t, testMotion)
	bak := goodBackup(t)
	rh := newRecoveryHarness(t, ptr("not json"), &bak)
	test.Tap(rh.r.restoreBtn)
	if rh.r.main == nil {
		t.Fatal("restore did not open the main window")
	}
	if got := len(rec.ofDuration(testMotion.Startup)); got != 0 {
		t.Errorf("%d startup sequences from the recovery window, want 0", got)
	}
	if !rh.r.main.frame.blockTexts[0].Visible() || !sameColor(rh.r.main.frame.blocks[0].FillColor, colPeri) {
		t.Error("the main window opened from recovery is not fully lit")
	}
}

func TestEverythingIsAtItsEndStateWhenAnimationsAreOff(t *testing.T) {
	rec := withMotion(t, Motion{Enabled: false, Flash: 200 * time.Millisecond, Startup: 300 * time.Millisecond, Fade: 110 * time.Millisecond})
	h := newHarness(t, sampleLibrary())
	f := h.m.frame
	for i, want := range blockColors {
		if !sameColor(f.blocks[i].FillColor, want) || !f.blockTexts[i].Visible() {
			t.Errorf("block %d is %v, want %v shown at once", i, f.blocks[i].FillColor, want)
		}
	}
	h.key(fyne.KeyDown)
	h.key(fyne.KeyReturn)
	test.Tap(h.m.copyBtn)
	sel := h.m.rows[h.mustRow(h.m.SelectedID()).row]
	if !sameColor(sel.pill.FillColor, colOrange) || !sameColor(h.m.pills[0].fill.FillColor, colOrange) || !sameColor(f.readout.FillColor, colRed) {
		t.Error("selection, Copy pill or readout is not at its resting color")
	}
	if len(rec.anims) != 0 {
		t.Errorf("%d animations started with animations off", len(rec.anims))
	}
}

func TestSelectionFadesInAndOutAndRapidMovesLeaveOneHighlight(t *testing.T) {
	rec := withMotion(t, testMotion)
	h := newHarness(t, sampleLibrary())
	rec.anims = nil
	first := h.m.SelectedID()
	h.key(fyne.KeyDown)
	second := h.m.SelectedID()
	fades := rec.ofDuration(testMotion.Fade)
	if len(fades) != 2 {
		t.Fatalf("one move started %d fades, want 2 (the row coming in and the row going out)", len(fades))
	}
	in, out := h.m.rows[h.mustRow(second).row], h.m.rows[h.mustRow(first).row]
	// Right after the move, before any tick, the new row has not jumped to full.
	if sameColor(in.pill.FillColor, colOrange) || !sameColor(out.pill.FillColor, colOrange) {
		t.Errorf("before the first tick: incoming %v, outgoing %v", in.pill.FillColor, out.pill.FillColor)
	}
	for _, a := range fades {
		a.Tick(0.5)
	}
	_, _, _, ia := in.pill.FillColor.RGBA()
	_, _, _, oa := out.pill.FillColor.RGBA()
	if ia == 0 || ia == 0xFFFF || oa == 0 || oa == 0xFFFF {
		t.Errorf("mid-fade alphas: incoming %#x, outgoing %#x, want both partway", ia, oa)
	}
	for _, a := range fades {
		a.Tick(1)
	}
	if !sameColor(in.pill.FillColor, colOrange) || out.pill.FillColor != color.Transparent {
		t.Errorf("fade ends: incoming %v, outgoing %v", in.pill.FillColor, out.pill.FillColor)
	}
	// Holding Down: every press is instant, nothing queues, and when the
	// animations finish (in any order) exactly one row is highlighted.
	rec.anims = nil
	for i := 0; i < 3; i++ {
		h.key(fyne.KeyDown)
	}
	if got := len(rec.anims); got > 3*2 {
		t.Errorf("3 key presses started %d animations, want at most 6", got)
	}
	for i := len(rec.anims) - 1; i >= 0; i-- { // newest first, then stale ones
		rec.anims[i].Tick(1)
	}
	lit := 0
	for _, r := range h.shownRows() {
		if sameColor(h.m.rows[r.row].pill.FillColor, colOrange) {
			lit++
			if r.id != h.m.SelectedID() {
				t.Errorf("the highlighted row is %s, the selection is %s", r.id, h.m.SelectedID())
			}
		}
	}
	if lit != 1 {
		t.Errorf("%d rows highlighted after rapid moves, want 1", lit)
	}
}

func TestMotionIsShortAndEndsAtRest(t *testing.T) {
	d := DefaultMotion()
	for name, v := range map[string]time.Duration{"flash": d.Flash, "startup": d.Startup, "fade": d.Fade} {
		if v <= 0 || v > 300*time.Millisecond {
			t.Errorf("%s lasts %v", name, v)
		}
	}
	if d.Flash > 250*time.Millisecond || d.Fade > 250*time.Millisecond {
		t.Errorf("flash %v or fade %v is over 250 ms", d.Flash, d.Fade)
	}
	_ = systemAnimations() // reads the Windows setting (or says yes elsewhere) without failing
	// The test driver runs an animation straight to its end state, so with the
	// real Run the window ends at rest.
	harnessMotion = Motion{Enabled: true, Flash: d.Flash, Startup: d.Startup, Fade: d.Fade}
	t.Cleanup(func() { harnessMotion = Motion{} })
	h := newHarness(t, sampleLibrary())
	h.key(fyne.KeyReturn)
	h.key(fyne.KeyDown)
	if !sameColor(h.m.frame.blocks[2].FillColor, colLavender) || !sameColor(h.m.pills[0].fill.FillColor, colOrange) {
		t.Error("animations did not end in their resting state")
	}
}
