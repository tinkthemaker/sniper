package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/tinkthemaker/sniper/internal/snippets"
)

func TestStartupFocusTypingAndSelection(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	if h.focused() != fyne.Focusable(h.m.search) {
		t.Fatalf("startup focus is %T, want the search entry", h.focused())
	}
	if want := []string{"Address", "Greeting", "Sign-off", "Zebra notes"}; !equalStrings(h.titles(), want) {
		t.Errorf("initial results = %v, want %v", h.titles(), want)
	}
	if h.selectedTitle() != "Address" {
		t.Errorf("startup selection = %q, want the first result", h.selectedTitle())
	}

	h.typeText("greet")
	if want := []string{"Greeting", "Zebra notes"}; !equalStrings(h.titles(), want) {
		t.Errorf("results for %q = %v, want %v", "greet", h.titles(), want)
	}
	if h.selectedTitle() != "Greeting" {
		t.Errorf("selection after typing = %q, want the first result", h.selectedTitle())
	}
}

func TestArrowKeysClampAndEnterCopies(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.key(fyne.KeyUp)
	if h.selectedTitle() != "Address" {
		t.Errorf("Up at the top moved selection to %q", h.selectedTitle())
	}
	h.key(fyne.KeyDown)
	h.key(fyne.KeyDown)
	if h.selectedTitle() != "Sign-off" {
		t.Errorf("after two Downs selection = %q", h.selectedTitle())
	}
	h.key(fyne.KeyPageDown)
	if h.selectedTitle() != "Zebra notes" {
		t.Errorf("PageDown must clamp at the end, got %q", h.selectedTitle())
	}
	h.key(fyne.KeyDown)
	if h.selectedTitle() != "Zebra notes" {
		t.Errorf("Down at the bottom must not wrap, got %q", h.selectedTitle())
	}
	h.key(fyne.KeyPageUp)
	if h.selectedTitle() != "Address" {
		t.Errorf("PageUp must clamp at the start, got %q", h.selectedTitle())
	}

	h.key(fyne.KeyDown) // Greeting
	h.key(fyne.KeyReturn)
	if h.cb.last() != "Hello,\nThanks for reaching out." {
		t.Errorf("clipboard = %q", h.cb.last())
	}
	if h.m.Status() != "Copied “Greeting”." {
		t.Errorf("status = %q", h.m.Status())
	}
	h.cb.copied = nil
	h.key(fyne.KeyEnter)
	if len(h.cb.copied) != 1 {
		t.Errorf("numpad Enter copied %d times", len(h.cb.copied))
	}
}

func TestListKeysMoveSelectionAndCopy(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.win.Canvas().Focus(h.m.list)
	h.key(fyne.KeyDown)
	h.key(fyne.KeyDown)
	if h.selectedTitle() != "Sign-off" {
		t.Errorf("list Down x2 selected %q", h.selectedTitle())
	}
	h.key(fyne.KeyUp)
	if h.selectedTitle() != "Greeting" {
		t.Errorf("list Up selected %q", h.selectedTitle())
	}
	h.key(fyne.KeyReturn)
	if h.cb.last() != "Hello,\nThanks for reaching out." {
		t.Errorf("clipboard = %q", h.cb.last())
	}
}

func TestTappingARowSelectsIt(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.m.list.Select(2)
	if h.selectedTitle() != "Sign-off" {
		t.Errorf("selected %q after selecting row 2", h.selectedTitle())
	}
	if h.m.previewTitle.Text != "Sign-off" || h.m.previewBody.Text != "Best regards,\nTink" {
		t.Errorf("preview = %q / %q", h.m.previewTitle.Text, h.m.previewBody.Text)
	}
}

func TestEscapeClearsQuery(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.typeText("zzz")
	if len(h.m.Results()) != 0 {
		t.Fatal("expected no results")
	}
	h.key(fyne.KeyEscape)
	if h.m.search.Text != "" || h.m.query != "" || len(h.m.Results()) != 4 {
		t.Errorf("after Escape: text=%q query=%q results=%d", h.m.search.Text, h.m.query, len(h.m.Results()))
	}
	h.key(fyne.KeyEscape) // no-op when already empty
	if len(h.m.Results()) != 4 {
		t.Error("Escape on an empty query changed the results")
	}
}

