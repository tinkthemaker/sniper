package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/tinkthemaker/sniper/internal/snippets"
	"github.com/tinkthemaker/sniper/internal/store"
)

func TestEditorTitleLimitIs200Runes(t *testing.T) {
	ok, over := fixtureFirstTitle(t, "title-200.json"), fixtureFirstTitle(t, "title-201.json")
	if n := len([]rune(ok)); n != 200 || len([]rune(over)) != 201 || len(ok) <= 200 {
		t.Fatalf("fixtures are %d and %d runes (%d bytes)", n, len([]rune(over)), len(ok))
	}
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutNew)
	ed := h.m.editor
	ed.title.SetText(over)
	ed.body.SetText("b")
	h.shortcut(shortcutSave)
	want := "Title is too long (201 characters; limit 200)."
	if h.m.editor == nil || ed.errLabel.Text != want || !ed.errLabel.Visible() {
		t.Fatalf("error = %q visible=%v", ed.errLabel.Text, ed.errLabel.Visible())
	}
	if h.focused() != fyne.Focusable(ed.title) {
		t.Errorf("focus = %T, want the title", h.focused())
	}
	if len(h.diskTitles()) != 4 {
		t.Error("an over-long title was written")
	}
	ed.title.SetText(ok)
	h.shortcut(shortcutSave)
	if h.m.editor != nil || h.selectedTitle() != ok || h.diskTitles()[4] != ok {
		t.Error("a 200-rune title must save")
	}
	if h.m.previewTitle.Text != up(ok) {
		t.Error("preview does not show the long title")
	}
}

func TestEditorBodyLimitIs65536Bytes(t *testing.T) {
	h := newHarness(t, snippets.NewLibrary())
	h.shortcut(shortcutNew)
	ed := h.m.editor
	ed.title.SetText("Big")
	ed.body.SetText(strings.Repeat("x", 65537))
	h.shortcut(shortcutSave)
	if want := "Body is too long (65,537 bytes; limit 65,536)."; ed.errLabel.Text != want {
		t.Fatalf("error = %q", ed.errLabel.Text)
	}
	if h.focused() != fyne.Focusable(ed.body) || h.primaryBytes() != "" {
		t.Errorf("focus %T, file %q", h.focused(), h.primaryBytes())
	}
	ed.body.SetText(strings.Repeat("x", 65536))
	h.shortcut(shortcutSave)
	if h.m.editor != nil || len(h.m.Library().Snippets[0].Body) != 65536 {
		t.Error("a body of exactly 65,536 bytes must save")
	}
	// Characters that need several bytes count as bytes, not runes.
	h.shortcut(shortcutNew)
	h.m.editor.title.SetText("Wide")
	h.m.editor.body.SetText(strings.Repeat("é", 32769)) // 65,538 bytes
	h.shortcut(shortcutSave)
	if want := "Body is too long (65,538 bytes; limit 65,536)."; h.m.editor.errLabel.Text != want {
		t.Errorf("error = %q", h.m.editor.errLabel.Text)
	}
}

