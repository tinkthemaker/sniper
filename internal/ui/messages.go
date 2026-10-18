package ui

import (
	"fmt"
	"strings"
)

// Every user-facing string lives here so tests can assert on it. “…” are U+201C and U+201D.
const (
	appTitle          = "Sniper"
	searchPlaceholder = "Search snippets (Ctrl+F)"

	labelDelete = "Delete"
	labelOK     = "OK"
	labelQuit   = "Quit"

	labelTitle       = "Title"
	labelBody        = "Body"
	labelCancel      = "Cancel"
	labelSave        = "Save"
	labelDiscard     = "Discard"
	labelKeepEditing = "Keep editing"

	editorTitleNew  = "New snippet"
	editorTitleEdit = "Edit snippet"
	discardTitle    = "Discard changes?"
	discardMessage  = "Your changes to this snippet will be lost."
	deleteTitle     = "Delete snippet?"

	menuFile       = "File"
	menuImport     = "Import…"
	menuExport     = "Export…"
	menuOpenFolder = "Open data folder"
	menuEdit       = "Edit"
	menuUndo       = "Undo"
	menuHelp       = "Help"
	menuShortcuts  = "Keyboard shortcuts"
	menuAbout      = "About"
	aboutTitle     = "About Sniper"
	aboutPrivacy   = "Snippets and exports are stored as plain text on this computer. Sniper makes no network connections."

	importTitle       = "Import snippets?"
	labelImport       = "Import"
	importEmptyText   = "The file contains no snippets."
	exportFileExt     = ".json"
	exportNamePrefix  = "sniper-export-"
	exportDateLayout  = "2006-01-02"
	importFilterExt   = ".json"
	shortcutsHelpText = `Search box
  Type — filter snippets
  Down / Up — move the selection
  PageDown / PageUp — move the selection by 10
  Enter — copy the selected snippet
  Escape — clear the search

Anywhere in the main window
  Ctrl+F — focus the search box
  Ctrl+N — new snippet
  Ctrl+E — edit the selected snippet
  Ctrl+Shift+D — delete the selected snippet
  Ctrl+Z — undo the last change
  Tab / Shift+Tab — move between controls

Mouse
  Click a snippet — select it
  Double-click a snippet — copy it

Editor
  Ctrl+S — save
  Escape — cancel`

	recoveryWindowTitle = "Sniper – Recover library"
	recoveryIntro       = "Sniper could not open your snippet library."
	recoveryUntouched   = "Your data has not been changed."
	labelRestoreBackup  = "Restore backup"
	noBackupCaption     = "No valid backup found."
	labelMoveAside      = "Move damaged file aside and start empty"
	labelOpenFolder     = "Open data folder"
	moveAsideTitle      = "Move damaged file aside?"
	startEmptyConfirm   = "snippets.json is missing, so there is nothing to move. Start with an empty library?"
	restoredStatus      = "Restored from backup."
	startedEmptyStatus  = "Started with an empty library."

	brandName        = "SNIPER"
	previewLabel     = "PREVIEW"
	searchBlockLabel = "SEARCH"
	findLabel        = "FIND"
	hintNewText      = "CTRL+N NEW"
	hintUndoText     = "CTRL+Z UNDO"

	// The main window's buttons are capitals, as on the frame around them.
	btnCopy   = "COPY"
	btnNew    = "NEW"
	btnEdit   = "EDIT"
	btnDelete = "DELETE"

	emptyLibraryText    = "No snippets yet. Press Ctrl+N to create one."
	previewEmptyText    = "Select a snippet to preview it."
	nothingToUndoStatus = "Nothing to undo."
	alreadyRunningText  = "Sniper is already running. Switch to it from the taskbar."
)

// unpunctuated returns err's text without a trailing period, so templates that
// follow it with ". …" do not produce "..".
func unpunctuated(err error) string { return strings.TrimSuffix(err.Error(), ".") }

func previewRecordLabel(code string) string { return previewLabel + " · RECORD " + code }

// versionText is the version block: V followed by the version in capitals, or
// VDEV when the build has none.
func versionText(version string) string {
	if version == "" {
		version = "dev"
	}
	return "V" + strings.ToUpper(version)
}

func countAllText(n int) string          { return fmt.Sprintf("%d IN LIBRARY", n) }
func countMatchText(n, total int) string { return fmt.Sprintf("%d OF %d", n, total) }

func noMatchesText(query string) string { return fmt.Sprintf("No snippets match “%s”.", query) }
func copiedStatus(title string) string  { return fmt.Sprintf("Copied “%s”.", title) }
func savedStatus(title string) string   { return fmt.Sprintf("Saved “%s”.", title) }
func deletedStatus(title string) string { return fmt.Sprintf("Deleted “%s”.", title) }

func saveFailedText(err error) string {
	return fmt.Sprintf("Could not save: %s. Your changes are still here; try again.", unpunctuated(err))
}
func deleteFailedText(err error) string { return fmt.Sprintf("Could not delete: %v", err) }

func importFailedText(err error) string {
	return fmt.Sprintf("Import failed: %s. Your library was not changed.", unpunctuated(err))
}

// snippetCount is "1 snippet" or "N snippets".
func snippetCount(n int) string {
	if n == 1 {
		return "1 snippet"
	}
	return fmt.Sprintf("%d snippets", n)
}

func importConfirmText(n int, name string) string {
	return fmt.Sprintf("Import %s from %s? Importing the same file twice creates duplicates.", snippetCount(n), name)
}
func importedStatus(n int) string { return fmt.Sprintf("Imported %s.", snippetCount(n)) }
func importSaveFailedText(err error) string {
	return fmt.Sprintf("Could not save the import: %s. Your library was not changed.", unpunctuated(err))
}
func exportedStatus(n int, name string) string {
	return fmt.Sprintf("Exported %s to %s.", snippetCount(n), name)
}
func exportFailedText(err error) string { return fmt.Sprintf("Export failed: %v", err) }
func openFolderFailedText(path string, err error) string {
	return fmt.Sprintf("Could not open %s: %v", path, err)
}
func aboutText(version, dir string) string {
	return fmt.Sprintf("Sniper %s\n\nData folder: %s\n\n%s", version, dir, aboutPrivacy)
}
func moveAsideConfirmText(name string) string {
	return fmt.Sprintf("Move snippets.json to %s and start with an empty library?", name)
}
func restoreFailedText(err error) string { return fmt.Sprintf("Could not restore the backup: %v", err) }
func moveAsideFailedText(err error) string {
	return fmt.Sprintf("Could not move the damaged file: %v", err)
}
func movedStatus(name string) string        { return fmt.Sprintf("Moved damaged file to %s.", name) }
func undidCreateStatus(title string) string { return fmt.Sprintf("Undid new snippet “%s”.", title) }
func undidEditStatus(title string) string   { return fmt.Sprintf("Undid edit of “%s”.", title) }
func undidDeleteStatus(title string) string { return fmt.Sprintf("Undid delete of “%s”.", title) }
func undidImportStatus(n int) string        { return fmt.Sprintf("Undid import of %s.", snippetCount(n)) }
func undoFailedText(err error) string       { return fmt.Sprintf("Could not undo: %v", err) }

func deleteConfirmText(title string) string {
	return fmt.Sprintf("Delete “%s”? This cannot be undone.", title)
}
