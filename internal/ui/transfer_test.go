package ui

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"fyne.io/fyne/v2/dialog"

	"github.com/tinkthemaker/sniper/internal/snippets"
	"github.com/tinkthemaker/sniper/internal/store"
)

const importJSON = `{"version":1,"snippets":[
  {"id":"x","title":"Imported one","body":"first\r\nbody"},
  {"id":"x","title":"Imported two","body":"second"}]}`

func TestImportValidFileThenConfirmAppends(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	before := h.selectedTitle()
	h.m.importFrom(strings.NewReader(importJSON), "notes.json")

	top := h.topDialog()
	want := "Import 2 snippets from notes.json? Importing the same file twice creates duplicates."
	if top == nil || top.message != want {
		t.Fatalf("confirmation = %+v, want message %q", top, want)
	}
	if len(h.m.Library().Snippets) != 4 {
		t.Fatal("library changed before the user confirmed")
	}
	h.confirmTop()

	lib := h.m.Library()
	if len(lib.Snippets) != 6 || lib.Snippets[4].Title != "Imported one" || lib.Snippets[5].Title != "Imported two" {
		t.Fatalf("library = %+v", lib.Snippets)
	}
	if lib.Snippets[4].ID == lib.Snippets[5].ID || lib.Snippets[4].ID == "x" {
		t.Error("imported snippets must get fresh, distinct IDs")
	}
	if lib.Snippets[4].Body != "first\nbody" {
		t.Errorf("body not normalized: %q", lib.Snippets[4].Body)
	}
	if h.m.Status() != "Imported 2 snippets." {
		t.Errorf("status = %q", h.m.Status())
	}
	if h.selectedTitle() != before {
		t.Errorf("selection changed from %q to %q although it is still a result", before, h.selectedTitle())
	}
	if got := h.diskTitles(); len(got) != 6 {
		t.Errorf("disk titles = %v", got)
	}
}

func TestImportSelectionFallsBackToFirstWhenHidden(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.typeText("greet") // Greeting, Zebra notes
	h.m.importFrom(strings.NewReader(importJSON), "f.json")
	h.confirmTop()
	// "greet" does not match the imported snippets; selection stays valid.
	if h.selectedTitle() != "Greeting" {
		t.Errorf("selected %q", h.selectedTitle())
	}
}

func TestImportCancelChangesNothing(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.importFrom(strings.NewReader(importJSON), "f.json")
	h.dismissTop()
	if len(h.m.Library().Snippets) != 4 || len(h.diskTitles()) != 4 {
		t.Error("cancelled import changed the library")
	}
}

func TestImportInvalidFilesShowErrorAndChangeNothing(t *testing.T) {
	cases := map[string]string{
		"malformed":      `{"version":1,`,
		"wrong version":  `{"version":2,"snippets":[]}`,
		"invalid record": `{"version":1,"snippets":[{"title":"ok","body":"ok"},{"title":"","body":"x"}]}`,
		"unknown field":  `{"version":1,"snippets":[],"extra":true}`,
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t, sampleLibrary())
			before := h.primaryBytes()
			h.m.importFrom(strings.NewReader(in), "bad.json")
			top := h.topDialog()
			if top == nil || !strings.HasPrefix(top.message, "Import failed: ") ||
				!strings.HasSuffix(top.message, ". Your library was not changed.") {
				t.Fatalf("dialog = %+v", top)
			}
			if _, isConfirm := top.dlg.(*dialog.ConfirmDialog); isConfirm {
				t.Error("an invalid file must not offer to import")
			}
			if len(h.m.Library().Snippets) != 4 || h.primaryBytes() != before {
				t.Error("library or file changed")
			}
		})
	}
	h := newHarness(t, sampleLibrary())
	h.m.importFrom(strings.NewReader(cases["invalid record"]), "bad.json")
	if !strings.Contains(h.topDialog().message, "record 2:") {
		t.Errorf("error should name the record: %q", h.topDialog().message)
	}
}

func TestImportEmptyFile(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.importFrom(strings.NewReader(`{"version":1,"snippets":[]}`), "empty.json")
	if top := h.topDialog(); top == nil || top.message != "The file contains no snippets." {
		t.Errorf("dialog = %+v", top)
	}
	if len(h.m.Library().Snippets) != 4 {
		t.Error("library changed")
	}
}

func TestImportOverTheLimitIsRefused(t *testing.T) {
	items := make([]snippets.Snippet, snippets.MaxSnippets-1)
	for i := range items {
		items[i] = snippets.Snippet{Title: "T", Body: "B"}
	}
	lib, err := snippets.Append(snippets.NewLibrary(), items)
	if err != nil {
		t.Fatal(err)
	}
	h := newHarness(t, lib)
	h.m.importFrom(strings.NewReader(importJSON), "big.json") // 2 more than fit
	top := h.topDialog()
	want := "Import failed: The library is full (limit 10,000 snippets). Your library was not changed."
	if top == nil || top.message != want {
		t.Errorf("dialog = %+v, want message %q", top, want)
	}
	if len(h.m.Library().Snippets) != snippets.MaxSnippets-1 {
		t.Error("library changed")
	}
}

