package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// searchEntry forwards app shortcuts and navigation keys to the main window,
// because a focused widget.Entry would otherwise swallow them.
type searchEntry struct {
	widget.Entry
	m *MainWindow
}

func newSearchEntry(m *MainWindow) *searchEntry {
	e := &searchEntry{m: m}
	e.ExtendBaseWidget(e)
	e.SetPlaceHolder(searchPlaceholder)
	return e
}

func (e *searchEntry) TypedShortcut(s fyne.Shortcut) {
	if e.m.handleShortcut(s) {
		return
	}
	e.Entry.TypedShortcut(s)
}

func (e *searchEntry) TypedKey(k *fyne.KeyEvent) {
	if k.Name == fyne.KeyEscape {
		e.m.clearQuery()
		return
	}
	if e.m.handleNavKey(k.Name) {
		return
	}
	e.Entry.TypedKey(k)
}
