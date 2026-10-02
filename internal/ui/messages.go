package ui

import "fmt"

// Every user-facing string lives here so tests can assert on it. “…” are U+201C and U+201D.
const (
	appTitle          = "Sniper"
	searchPlaceholder = "Search snippets (Ctrl+F)"

	labelCopy   = "Copy"
	labelNew    = "New"
	labelEdit   = "Edit"
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

	emptyLibraryText   = "No snippets yet. Press Ctrl+N to create one."
	previewEmptyText   = "Select a snippet to preview it."
	alreadyRunningText = "Sniper is already running. Switch to it from the taskbar."
)

func noMatchesText(query string) string { return fmt.Sprintf("No snippets match “%s”.", query) }
func copiedStatus(title string) string  { return fmt.Sprintf("Copied “%s”.", title) }
func savedStatus(title string) string   { return fmt.Sprintf("Saved “%s”.", title) }
func deletedStatus(title string) string { return fmt.Sprintf("Deleted “%s”.", title) }

func saveFailedText(err error) string {
	return fmt.Sprintf("Could not save: %v. Your changes are still here; try again.", err)
}
func deleteFailedText(err error) string { return fmt.Sprintf("Could not delete: %v", err) }
func deleteConfirmText(title string) string {
	return fmt.Sprintf("Delete “%s”? This cannot be undone.", title)
}
