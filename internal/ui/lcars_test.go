package ui

import (
	"image"
	"math"
	"regexp"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/tinkthemaker/sniper/internal/snippets"
)

var codeFormat = regexp.MustCompile(`^\d\d-\d{4}$`)

func TestRecordCodesAreStablePerID(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	codes := map[string]string{}
	for _, s := range h.m.Library().Snippets {
		c := recordCode(s.ID)
		if !codeFormat.MatchString(c) {
			t.Errorf("code %q of %q is not NN-NNNN", c, s.Title)
		}
		if again := recordCode(s.ID); again != c {
			t.Errorf("code of %q changed: %q then %q", s.Title, c, again)
		}
		codes[s.ID] = c
	}
	if recordCode("a") == recordCode("b") && recordCode("a") == recordCode("c") {
		t.Error("codes do not depend on the ID")
	}
	// The row shows the code of its own snippet however the list is filtered,
	// and the preview label names the selected snippet's code.
	for _, query := range []string{"", "greet", "o", "zebra"} {
		h.m.search.SetText(query)
		for _, r := range h.shownRows() {
			if got := h.m.rows[r.row].code.Text; got != codes[r.id] {
				t.Errorf("query %q: row of %s shows %q, want %q", query, r.id, got, codes[r.id])
			}
		}
		if id := h.m.SelectedID(); id != "" {
			if want := previewRecordLabel(codes[id]); h.m.previewHead.Text != want {
				t.Errorf("query %q: preview label = %q, want %q", query, h.m.previewHead.Text, want)
			}
		}
	}
	// A different order does not change codes either: reverse the library.
	lib := sampleLibrary()
	for i, j := 0, len(lib.Snippets)-1; i < j; i, j = i+1, j-1 {
		lib.Snippets[i], lib.Snippets[j] = lib.Snippets[j], lib.Snippets[i]
	}
	for _, s := range lib.Snippets {
		if recordCode(s.ID) != codes[s.ID] && codes[s.ID] != "" {
			t.Errorf("code of %q depends on position", s.Title)
		}
	}
}

func TestPreviewLabelNamesTheRecord(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	want := "PREVIEW · RECORD " + recordCode(h.m.SelectedID())
	if h.m.previewHead.Text != want {
		t.Errorf("preview label = %q, want %q", h.m.previewHead.Text, want)
	}
	h.typeText("nomatch")
	if h.m.previewHead.Text != "PREVIEW" {
		t.Errorf("with no selection the label = %q, want just PREVIEW", h.m.previewHead.Text)
	}
}

func TestCapitalsAreDisplayOnly(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	before := titlesOf(h.m.Library())
	// What is drawn is in capitals...
	for _, r := range h.shownRows() {
		shown := h.m.rows[r.row].title.Text
		if shown != strings.ToUpper(shown) || shown == "" {
			t.Errorf("row shows %q, want capitals", shown)
		}
	}
	if h.m.previewTitle.Text != "ADDRESS" {
		t.Errorf("preview title = %q, want ADDRESS", h.m.previewTitle.Text)
	}
	// ...and nothing stored or searched is.
	if got := titlesOf(h.m.Library()); !equalStrings(got, before) {
		t.Errorf("titles changed: %v -> %v", before, got)
	}
	if !equalStrings(before, []string{"Greeting", "Sign-off", "Address", "Zebra notes"}) {
		t.Errorf("sample titles = %v", before)
	}
	h.typeText("zebra")
	if !equalStrings(h.titles(), []string{"Zebra notes"}) {
		t.Errorf("a lower-case query found %v", h.titles())
	}
	h.key(fyne.KeyEscape)
	h.typeText("ZEBRA")
	if !equalStrings(h.titles(), []string{"Zebra notes"}) {
		t.Errorf("an upper-case query found %v", h.titles())
	}
	h.m.search.SetText("")
	h.shortcut(shortcutNew)
	h.typeText("Mixed Case Title")
	h.m.editor.body.SetText("body")
	h.shortcut(shortcutSave)
	if got := h.diskTitles(); got[len(got)-1] != "Mixed Case Title" {
		t.Errorf("saved titles = %v, want the title as typed", got)
	}
	if got := titlesOf(h.m.Library()); got[len(got)-1] != "Mixed Case Title" {
		t.Errorf("library titles = %v", got)
	}
}

