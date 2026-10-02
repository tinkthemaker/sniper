package ui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"

	"sniper/internal/snippets"
	"sniper/internal/store"
)

type fakeClipboard struct{ copied []string }

func (f *fakeClipboard) SetText(s string) { f.copied = append(f.copied, s) }

func (f *fakeClipboard) last() string {
	if len(f.copied) == 0 {
		return ""
	}
	return f.copied[len(f.copied)-1]
}

type harness struct {
	t      *testing.T
	app    fyne.App
	store  *store.Store
	cb     *fakeClipboard
	deps   Deps
	m      *MainWindow
	opened []string
}

var fixedNow = time.Date(2026, 10, 1, 15, 30, 0, 0, time.UTC)

func sampleLibrary() snippets.Library {
	lib := snippets.NewLibrary()
	for _, s := range []struct{ title, body string }{
		{"Greeting", "Hello,\nThanks for reaching out."},
		{"Sign-off", "Best regards,\nTalmage"},
		{"Address", "1 Main St\nSpringfield"},
		{"Zebra notes", "Mentions greeting in the body"},
	} {
		var err error
		if lib, _, err = snippets.Add(lib, s.title, s.body); err != nil {
			panic(err)
		}
	}
	return lib
}

// newHarness builds a main window over a real store in a temp dir. If lib has
// snippets they are saved first, so the file on disk matches the window.
func newHarness(t *testing.T, lib snippets.Library) *harness {
	t.Helper()
	h := &harness{t: t, app: test.NewTempApp(t), store: store.New(t.TempDir()), cb: &fakeClipboard{}}
	if len(lib.Snippets) > 0 {
		if err := h.store.Save(lib); err != nil {
			t.Fatal(err)
		}
	}
	h.deps = Deps{
		App: h.app, Store: h.store, Clipboard: h.cb, Now: func() time.Time { return fixedNow },
		OpenFolder: func(p string) error { h.opened = append(h.opened, p); return nil },
	}
	h.m = ShowMain(h.deps, lib)
	return h
}

func (h *harness) focused() fyne.Focusable { return h.m.win.Canvas().Focused() }

func (h *harness) typeText(s string) {
	h.t.Helper()
	f := h.focused()
	if f == nil {
		h.t.Fatalf("no focused widget to type %q into", s)
	}
	for _, r := range s {
		f.TypedRune(r)
	}
}

func (h *harness) key(name fyne.KeyName) {
	h.t.Helper()
	f := h.focused()
	if f == nil {
		h.t.Fatalf("no focused widget for key %s", name)
	}
	f.TypedKey(&fyne.KeyEvent{Name: name})
}

func (h *harness) shortcut(sc *desktop.CustomShortcut) {
	h.t.Helper()
	f, ok := h.focused().(fyne.Shortcutable)
	if !ok {
		h.t.Fatalf("focused widget %T is not Shortcutable", h.focused())
	}
	f.TypedShortcut(sc)
}

func (h *harness) titles() []string {
	var out []string
	for _, s := range h.m.Results() {
		out = append(out, s.Title)
	}
	return out
}

func (h *harness) selectedTitle() string {
	for _, s := range h.m.Results() {
		if s.ID == h.m.SelectedID() {
			return s.Title
		}
	}
	return ""
}

func equalStrings(a, b []string) bool { return strings.Join(a, "\x00") == strings.Join(b, "\x00") }