func TestEditorNormalizesAndValidatesWhatWasTyped(t *testing.T) {
	h := newHarness(t, snippets.NewLibrary())
	h.shortcut(shortcutNew)
	ed := h.m.editor
	if ed.errLabel.Visible() {
		t.Error("the error label must start hidden")
	}

	ed.title.SetText("  \t  ")
	ed.body.SetText("body")
	h.shortcut(shortcutSave)
	if ed.errLabel.Text != "Title is required." {
		t.Errorf("blank title: %q", ed.errLabel.Text)
	}
	ed.title.SetText("Title")
	ed.body.SetText(" \n\t \r\n")
	h.shortcut(shortcutSave)
	if ed.errLabel.Text != "Body is required." || h.focused() != fyne.Focusable(ed.body) {
		t.Errorf("blank body: %q focus %T", ed.errLabel.Text, h.focused())
	}
	ed.title.SetText("bad\x00title")
	ed.body.SetText("body")
	h.shortcut(shortcutSave)
	if ed.errLabel.Text != "Title contains characters that cannot be saved." {
		t.Errorf("NUL in title: %q", ed.errLabel.Text)
	}
	ed.title.SetText("ok")
	ed.body.SetText("bad\x00body")
	h.shortcut(shortcutSave)
	if ed.errLabel.Text != "Body contains characters that cannot be saved." {
		t.Errorf("NUL in body: %q", ed.errLabel.Text)
	}
	if len(h.m.Library().Snippets) != 0 || h.primaryBytes() != "" {
		t.Fatal("something was saved despite the errors")
	}

	// Title: line breaks become spaces, edges are trimmed. Body: kept as typed.
	ed.title.SetText("  two\r\nlines\nhere \r ")
	ed.body.SetText("  keep \r\n\ttabs  \r")
	h.shortcut(shortcutSave)
	if h.m.editor != nil {
		t.Fatalf("valid input did not save: %q", ed.errLabel.Text)
	}
	got := store.New(h.store.Dir()).Load().Library.Snippets[0]
	if got.Title != "two  lines here" || got.Body != "  keep \n\ttabs  \n" {
		t.Errorf("saved %q / %q", got.Title, got.Body)
	}

	h.shortcut(shortcutNew) // a fresh editor starts with no error showing
	if h.m.editor.errLabel.Visible() {
		t.Error("the error from the previous editor is still showing")
	}
}

func TestEditorButtonsSaveAndCancel(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutNew)
	u := h.overlay()
	if got := u.buttonLabels(); !equalStrings(got, []string{"Cancel", "Save"}) {
		t.Fatalf("editor buttons = %v", got)
	}
	if u.button("Save").Importance != widget.HighImportance || u.button("Cancel").Importance == widget.HighImportance {
		t.Error("Save must be the primary button and Cancel secondary")
	}
	h.m.editor.title.SetText("Via button")
	h.m.editor.body.SetText("body")
	test.Tap(u.button("Save"))
	if h.m.editor != nil || h.m.Status() != "Saved “Via button”." || len(h.diskTitles()) != 5 {
		t.Errorf("Save button: editor open=%v status %q", h.m.editor != nil, h.m.Status())
	}

	h.shortcut(shortcutNew)
	test.Tap(h.overlay().button("Cancel"))
	if h.m.editor != nil || len(h.diskTitles()) != 5 {
		t.Error("Cancel on an untouched editor must close it")
	}
}

func TestEditorHeadingsAndStartingValues(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutNew)
	if !h.overlay().hasText("New snippet") || !h.overlay().hasText("Title") || !h.overlay().hasText("Body") {
		t.Errorf("new editor texts = %q", h.overlay().texts)
	}
	if h.m.editor.title.Text != "" || h.m.editor.body.Text != "" || !h.m.editor.body.MultiLine || h.m.editor.title.MultiLine {
		t.Error("new editor must start empty with a single-line title and a multiline body")
	}
	if h.m.editor.body.Wrapping != fyne.TextWrapWord {
		t.Error("the body must wrap at words")
	}
	h.key(fyne.KeyEscape)
	h.shortcut(shortcutEdit)
	if !h.overlay().hasText("Edit snippet") || h.m.editor.title.Text != "Address" || h.m.editor.body.Text != "1 Main St\nSpringfield" {
		t.Errorf("edit editor: %q %q", h.m.editor.title.Text, h.m.editor.body.Text)
	}
}

// The editor is the window minus a 40 px margin. Fyne clamps a dialog to its
// window, so the 420x380 floor only shows on windows at least that large.
func TestEditorSizeIsTheWindowMinusAMargin(t *testing.T) {
	for _, win := range []fyne.Size{fyne.NewSize(520, 600), fyne.NewSize(900, 700)} {
		h := newHarness(t, sampleLibrary())
		h.m.win.Resize(win)
		h.shortcut(shortcutNew)
		var pop *widget.PopUp
		walk(h.m.win.Canvas().Overlays().Top(), func(o fyne.CanvasObject) {
			if p, ok := o.(*widget.PopUp); ok {
				pop = p
			}
		})
		if pop == nil {
			t.Fatal("no popup in the overlay")
		}
		want := fyne.NewSize(win.Width-40, win.Height-40)
		if got := pop.Content.Size(); got != want {
			t.Errorf("window %v: editor is %v, want %v", win, got, want)
		}
	}
}

