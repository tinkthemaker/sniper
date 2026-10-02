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

	emptyLibraryText   = "No snippets yet. Press Ctrl+N to create one."
	previewEmptyText   = "Select a snippet to preview it."
	alreadyRunningText = "Sniper is already running. Switch to it from the taskbar."
)

func noMatchesText(query string) string { return fmt.Sprintf("No snippets match “%s”.", query) }
func copiedStatus(title string) string  { return fmt.Sprintf("Copied “%s”.", title) }
