package ui

import (
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"github.com/tinkthemaker/sniper/internal/snippets"
	"github.com/tinkthemaker/sniper/internal/store"
)

// fileDialogSize leaves a 40 px margin around the window; Fyne's default is cramped.
func (m *MainWindow) fileDialogSize() fyne.Size {
	s := m.win.Canvas().Size()
	return fyne.NewSize(max(s.Width-40, 420), max(s.Height-40, 380))
}

func (m *MainWindow) importFile() {
	d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil {
			m.showError(importFailedText(err))
			return
		}
		if r == nil { // cancelled
			return
		}
		defer r.Close()
		m.importFrom(r, r.URI().Name())
	}, m.win)
	d.SetFilter(storage.NewExtensionFileFilter([]string{importFilterExt}))
	m.present(d, menuImport, "")
	d.Resize(m.fileDialogSize()) // only valid once shown
}

// importFrom decodes and validates everything before asking to import, so a bad
// file changes nothing.
func (m *MainWindow) importFrom(r io.Reader, name string) {
	items, err := store.DecodeImport(r)
	if err == nil && len(items) == 0 {
		m.present(dialog.NewInformation(menuImport, importEmptyText, m.win), menuImport, importEmptyText)
		return
	}
	if err == nil {
		_, err = snippets.Append(m.lib, items) // surfaces the library size limit early
	}
	if err != nil {
		m.showError(importFailedText(err))
		return
	}
	msg := importConfirmText(len(items), name)
	d := dialog.NewConfirm(importTitle, msg, func(ok bool) {
		if ok {
			m.applyImport(items)
		}
		m.win.Canvas().Focus(m.search)
	}, m.win)
	d.SetConfirmText(labelImport)
	d.SetDismissText(labelCancel)
	m.present(d, importTitle, msg)
}

func (m *MainWindow) applyImport(items []snippets.Snippet) {
	next, err := snippets.Append(m.lib, items)
	if err == nil {
		err = m.deps.Store.Save(next)
	}
	if err != nil {
		m.showError(importSaveFailedText(err))
		return
	}
	m.pushUndo(undoStep{kind: undoImport, prev: m.lib, n: len(items)})
	m.lib = next
	m.refresh(m.selectedID) // keeps the selection if it is still a result
	m.setStatus(importedStatus(len(items)))
}

func (m *MainWindow) exportFile() {
	d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil {
			m.showError(exportFailedText(err))
			return
		}
		if w == nil { // cancelled
			return
		}
		name := w.URI().Name()
		ok := m.exportTo(w, name)
		if err := w.Close(); err != nil && ok {
			m.showError(exportFailedText(err))
		}
	}, m.win)
	d.SetFileName(exportNamePrefix + m.deps.Now().Format(exportDateLayout) + exportFileExt)
	d.SetFilter(storage.NewExtensionFileFilter([]string{exportFileExt}))
	m.present(d, menuExport, "")
	d.Resize(m.fileDialogSize()) // only valid once shown
}

// exportTo writes the whole library to w and reports whether that succeeded.
func (m *MainWindow) exportTo(w io.Writer, name string) bool {
	if err := store.Export(w, m.lib); err != nil {
		m.showError(exportFailedText(err))
		return false
	}
	m.setStatus(exportedStatus(len(m.lib.Snippets), name))
	return true
}

func (m *MainWindow) openDataFolder() {
	dir := m.deps.Store.Dir()
	if err := m.deps.OpenFolder(dir); err != nil {
		m.showError(openFolderFailedText(dir, err))
	}
}