func titlesOf(lib snippets.Library) []string {
	var out []string
	for _, s := range lib.Snippets {
		out = append(out, s.Title)
	}
	return out
}

func TestStatusReadoutShowsStatusMessages(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.win.Resize(fyne.NewSize(520, 600))
	if h.m.status.Text != "" {
		t.Fatalf("status starts as %q", h.m.status.Text)
	}
	h.key(fyne.KeyReturn)
	want := copiedStatus("Address")
	if h.m.status.Text != want || !strings.Contains(inspect(h.m.win.Content()).joined(), want) {
		t.Errorf("status = %q; the window shows %q", h.m.status.Text, inspect(h.m.win.Content()).texts)
	}
	// The readout is the widest segment of the bottom bar, and sits in it.
	f := h.m.frame
	if !f.readout.Visible() || f.readout.Size().Width <= f.hintNew.Size().Width+f.hintUndo.Size().Width {
		t.Errorf("readout %v is not the widest segment (hints %v, %v)", f.readout.Size(), f.hintNew.Size(), f.hintUndo.Size())
	}
	sp, rp := absPos(h.m.status), absPos(f.readout)
	mid := sp.Y + h.m.status.Size().Height/2 // the label is a little taller than the bar; its text is centered
	if sp.X < rp.X || sp.X+h.m.status.Size().Width > rp.X+f.readout.Size().Width || mid < rp.Y || mid > rp.Y+f.readout.Size().Height {
		t.Errorf("status label at %v %v is outside the readout at %v %v", sp, h.m.status.Size(), rp, f.readout.Size())
	}
	// It is ellipsized, never wrapped, when the message is too long.
	h.m.setStatus(strings.Repeat("long ", 80))
	if h.m.status.Truncation != fyne.TextTruncateEllipsis || h.m.status.Size().Width > f.readout.Size().Width {
		t.Errorf("long status: truncation %v, width %v in a readout %v wide", h.m.status.Truncation, h.m.status.Size().Width, f.readout.Size().Width)
	}
	if got := f.hintNewText.String() + "|" + f.hintUndoText.String(); got != "CTRL+N NEW|CTRL+Z UNDO" {
		t.Errorf("hints = %q", got)
	}
}

func (u uiOf) joined() string { return strings.Join(u.texts, "\n") }

func TestColumnCountFollowsTheQueryRules(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.win.Resize(fyne.NewSize(520, 600))
	count := func() string { return h.m.count.String() }
	if count() != "4 IN LIBRARY" {
		t.Errorf("count = %q", count())
	}
	h.typeText("greet")
	if count() != "2 OF 4" {
		t.Errorf("count with a query = %q", count())
	}
	h.key(fyne.KeyEscape)
	h.typeText("   ")
	if count() != "4 IN LIBRARY" {
		t.Errorf("a spaces-only query counts as no query, got %q", count())
	}
	// The count is the first block of the left column, labels at its bottom right.
	f := h.m.frame
	if x := absPos(h.m.count).X; x < 0 || x+h.m.count.Size().Width > absPos(f.blocks[0]).X+f.blocks[0].Size().Width {
		t.Errorf("count at x=%v is outside the left column", x)
	}
	if y := absPos(h.m.count).Y + h.m.count.Size().Height; y > absPos(f.blocks[0]).Y+f.blocks[0].Size().Height {
		t.Errorf("count sits below its block")
	}
	want := []string{"4 IN LIBRARY", "SEARCH", "PREVIEW", "VDEV"}
	for i, b := range f.blocks {
		if got := f.blockTexts[i].String(); got != want[i] {
			t.Errorf("block %d reads %q, want %q", i, got, want[i])
		}
		if i > 0 {
			prev := f.blocks[i-1]
			if gap := absPos(b).Y - (absPos(prev).Y + prev.Size().Height); gap != 4 {
				t.Errorf("gap above block %d = %v, want 4", i, gap)
			}
		}
	}
}

