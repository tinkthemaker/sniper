package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// resultList moves the selection itself on Up, Down, PageUp, PageDown and Enter;
// Fyne's own list keys only move a focus highlight.
type resultList struct {
	widget.List
	m *MainWindow
}

func newResultList(m *MainWindow) *resultList {
	l := &resultList{m: m}
	l.Length = func() int { return len(m.results) }
	l.CreateItem = m.newRow
	l.UpdateItem = m.updateRow
	l.OnSelected = m.onListSelected
	l.ExtendBaseWidget(l)
	return l
}

func (l *resultList) TypedKey(k *fyne.KeyEvent) {
	if l.m.handleNavKey(k.Name) {
		return
	}
	l.List.TypedKey(k)
}
