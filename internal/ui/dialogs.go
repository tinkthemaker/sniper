package ui

import (
	"errors"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"sniper/internal/snippets"
)

// shownDialog records an open dialog so the modal guard and tests can see it.
type shownDialog struct {
	title, message string
	dlg            presentable
}

// presentable is what present needs; fyne's file dialogs satisfy it but not dialog.Dialog.
type presentable interface {
	Show()
	SetOnClosed(closed func())
}

// present shows d over the main window and tracks it until it closes.
func (m *MainWindow) present(d presentable, title, message string) {
	sd := &shownDialog{title: title, message: message, dlg: d}
	m.dialogs = append(m.dialogs, sd)
	d.SetOnClosed(func() { m.forget(sd) })
	d.Show()
}

func (m *MainWindow) forget(sd *shownDialog) {
	for i, x := range m.dialogs {
		if x == sd {
			m.dialogs = append(m.dialogs[:i:i], m.dialogs[i+1:]...)
			return
		}
	}
}

// modalOpen reports whether any dialog is open over the main window. While one
// is, main-window shortcuts must do nothing.
func (m *MainWindow) modalOpen() bool { return len(m.dialogs) > 0 }

func (m *MainWindow) showError(text string) {
	m.present(dialog.NewError(errors.New(text), m.win), "Error", text)
}

func (m *MainWindow) deleteSelected() {
	s, ok := m.selected()
	if !ok {
		return
	}
	msg := deleteConfirmText(s.Title)
	d := dialog.NewConfirm(deleteTitle, msg, func(confirmed bool) {
		if confirmed {
			m.deleteSnippet(s)
		}
		m.win.Canvas().Focus(m.search)
	}, m.win)
	d.SetConfirmText(labelDelete)
	d.SetDismissText(labelCancel)
	d.SetConfirmImportance(widget.DangerImportance)
	m.present(d, deleteTitle, msg)
}

// deleteSnippet removes s and saves. On a failed save the library is unchanged.
func (m *MainWindow) deleteSnippet(s snippets.Snippet) {
	at := m.indexOf(s.ID)
	next, err := snippets.Delete(m.lib, s.ID)
	if err == nil {
		err = m.deps.Store.Save(next)
	}
	if err != nil {
		m.showError(deleteFailedText(err))
		return
	}
	m.lib = next

	// The result now at the deleted index, else the one before it, else none.
	remaining := snippets.Search(m.lib.Snippets, m.query)
	pick := ""
	switch {
	case at >= 0 && at < len(remaining):
		pick = remaining[at].ID
	case len(remaining) > 0:
		pick = remaining[len(remaining)-1].ID
	}
	m.refresh(pick)
	m.setStatus(deletedStatus(s.Title))
}
