package ui

import (
	"fmt"
	"strings"
	"testing"

	"fyne.io/fyne/v2"

	"github.com/tinkthemaker/sniper/internal/store"
)

var sampleTitles = []string{"Greeting", "Sign-off", "Address", "Zebra notes"}

// undoKey delivers Ctrl+Z the way the driver does: to the focused widget if it
// takes shortcuts, otherwise to the canvas.
func (h *harness) undoKey() {
	h.t.Helper()
	sc := &fyne.ShortcutUndo{}
	if f, ok := h.focused().(fyne.Shortcutable); ok {
		f.TypedShortcut(sc)
		return
	}
	h.canvasShortcut(sc)
}

func (h *harness) libTitles() []string {
	var out []string
	for _, s := range h.m.Library().Snippets {
		out = append(out, s.Title)
	}
	return out
}

// wantLibrary checks both the window's library and the file on disk.
func (h *harness) wantLibrary(want []string) {
	h.t.Helper()
	if got := h.libTitles(); !equalStrings(got, want) {
		h.t.Errorf("library = %v, want %v", got, want)
	}
	if got := h.diskTitles(); !equalStrings(got, want) {
		h.t.Errorf("file on disk = %v, want %v", got, want)
	}
}

func (h *harness) editTitle(to string) {
	h.t.Helper()
	h.shortcut(shortcutEdit)
	h.m.editor.title.SetText(to)
	h.shortcut(shortcutSave)
}

func TestUndoCreate(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	createViaKeyboard(h, "Phone", "555")
	if h.selectedTitle() != "Phone" {
		t.Fatalf("selected %q", h.selectedTitle())
	}
	h.undoKey()
	h.wantLibrary(sampleTitles)
	if h.m.Status() != "Undid new snippet “Phone”." {
		t.Errorf("status = %q", h.m.Status())
	}
	// Results were Address, Greeting, Phone, Sign-off, Zebra notes: the delete rule picks the next one.
	if h.selectedTitle() != "Sign-off" {
		t.Errorf("selected %q, want Sign-off", h.selectedTitle())
	}
}

func TestUndoEdit(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.editTitle("Street")
	if h.m.Status() != "Saved “Street”." {
		t.Fatalf("status = %q", h.m.Status())
	}
	h.undoKey()
	h.wantLibrary(sampleTitles)
	if h.m.Status() != "Undid edit of “Street”." || h.selectedTitle() != "Address" {
		t.Errorf("status = %q, selected %q", h.m.Status(), h.selectedTitle())
	}
}

func TestUndoDelete(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutDelete)
	h.confirmTop()
	if h.selectedTitle() != "Greeting" {
		t.Fatalf("selected %q after delete", h.selectedTitle())
	}
	h.undoKey()
	h.wantLibrary(sampleTitles)
	if h.m.Status() != "Undid delete of “Address”." || h.selectedTitle() != "Address" {
		t.Errorf("status = %q, selected %q", h.m.Status(), h.selectedTitle())
	}
}

func TestUndoDeleteClearsQueryThatHidesIt(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutDelete)
	h.confirmTop() // Address
	h.typeText("zebra")
	h.undoKey()
	if h.m.search.Text != "" || h.selectedTitle() != "Address" || len(h.m.Results()) != 4 {
		t.Errorf("query %q, selected %q, %d results", h.m.search.Text, h.selectedTitle(), len(h.m.Results()))
	}
}

func TestUndoImport(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.importFrom(strings.NewReader(importJSON), "notes.json")
	h.confirmTop()
	if len(h.m.Library().Snippets) != 6 {
		t.Fatal("import did not apply")
	}
	h.undoKey()
	h.wantLibrary(sampleTitles)
	if h.m.Status() != "Undid import of 2 snippets." || h.selectedTitle() != "Address" {
		t.Errorf("status = %q, selected %q", h.m.Status(), h.selectedTitle())
	}
}

func TestUndoImportOfOneSnippetIsSingular(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.importFrom(strings.NewReader(`{"version":1,"snippets":[{"id":"x","title":"Solo","body":"b"}]}`), "one.json")
	h.confirmTop()
	h.undoKey()
	if h.m.Status() != "Undid import of 1 snippet." {
		t.Errorf("status = %q", h.m.Status())
	}
}

func TestUndoStepsBackThroughSeveralChanges(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	createViaKeyboard(h, "Phone", "555")
	h.editTitle("Mobile")
	h.shortcut(shortcutDelete)
	h.confirmTop()
	h.undoKey()
	h.undoKey()
	h.undoKey()
	h.wantLibrary(sampleTitles)
	h.undoKey()
	if h.m.Status() != "Nothing to undo." {
		t.Errorf("status = %q", h.m.Status())
	}
}

func TestUndoIsCappedAtTwentySteps(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	for i := 0; i < 22; i++ {
		createViaKeyboard(h, fmt.Sprintf("N%02d", i), "b")
	}
	if len(h.m.undo) != 20 {
		t.Fatalf("stack holds %d steps", len(h.m.undo))
	}
	for i := 0; i < 20; i++ {
		h.undoKey()
		if strings.HasPrefix(h.m.Status(), "Nothing") {
			t.Fatalf("ran out after %d undos", i)
		}
	}
	h.undoKey()
	if h.m.Status() != "Nothing to undo." {
		t.Errorf("status = %q", h.m.Status())
	}
	// The two oldest creations are no longer undoable.
	if got := h.libTitles(); len(got) != 6 || got[4] != "N00" || got[5] != "N01" {
		t.Errorf("library = %v", got)
	}
	h.wantLibrary(h.libTitles())
}

