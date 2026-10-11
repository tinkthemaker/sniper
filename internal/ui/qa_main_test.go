package ui

import (
	"fmt"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"

	"github.com/tinkthemaker/sniper/internal/snippets"
)

func TestPageKeysMoveByTenFromSearchAndList(t *testing.T) {
	h := newHarness(t, bigLibrary(t, 300))
	check := func(want string) {
		t.Helper()
		if h.selectedTitle() != want {
			t.Fatalf("selected %q, want %q", h.selectedTitle(), want)
		}
	}
	h.key(fyne.KeyPageDown)
	check("Snippet 010")
	h.key(fyne.KeyPageDown)
	check("Snippet 020")
	h.key(fyne.KeyPageUp)
	check("Snippet 010")
	h.focus(h.m.list)
	h.key(fyne.KeyPageDown)
	check("Snippet 020")
	h.key(fyne.KeyPageUp)
	h.key(fyne.KeyPageUp)
	h.key(fyne.KeyPageUp) // clamps at the top
	check("Snippet 000")
	for i := 0; i < 40; i++ {
		h.key(fyne.KeyPageDown)
	}
	check("Snippet 299")
	if got := h.m.list.Length(); got != 300 {
		t.Errorf("list length %d", got)
	}
}

func TestStatusStaysUntilTheNextAction(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.key(fyne.KeyReturn)
	copied := h.m.Status()
	if copied != "Copied “Address”." {
		t.Fatalf("status = %q", copied)
	}
	h.typeText("sign")
	h.key(fyne.KeyDown)
	h.key(fyne.KeyEscape)
	h.shortcut(shortcutFind)
	if h.m.Status() != copied {
		t.Errorf("status changed to %q without an action", h.m.Status())
	}
	h.key(fyne.KeyDown)
	h.key(fyne.KeyReturn)
	if h.m.Status() != "Copied “Greeting”." {
		t.Errorf("next copy did not replace the status: %q", h.m.Status())
	}
}

func TestQueryChangeSelectsFirstResult(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.key(fyne.KeyDown)
	h.key(fyne.KeyDown)
	if h.selectedTitle() != "Sign-off" {
		t.Fatalf("setup: %q", h.selectedTitle())
	}
	h.typeText("e") // Address, Greeting, Sign-off, Zebra notes all contain "e"
	if h.selectedTitle() != "Address" {
		t.Errorf("selection after the query changed = %q, want the first result", h.selectedTitle())
	}
	h.key(fyne.KeyDown)
	h.key(fyne.KeyEscape)
	if h.selectedTitle() != "Address" {
		t.Errorf("selection after clearing the query = %q", h.selectedTitle())
	}
}

func TestSearchRankingAndMatchingThroughTheUI(t *testing.T) {
	lib := snippets.NewLibrary()
	for _, s := range [][2]string{
		{"Alpha", "has a needle in the body"},
		{"Needle tool", "unrelated"},
		{"Beta", "needle again, and thread"},
		{"Thread needle", "x"},
	} {
		var err error
		if lib, _, err = snippets.Add(lib, s[0], s[1]); err != nil {
			t.Fatal(err)
		}
	}
	h := newHarness(t, lib)
	h.typeText("NEEDLE")
	if want := []string{"Needle tool", "Thread needle", "Alpha", "Beta"}; !equalStrings(h.titles(), want) {
		t.Errorf("title matches must rank first, case-insensitively: %v", h.titles())
	}
	if h.selectedTitle() != "Needle tool" {
		t.Errorf("selected %q", h.selectedTitle())
	}
	h.key(fyne.KeyEscape)
	h.typeText("needle  thread") // two terms, each may match in either field
	if want := []string{"Thread needle", "Beta"}; !equalStrings(h.titles(), want) {
		t.Errorf("two-term query results = %v", h.titles())
	}
	if h.m.countSeg.Text != "2 OF 4" {
		t.Errorf("count = %q", h.m.countSeg.Text)
	}
}

func TestWhitespaceOnlyQueryListsEverything(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.typeText("   ")
	if len(h.m.Results()) != 4 || h.m.emptyLabel.Visible() || h.selectedTitle() != "Address" {
		t.Errorf("results=%v empty=%v selected=%q", h.titles(), h.m.emptyLabel.Visible(), h.selectedTitle())
	}
	// Spaces alone are no query, so the header shows the library count, not "4 OF 4".
	if h.m.countSeg.Text != "4 IN LIBRARY" {
		t.Errorf("count = %q, want 4 IN LIBRARY", h.m.countSeg.Text)
	}
	h.typeText(" zebra ")
	if h.m.countSeg.Text != "1 OF 4" {
		t.Errorf("count with a real query = %q, want 1 OF 4", h.m.countSeg.Text)
	}
}

