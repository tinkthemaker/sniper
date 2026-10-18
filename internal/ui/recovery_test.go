package ui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/test"

	"github.com/tinkthemaker/sniper/internal/snippets"
	"github.com/tinkthemaker/sniper/internal/store"
)

type recoveryHarness struct {
	*harness
	r   *recoveryWindow
	dir string
}

// newRecoveryHarness lays out a data directory (nil means the file is absent),
// loads it as the app would, and shows the recovery window.
func newRecoveryHarness(t *testing.T, primary, backup *string) *recoveryHarness {
	t.Helper()
	dir := t.TempDir()
	put := func(name string, content *string) {
		if content != nil {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(*content), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	put(store.FileName, primary)
	put(store.BackupName, backup)

	h := &harness{t: t, app: test.NewTempApp(t), store: store.New(dir), cb: &fakeClipboard{}}
	h.app.Settings().SetTheme(NewTheme())
	h.deps = Deps{
		App: h.app, Store: h.store, Clipboard: h.cb, Now: func() time.Time { return fixedNow },
		OpenFolder: func(p string) error { h.opened = append(h.opened, p); return nil },
		Motion:     harnessMotion,
	}
	res := h.store.Load()
	if res.State != store.StateDamaged {
		t.Fatalf("setup should be damaged, got %v", res.State)
	}
	rh := &recoveryHarness{harness: h, dir: dir}
	rh.r = showRecovery(h.deps, res)
	return rh
}

func (rh *recoveryHarness) read(name string) (string, bool) {
	b, err := os.ReadFile(filepath.Join(rh.dir, name))
	return string(b), err == nil
}

func (rh *recoveryHarness) confirm() {
	rh.t.Helper()
	d, ok := rh.r.dialog.(*dialog.ConfirmDialog)
	if !ok {
		rh.t.Fatalf("no confirmation is showing: %T", rh.r.dialog)
	}
	d.Confirm()
}

func (rh *recoveryHarness) windowOpen(w fyne.Window) bool {
	for _, x := range rh.app.Driver().AllWindows() {
		if x == w {
			return true
		}
	}
	return false
}

func goodBackup(t *testing.T) string {
	t.Helper()
	lib, _, err := snippets.Add(snippets.NewLibrary(), "From backup", "backup body")
	if err != nil {
		t.Fatal(err)
	}
	b, err := store.Encode(lib)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRecoveryChangesNothingUntilAnActionIsChosen(t *testing.T) {
	bak := goodBackup(t)
	rh := newRecoveryHarness(t, ptr("xx"), &bak)
	if got, _ := rh.read(store.FileName); got != "xx" {
		t.Errorf("primary = %q", got)
	}
	if got, _ := rh.read(store.BackupName); got != bak {
		t.Error("backup changed")
	}
	if rh.r.restoreBtn.Disabled() {
		t.Error("Restore backup should be enabled when a valid backup exists")
	}
	if rh.r.main != nil {
		t.Error("main window opened by itself")
	}
}

func ptr(s string) *string { return &s }

func TestRecoveryRestoreOpensMainWindowWithBackup(t *testing.T) {
	bak := goodBackup(t)
	rh := newRecoveryHarness(t, ptr("xx"), &bak)
	test.Tap(rh.r.restoreBtn)

	m := rh.r.main
	if m == nil {
		t.Fatal("main window not opened")
	}
	if len(m.Library().Snippets) != 1 || m.Library().Snippets[0].Title != "From backup" {
		t.Errorf("library = %+v", m.Library().Snippets)
	}
	if m.Status() != "Restored from backup." {
		t.Errorf("status = %q", m.Status())
	}
	if got, _ := rh.read(store.FileName); got != bak {
		t.Error("primary should now equal the backup")
	}
	if rh.windowOpen(rh.r.win) || !rh.windowOpen(m.Window()) {
		t.Error("recovery window should be closed and the main window open")
	}
	if res := store.New(rh.dir).Load(); res.State != store.StateOK {
		t.Errorf("reload after restore: %+v", res)
	}
}

func TestRecoveryRestoreDisabledWithoutBackup(t *testing.T) {
	for name, bak := range map[string]*string{"no backup": nil, "corrupt backup": ptr("nope")} {
		t.Run(name, func(t *testing.T) {
			rh := newRecoveryHarness(t, ptr("xx"), bak)
			if !rh.r.restoreBtn.Disabled() {
				t.Error("Restore backup must be disabled")
			}
		})
	}
}

func TestRecoveryRestoreFailureKeepsWindowOpen(t *testing.T) {
	bak := goodBackup(t)
	rh := newRecoveryHarness(t, ptr("xx"), &bak)
	if err := os.WriteFile(filepath.Join(rh.dir, store.BackupName), []byte("now corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	test.Tap(rh.r.restoreBtn)
	if rh.r.main != nil || !rh.windowOpen(rh.r.win) {
		t.Error("recovery window must stay open after a failed restore")
	}
	if rh.r.dialog == nil {
		t.Error("expected an error dialog")
	}
	if got, _ := rh.read(store.FileName); got != "xx" {
		t.Error("primary changed by a failed restore")
	}
}

func TestRecoveryMoveAsideAfterConfirmation(t *testing.T) {
	bak := goodBackup(t)
	rh := newRecoveryHarness(t, ptr("xx"), &bak)
	test.Tap(rh.r.moveBtn)
	if rh.r.dialog == nil {
		t.Fatal("expected a confirmation first")
	}
	if got, _ := rh.read(store.FileName); got != "xx" {
		t.Fatal("file moved before the user confirmed")
	}
	rh.confirm()

	name := "snippets.damaged-20261001-153000.json"
	if got, ok := rh.read(name); !ok || got != "xx" {
		t.Errorf("damaged file = %q (exists=%v)", got, ok)
	}
	if _, ok := rh.read(store.FileName); ok {
		t.Error("primary should have been moved aside")
	}
	m := rh.r.main
	if m == nil || len(m.Library().Snippets) != 0 {
		t.Fatalf("expected an empty main window, got %+v", m)
	}
	if m.Status() != "Moved damaged file to "+name+"." {
		t.Errorf("status = %q", m.Status())
	}
	if rh.windowOpen(rh.r.win) {
		t.Error("recovery window should close")
	}

	// The first save in the new empty library writes a fresh primary and keeps the damaged copy.
	h := &harness{t: t, m: m}
	h.app = rh.app
	m.win.Canvas().Focus(m.search)
	m.openEditor(nil)
	m.editor.title.SetText("Fresh")
	m.editor.body.SetText("start")
	m.editor.save()
	if res := store.New(rh.dir).Load(); res.State != store.StateOK || len(res.Library.Snippets) != 1 {
		t.Errorf("after first save: %+v", res)
	}
	if got, ok := rh.read(name); !ok || got != "xx" {
		t.Error("damaged copy must survive later saves")
	}
}

func TestRecoveryMoveAsideConfirmText(t *testing.T) {
	rh := newRecoveryHarness(t, ptr("xx"), nil)
	test.Tap(rh.r.moveBtn)
	want := "Move snippets.json to snippets.damaged-20261001-153000.json and start with an empty library?"
	if rh.r.text != want {
		t.Errorf("text = %q, want %q", rh.r.text, want)
	}
}

func TestRecoveryMoveAsideCancelKeepsFile(t *testing.T) {
	rh := newRecoveryHarness(t, ptr("xx"), nil)
	test.Tap(rh.r.moveBtn)
	rh.r.dialog.Dismiss()
	if got, _ := rh.read(store.FileName); got != "xx" || rh.r.main != nil {
		t.Error("cancel must leave everything as it was")
	}
}

func TestRecoveryMoveAsideWhenPrimaryIsMissing(t *testing.T) {
	bak := goodBackup(t)
	rh := newRecoveryHarness(t, nil, &bak) // missing primary, valid backup
	if rh.r.restoreBtn.Disabled() {
		t.Error("backup should be restorable when the primary is missing")
	}
	test.Tap(rh.r.moveBtn)
	rh.confirm()
	if rh.r.main == nil || rh.r.main.Status() != "Started with an empty library." {
		t.Fatalf("main = %+v", rh.r.main)
	}
	if got, _ := rh.read(store.BackupName); got != bak {
		t.Error("backup must not be touched")
	}
}

func TestRecoveryMoveAsideFailureKeepsWindowOpen(t *testing.T) {
	rh := newRecoveryHarness(t, ptr("xx"), nil)
	rh.r.d.Store = renameFailStorage{rh.store}
	test.Tap(rh.r.moveBtn)
	rh.confirm()
	if rh.r.main != nil || !rh.windowOpen(rh.r.win) {
		t.Error("window must stay open when the move fails")
	}
	if got, _ := rh.read(store.FileName); got != "xx" {
		t.Error("primary changed")
	}
}

type renameFailStorage struct{ *store.Store }

func (renameFailStorage) QuarantinePrimary(time.Time) (string, error) {
	return "", errors.New("access denied")
}

func TestRecoveryOpenFolder(t *testing.T) {
	rh := newRecoveryHarness(t, ptr("xx"), nil)
	rh.r.openFolder()
	if len(rh.opened) != 1 || rh.opened[0] != rh.dir {
		t.Errorf("opened = %v", rh.opened)
	}
}
