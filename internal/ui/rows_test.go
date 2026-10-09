package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/tinkthemaker/sniper/internal/snippets"
	"github.com/tinkthemaker/sniper/internal/store"
)

// fixtureLibrary loads an import-format file into a library, as the app's
// Import does.
func fixtureLibrary(t *testing.T, path ...string) snippets.Library {
	t.Helper()
	f, err := os.Open(filepath.Join(path...))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	items, err := store.DecodeImport(f)
	if err != nil {
		t.Fatal(err)
	}
	lib, err := snippets.Append(snippets.NewLibrary(), items)
	if err != nil {
		t.Fatal(err)
	}
	return lib
}

func fixture(t *testing.T, name string) snippets.Library {
	return fixtureLibrary(t, "testdata", "fixtures", name)
}

func corpus(t *testing.T) snippets.Library {
	return fixtureLibrary(t, "..", "..", "testdata", "sample-import.json")
}

// bigLibrary is n deterministic snippets with distinct titles and bodies.
func bigLibrary(t *testing.T, n int) snippets.Library {
	t.Helper()
	items := make([]snippets.Snippet, n)
	for i := range items {
		items[i] = snippets.Snippet{Title: fmt.Sprintf("Snippet %03d", i), Body: fmt.Sprintf("Body of snippet %03d\nsecond line", i)}
	}
	lib, err := snippets.Append(snippets.NewLibrary(), items)
	if err != nil {
		t.Fatal(err)
	}
	return lib
}

func absPos(o fyne.CanvasObject) fyne.Position {
	return fyne.CurrentApp().Driver().AbsolutePositionForObject(o)
}

// shownRow is a list row currently in the window's tree.
type shownRow struct {
	row  *listRow
	id   string
	pos  fyne.Position
	full bool // entirely inside the list's visible area
}

