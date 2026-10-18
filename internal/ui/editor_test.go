package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/tinkthemaker/sniper/internal/snippets"
	"github.com/tinkthemaker/sniper/internal/store"
)

// createViaKeyboard drives Ctrl+N, title, body, Ctrl+S exactly as a user would.
func createViaKeyboard(h *harness, title, body string) {
	h.t.Helper()
	h.shortcut(shortcutNew)
	if h.m.editor == nil {
		h.t.Fatal("Ctrl+N did not open the editor")
	}
	h.typeText(title)
	h.m.win.Canvas().Focus(h.m.editor.body)
	h.typeText(body)
	h.shortcut(shortcutSave)
}

func TestCtrlNOpensEditorAndCtrlSSavesFromBody(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutNew) // search has focus: the routing regression test
	ed := h.m.editor
	if ed == nil {
		t.Fatal("Ctrl+N with search focused did not open the editor")
	}
	if h.focused() != fyne.Focusable(ed.title) {
		t.Errorf("focus after New = %T, want the title entry", h.focused())
	}
	if h.topDialog().title != "New snippet" {
		t.Errorf("dialog title = %q", h.topDialog().title)
	}

	h.typeText("Phone")
	h.m.win.Canvas().Focus(ed.body)
	h.typeText("555-0100")
	h.key(fyne.KeyReturn)
	h.typeText("ext 12")
	h.shortcut(shortcutSave) // body entry has focus

	if h.m.editor != nil || h.m.modalOpen() {
		t.Error("editor should close after a successful save")
	}
	if h.focused() != fyne.Focusable(h.m.search) {
		t.Errorf("focus after save = %T, want search", h.focused())
	}
	if h.selectedTitle() != "Phone" {
		t.Errorf("selected = %q, want the new snippet", h.selectedTitle())
	}
	if h.m.Status() != "Saved “Phone”." {
		t.Errorf("status = %q", h.m.Status())
	}
	if got := h.diskTitles(); got[len(got)-1] != "Phone" || len(got) != 5 {
		t.Errorf("disk titles = %v", got)
	}
	res := store.New(h.store.Dir()).Load()
	if body := res.Library.Snippets[4].Body; body != "555-0100\next 12" {
		t.Errorf("saved body = %q", body)
	}
}

func TestCtrlSWorksFromTitleToo(t *testing.T) {
	h := newHarness(t, snippets.NewLibrary())
	h.shortcut(shortcutNew)
	h.typeText("Only title")
	h.m.win.Canvas().Focus(h.m.editor.body)
	h.typeText("b")
	h.m.win.Canvas().Focus(h.m.editor.title)
	h.shortcut(shortcutSave)
	if h.m.editor != nil || len(h.m.Library().Snippets) != 1 {
		t.Error("Ctrl+S from the title entry did not save")
	}
}

func TestValidationShowsErrorAndWritesNothing(t *testing.T) {
	h := newHarness(t, snippets.NewLibrary())
	h.shortcut(shortcutNew)
	h.shortcut(shortcutSave)
	ed := h.m.editor
	if ed == nil {
		t.Fatal("dialog closed on a validation error")
	}
	if ed.errLabel.Text != "Title is required." || !ed.errLabel.Visible() {
		t.Errorf("error label = %q visible=%v", ed.errLabel.Text, ed.errLabel.Visible())
	}
	if h.focused() != fyne.Focusable(ed.title) {
		t.Errorf("focus = %T, want the offending title field", h.focused())
	}
	if h.primaryBytes() != "" {
		t.Error("a file was written for an invalid snippet")
	}

	h.typeText("Has title")
	h.shortcut(shortcutSave)
	if ed.errLabel.Text != "Body is required." || h.focused() != fyne.Focusable(ed.body) {
		t.Errorf("error = %q, focus = %T; want body error and body focus", ed.errLabel.Text, h.focused())
	}
	if len(h.m.Library().Snippets) != 0 {
		t.Error("library changed on validation error")
	}
}