func TestCtrlFFocusesAndSelectsSearch(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.typeText("greet")
	h.m.win.Canvas().Focus(h.m.list)
	// With the list focused, the canvas shortcut path is used.
	if !h.m.handleShortcut(shortcutFind) {
		t.Fatal("Ctrl+F not handled")
	}
	if h.focused() != fyne.Focusable(h.m.search) {
		t.Errorf("focus = %T, want search", h.focused())
	}
	if h.m.search.SelectedText() != "greet" {
		t.Errorf("selected text = %q, want the whole query", h.m.search.SelectedText())
	}
	// With search already focused, the entry's own TypedShortcut must route it.
	h.m.win.Canvas().Focus(h.m.search)
	h.shortcut(shortcutFind)
	if h.m.search.SelectedText() != "greet" {
		t.Errorf("selected text via entry = %q", h.m.search.SelectedText())
	}
}

func TestEmptyStatesAndDisabledButtons(t *testing.T) {
	h := newHarness(t, snippets.NewLibrary())
	if !h.m.emptyLabel.Visible() || h.m.emptyLabel.Text != "No snippets yet. Press Ctrl+N to create one." {
		t.Errorf("empty library label: visible=%v %q", h.m.emptyLabel.Visible(), h.m.emptyLabel.Text)
	}
	if h.m.previewBody.Text != "Select a snippet to preview it." {
		t.Errorf("preview = %q", h.m.previewBody.Text)
	}
	for name, b := range map[string]interface{ Disabled() bool }{
		"Copy": h.m.copyBtn, "Edit": h.m.editBtn, "Delete": h.m.deleteBtn,
	} {
		if !b.Disabled() {
			t.Errorf("%s should be disabled with nothing selected", name)
		}
	}
	if h.m.newBtn.Disabled() {
		t.Error("New must always be enabled")
	}
	h.key(fyne.KeyReturn) // Enter with nothing selected must be harmless
	if len(h.cb.copied) != 0 {
		t.Error("copied with no selection")
	}

	h2 := newHarness(t, sampleLibrary())
	h2.typeText("nomatch")
	if !h2.m.emptyLabel.Visible() || h2.m.emptyLabel.Text != "No snippets match “nomatch”." {
		t.Errorf("no-match label: visible=%v %q", h2.m.emptyLabel.Visible(), h2.m.emptyLabel.Text)
	}
	if !h2.m.copyBtn.Disabled() {
		t.Error("Copy should be disabled when nothing matches")
	}
	h2.key(fyne.KeyEscape)
	if h2.m.emptyLabel.Visible() || h2.m.copyBtn.Disabled() {
		t.Error("empty state should clear with the query")
	}
}

func TestCopyButton(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	test.Tap(h.m.copyBtn)
	if h.cb.last() != "1 Main St\nSpringfield" || h.m.Status() != "Copied “Address”." {
		t.Errorf("clipboard=%q status=%q", h.cb.last(), h.m.Status())
	}
}

func TestShowAlreadyRunningAndFatal(t *testing.T) {
	a := test.NewTempApp(t)
	before := len(a.Driver().AllWindows())
	ShowAlreadyRunning(a)
	ShowFatal(a, "Title", "detail text")
	if n := len(a.Driver().AllWindows()) - before; n != 2 {
		t.Errorf("opened %d windows, want 2", n)
	}
}

func TestHeaderCountFollowsTheQuery(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	if h.m.countSeg.Text != "4 IN LIBRARY" {
		t.Errorf("count = %q", h.m.countSeg.Text)
	}
	h.typeText("greet")
	if h.m.countSeg.Text != "2 OF 4" {
		t.Errorf("count after query = %q", h.m.countSeg.Text)
	}
	h.key(fyne.KeyEscape)
	if h.m.countSeg.Text != "4 IN LIBRARY" {
		t.Errorf("count after clearing = %q", h.m.countSeg.Text)
	}
}

func TestRowShowsSelectionBar(t *testing.T) {
	h := newHarness(t, sampleLibrary()) // first result (Address) is selected
	barFor := func(i int) color.Color {
		row := h.m.newRow()
		h.m.updateRow(i, row)
		return h.m.rows[row].bar.FillColor
	}
	if got := barFor(0); got != bloodRed {
		t.Errorf("selected row bar = %v, want blood red", got)
	}
	if got := barFor(1); got != color.Color(color.Transparent) {
		t.Errorf("unselected row bar = %v, want transparent", got)
	}
	h.key(fyne.KeyDown)
	if got := barFor(1); got != bloodRed {
		t.Errorf("bar did not follow the selection: %v", got)
	}
}
