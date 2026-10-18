package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

const pageStep = 10

func ctrl(key fyne.KeyName) *desktop.CustomShortcut {
	return &desktop.CustomShortcut{KeyName: key, Modifier: fyne.KeyModifierShortcutDefault}
}

var (
	shortcutFind   = ctrl(fyne.KeyF)
	shortcutNew    = ctrl(fyne.KeyN)
	shortcutEdit   = ctrl(fyne.KeyE)
	shortcutSave   = ctrl(fyne.KeyS)
	shortcutDelete = &desktop.CustomShortcut{
		KeyName:  fyne.KeyD,
		Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierShift,
	}
)

type shortcutAction struct {
	shortcut fyne.Shortcut
	run      func()
}

func (m *MainWindow) shortcutActions() []shortcutAction {
	return []shortcutAction{
		{shortcutFind, m.focusSearch},
		{shortcutNew, m.newSnippet},
		{shortcutEdit, m.editSelected},
		{shortcutDelete, m.deleteSelected},
		{&fyne.ShortcutUndo{}, m.undoLast},
	}
}

func (m *MainWindow) registerShortcuts() {
	for _, a := range m.shortcutActions() {
		m.win.Canvas().AddShortcut(a.shortcut, func(s fyne.Shortcut) { m.handleShortcut(s) })
	}
	m.win.Canvas().AddShortcut(shortcutSave, func(fyne.Shortcut) { m.saveOpenEditor() })
}

// handleShortcut runs the main-window action bound to s and reports whether s
// was one of them. While a dialog is open the action is swallowed, not run.
func (m *MainWindow) handleShortcut(s fyne.Shortcut) bool {
	for _, a := range m.shortcutActions() {
		if a.shortcut.ShortcutName() != s.ShortcutName() {
			continue
		}
		if !m.modalOpen() {
			a.run()
		}
		return true
	}
	return false
}

// handleNavKey implements the keys shared by the search box and the result list.
func (m *MainWindow) handleNavKey(k fyne.KeyName) bool {
	switch k {
	case fyne.KeyDown:
		m.moveSelection(1)
	case fyne.KeyUp:
		m.moveSelection(-1)
	case fyne.KeyPageDown:
		m.moveSelection(pageStep)
	case fyne.KeyPageUp:
		m.moveSelection(-pageStep)
	case fyne.KeyReturn, fyne.KeyEnter:
		m.copySelected()
	default:
		return false
	}
	return true
}

// saveOpenEditor makes Ctrl+S work when focus rests on one of the editor's
// buttons, where no entry is there to route it (Tab can land there). It acts
// only while the editor itself is the top dialog, never over its discard
// confirmation.
func (m *MainWindow) saveOpenEditor() {
	n := len(m.dialogs)
	if m.editor != nil && n > 0 && m.dialogs[n-1].dlg == presentable(m.editor.dlg) {
		m.editor.save()
	}
}
