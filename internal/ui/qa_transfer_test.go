package ui

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/tinkthemaker/sniper/internal/snippets"
	"github.com/tinkthemaker/sniper/internal/store"
)

// importFixture feeds a fixture file to the import flow like the file dialog does.
func (h *harness) importFixture(name string) {
	h.t.Helper()
	h.m.importFrom(fixtureReader(h.t, name), name)
}

func (h *harness) mustNotHaveChanged(before string, count int) {
	h.t.Helper()
	if len(h.m.Library().Snippets) != count || h.primaryBytes() != before {
		h.t.Error("the library or the file changed")
	}
}

func TestImportRefusesBadFixtures(t *testing.T) {
	cases := []struct {
		file string
		want string // exact dialog text
	}{
		{"invalid-import.json", "Import failed: record 3: Title is required. Your library was not changed."},
		{"wrong-version.json", "Import failed: file was created by an unsupported version of Sniper (version 2). Your library was not changed."},
		{"title-201.json", "Import failed: record 1: Title is too long (201 characters; limit 200). Your library was not changed."},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			h := newHarness(t, sampleLibrary())
			before := h.primaryBytes()
			h.importFixture(c.file)
			top := h.topDialog()
			if top == nil || top.message != c.want {
				t.Fatalf("dialog = %+v\nwant %q", top, c.want)
			}
			if _, confirm := top.dlg.(*dialog.ConfirmDialog); confirm {
				t.Error("must not offer to import")
			}
			if h.m.Status() != "" {
				t.Errorf("status = %q", h.m.Status())
			}
			h.mustNotHaveChanged(before, 4)
		})
	}

	t.Run("malformed.json", func(t *testing.T) {
		h := newHarness(t, sampleLibrary())
		before := h.primaryBytes()
		h.importFixture("malformed.json")
		top := h.topDialog()
		if top == nil || !strings.HasPrefix(top.message, "Import failed: invalid JSON: ") ||
			!strings.HasSuffix(top.message, ". Your library was not changed.") {
			t.Fatalf("dialog = %+v", top)
		}
		h.mustNotHaveChanged(before, 4)
	})
}

func TestImportDialogsCanBeDismissedAndTheWindowStaysUsable(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.importFixture("invalid-import.json")
	h.dismissTop()
	if h.m.modalOpen() {
		t.Fatal("the error dialog did not close")
	}
	h.key(fyne.KeyDown)
	h.key(fyne.KeyReturn)
	if h.cb.last() != "Hello,\nThanks for reaching out." {
		t.Errorf("after dismissing, copy gave %q", h.cb.last())
	}
}

func TestImportOfAnEmptyFileShowsTheInfoDialog(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	before := h.primaryBytes()
	h.importFixture("empty-import.json")
	top := h.topDialog()
	if top == nil || top.message != "The file contains no snippets." {
		t.Fatalf("dialog = %+v", top)
	}
	if _, confirm := top.dlg.(*dialog.ConfirmDialog); confirm {
		t.Error("an empty file must not ask to import")
	}
	if got := h.overlay().buttonLabels(); !equalStrings(got, []string{"OK"}) {
		t.Errorf("info dialog buttons = %v", got)
	}
	h.mustNotHaveChanged(before, 4)
}

func TestImportConfirmationWording(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.importFixture("duplicates.json")
	top := h.topDialog()
	if top.title != "Import snippets?" {
		t.Errorf("title = %q", top.title)
	}
	if got := h.overlay().buttonLabels(); !equalStrings(got, []string{"Cancel", "Import"}) {
		t.Errorf("buttons = %v", got)
	}
	want := "Import 3 snippets from duplicates.json? Importing the same file twice creates duplicates."
	if top.message != want {
		t.Errorf("message = %q", top.message)
	}
}

