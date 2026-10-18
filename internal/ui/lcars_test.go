package ui

import (
	"regexp"
	"strings"
	"testing"

	"fyne.io/fyne/v2"

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
