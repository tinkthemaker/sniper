package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/tinkthemaker/sniper/internal/snippets"
	"github.com/tinkthemaker/sniper/internal/store"
)

// recoveryWindow offers ways out of an unreadable library. It never changes
// anything on disk until the user picks an action.
type recoveryWindow struct {
	d    Deps
	win  fyne.Window
	path string

	restoreBtn *widget.Button
	moveBtn    *widget.Button

	main   *MainWindow   // set once the library has been opened
	dialog dialog.Dialog // the confirmation currently shown, if any
	text   string        // that confirmation's message
}

// ShowRecovery shows the recovery window; on success it opens the main window and closes itself.
func ShowRecovery(d Deps, res store.LoadResult) { showRecovery(d, res) }

func showRecovery(d Deps, res store.LoadResult) *recoveryWindow {
	r := &recoveryWindow{d: d, path: filepath.Join(d.Store.Dir(), store.FileName)}
	r.win = d.App.NewWindow(recoveryWindowTitle)

	why := widget.NewLabel(res.Err.Error())
	why.Selectable = true
	why.Wrapping = fyne.TextWrapWord
	where := widget.NewLabel(r.path)
	where.Selectable = true
	where.Wrapping = fyne.TextWrapWord

	r.restoreBtn = widget.NewButton(labelRestoreBackup, r.restore)
	r.restoreBtn.Importance = widget.HighImportance
	r.moveBtn = widget.NewButton(labelMoveAside, r.moveAside)
	caption := widget.NewLabel(noBackupCaption)
	caption.Importance = widget.LowImportance
	if res.BackupAvailable {
		caption.Hide()
	} else {
		r.restoreBtn.Disable()
	}
	open := widget.NewButton(labelOpenFolder, r.openFolder)
	quit := widget.NewButton(labelQuit, d.App.Quit)

	open.Importance, quit.Importance, r.moveBtn.Importance = widget.MediumImportance, widget.MediumImportance, widget.MediumImportance

	// Same grid as the main window (chrome.go). "Your data has not been
	// changed" sits right under the error it qualifies, and the actions form
	// one group at the bottom, 8 px apart.
	intro := widget.NewLabelWithStyle(recoveryIntro, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	untouched := widget.NewLabel(recoveryUntouched)
	text := container.NewVScroll(container.New(layout.NewCustomPaddedVBoxLayout(0), intro, where, why, untouched))
	pair := container.New(&equalColumns{gap: gapItem}, open, quit)
	buttons := container.New(layout.NewCustomPaddedVBoxLayout(gapItem), r.restoreBtn, caption, r.moveBtn, pair)
	content := container.NewBorder(nil, inset(buttons, gapArea, 0, 0, 0), nil, nil, text)
	r.win.SetPadded(false)
	r.win.SetContent(inset(content, boxLine, boxLine, boxLine, boxLine))
	r.win.Resize(fyne.NewSize(520, 320))
	r.win.Show()
	return r
}

func (r *recoveryWindow) restore() {
	lib, err := r.d.Store.RestoreBackup()
	if err != nil {
		r.showError(restoreFailedText(err))
		return
	}
	r.open(lib, restoredStatus)
}

func (r *recoveryWindow) moveAside() {
	_, statErr := os.Stat(r.path)
	missing := errors.Is(statErr, os.ErrNotExist)
	msg := startEmptyConfirm
	if !missing {
		name := fmt.Sprintf("snippets.damaged-%s.json", r.d.Now().Format("20060102-150405"))
		msg = moveAsideConfirmText(name)
	}
	d := dialog.NewConfirm(moveAsideTitle, msg, func(ok bool) {
		r.dialog = nil
		if ok {
			r.confirmMoveAside()
		}
	}, r.win)
	d.SetConfirmText(labelMoveAside)
	d.SetDismissText(labelCancel)
	r.dialog, r.text = d, msg
	d.Show()
}

func (r *recoveryWindow) confirmMoveAside() {
	name, err := r.d.Store.QuarantinePrimary(r.d.Now())
	if err != nil {
		r.showError(moveAsideFailedText(err))
		return
	}
	status := startedEmptyStatus
	if name != "" {
		status = movedStatus(name)
	}
	r.open(snippets.NewLibrary(), status)
}

func (r *recoveryWindow) openFolder() {
	dir := r.d.Store.Dir()
	if err := r.d.OpenFolder(dir); err != nil {
		r.showError(openFolderFailedText(dir, err))
	}
}

// open shows the main window first, then closes this one, so the app never has
// zero windows in between.
func (r *recoveryWindow) open(lib snippets.Library, status string) {
	r.main = ShowMain(r.d, lib)
	r.main.setStatus(status)
	r.win.Close()
}

func (r *recoveryWindow) showError(text string) {
	d := dialog.NewError(errors.New(text), r.win)
	r.dialog = d
	d.SetOnClosed(func() { r.dialog = nil })
	d.Show()
}