func TestImportSaveFailureLeavesLibraryUnchanged(t *testing.T) {
	var fs *failingStorage
	h := newHarnessWrapping(t, sampleLibrary(), func(s *store.Store) Storage {
		fs = &failingStorage{Store: s}
		return fs
	})
	before := h.primaryBytes()
	h.m.importFrom(strings.NewReader(importJSON), "f.json")
	fs.fail = true
	h.confirmTop()
	if len(h.m.Library().Snippets) != 4 || h.primaryBytes() != before {
		t.Error("failed import changed data")
	}
	top := h.topDialog()
	if top == nil || top.message != "Could not save the import: disk full. Your library was not changed." {
		t.Errorf("dialog = %+v", top)
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, errors.New("broken pipe") }

func TestExportWritesDecodableJSON(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	var buf bytes.Buffer
	if !h.m.exportTo(&buf, "out.json") {
		t.Fatal("export failed")
	}
	if h.m.Status() != "Exported 4 snippets to out.json." {
		t.Errorf("status = %q", h.m.Status())
	}
	lib, err := store.Decode(&buf)
	if err != nil {
		t.Fatalf("export is not a valid library file: %v", err)
	}
	if len(lib.Snippets) != 4 || lib.Snippets[0].Title != "Greeting" || lib.Snippets[0].Body != "Hello,\nThanks for reaching out." {
		t.Errorf("exported = %+v", lib.Snippets)
	}
}

func TestExportThenImportRoundTrips(t *testing.T) {
	src := newHarness(t, sampleLibrary())
	var buf bytes.Buffer
	src.m.exportTo(&buf, "x.json")

	dst := newHarness(t, snippets.NewLibrary())
	dst.m.importFrom(&buf, "x.json")
	dst.confirmTop()
	got, want := dst.m.Library().Snippets, src.m.Library().Snippets
	if len(got) != len(want) {
		t.Fatalf("got %d snippets", len(got))
	}
	for i := range want {
		if got[i].Title != want[i].Title || got[i].Body != want[i].Body {
			t.Errorf("snippet %d differs: %+v vs %+v", i, got[i], want[i])
		}
	}
}

func TestExportWriteFailureShowsError(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	if h.m.exportTo(failWriter{}, "x.json") {
		t.Error("export reported success")
	}
	top := h.topDialog()
	if top == nil || !strings.HasPrefix(top.message, "Export failed: ") {
		t.Errorf("dialog = %+v", top)
	}
	if strings.HasPrefix(h.m.Status(), "Exported") {
		t.Error("status claims success")
	}
}

func TestExportFileDialogDefaultName(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.exportFile()
	if top := h.topDialog(); top == nil || top.title != "Export…" {
		t.Fatalf("dialog = %+v", top)
	}
	h.m.importFile() // modal guard is only for shortcuts/menus; opening directly must still work
	if len(h.m.dialogs) != 2 {
		t.Errorf("dialogs = %d", len(h.m.dialogs))
	}
}

func TestMenusAndHelpDialogs(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	menu := h.m.win.MainMenu()
	if menu == nil || len(menu.Items) != 3 || menu.Items[0].Label != "File" || menu.Items[1].Label != "Edit" || menu.Items[2].Label != "Help" {
		t.Fatalf("main menu = %+v", menu)
	}
	labels := func(i int) (out []string) {
		for _, it := range menu.Items[i].Items {
			out = append(out, it.Label)
		}
		return
	}
	if !equalStrings(labels(0), []string{"Import…", "Export…", "Open data folder"}) {
		t.Errorf("File menu = %v", labels(0))
	}
	if !equalStrings(labels(1), []string{"Undo"}) {
		t.Errorf("Edit menu = %v", labels(1))
	}
	if !equalStrings(labels(2), []string{"Keyboard shortcuts", "About"}) {
		t.Errorf("Help menu = %v", labels(2))
	}

	menu.Items[0].Items[2].Action() // Open data folder
	if len(h.opened) != 1 || h.opened[0] != h.store.Dir() {
		t.Errorf("OpenFolder calls = %v, want [%s]", h.opened, h.store.Dir())
	}

	menu.Items[2].Items[1].Action() // About
	top := h.topDialog()
	if top == nil || !strings.Contains(top.message, "Sniper dev") || !strings.Contains(top.message, h.store.Dir()) ||
		!strings.Contains(top.message, "plain text") || !strings.Contains(top.message, "no network connections") {
		t.Fatalf("About = %+v", top)
	}
	menu.Items[0].Items[2].Action() // guarded while a dialog is open
	if len(h.opened) != 1 {
		t.Error("menu action ran while a dialog was open")
	}
	h.dismissTop()

	menu.Items[2].Items[0].Action() // Keyboard shortcuts
	top = h.topDialog()
	if top == nil || !strings.Contains(top.message, "Ctrl+Shift+D") || !strings.Contains(top.message, "Ctrl+S") || !strings.Contains(top.message, "Ctrl+Z") {
		t.Errorf("shortcuts help = %+v", top)
	}
}

func TestOpenFolderFailureShowsError(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.deps.OpenFolder = nil
	h.m.deps.OpenFolder = func(string) error { return errors.New("no explorer") }
	h.m.openDataFolder()
	want := "Could not open " + h.store.Dir() + ": no explorer"
	if top := h.topDialog(); top == nil || top.message != want {
		t.Errorf("dialog = %+v, want message %q", top, want)
	}
}