func TestCountWordingIsSingularForOne(t *testing.T) {
	for _, c := range []struct{ got, want string }{
		{importConfirmText(1, "a.json"), "Import 1 snippet from a.json? Importing the same file twice creates duplicates."},
		{importConfirmText(2, "a.json"), "Import 2 snippets from a.json? Importing the same file twice creates duplicates."},
		{importedStatus(1), "Imported 1 snippet."},
		{importedStatus(3), "Imported 3 snippets."},
		{exportedStatus(1, "out.json"), "Exported 1 snippet to out.json."},
		{exportedStatus(0, "out.json"), "Exported 0 snippets to out.json."},
	} {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}

func TestPreviewFollowsSelection(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	if !h.m.previewBody.Selectable || h.m.previewBody.Wrapping != fyne.TextWrapWord {
		t.Error("the preview body must be selectable and word-wrapped")
	}
	if h.m.previewTitle.Text != "Address" || h.m.previewBody.Text != "1 Main St\nSpringfield" {
		t.Errorf("preview = %q / %q", h.m.previewTitle.Text, h.m.previewBody.Text)
	}
	h.key(fyne.KeyDown)
	if h.m.previewTitle.Text != "Greeting" || h.m.previewBody.Text != "Hello,\nThanks for reaching out." {
		t.Errorf("preview after Down = %q / %q", h.m.previewTitle.Text, h.m.previewBody.Text)
	}
	h.typeText("zzz")
	if h.m.previewTitle.Text != "" || h.m.previewBody.Text != "Select a snippet to preview it." {
		t.Errorf("preview with no result = %q / %q", h.m.previewTitle.Text, h.m.previewBody.Text)
	}
}

func TestCopyPreservesTheUnicodeCorpus(t *testing.T) {
	lib := corpus(t)
	h := newHarness(t, lib)
	n := len(h.m.Results())
	if n != 12 {
		t.Fatalf("corpus has %d snippets", n)
	}
	for i := 0; i < n; i++ {
		s := h.m.Results()[i]
		if h.m.SelectedID() != s.ID {
			t.Fatalf("row %d: selection is %q", i, h.selectedTitle())
		}
		h.key(fyne.KeyReturn)
		if h.cb.last() != s.Body {
			t.Errorf("%q: clipboard %q, want %q", s.Title, h.cb.last(), s.Body)
		}
		if h.m.Status() != "Copied “"+s.Title+"”." {
			t.Errorf("%q: status %q", s.Title, h.m.Status())
		}
		if h.m.previewBody.Text != s.Body {
			t.Errorf("%q: preview %q", s.Title, h.m.previewBody.Text)
		}
		h.key(fyne.KeyDown)
	}
}

func TestShortcutsReachTheWindowFromAListAndFromButtons(t *testing.T) {
	for name, focus := range map[string]func(*harness) fyne.Focusable{
		"list":        func(h *harness) fyne.Focusable { return h.m.list },
		"copy button": func(h *harness) fyne.Focusable { return h.m.copyBtn },
		"new button":  func(h *harness) fyne.Focusable { return h.m.newBtn },
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t, sampleLibrary())
			h.focus(focus(h))
			h.canvasShortcut(shortcutNew)
			if h.m.editor == nil || h.m.editor.id != "" {
				t.Fatal("Ctrl+N did not open a new-snippet editor")
			}
			h.key(fyne.KeyEscape)
			h.focus(focus(h))
			h.canvasShortcut(shortcutEdit)
			if h.m.editor == nil || h.m.editor.id != h.m.SelectedID() {
				t.Fatal("Ctrl+E did not open the selected snippet")
			}
			h.key(fyne.KeyEscape)
			h.focus(focus(h))
			h.canvasShortcut(shortcutDelete)
			if top := h.topDialog(); top == nil || top.title != "Delete snippet?" {
				t.Fatal("Ctrl+Shift+D did not ask to delete")
			}
			h.dismissTop()
			h.focus(focus(h))
			h.canvasShortcut(shortcutFind)
			if h.focused() != fyne.Focusable(h.m.search) {
				t.Errorf("Ctrl+F focused %T", h.focused())
			}
		})
	}
}

func TestPlainCtrlDDoesNotDelete(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.shortcut(ctrl(fyne.KeyD))
	h.canvasShortcut(ctrl(fyne.KeyD))
	if h.m.modalOpen() || len(h.m.Library().Snippets) != 4 {
		t.Error("Ctrl+D without Shift must not delete")
	}
}

