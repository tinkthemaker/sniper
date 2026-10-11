package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/tinkthemaker/sniper/internal/snippets"
)

// walk visits o and everything inside it, containers and widget renderers alike.
func walk(o fyne.CanvasObject, fn func(fyne.CanvasObject)) {
	if o == nil || !o.Visible() { // what is hidden cannot be read or pressed
		return
	}
	fn(o)
	switch v := o.(type) {
	case *fyne.Container:
		for _, c := range v.Objects {
			walk(c, fn)
		}
	case fyne.Widget:
		for _, c := range test.WidgetRenderer(v).Objects() {
			walk(c, fn)
		}
	}
}

// uiOf is what a user can read and press inside one part of a window.
type uiOf struct {
	texts   []string
	buttons []*widget.Button
	entries []*widget.Entry
}

func inspect(root fyne.CanvasObject) uiOf {
	var u uiOf
	walk(root, func(o fyne.CanvasObject) {
		switch v := o.(type) {
		case *widget.Label:
			if v.Visible() && v.Text != "" {
				u.texts = append(u.texts, v.Text)
			}
		case *widget.RichText:
			if v.Visible() && v.String() != "" {
				u.texts = append(u.texts, v.String())
			}
		case *widget.Button:
			u.buttons = append(u.buttons, v)
		case *widget.Entry:
			u.entries = append(u.entries, v)
		}
	})
	return u
}

func (u uiOf) button(label string) *widget.Button {
	for _, b := range u.buttons {
		if b.Text == label {
			return b
		}
	}
	return nil
}

func (u uiOf) buttonLabels() []string {
	var out []string
	for _, b := range u.buttons {
		out = append(out, b.Text)
	}
	return out
}

func (u uiOf) hasText(s string) bool {
	for _, t := range u.texts {
		if t == s {
			return true
		}
	}
	return false
}

// overlay inspects the dialog currently on top of the window.
func (h *harness) overlay() uiOf {
	h.t.Helper()
	top := h.m.win.Canvas().Overlays().Top()
	if top == nil {
		h.t.Fatal("no dialog is open")
	}
	return inspect(top)
}

// canvasShortcut delivers a shortcut the way the real driver does: to the
// canvas, which runs the handlers registered with AddShortcut. This is the
// path a key press takes when a list or a button has focus.
func (h *harness) canvasShortcut(s fyne.Shortcut) {
	h.t.Helper()
	// The test canvas hides its shortcut handler behind an embedded interface,
	// so reach the promoted method by reflection.
	c := reflect.ValueOf(h.m.win.Canvas()).Elem().FieldByName("WindowlessCanvas").Elem()
	m := c.MethodByName("TypedShortcut")
	if !m.IsValid() {
		h.t.Fatalf("canvas %s has no TypedShortcut", c.Type())
	}
	m.Call([]reflect.Value{reflect.ValueOf(s)})
}

func (h *harness) focus(o fyne.Focusable) { h.m.win.Canvas().Focus(o) }

// fixtureBytes reads a file from testdata/fixtures.
func fixtureBytes(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "fixtures", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func fixtureReader(t *testing.T, name string) *strings.Reader {
	return strings.NewReader(string(fixtureBytes(t, name)))
}

// fixtureFirstTitle returns the first record's raw title, without validation.
func fixtureFirstTitle(t *testing.T, name string) string {
	t.Helper()
	var f struct {
		Snippets []snippets.Snippet `json:"snippets"`
	}
	if err := json.Unmarshal(fixtureBytes(t, name), &f); err != nil || len(f.Snippets) == 0 {
		t.Fatalf("fixture %s: %v", name, err)
	}
	return f.Snippets[0].Title
}

// quitApp counts Quit calls, because the test app's Quit does nothing.
type quitApp struct {
	fyne.App
	quits int
}

func (q *quitApp) Quit() { q.quits++ }
