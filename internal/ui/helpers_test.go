package ui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"

	"github.com/tinkthemaker/sniper/internal/snippets"
	"github.com/tinkthemaker/sniper/internal/store"
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
		{"Sign-off", "Best regards,\nTink"},
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
	return newHarnessWrapping(t, lib, nil)
}

// newHarnessWrapping is newHarness with the store wrapped, to inject failures.
func newHarnessWrapping(t *testing.T, lib snippets.Library, wrap func(*store.Store) Storage) *harness {
	t.Helper()
	h := &harness{t: t, app: test.NewTempApp(t), store: store.New(t.TempDir()), cb: &fakeClipboard{}}
	h.app.Settings().SetTheme(NewTheme())
	if len(lib.Snippets) > 0 {
		if err := h.store.Save(lib); err != nil {
			t.Fatal(err)
		}
	}
	var st Storage = h.store
	if wrap != nil {
		st = wrap(h.store)
	}
	h.deps = Deps{
		App: h.app, Store: st, Clipboard: h.cb, Now: func() time.Time { return fixedNow },
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

// failingStorage fails Save while fail is set, leaving the real store untouched.
type failingStorage struct {
	*store.Store
	fail bool
}

func (f *failingStorage) Save(lib snippets.Library) error {
	if f.fail {
		return errors.New("disk full")
	}
	return f.Store.Save(lib)
}

// topDialog returns the newest open dialog, or nil.
func (h *harness) topDialog() *shownDialog {
	if n := len(h.m.dialogs); n > 0 {
		return h.m.dialogs[n-1]
	}
	return nil
}

func (h *harness) confirmTop() {
	h.t.Helper()
	d, ok := h.topDialog().dlg.(*dialog.ConfirmDialog)
	if !ok {
		h.t.Fatalf("top dialog is %T, not a confirm dialog", h.topDialog().dlg)
	}
	d.Confirm()
}

func (h *harness) dismissTop() {
	h.t.Helper()
	h.topDialog().dlg.(interface{ Dismiss() }).Dismiss()
}

// diskTitles reloads the library from disk, as a restart would.
func (h *harness) diskTitles() []string {
	h.t.Helper()
	res := store.New(h.store.Dir()).Load()
	if res.State != store.StateOK {
		h.t.Fatalf("reload state %v: %v", res.State, res.Err)
	}
	var out []string
	for _, s := range res.Library.Snippets {
		out = append(out, s.Title)
	}
	return out
}

func (h *harness) primaryBytes() string {
	h.t.Helper()
	b, err := os.ReadFile(filepath.Join(h.store.Dir(), store.FileName))
	if err != nil {
		return ""
	}
	return string(b)
}

// up is how the main window displays a title: in capitals, and only for display.
func up(s string) string { return strings.ToUpper(s) }