func TestSaveFailureKeepsEditorOpenAndDataUnchanged(t *testing.T) {
	var fs *failingStorage
	h := newHarnessWrapping(t, sampleLibrary(), func(s *store.Store) Storage {
		fs = &failingStorage{Store: s}
		return fs
	})
	before := h.primaryBytes()
	libBefore := h.m.Library()
	fs.fail = true

	h.shortcut(shortcutNew)
	h.typeText("Will not save")
	h.m.win.Canvas().Focus(h.m.editor.body)
	h.typeText("body")
	h.shortcut(shortcutSave)

	ed := h.m.editor
	if ed == nil {
		t.Fatal("editor closed after a failed save")
	}
	want := "Could not save: disk full. Your changes are still here; try again."
	if ed.errLabel.Text != want || !ed.errLabel.Visible() {
		t.Errorf("error label = %q, want %q", ed.errLabel.Text, want)
	}
	if len(h.m.Library().Snippets) != len(libBefore.Snippets) {
		t.Error("in-memory library changed after a failed save")
	}
	if h.primaryBytes() != before {
		t.Error("file on disk changed after a failed save")
	}
	if ed.title.Text != "Will not save" || ed.body.Text != "body" {
		t.Error("the user's text was lost")
	}

	fs.fail = false // retry succeeds from the same form
	h.shortcut(shortcutSave)
	if h.m.editor != nil || len(h.m.Library().Snippets) != 5 {
		t.Error("retry did not save")
	}
}

func TestEditViaCtrlE(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.key(fyne.KeyDown) // Greeting
	h.shortcut(shortcutEdit)
	ed := h.m.editor
	if ed == nil {
		t.Fatal("Ctrl+E did not open the editor")
	}
	if h.topDialog().title != "Edit snippet" || ed.title.Text != "Greeting" {
		t.Errorf("dialog %q title entry %q", h.topDialog().title, ed.title.Text)
	}
	if h.focused() != fyne.Focusable(ed.body) {
		t.Errorf("focus after Edit = %T, want the body entry", h.focused())
	}
	ed.body.SetText("Changed body")
	h.shortcut(shortcutSave)

	if h.m.previewBody.Text != "Changed body" || h.m.previewTitle.Text != up("Greeting") {
		t.Errorf("preview = %q / %q", h.m.previewTitle.Text, h.m.previewBody.Text)
	}
	if h.m.Status() != "Saved “Greeting”." {
		t.Errorf("status = %q", h.m.Status())
	}
	res := store.New(h.store.Dir()).Load()
	if res.Library.Snippets[0].Body != "Changed body" || len(res.Library.Snippets) != 4 {
		t.Errorf("saved library = %+v", res.Library.Snippets)
	}
	if res.Library.Snippets[0].ID != h.m.SelectedID() {
		t.Error("edit must keep the snippet's ID and selection")
	}
}

func TestEditWithNothingSelectedDoesNothing(t *testing.T) {
	h := newHarness(t, snippets.NewLibrary())
	h.shortcut(shortcutEdit)
	h.shortcut(shortcutDelete)
	if h.m.editor != nil || h.m.modalOpen() {
		t.Error("Ctrl+E / Ctrl+Shift+D opened something with no selection")
	}
}

func TestSaveClearsQueryThatHidesTheSnippet(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.typeText("greet")
	createViaKeyboard(h, "Brand new", "nothing in common")
	if h.m.search.Text != "" || h.m.query != "" {
		t.Errorf("query %q should have been cleared", h.m.search.Text)
	}
	if h.selectedTitle() != "Brand new" || len(h.m.Results()) != 5 {
		t.Errorf("selected %q among %d results", h.selectedTitle(), len(h.m.Results()))
	}

	// Editing so that the snippet no longer matches the query clears it too.
	h.typeText("brand")
	h.shortcut(shortcutEdit)
	h.m.editor.title.SetText("Renamed")
	h.m.editor.body.SetText("different")
	h.shortcut(shortcutSave)
	if h.m.search.Text != "" || h.selectedTitle() != "Renamed" {
		t.Errorf("query=%q selected=%q", h.m.search.Text, h.selectedTitle())
	}

	// A query that still matches is kept.
	h.typeText("renam")
	h.shortcut(shortcutEdit)
	h.m.editor.body.SetText("still here")
	h.shortcut(shortcutSave)
	if h.m.search.Text != "renam" || h.selectedTitle() != "Renamed" {
		t.Errorf("query=%q selected=%q", h.m.search.Text, h.selectedTitle())
	}
}

