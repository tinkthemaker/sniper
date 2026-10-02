package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// buildMenu creates the File and Help menus. Fyne adds Quit to File itself.
// Menu actions do nothing while a dialog is open, like the keyboard shortcuts.
func (m *MainWindow) buildMenu() *fyne.MainMenu {
	item := func(label string, action func()) *fyne.MenuItem {
		return fyne.NewMenuItem(label, func() {
			if !m.modalOpen() {
				action()
			}
		})
	}
	file := fyne.NewMenu(menuFile,
		item(menuImport, m.importFile),
		item(menuExport, m.exportFile),
		item(menuOpenFolder, m.openDataFolder),
	)
	help := fyne.NewMenu(menuHelp,
		item(menuShortcuts, m.showShortcuts),
		item(menuAbout, m.showAbout),
	)
	return fyne.NewMainMenu(file, help)
}

func (m *MainWindow) showShortcuts() {
	m.present(dialog.NewInformation(menuShortcuts, shortcutsHelpText, m.win), menuShortcuts, shortcutsHelpText)
}

func (m *MainWindow) showAbout() {
	version := m.deps.App.Metadata().Version
	if version == "" {
		version = "dev"
	}
	text := aboutText(version, m.deps.Store.Dir())
	m.present(dialog.NewInformation(aboutTitle, text, m.win), aboutTitle, text)
}
