package ui

import (
	"fmt"
	"testing"

	"fyne.io/fyne/v2"
)

// The real driver turns Tab into Canvas.FocusNext, so walking it shows the
// order a keyboard user gets. Fyne orders by child position and NewBorder puts
// its center first, which once put the list before the search box.
func TestMainTabOrderFollowsTheSpec(t *testing.T) {
	h := newHarness(t, sampleLibrary())
	c := h.m.win.Canvas()
	if c.Focused() != fyne.Focusable(h.m.search) {
		t.Fatalf("start focus %T", c.Focused())
	}
	// The preview's selectable text is a Fyne-internal focus target.
	preview := "*widget.focusSelectable"
	want := []struct {
		name string
		obj  fyne.Focusable
	}{
		{"list", h.m.list}, {preview, nil}, {"Copy", h.m.copyBtn},
		{"New", h.m.newBtn}, {"Edit", h.m.editBtn}, {"Delete", h.m.deleteBtn},
	}
	same := func(w struct {
		name string
		obj  fyne.Focusable
	}) bool {
		if w.obj == nil {
			return fmt.Sprintf("%T", c.Focused()) == w.name
		}
		return c.Focused() == w.obj
	}
	for _, w := range want {
		c.FocusNext()
		if !same(w) {
			t.Fatalf("Tab should reach %s, got %T", w.name, c.Focused())
		}
	}
	for i := len(want) - 2; i >= 0; i-- {
		c.FocusPrevious()
		if !same(want[i]) {
			t.Fatalf("Shift+Tab should reach %s, got %T", want[i].name, c.Focused())
		}
	}
	c.FocusPrevious()
	if c.Focused() != fyne.Focusable(h.m.search) {
		t.Errorf("Shift+Tab from the list reached %T, want search", c.Focused())
	}
}