func TestDeleteSelectionRules(t *testing.T) {
	h := newHarness(t, sampleLibrary()) // Address, Greeting, Sign-off, Zebra notes
	h.key(fyne.KeyDown)                 // Greeting
	h.shortcut(shortcutDelete)
	top := h.topDialog()
	if top == nil || top.title != "Delete snippet?" || top.message != "Delete “Greeting”? This cannot be undone." {
		t.Fatalf("confirmation = %+v", top)
	}
	h.confirmTop()
	if h.selectedTitle() != "Sign-off" {
		t.Errorf("after deleting the 2nd row, selection = %q; want the row now at that index", h.selectedTitle())
	}
	if h.m.Status() != "Deleted “Greeting”." || h.m.modalOpen() {
		t.Errorf("status=%q modal=%v", h.m.Status(), h.m.modalOpen())
	}
	if want := []string{"Sign-off", "Address", "Zebra notes"}; !equalStrings(h.diskTitles(), want) {
		t.Errorf("disk = %v, want %v (creation order)", h.diskTitles(), want)
	}

	h.key(fyne.KeyPageDown) // last row
	h.shortcut(shortcutDelete)
	h.confirmTop()
	if h.selectedTitle() != "Sign-off" {
		t.Errorf("after deleting the last row, selection = %q; want the previous one", h.selectedTitle())
	}

	h.shortcut(shortcutDelete)
	h.confirmTop()
	h.shortcut(shortcutDelete)
	h.confirmTop()
	if h.m.SelectedID() != "" || len(h.m.Library().Snippets) != 0 {
		t.Errorf("selection %q after deleting everything", h.m.SelectedID())
	}
	if !h.m.emptyLabel.Visible() || !h.m.deleteBtn.Disabled() {
		t.Error("empty state not shown after deleting everything")
	}
	if len(h.diskTitles()) != 0 {
		t.Errorf("disk = %v", h.diskTitles())
	}
}

func TestDeleteCancelKeepsSnippet(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutDelete)
	h.dismissTop()
	if len(h.m.Library().Snippets) != 4 || h.m.modalOpen() {
		t.Error("cancelling the confirmation must keep the snippet")
	}
	if h.focused() != fyne.Focusable(h.m.search) {
		t.Errorf("focus = %T, want search", h.focused())
	}
}

func TestDeleteSaveFailureLeavesLibraryUnchanged(t *testing.T) {
	var fs *failingStorage
	h := newHarnessWrapping(t, sampleLibrary(), func(s *store.Store) Storage {
		fs = &failingStorage{Store: s}
		return fs
	})
	before := h.primaryBytes()
	fs.fail = true
	h.shortcut(shortcutDelete)
	h.confirmTop()
	if len(h.m.Library().Snippets) != 4 || h.primaryBytes() != before {
		t.Error("failed delete changed data")
	}
	top := h.topDialog()
	if top == nil || top.message != "Could not delete: disk full" {
		t.Errorf("error dialog = %+v", top)
	}
	if h.m.Status() == "Deleted “Address”." {
		t.Error("status claims the delete succeeded")
	}
}

func TestCancelWithChangesAsksToDiscard(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutNew)
	h.typeText("draft")
	h.key(fyne.KeyEscape)
	top := h.topDialog()
	if top == nil || top.title != "Discard changes?" {
		t.Fatalf("expected the discard confirmation, got %+v", top)
	}
	if h.m.editor == nil {
		t.Fatal("editor closed before the user chose")
	}

	h.dismissTop() // Keep editing
	if h.m.editor == nil || h.m.editor.title.Text != "draft" || len(h.m.dialogs) != 1 {
		t.Error("Keep editing must leave the editor open with its text")
	}
	h.key(fyne.KeyEscape)
	h.confirmTop() // Discard
	if h.m.editor != nil || h.m.modalOpen() {
		t.Error("Discard should close the editor")
	}
	if len(h.m.Library().Snippets) != 4 {
		t.Error("discarded draft was saved")
	}
	if h.focused() != fyne.Focusable(h.m.search) {
		t.Errorf("focus = %T, want search", h.focused())
	}
}

func TestCancelWithoutChangesClosesAtOnce(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutNew)
	h.key(fyne.KeyEscape)
	if h.m.editor != nil || h.m.modalOpen() {
		t.Error("untouched editor should close on Escape without asking")
	}

	h.key(fyne.KeyDown)
	h.shortcut(shortcutEdit)
	h.key(fyne.KeyEscape) // body unchanged from its initial value
	if h.m.editor != nil || h.m.modalOpen() {
		t.Error("unchanged edit should close on Escape without asking")
	}

	h.shortcut(shortcutNew)
	h.typeText("x")
	test.Tap(h.m.editor.cancelBtn)
	if h.topDialog() == nil || h.topDialog().title != "Discard changes?" {
		t.Error("Cancel button should ask about changes too")
	}
}