func TestVersionBlockShowsVDEVInTests(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	if got := h.m.frame.blockTexts[3].String(); got != "VDEV" {
		t.Errorf("version block = %q, want VDEV", got)
	}
	for in, want := range map[string]string{"": "VDEV", "dev": "VDEV", "1.2.3": "V1.2.3", "0.1.0-rc1": "V0.1.0-RC1"} {
		if got := versionText(in); got != want {
			t.Errorf("versionText(%q) = %q, want %q", in, got, want)
		}
	}
}

// firstInk is the top-left of the first bright (text-colored) pixel found
// scanning columns, then rows, inside the box.
func firstInk(img image.Image, x0, y0, x1, y1 int) (int, int, bool) {
	for x := x0; x < x1; x++ {
		for y := y0; y < y1; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if (r+g+b)>>8 > 600 {
				return x, y, true
			}
		}
	}
	return 0, 0, false
}

// cornerClearance is how far (x, y), measured from the field's top-left, is
// inside a corner of the given radius: 0 on the curve, more further in.
func cornerClearance(x, y, radius float64) float64 {
	if x >= radius || y >= radius {
		return min(x, y) // on a straight edge
	}
	return radius - math.Hypot(radius-x, radius-y)
}

func TestFieldTextAndCaretClearTheRoundedCorners(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutNew)
	ed := h.m.editor
	h.typeText("H") // into the title
	h.m.win.Canvas().Focus(ed.body)
	h.typeText("H")
	img := h.m.win.Canvas().Capture()
	const want = 4.0 // clear of the curve, in px
	for name, e := range map[string]*editorEntry{"title": ed.title, "body": ed.body} {
		p, s := absPos(e), e.Size()
		x, y, ok := firstInk(img, int(p.X)+2, int(p.Y)+2, int(p.X)+60, int(p.Y)+60)
		if !ok {
			t.Fatalf("%s: no text drawn", name)
		}
		// The caret starts a pixel left of the glyph and 3 px above its top.
		cx, cy := float64(x-int(p.X))-1, float64(y-int(p.Y))-3
		r := float64(min(fieldRadius, s.Height/2))
		if got := cornerClearance(cx, cy, r); got < want {
			t.Errorf("%s: the caret at (%.0f, %.0f) is %.1f px from the corner curve (radius %.0f), want at least %v", name, cx, cy, got, r, want)
		}
	}
	// One-line fields do not grow: the title is as tall as a plain entry.
	if got, plain := ed.title.MinSize().Height, widget.NewEntry().MinSize().Height; got != plain {
		t.Errorf("title field is %v tall, a plain entry is %v", got, plain)
	}
}

func TestSearchFieldClearsItsRoundedEnd(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.win.Resize(fyne.NewSize(520, 600))
	h.typeText(strings.Repeat("w", 80))
	p := h.m.parts
	if h.m.search.Size().Height != searchHeight && absPos(p.outline).Y != absPos(p.findPill).Y {
		t.Errorf("search is not as tall as the pill")
	}
	if p.outline.Size().Height != searchHeight || p.findPill.Size().Height != searchHeight {
		t.Errorf("field %v and pill %v should be %v tall", p.outline.Size().Height, p.findPill.Size().Height, searchHeight)
	}
	end := absPos(p.outline).X + p.outline.Size().Width
	if got := absPos(h.m.search).X + h.m.search.Size().Width; got > end-searchHeight/2 { // its own 4 px padding keeps the text that far again from the curve
		t.Errorf("search text can run to x=%v, inside the curve that starts at %v", got, end-searchHeight/2)
	}
	if got := absPos(h.m.search).X - absPos(p.outline).X; got < 4 {
		t.Errorf("search text starts %v px from the field's left edge", got)
	}
}