func TestImportingTheSameFileTwiceCreatesDistinctDuplicates(t *testing.T) {
	h := newHarness(t, snippets.NewLibrary())
	for i := 1; i <= 2; i++ {
		h.importFixture("duplicates.json")
		h.confirmTop()
		if want := i * 3; len(h.m.Library().Snippets) != want || len(h.diskTitles()) != want {
			t.Fatalf("after import %d: %d snippets in memory, %d on disk", i, len(h.m.Library().Snippets), len(h.diskTitles()))
		}
	}
	ids := map[string]bool{}
	for _, s := range h.m.Library().Snippets {
		ids[s.ID] = true
	}
	if len(ids) != 6 || h.m.countSeg.Text != "6 IN LIBRARY" || h.m.Status() != "Imported 3 snippets." {
		t.Errorf("ids=%d count=%q status=%q", len(ids), h.m.countSeg.Text, h.m.Status())
	}
	// All six rows read "Greeting"; the preview tells them apart.
	seen := map[string]bool{}
	for i := 0; i < 6; i++ {
		if h.m.previewTitle.Text != up("Greeting") {
			t.Fatalf("row %d title %q", i, h.m.previewTitle.Text)
		}
		seen[h.m.previewBody.Text] = true
		h.key(fyne.KeyDown)
	}
	if len(seen) != 3 {
		t.Errorf("previews seen: %d distinct bodies, want 3 (each twice)", len(seen))
	}
}

func TestImportNormalizesLineEndings(t *testing.T) {
	h := newHarness(t, snippets.NewLibrary())
	h.importFixture("crlf-bodies.json")
	h.confirmTop()
	want := map[string]string{
		"Windows lines":     "line one\nline two\n\nline four",
		"Old Mac lines":     "alpha\nbeta\ngamma",
		"Title with  break": "mixed\nand\nand\nall", // CR and LF each become a space in a title
	}
	disk := store.New(h.store.Dir()).Load().Library.Snippets
	if len(disk) != 3 {
		t.Fatalf("disk has %d snippets", len(disk))
	}
	for _, s := range disk {
		if want[s.Title] != s.Body {
			t.Errorf("%q saved as %q, want %q", s.Title, s.Body, want[s.Title])
		}
	}
	for range disk { // and the clipboard gets the stored text
		h.key(fyne.KeyReturn)
		if h.cb.last() != want[h.selectedTitle()] {
			t.Errorf("%q copied as %q", h.selectedTitle(), h.cb.last())
		}
		h.key(fyne.KeyDown)
	}
}

func TestImportAcceptsATitleOf200RunesAndShowsItWhole(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.importFixture("title-200.json")
	h.confirmTop()
	long := fixtureFirstTitle(t, "title-200.json")
	if got := h.diskTitles(); got[len(got)-1] != long {
		t.Error("the 200-rune title was not saved whole")
	}
	h.typeText("éééé") // matches the title only
	if h.selectedTitle() != long || h.m.previewTitle.Text != up(long) {
		t.Errorf("selected %q", h.selectedTitle())
	}
}

func TestImportSelectionRules(t *testing.T) {
	// Nothing selected (empty library): the first result gets selected.
	h := newHarness(t, snippets.NewLibrary())
	h.importFixture("crlf-bodies.json")
	h.confirmTop()
	if h.selectedTitle() != "Old Mac lines" {
		t.Errorf("selected %q, want the first result", h.selectedTitle())
	}

	// A selection that is still in the results stays, even if new ones sort before it.
	h2 := newHarness(t, sampleLibrary())
	h2.key(fyne.KeyPageDown) // Zebra notes
	h2.importFixture("crlf-bodies.json")
	h2.confirmTop()
	if h2.selectedTitle() != "Zebra notes" {
		t.Errorf("selection moved to %q", h2.selectedTitle())
	}

	// With a query active, the count follows and the query stays.
	h3 := newHarness(t, sampleLibrary())
	h3.typeText("lines")
	if len(h3.m.Results()) != 0 {
		t.Fatal("setup")
	}
	h3.importFixture("crlf-bodies.json")
	h3.confirmTop()
	if h3.m.search.Text != "lines" || h3.m.countSeg.Text != "2 OF 7" || h3.selectedTitle() != "Old Mac lines" {
		t.Errorf("query %q count %q selected %q", h3.m.search.Text, h3.m.countSeg.Text, h3.selectedTitle())
	}
	if h3.m.emptyLabel.Visible() {
		t.Error("the empty state should clear once results appear")
	}
}

func TestImportOfAnOversizedFileIsRefused(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	before := h.primaryBytes()
	h.m.importFrom(strings.NewReader(`{"version":1,"snippets":[],"pad":"`+strings.Repeat("x", store.MaxFileBytes)+`"}`), "huge.json")
	want := "Import failed: file is larger than 16 MiB. Your library was not changed."
	if top := h.topDialog(); top == nil || top.message != want {
		t.Errorf("dialog = %+v", top)
	}
	h.mustNotHaveChanged(before, 4)
}

