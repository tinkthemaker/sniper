package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/tinkthemaker/sniper/internal/store"
)

func (rh *recoveryHarness) content() uiOf { return inspect(rh.r.win.Content()) }

// errorDialogText returns what the recovery window's current dialog says.
func (rh *recoveryHarness) dialogTexts() uiOf {
	rh.t.Helper()
	top := rh.r.win.Canvas().Overlays().Top()
	if top == nil {
		rh.t.Fatal("no dialog is open over the recovery window")
	}
	return inspect(top)
}

func TestRecoveryWindowShowsWhatHappenedAndWhere(t *testing.T) {
	bak := goodBackup(t)
	rh := newRecoveryHarness(t, ptr(string(fixtureBytes(t, "malformed.json"))), &bak)
	u := rh.content()
	if rh.r.win.Title() != "Sniper – Recover library" {
		t.Errorf("title = %q", rh.r.win.Title())
	}
	for _, want := range []string{
		"Sniper could not open your snippet library.",
		filepath.Join(rh.dir, "snippets.json"),
		"Your data has not been changed.",
	} {
		if !u.hasText(want) {
			t.Errorf("window does not show %q; texts: %q", want, u.texts)
		}
	}
	reason := rh.store.Load().Err.Error()
	if !u.hasText(reason) || !strings.Contains(reason, "invalid JSON") {
		t.Errorf("window does not show the error %q", reason)
	}
	if u.hasText("No valid backup found.") {
		t.Error("the no-backup caption must be hidden when a backup exists")
	}
	if got := u.buttonLabels(); !equalStrings(got, []string{"Restore backup", "Move damaged file aside and start empty", "Open data folder", "Quit"}) {
		t.Errorf("buttons = %v", got)
	}
	if u.button("Restore backup").Importance != widget.HighImportance {
		t.Error("Restore backup is the primary action")
	}
}

func TestRecoveryWithoutABackupSaysSo(t *testing.T) {
	rh := newRecoveryHarness(t, ptr(string(fixtureBytes(t, "wrong-version.json"))), nil)
	u := rh.content()
	if !u.hasText("No valid backup found.") || !u.button("Restore backup").Disabled() {
		t.Errorf("no-backup state: texts %q, disabled=%v", u.texts, u.button("Restore backup").Disabled())
	}
	if !u.hasText(rh.store.Load().Err.Error()) {
		t.Errorf("the version error is not shown: %q", u.texts)
	}
	test.Tap(u.button("Restore backup")) // a disabled button does nothing
	if rh.r.main != nil || rh.r.dialog != nil {
		t.Error("a disabled Restore backup acted")
	}
}

func TestRecoveryAcceptsEveryKindOfDamagedFile(t *testing.T) {
	for _, name := range []string{"malformed.json", "wrong-version.json", "invalid-import.json", "title-201.json"} {
		t.Run(name, func(t *testing.T) {
			bak := goodBackup(t)
			raw := string(fixtureBytes(t, name))
			rh := newRecoveryHarness(t, &raw, &bak) // newRecoveryHarness itself checks the file is damaged
			if !rh.content().hasText(rh.store.Load().Err.Error()) {
				t.Error("the reason is not shown")
			}
			if got, _ := rh.read(store.FileName); got != raw {
				t.Error("showing the window changed the file")
			}
			test.Tap(rh.r.restoreBtn)
			if rh.r.main == nil || rh.r.main.Library().Snippets[0].Title != "From backup" {
				t.Fatal("restore did not open the backup")
			}
		})
	}
}

func TestRecoveryButtonWording(t *testing.T) {
	rh := newRecoveryHarness(t, ptr("xx"), nil)
	test.Tap(rh.r.moveBtn)
	u := rh.dialogTexts()
	if got := u.buttonLabels(); !equalStrings(got, []string{"Cancel", "Move damaged file aside and start empty"}) {
		t.Errorf("confirmation buttons = %v", got)
	}
	if !u.hasText("Move snippets.json to snippets.damaged-20261001-153000.json and start with an empty library?") {
		t.Errorf("confirmation texts = %q", u.texts)
	}
	rh.r.dialog.Dismiss()

	// With no file there is nothing to move, and the text says so.
	bak := goodBackup(t)
	none := newRecoveryHarness(t, nil, &bak)
	test.Tap(none.r.moveBtn)
	if !none.dialogTexts().hasText("snippets.json is missing, so there is nothing to move. Start with an empty library?") {
		t.Errorf("missing-file confirmation = %q", none.dialogTexts().texts)
	}
}