func TestModalGuardBlocksMainShortcuts(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutNew)
	ed := h.m.editor

	h.shortcut(shortcutNew) // routed through the focused editor entry
	h.m.handleShortcut(shortcutNew)
	h.m.handleShortcut(shortcutEdit)
	h.m.handleShortcut(shortcutDelete)
	h.m.handleShortcut(shortcutFind)
	if h.m.editor != ed || len(h.m.dialogs) != 1 {
		t.Errorf("shortcuts opened more dialogs: editor same=%v dialogs=%d", h.m.editor == ed, len(h.m.dialogs))
	}
	if !h.m.handleShortcut(shortcutNew) {
		t.Error("a main shortcut should still be reported as handled (swallowed) while modal")
	}

	// The same guard covers other dialogs, e.g. the delete confirmation.
	ed.title.SetText("")
	h.key(fyne.KeyEscape)
	h.key(fyne.KeyDown)
	h.shortcut(shortcutDelete)
	h.m.handleShortcut(shortcutDelete)
	h.m.handleShortcut(shortcutNew)
	if len(h.m.dialogs) != 1 || h.m.editor != nil {
		t.Errorf("dialogs=%d editor=%v", len(h.m.dialogs), h.m.editor != nil)
	}
}

func TestToolbarButtonsOpenEditorsAndDelete(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	test.Tap(h.m.newBtn)
	if h.m.editor == nil {
		t.Fatal("New button did not open the editor")
	}
	h.key(fyne.KeyEscape)
	test.Tap(h.m.editBtn)
	if h.m.editor == nil || h.m.editor.id != h.m.SelectedID() {
		t.Fatal("Edit button did not open the selected snippet")
	}
	h.key(fyne.KeyEscape)
	test.Tap(h.m.deleteBtn)
	if top := h.topDialog(); top == nil || top.title != "Delete snippet?" {
		t.Fatal("Delete button did not ask for confirmation")
	}
}

// The real driver turns Tab into Canvas.FocusNext for widgets that don't accept
// tabs, so walking FocusNext reproduces what a keyboard user gets.
func TestEditorTabOrderIsTitleBodyButtons(t *testing.T) {
	h := newHarness(t, snippets.NewLibrary())
	h.shortcut(shortcutNew)
	c := h.m.win.Canvas()
	c.Focus(h.m.editor.title)
	c.FocusNext()
	if c.Focused() != h.m.editor.body {
		t.Fatalf("Tab from title focused %T, want the body", c.Focused())
	}
	c.FocusNext()
	if c.Focused() != h.m.editor.cancelBtn {
		t.Fatalf("after the body, focus went to %T, want the Cancel button", c.Focused())
	}
	c.FocusPrevious()
	c.FocusPrevious()
	if c.Focused() != h.m.editor.title {
		t.Errorf("Shift+Tab back from Cancel ended on %T, want the title", c.Focused())
	}
}

func TestTabInBodyInsertsATab(t *testing.T) {
	h := newHarness(t, snippets.NewLibrary())
	h.shortcut(shortcutNew)
	h.m.win.Canvas().Focus(h.m.editor.body)
	h.typeText("a")
	h.key(fyne.KeyTab)
	h.typeText("b")
	if h.m.editor.body.Text != "a\tb" {
		t.Errorf("body = %q; Fyne's Entry inserts a tab for Tab in a multiline field", h.m.editor.body.Text)
	}
	h.shortcut(shortcutSave) // Ctrl+S still reaches the editor from the body
	if h.m.editor == nil || h.m.editor.errLabel.Text != "Title is required." {
		t.Error("Ctrl+S did not run from the body after a Tab")
	}
}

func TestPastedTabsAndOddTextSurviveSave(t *testing.T) {
	h := newHarness(t, snippets.NewLibrary())
	for i, s := range []string{"tab\tseparated", "emoji \U0001F44B\U0001F3FD", "日本語\n  indented  \n"} {
		h.shortcut(shortcutNew)
		h.m.editor.title.SetText("T" + string(rune('A'+i)))
		h.m.editor.body.SetText(strings.ReplaceAll(s, "\n", "\r\n"))
		h.shortcut(shortcutSave)
		got := store.New(h.store.Dir()).Load().Library.Snippets[i].Body
		if got != s {
			t.Errorf("snippet %d body = %q, want %q", i, got, s)
		}
	}
}