func TestImportBOMIsIgnored(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.importFrom(strings.NewReader("\xEF\xBB\xBF"+importJSON), "bom.json")
	if top := h.topDialog(); top == nil || !strings.HasPrefix(top.message, "Import 2 snippets from bom.json?") {
		t.Errorf("dialog = %+v", top)
	}
}

func TestExportFileRoundTripsThroughDisk(t *testing.T) {
	h := newHarness(t, corpus(t))
	path := filepath.Join(t.TempDir(), "out.json")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if !h.m.exportTo(f, "out.json") {
		t.Fatal("export failed")
	}
	f.Close()
	if h.m.Status() != "Exported 12 snippets to out.json." {
		t.Errorf("status = %q", h.m.Status())
	}
	raw, _ := os.ReadFile(path)
	if !bytes.HasSuffix(raw, []byte("}\n")) || bytes.HasPrefix(raw, []byte("\xEF\xBB\xBF")) || !bytes.Contains(raw, []byte("\n  \"version\": 1,")) {
		t.Errorf("export is not in the file format: %.80q", raw)
	}

	dst := newHarness(t, snippets.NewLibrary())
	g, _ := os.Open(path)
	defer g.Close()
	dst.m.importFrom(g, "out.json")
	dst.confirmTop()
	want := h.m.Library().Snippets
	got := store.New(dst.store.Dir()).Load().Library.Snippets
	if len(got) != len(want) {
		t.Fatalf("round trip lost snippets: %d of %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Title != want[i].Title || got[i].Body != want[i].Body {
			t.Errorf("snippet %d changed: %q/%q", i, got[i].Title, got[i].Body)
		}
	}
}

func TestExportOfAnEmptyLibraryStillWorks(t *testing.T) {
	h := newHarness(t, snippets.NewLibrary())
	var buf bytes.Buffer
	if !h.m.exportTo(&buf, "none.json") || h.m.Status() != "Exported 0 snippets to none.json." {
		t.Errorf("status = %q", h.m.Status())
	}
	if lib, err := store.Decode(&buf); err != nil || len(lib.Snippets) != 0 {
		t.Errorf("decode: %v %+v", err, lib)
	}
}

func TestExportDialogOffersTheDatedFileName(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.exportFile()
	found := false
	for _, e := range h.overlay().entries {
		if e.Text == "sniper-export-2026-10-01.json" {
			found = true
		}
	}
	if !found {
		t.Error("the save dialog does not offer sniper-export-2026-10-01.json")
	}
}

func TestMenuItemsOpenTheFileDialogs(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	menu := h.m.win.MainMenu()
	menu.Items[0].Items[0].Action()
	if top := h.topDialog(); top == nil || top.title != "Import…" {
		t.Fatalf("Import… opened %+v", top)
	}
	h.dismissTop()
	menu.Items[0].Items[1].Action()
	if top := h.topDialog(); top == nil || top.title != "Export…" {
		t.Fatalf("Export… opened %+v", top)
	}
}

func TestHelpDialogsListEveryDocumentedKey(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.win.MainMenu().Items[2].Items[0].Action()
	top := h.topDialog()
	for _, key := range []string{
		"Down / Up", "PageDown / PageUp", "Enter", "Escape", "Ctrl+F", "Ctrl+N", "Ctrl+E",
		"Ctrl+Shift+D", "Ctrl+Z", "Tab / Shift+Tab", "Ctrl+S", "Double-click",
	} {
		if !strings.Contains(top.message, key) {
			t.Errorf("the shortcuts help does not mention %q", key)
		}
	}
	if got := h.overlay().buttonLabels(); !equalStrings(got, []string{"OK"}) {
		t.Errorf("buttons = %v", got)
	}
}

func TestDeleteConfirmationWording(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutDelete)
	u := h.overlay()
	if got := u.buttonLabels(); !equalStrings(got, []string{"Cancel", "Delete"}) {
		t.Fatalf("buttons = %v", got)
	}
	if u.button("Delete").Importance != widget.DangerImportance {
		t.Error("the confirm button must be the red danger button")
	}
	test.Tap(u.button("Delete"))
	if h.m.Status() != "Deleted “Address”." || len(h.diskTitles()) != 3 || h.focused() != fyne.Focusable(h.m.search) {
		t.Errorf("status %q focus %T", h.m.Status(), h.focused())
	}
}