func TestModalGuardAppliesToEveryDialogOnTheCanvasPath(t *testing.T) {
	openers := map[string]func(*harness){
		"editor":      func(h *harness) { h.m.newSnippet() },
		"delete":      func(h *harness) { h.m.deleteSelected() },
		"about":       func(h *harness) { h.m.showAbout() },
		"shortcuts":   func(h *harness) { h.m.showShortcuts() },
		"import file": func(h *harness) { h.m.importFile() },
		"export file": func(h *harness) { h.m.exportFile() },
	}
	for name, open := range openers {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t, sampleLibrary())
			open(h)
			n := len(h.m.dialogs)
			if n != 1 {
				t.Fatalf("opened %d dialogs", n)
			}
			for _, s := range []*desktop.CustomShortcut{shortcutNew, shortcutEdit, shortcutDelete, shortcutFind} {
				h.canvasShortcut(s)
			}
			if len(h.m.dialogs) != n {
				t.Errorf("a shortcut opened another dialog: %d open", len(h.m.dialogs))
			}
			if len(h.m.Library().Snippets) != 4 {
				t.Error("library changed")
			}
		})
	}
}

func TestMenuItemsAreBlockedWhileADialogIsOpen(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	menu := h.m.win.MainMenu()
	h.m.showAbout()
	for _, i := range []struct{ menu, item int }{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}} {
		menu.Items[i.menu].Items[i.item].Action()
	}
	if len(h.m.dialogs) != 1 || len(h.opened) != 0 {
		t.Errorf("dialogs=%d opened=%v", len(h.m.dialogs), h.opened)
	}
}

func TestDeleteWithAnActiveQueryPicksAmongTheResults(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	h.typeText("greet") // Greeting, Zebra notes
	h.key(fyne.KeyDown)
	h.shortcut(shortcutDelete)
	h.confirmTop()
	if h.selectedTitle() != "Greeting" || h.m.countSeg.Text != "1 OF 3" {
		t.Errorf("selected %q count %q", h.selectedTitle(), h.m.countSeg.Text)
	}
	if h.m.search.Text != "greet" {
		t.Errorf("delete changed the query to %q", h.m.search.Text)
	}
	h.shortcut(shortcutDelete)
	h.confirmTop()
	if h.m.SelectedID() != "" || !h.m.emptyLabel.Visible() || h.m.emptyLabel.Text != "No snippets match “greet”." {
		t.Errorf("selected %q, empty label %q", h.m.SelectedID(), h.m.emptyLabel.Text)
	}
	if len(h.m.Library().Snippets) != 2 {
		t.Errorf("library has %d snippets", len(h.m.Library().Snippets))
	}
}

func TestHeaderCountTracksCreateAndDelete(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	createViaKeyboard(h, "Fifth", "five")
	if h.m.countSeg.Text != "5 IN LIBRARY" {
		t.Errorf("after create: %q", h.m.countSeg.Text)
	}
	h.shortcut(shortcutDelete)
	h.confirmTop()
	if h.m.countSeg.Text != "4 IN LIBRARY" {
		t.Errorf("after delete: %q", h.m.countSeg.Text)
	}
}

func TestEmptyLibraryStartsWithNoSelectionAndNewWorks(t *testing.T) {
	h := newHarness(t, snippets.NewLibrary())
	if h.m.SelectedID() != "" || h.m.countSeg.Text != "0 IN LIBRARY" {
		t.Errorf("selected %q count %q", h.m.SelectedID(), h.m.countSeg.Text)
	}
	test.Tap(h.m.newBtn)
	if h.m.editor == nil {
		t.Fatal("New is always enabled")
	}
	h.key(fyne.KeyEscape)
	h.typeText("x") // search regained focus
	if h.m.emptyLabel.Text != "No snippets yet. Press Ctrl+N to create one." {
		t.Errorf("an empty library shows %q even with a query", h.m.emptyLabel.Text)
	}
}

func TestCRLFFixtureLoadsNormalized(t *testing.T) {
	h := newHarness(t, fixture(t, "crlf-bodies.json"))
	for _, s := range h.m.Results() {
		if strings.ContainsAny(s.Body, "\r") || strings.ContainsAny(s.Title, "\r\n") {
			t.Errorf("%q still has CR or LF: %q", s.Title, s.Body)
		}
	}
	if got := fmt.Sprint(h.titles()); got != "[Old Mac lines Title with  break Windows lines]" {
		t.Errorf("titles = %s", got)
	}
}