func TestRecoveryErrorsAreShownAndTheWindowStays(t *testing.T) {
	bak := goodBackup(t)
	rh := newRecoveryHarness(t, ptr("xx"), &bak)
	if err := os.WriteFile(filepath.Join(rh.dir, store.BackupName), []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	test.Tap(rh.r.restoreBtn)
	var found string
	for _, s := range rh.dialogTexts().texts {
		if strings.HasPrefix(s, "Could not restore the backup: ") {
			found = s
		}
	}
	if found == "" {
		t.Errorf("restore error not shown: %q", rh.dialogTexts().texts)
	}
	rh.r.dialog.Dismiss()

	rh.r.d.Store = renameFailStorage{rh.store}
	test.Tap(rh.r.moveBtn)
	rh.confirm()
	if got := rh.dialogTexts().texts; !containsText(got, "Could not move the damaged file: access denied") {
		t.Errorf("move error not shown: %q", got)
	}
	if rh.r.main != nil || !rh.windowOpen(rh.r.win) {
		t.Error("the recovery window must stay open")
	}
}

func containsText(texts []string, want string) bool {
	for _, s := range texts {
		if s == want {
			return true
		}
	}
	return false
}

func TestRecoveryOpenFolderFailureShowsTheError(t *testing.T) {
	rh := newRecoveryHarness(t, ptr("xx"), nil)
	rh.r.d.OpenFolder = func(string) error { return os.ErrPermission }
	test.Tap(rh.content().button("Open data folder"))
	want := "Could not open " + rh.dir + ": " + os.ErrPermission.Error()
	if !containsText(rh.dialogTexts().texts, want) {
		t.Errorf("texts = %q, want %q", rh.dialogTexts().texts, want)
	}
}

func TestRecoveryQuitButtonQuits(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, store.FileName), []byte("xx"), 0o600); err != nil {
		t.Fatal(err)
	}
	st := store.New(dir)
	app := &quitApp{App: test.NewTempApp(t)}
	app.Settings().SetTheme(NewTheme())
	r := showRecovery(Deps{App: app, Store: st, Clipboard: &fakeClipboard{}, Now: time.Now,
		OpenFolder: func(string) error { return nil }}, st.Load())
	test.Tap(inspect(r.win.Content()).button("Quit"))
	if app.quits != 1 {
		t.Errorf("Quit pressed, App.Quit called %d times", app.quits)
	}
}

func TestMainWindowAfterRecoveryIsFullyUsable(t *testing.T) {
	bak := goodBackup(t)
	rh := newRecoveryHarness(t, ptr("xx"), &bak)
	test.Tap(rh.r.restoreBtn)
	h := &harness{t: t, app: rh.app, store: rh.store, cb: rh.cb, deps: rh.deps, m: rh.r.main}
	if h.focused() != fyne.Focusable(h.m.search) || h.selectedTitle() != "From backup" {
		t.Errorf("focus %T, selected %q", h.focused(), h.selectedTitle())
	}
	h.key(fyne.KeyReturn)
	if h.cb.last() != "backup body" {
		t.Errorf("clipboard %q", h.cb.last())
	}
	createViaKeyboard(h, "After", "restore")
	if got := h.diskTitles(); !equalStrings(got, []string{"From backup", "After"}) {
		t.Errorf("disk = %v", got)
	}
}

func TestAlreadyRunningWindow(t *testing.T) {
	app := &quitApp{App: test.NewTempApp(t)}
	ShowAlreadyRunning(app)
	ws := app.Driver().AllWindows()
	u := inspect(ws[len(ws)-1].Content())
	if !u.hasText("Sniper is already running. Switch to it from the taskbar.") {
		t.Errorf("texts = %q", u.texts)
	}
	if got := u.buttonLabels(); !equalStrings(got, []string{"OK"}) {
		t.Fatalf("buttons = %v", got)
	}
	test.Tap(u.button("OK"))
	if app.quits != 1 {
		t.Errorf("OK called App.Quit %d times", app.quits)
	}
}

func TestFatalWindow(t *testing.T) {
	app := &quitApp{App: test.NewTempApp(t)}
	ShowFatal(app, "Could not create the data folder", `C:\x\Sniper: access is denied`)
	ws := app.Driver().AllWindows()
	u := inspect(ws[len(ws)-1].Content())
	if !u.hasText("Could not create the data folder") || !u.hasText(`C:\x\Sniper: access is denied`) {
		t.Errorf("texts = %q", u.texts)
	}
	var detail *widget.Label
	walk(ws[len(ws)-1].Content(), func(o fyne.CanvasObject) {
		if l, ok := o.(*widget.Label); ok && l.Text == `C:\x\Sniper: access is denied` {
			detail = l
		}
	})
	if detail == nil || !detail.Selectable {
		t.Error("the detail text must be selectable")
	}
	test.Tap(u.button("Quit"))
	if app.quits != 1 {
		t.Errorf("Quit called App.Quit %d times", app.quits)
	}
}