// shownRows returns the rows at least partly inside the list, top to bottom.
func (h *harness) shownRows() []shownRow {
	listPos, listSize := absPos(h.m.list), h.m.list.Size()
	var out []shownRow
	for o, refs := range h.m.rows {
		row, ok := o.(*listRow)
		if !ok || refs.id == "" {
			continue
		}
		p := absPos(row)
		if p.Y+row.Size().Height <= listPos.Y || p.Y >= listPos.Y+listSize.Height {
			continue // pooled or scrolled out
		}
		out = append(out, shownRow{row, refs.id, p, p.Y >= listPos.Y && p.Y+row.Size().Height <= listPos.Y+listSize.Height})
	}
	for i := range out { // insertion sort by Y
		for j := i; j > 0 && out[j].pos.Y < out[j-1].pos.Y; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func (h *harness) rowFor(id string) (shownRow, bool) {
	for _, r := range h.shownRows() {
		if r.id == id {
			return r, true
		}
	}
	return shownRow{}, false
}

func (h *harness) mustRow(id string) shownRow {
	h.t.Helper()
	r, ok := h.rowFor(id)
	if !ok {
		h.t.Fatalf("no row shows snippet %s", id)
	}
	return r
}

func (h *harness) idOfTitle(title string, nth int) string {
	h.t.Helper()
	for _, s := range h.m.Results() {
		if s.Title == title {
			if nth == 0 {
				return s.ID
			}
			nth--
		}
	}
	h.t.Fatalf("no result titled %q", title)
	return ""
}

func (h *harness) body(id string) string {
	for _, s := range h.m.Library().Snippets {
		if s.ID == id {
			return s.Body
		}
	}
	return ""
}

func (h *harness) click(id string)       { test.Tap(h.mustRow(id).row) }
func (h *harness) doubleClick(id string) { h.click(id); h.click(id) }

func TestRowsShowOnlyATruncatedTitle(t *testing.T) {
	h := newHarness(t, fixture(t, "long-titles.json"))
	rows := h.shownRows()
	if len(rows) != 4 {
		t.Fatalf("got %d rows, want 4", len(rows))
	}
	for _, r := range rows {
		kids := r.row.content.(*fyne.Container).Objects
		if len(kids) != 2 {
			t.Errorf("row has %d parts, want the bar and the title", len(kids))
		}
		title := h.m.rows[r.row].title
		if title.Truncation != fyne.TextTruncateEllipsis {
			t.Errorf("title truncation = %v, want ellipsis", title.Truncation)
		}
		if title.MinSize().Width > 100 {
			t.Errorf("title %q cannot shrink to the row: min %v", title.Text, title.MinSize())
		}
		if got := r.row.Size().Width; got != 496 {
			t.Errorf("row width = %v, want 496", got)
		}
	}
}

func TestDoubleClickSelectsAndCopies(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	id := h.idOfTitle("Sign-off", 0)
	if h.m.SelectedID() == id {
		t.Fatal("test needs an unselected row")
	}
	h.doubleClick(id)
	if h.m.SelectedID() != id {
		t.Errorf("selected %q, want Sign-off", h.selectedTitle())
	}
	if h.cb.last() != "Best regards,\nTink" || len(h.cb.copied) != 1 {
		t.Errorf("clipboard = %q", h.cb.copied)
	}
	if want := copiedStatus("Sign-off"); h.m.Status() != want {
		t.Errorf("status = %q, want %q", h.m.Status(), want)
	}
	if h.m.previewTitle.Text != "Sign-off" {
		t.Errorf("preview shows %q", h.m.previewTitle.Text)
	}
}

func TestSingleClickSelectsWithoutCopying(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	id := h.idOfTitle("Zebra notes", 0)
	h.click(id)
	if h.m.SelectedID() != id {
		t.Errorf("selected %q, want Zebra notes", h.selectedTitle())
	}
	if len(h.cb.copied) != 0 || h.m.Status() != "" {
		t.Errorf("single click copied %q, status %q", h.cb.copied, h.m.Status())
	}
	// Two slow clicks are two single clicks.
	now := fixedNow
	h.m.deps.Now = func() time.Time { return now }
	other := h.idOfTitle("Address", 0)
	h.click(other)
	now = now.Add(doubleClickWindow + time.Millisecond)
	h.click(other)
	if len(h.cb.copied) != 0 {
		t.Errorf("slow second click copied %q", h.cb.copied)
	}
	// Clicks on different rows are not a double-click either.
	h.click(id)
	h.click(other)
	if len(h.cb.copied) != 0 {
		t.Errorf("clicks on different rows copied %q", h.cb.copied)
	}
}

func TestDoubleClickAfterFilteringCopiesTheClickedRow(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.typeText("greet")
	if want := []string{"Greeting", "Zebra notes"}; !equalStrings(h.titles(), want) {
		t.Fatalf("results = %v", h.titles())
	}
	h.doubleClick(h.idOfTitle("Zebra notes", 0))
	if h.cb.last() != "Mentions greeting in the body" {
		t.Errorf("clipboard = %q", h.cb.last())
	}
	if h.selectedTitle() != "Zebra notes" || h.m.Status() != copiedStatus("Zebra notes") {
		t.Errorf("selected %q, status %q", h.selectedTitle(), h.m.Status())
	}
}

func TestDoubleClickInALargeScrolledLibrary(t *testing.T) {
	h := newHarness(t, bigLibrary(t, 300))
	h.m.win.Resize(fyne.NewSize(520, 600))
	h.m.list.ScrollTo(200)
	rows := h.shownRows()
	if len(rows) < 5 || rows[0].id == h.m.Results()[0].ID {
		t.Fatalf("list did not scroll: %d rows shown", len(rows))
	}
	// Click the lowest row, which may be cut off by the bottom of the list.
	target := rows[len(rows)-1].id
	h.doubleClick(target)
	if h.m.SelectedID() != target || h.cb.last() != h.body(target) || len(h.cb.copied) != 1 {
		t.Errorf("selected %s, clipboard %q, want body of %s", h.m.SelectedID(), h.cb.copied, target)
	}
	r, ok := h.rowFor(target)
	if !ok || !r.full {
		t.Errorf("clicked row visible=%v fully=%v after the double-click", ok, r.full)
	}
}

func TestDoubleClickIsIgnoredWhileADialogIsOpen(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	id := h.idOfTitle("Sign-off", 0)
	before := h.m.SelectedID()
	check := func(what string) {
		t.Helper()
		h.doubleClick(id)
		if len(h.cb.copied) != 0 || h.m.SelectedID() != before || h.m.Status() != "" {
			t.Errorf("%s: double-click copied %q, selected %q", what, h.cb.copied, h.selectedTitle())
		}
	}
	h.shortcut(shortcutNew)
	if h.topDialog() == nil {
		t.Fatal("no editor open")
	}
	check("editor open")
	h.dismissTop()
	if h.topDialog() != nil {
		t.Fatal("editor did not close")
	}
	h.shortcut(shortcutDelete)
	if h.topDialog() == nil {
		t.Fatal("no confirmation open")
	}
	check("confirmation open")
}

func TestDuplicateTitlesStayDistinct(t *testing.T) {
	h := newHarness(t, fixture(t, "duplicates.json"))
	ids := map[string]bool{}
	for n := 0; n < 3; n++ {
		id := h.idOfTitle("Greeting", n)
		ids[id] = true
		h.click(id)
		if h.m.previewBody.Text != h.body(id) || h.m.previewTitle.Text != "Greeting" {
			t.Errorf("preview of Greeting #%d shows %q", n, h.m.previewBody.Text)
		}
	}
	if len(ids) != 3 {
		t.Fatalf("results share IDs: %v", ids)
	}
	for n := 2; n >= 0; n-- {
		id := h.idOfTitle("Greeting", n)
		h.doubleClick(id)
		if h.cb.last() != h.body(id) {
			t.Errorf("double-click on Greeting #%d copied %q, want %q", n, h.cb.last(), h.body(id))
		}
	}
	if len(h.cb.copied) != 3 {
		t.Errorf("copied %d times, want 3", len(h.cb.copied))
	}
}

func TestEveryRowHasThePitchHeight(t *testing.T) {
	libs := map[string]snippets.Library{
		"corpus":      corpus(t),
		"long titles": fixture(t, "long-titles.json"),
		"duplicates":  fixture(t, "duplicates.json"),
		"large":       bigLibrary(t, 300),
	}
	for name, lib := range libs {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t, lib)
			rows := h.shownRows()
			if len(rows) == 0 {
				t.Fatal("no rows")
			}
			for i, r := range rows {
				if got := r.row.Size().Height; got != rowHeight {
					t.Errorf("row %d height = %v, want %v", i, got, rowHeight)
				}
				if i > 0 {
					if got := r.pos.Y - rows[i-1].pos.Y; got != rowHeight {
						t.Errorf("pitch between rows %d and %d = %v, want %v", i-1, i, got, rowHeight)
					}
				}
			}
		})
	}
}