func TestNothingToUndoSavesNothing(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	before := h.primaryBytes()
	h.undoKey()
	if h.m.Status() != "Nothing to undo." || h.primaryBytes() != before || h.m.modalOpen() {
		t.Errorf("status %q, dialog open %v", h.m.Status(), h.m.modalOpen())
	}
}

func TestUndoSaveFailureKeepsStackAndLibrary(t *testing.T) {
	var fs *failingStorage
	h := newHarnessWrapping(t, sampleLibrary(), func(s *store.Store) Storage {
		fs = &failingStorage{Store: s}
		return fs
	})
	createViaKeyboard(h, "Phone", "555")
	before := h.primaryBytes()
	status := h.m.Status()
	fs.fail = true
	h.undoKey()
	top := h.topDialog()
	if top == nil || top.message != "Could not undo: disk full" {
		t.Fatalf("dialog = %+v", top)
	}
	if len(h.m.undo) != 1 || len(h.m.Library().Snippets) != 5 || h.primaryBytes() != before || h.m.Status() != status {
		t.Error("failed undo changed the stack, library, file or status")
	}
	if h.m.undoItem.Disabled {
		t.Error("Undo menu item disabled although the stack is intact")
	}
	h.dismissTop()
	fs.fail = false
	h.focus(h.m.search)
	h.undoKey()
	h.wantLibrary(sampleTitles)
}

func TestUndoFromSearchListAndButton(t *testing.T) {
	for name, focus := range map[string]func(h *harness){
		"search": func(h *harness) { h.focus(h.m.search) },
		"list":   func(h *harness) { h.focus(h.m.list) },
		"button": func(h *harness) { h.focus(h.m.newBtn) },
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t, sampleLibrary())
			createViaKeyboard(h, "Phone", "555")
			focus(h)
			h.undoKey()
			h.wantLibrary(sampleTitles)
		})
	}
}

func TestUndoFromSearchBoxLeavesQueryAlone(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	createViaKeyboard(h, "Phone", "555")
	h.typeText("gre")
	h.undoKey()
	if h.m.search.Text != "gre" || len(h.m.Library().Snippets) != 4 {
		t.Errorf("query %q, %d snippets", h.m.search.Text, len(h.m.Library().Snippets))
	}
}

func TestUndoViaMenuAndItsEnabledState(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	menu := h.m.win.MainMenu()
	if len(menu.Items) != 3 || menu.Items[1].Label != "Edit" || len(menu.Items[1].Items) != 1 {
		t.Fatalf("menus = %+v", menu.Items)
	}
	item := menu.Items[1].Items[0]
	if item.Label != "Undo" || !item.Disabled {
		t.Fatalf("Undo item = %+v, want disabled", item)
	}
	createViaKeyboard(h, "Phone", "555")
	if item.Disabled {
		t.Error("Undo should be enabled after a change")
	}
	item.Action()
	h.wantLibrary(sampleTitles)
	if !item.Disabled {
		t.Error("Undo should be disabled once the stack is empty")
	}
}

func TestUndoMenuIgnoredWhileDialogOpen(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	createViaKeyboard(h, "Phone", "555")
	h.m.win.MainMenu().Items[2].Items[1].Action() // About
	h.m.win.MainMenu().Items[1].Items[0].Action()
	if len(h.m.Library().Snippets) != 5 {
		t.Error("menu undo ran while a dialog was open")
	}
}

func TestUndoIgnoredWhileADialogIsOpen(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	createViaKeyboard(h, "Phone", "555")
	h.shortcut(shortcutDelete) // the confirm dialog stays open
	status := h.m.Status()
	h.undoKey() // the search box still has focus
	h.canvasShortcut(&fyne.ShortcutUndo{})
	if len(h.m.Library().Snippets) != 5 || len(h.m.undo) != 1 || h.m.Status() != status {
		t.Error("undo ran while a dialog was open")
	}
}

func TestCtrlZInsideTheEditorIsTheEntrysOwnUndo(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	createViaKeyboard(h, "Phone", "555")
	h.shortcut(shortcutNew)
	h.typeText("abc")
	h.undoKey() // the title entry has focus
	if len(h.m.Library().Snippets) != 5 || len(h.m.undo) != 1 {
		t.Error("app undo fired inside the editor")
	}
	if h.m.editor == nil || h.m.editor.title.Text == "abc" {
		t.Errorf("entry text undo did not run: %+v", h.m.editor)
	}
	h.m.win.Canvas().Focus(h.m.editor.body)
	h.undoKey()
	h.canvasShortcut(&fyne.ShortcutUndo{}) // as when a button in the editor has focus
	if len(h.m.Library().Snippets) != 5 || len(h.m.undo) != 1 {
		t.Error("app undo fired inside the editor")
	}
}
