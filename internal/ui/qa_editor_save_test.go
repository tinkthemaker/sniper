package ui

import (
	"testing"

	"fyne.io/fyne/v2"
)

// After Tab, focus can rest on Cancel or Save, which are not entries, so the
// editor's own routing never sees Ctrl+S; the canvas handler must.
func TestCtrlSSavesWhileAnEditorButtonHasFocus(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(shortcutNew)
	h.m.editor.title.SetText("From a button")
	h.m.editor.body.SetText("body")
	h.focus(h.m.editor.cancelBtn)
	h.canvasShortcut(shortcutSave)
	if h.m.editor != nil || len(h.diskTitles()) != 5 || h.m.Status() != "Saved “From a button”." {
		t.Error("Ctrl+S did nothing while a button had focus")
	}
}

func TestCtrlSDoesNothingWithoutAnEditorOrOverItsDiscardDialog(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.canvasShortcut(shortcutSave)
	if h.m.modalOpen() || len(h.m.Library().Snippets) != 4 {
		t.Fatal("Ctrl+S with no editor did something")
	}
	h.shortcut(shortcutNew)
	h.m.editor.title.SetText("Draft")
	h.m.editor.body.SetText("draft body")
	h.key(fyne.KeyEscape) // discard confirmation now sits over the editor
	h.canvasShortcut(shortcutSave)
	if len(h.m.Library().Snippets) != 4 || h.m.editor == nil {
		t.Error("Ctrl+S saved the draft while the discard confirmation was showing")
	}
}