func TestCopyButtonIsChunky(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.win.Resize(fyne.NewSize(520, 600))
	m := h.m
	cs := m.copyBtn.Size()
	if cs.Width < 112 || cs.Height < 40 {
		t.Errorf("Copy is %v, want at least 112×40", cs)
	}
	for _, b := range []*widget.Button{m.newBtn, m.editBtn, m.deleteBtn} {
		if b.Size().Height >= cs.Height {
			t.Errorf("%q (%v) is not smaller than Copy (%v)", b.Text, b.Size(), cs)
		}
	}
	if got := m.copyBtn.Theme().Size(theme.SizeNameText); got != 16 {
		t.Errorf("Copy text size = %v, want 16", got)
	}
	if got := m.newBtn.Theme().Size(theme.SizeNameText); got != 14 {
		t.Errorf("New text size = %v, want 14", got)
	}
	if x := absPos(m.copyBtn).X; x != 12 {
		t.Errorf("Copy starts at x = %v, want 12", x)
	}
	prev := fyne.CanvasObject(m.copyBtn)
	cy := absPos(m.copyBtn).Y + cs.Height/2
	for _, b := range []*widget.Button{m.newBtn, m.editBtn, m.deleteBtn} {
		if gap := absPos(b).X - (absPos(prev).X + prev.Size().Width); gap != 8 {
			t.Errorf("gap before %q = %v, want 8", b.Text, gap)
		}
		if mid := absPos(b).Y + b.Size().Height/2; mid < cy-0.5 || mid > cy+0.5 {
			t.Errorf("%q is centered at y = %v, Copy at %v", b.Text, mid, cy)
		}
		prev = b
	}
	if mid := absPos(m.status).Y + m.status.Size().Height/2; mid < cy-0.5 || mid > cy+0.5 {
		t.Errorf("status is centered at y = %v, Copy at %v", mid, cy)
	}
	// The Copy row sits 12 px above the key strip (less the 1 px rule above it).
	if below := absPos(m.keyStrip).Y - (absPos(m.copyBtn).Y + cs.Height); below != 12 {
		t.Errorf("space between Copy and the key strip = %v, want 12", below)
	}
}

func TestMainWindowGrid(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.win.Resize(fyne.NewSize(520, 600))
	m := h.m
	w := m.win.Canvas().Size().Width
	if w != 520 {
		t.Fatalf("canvas width = %v", w)
	}
	for name, o := range map[string]fyne.CanvasObject{"search box": m.search, "list": m.list, "Copy": m.copyBtn} {
		if x := absPos(o).X; x != 12 {
			t.Errorf("%s starts at x = %v, want 12", name, x)
		}
	}
	for name, o := range map[string]fyne.CanvasObject{"search box": m.search, "list": m.list} {
		if end := absPos(o).X + o.Size().Width; end != w-12 {
			t.Errorf("%s ends at x = %v, want %v", name, end, w-12)
		}
	}
	for name, o := range map[string]fyne.CanvasObject{"header strip": m.headerStrip, "key strip": m.keyStrip} {
		if x, end := absPos(o).X, absPos(o).X+o.Size().Width; x != 0 || end != w {
			t.Errorf("%s spans %v..%v, want 0..%v", name, x, end, w)
		}
	}
}