func TestDiscardDialogWording(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutNew)
	h.typeText("x")
	h.key(fyne.KeyEscape)
	top := h.topDialog()
	if top.title != "Discard changes?" || top.message != "Your changes to this snippet will be lost." {
		t.Errorf("discard dialog = %+v", top)
	}
	if got := h.overlay().buttonLabels(); !equalStrings(got, []string{"Keep editing", "Discard"}) {
		t.Errorf("discard buttons = %v", got)
	}
}

func TestEditCancelAndRevertLeaveTheLibraryAlone(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	before := h.primaryBytes()
	id := h.m.SelectedID()

	h.shortcut(shortcutEdit)
	h.m.editor.title.SetText("Changed")
	h.m.editor.body.SetText("changed body")
	h.key(fyne.KeyEscape)
	h.confirmTop() // Discard
	if h.primaryBytes() != before || h.m.SelectedID() != id || h.selectedTitle() != "Address" {
		t.Error("discarding an edit changed the library or the selection")
	}

	h.shortcut(shortcutEdit)
	h.m.editor.title.SetText("Changed")
	h.m.editor.title.SetText("Address") // typed back to the original: no longer a change
	h.key(fyne.KeyEscape)
	if h.m.editor != nil || h.m.modalOpen() {
		t.Error("an edit typed back to its original value should close without asking")
	}
}

func TestEditSaveFailureKeepsEditorAndPreview(t *testing.T) {
	var fs *failingStorage
	h := newHarness2(t, &fs)
	before := h.primaryBytes()
	fs.fail = true
	h.shortcut(shortcutEdit)
	h.m.editor.body.SetText("will not stick")
	h.shortcut(shortcutSave)
	if h.m.editor == nil || h.m.editor.errLabel.Text != "Could not save: disk full. Your changes are still here; try again." {
		t.Fatal("expected the save error in the open editor")
	}
	if h.m.previewBody.Text != "1 Main St\nSpringfield" || h.primaryBytes() != before || h.m.Status() != "" {
		t.Error("the window or file changed after a failed save")
	}
	fs.fail = false
	test.Tap(h.overlay().button("Save"))
	if h.m.editor != nil || h.m.previewBody.Text != "will not stick" || h.m.Status() != "Saved “Address”." {
		t.Error("retry from the same form did not save")
	}
}

func newHarness2(t *testing.T, out **failingStorage) *harness {
	return newHarnessWrapping(t, sampleLibrary(), func(s *store.Store) Storage {
		*out = &failingStorage{Store: s}
		return *out
	})
}

func TestEditingOneOfSeveralSameTitledSnippets(t *testing.T) {
	h := newHarness(t, fixture(t, "duplicates.json"))
	h.key(fyne.KeyDown) // second Greeting
	second := h.m.SelectedID()
	h.shortcut(shortcutEdit)
	if h.m.editor.body.Text != "Hi there!\nQuick question about my order." {
		t.Fatalf("editing the wrong duplicate: %q", h.m.editor.body.Text)
	}
	h.m.editor.body.SetText("Edited")
	h.shortcut(shortcutSave)
	if h.m.SelectedID() != second || h.body(second) != "Edited" {
		t.Error("the edited duplicate is not the selected one")
	}
	bodies := []string{}
	for _, s := range h.m.Library().Snippets {
		bodies = append(bodies, s.Body)
	}
	if bodies[0] != "Hello,\nThanks for reaching out." || bodies[1] != "Edited" || bodies[2] != "Dear team,\nPlease find the report attached." {
		t.Errorf("bodies = %q", bodies)
	}
}
