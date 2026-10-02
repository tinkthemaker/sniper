package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

const pageStep = 10

func ctrl(key fyne.KeyName) *desktop.CustomShortcut {
	return &desktop.CustomShortcut{KeyName: key, Modifier: fyne.KeyModifierShortcutDefault}
}

var shortcutFind = ctrl(fyne.KeyF)

type shortcutAction struct {
	shortcut *desktop.CustomShortcut
	run      func()
}

func (m *MainWindow) shortcutActions() []shortcutAction {
	return []shortcutAction{
		{shortcutFind, m.focusSearch},
	}
}

func (m *MainWindow) registerShortcuts() {
	for _, a := range m.shortcutActions() {
		m.win.Canvas().AddShortcut(a.shortcut, func(s fyne.Shortcut) { m.handleShortcut(s) })
	}
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
