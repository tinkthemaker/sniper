# UI specification

Exact screens, behaviors, keys, and text. Put every user-facing string in `internal/ui/messages.go` so tests can assert on them. `“…”` are typographic quotes (U+201C and U+201D).

## Main window

Title `Sniper`. Initial size 520×600. Resizable. Closing the window quits the app.

```text
┌ File  Help ─────────────────────────────────────┐
│ [ Search snippets (Ctrl+F)                    ] │  ← searchEntry (focused at start)
│ ┌─────────────────────────────────────────────┐ │
│ │ Greeting                          (bold)    │ │  ← widget.List, row = title +
│ │ Hello, thanks for reaching out…   (dim)     │ │    body excerpt (both truncated
│ │ Sign-off                                    │ │    with ellipsis, one line each)
│ │ Best regards,                               │ │
│ ├──────────────── VSplit (0.45) ──────────────┤ │
│ │ Greeting                          (bold)    │ │  ← preview title
│ │ Hello,                                      │ │  ← preview body: Label, Selectable,
│ │ Thanks for reaching out. …                  │ │    TextWrapWord, inside VScroll
│ └─────────────────────────────────────────────┘ │
│ [ Copy ] [ New ] [ Edit ] [ Delete ]            │  ← Copy = HighImportance,
│ Copied “Greeting”.                              │    Delete = DangerImportance
└─────────────────────────────────────────────────┘  ← status label
```

- Row excerpt: `snippets.Excerpt(body, 80)`.
- **Copy**, **Edit**, and **Delete** are disabled when nothing is selected. **New** is always enabled.
- Empty states replace the list content, with a centered label in the list area:
  - Library empty: `No snippets yet. Press Ctrl+N to create one.`
  - No matches: `No snippets match “<query>”.`
- The preview with no selection shows `Select a snippet to preview it.`
- The status line keeps its message until the next action replaces it. No timers.

### Selection rules

| Event | Selection afterwards |
|---|---|
| Startup | First result, if any |
| Query changes | First result, if any |
| Save of a new snippet | The new snippet. If the current query hides it, clear the query first. |
| Save of an edit | The edited snippet. If the query now hides it, clear the query first. |
| Delete | The result that now sits at the deleted index; otherwise the previous one; otherwise none. |
| Import | Unchanged if the selected snippet is still in the results; otherwise the first result. |

Calling `list.Select(i)` fires `OnSelected`, so guard against re-entrancy with a `syncing bool` flag. Always scroll the selected row into view (`list.ScrollTo(i)`).

## Keyboard map

| Key | Context | Action |
|---|---|---|
| Typing | Search focused | Filters results |
| Down / Up | Search focused | Moves selection ±1, clamped (no wrap) |
| PageDown / PageUp | Search focused | Moves selection ±10, clamped |
| Enter | Search focused | Copies selected snippet |
| Escape | Search focused | Clears query; no-op if already empty |
| Up / Down, PageUp / PageDown | List focused | Same as in search: moves the **selection** directly, not just Fyne's focus highlight |
| Enter | List focused | Copies selected snippet |
| Ctrl+F | Anywhere in main window | Focuses search and selects all of its text |
| Ctrl+N | Anywhere in main window | New snippet |
| Ctrl+E | Anywhere in main window | Edit selected snippet |
| Ctrl+Shift+D | Anywhere in main window | Delete selected snippet (with confirmation) |
| Tab / Shift+Tab | Everywhere | Moves focus: search → list → preview → Copy → New → Edit → Delete |
| Ctrl+S | Editor | Save |
| Escape | Editor | Cancel (with discard confirmation if anything changed) |

Ctrl+C keeps its normal meaning everywhere (copy selected text), so the Copy action deliberately has no Ctrl+C binding.

### Shortcut routing

A focused `widget.Entry` receives every shortcut and swallows the ones it does not know, so `Canvas.AddShortcut` alone is **not enough**. Implement it like this:

1. `shortcuts.go` defines the app shortcuts as `*desktop.CustomShortcut` values with `Modifier: fyne.KeyModifierShortcutDefault` (Ctrl on Windows and Linux, Cmd on macOS; Ctrl+Shift+D adds `fyne.KeyModifierShift`), and has a `func (m *MainWindow) handleShortcut(s fyne.Shortcut) bool`. It matches on `ShortcutName()`, runs the action, and returns true if it handled the shortcut.
2. Register each shortcut on the window canvas (`Canvas().AddShortcut`) so it works when the list or a button has focus.
3. `searchEntry` embeds `widget.Entry` (call `ExtendBaseWidget`) and overrides `TypedShortcut`: if `handleShortcut` returns true, stop; otherwise call `Entry.TypedShortcut`. It also overrides `TypedKey` for Down, Up, PageDown, PageUp, Escape, and Return/Enter (both `KeyReturn` and `KeyEnter`), and delegates every other key.
4. The editor's entries use the same pattern with an editor handler (Ctrl+S, Escape).
5. `resultList` embeds `widget.List` and overrides `TypedKey` for Up, Down, PageUp, PageDown, and Return/Enter, so they act on the selection the same way as in search. It delegates every other key.
6. **Modal guard.** Fyne dialogs are overlays on the *same* window canvas, so canvas shortcuts still fire while a dialog is open. `MainWindow` tracks `modalOpen bool`, set when any dialog opens and cleared in its close callback. While it is true, `handleShortcut` ignores main-window actions (Ctrl+F/N/E/Shift+D), so Ctrl+N inside the editor cannot open a second editor.

Headless tests must prove that Ctrl+N works while search has focus, that Ctrl+S works while the body entry has focus, and that Ctrl+N does nothing while the editor is open.

**Editor Tab order:** Title → Body → Cancel → Save. Fyne's Tab order follows the order of a container's children, and `container.NewBorder` puts the center object first. So build the editor with `container.New(layout.NewBorderLayout(top, bottom, nil, nil), top, bodyRow, bottom)`, listing the children in visual order (D20).

**Tab inside the multiline body:** keep Fyne's native behavior, whether it moves focus or inserts a tab. Tabs in pasted text are always preserved. Record which one Fyne does in the decision log. Ctrl+S and Escape must work in either case.

## Editor dialog

A modal custom dialog without built-in buttons (`dialog.NewCustomWithoutButtons`), titled `New snippet` or `Edit snippet`. Size: the window size minus a 40 px margin, and at least 420×380.

```text
Title  [ single-line editorEntry                 ]
Body   [ multiline editorEntry, TextWrapWord,     ]
       [ fills remaining height                   ]
<error label, DangerImportance, hidden when empty>
                                [ Cancel ] [ Save ]
```

- Focus starts on Title for New and on Body for Edit.
- **Save** (or Ctrl+S): normalize, then validate. A validation error is shown in the error label, and focus moves to the offending field. Then mutate and call `Store.Save`. If the save fails, show `Could not save: <err>. Your changes are still here; try again.` and keep the dialog open. On success, close the dialog, set the status to `Saved “<title>”.`, and refocus search.
- **Cancel** (or Escape): if the title or body differs from its initial value, confirm `Discard changes?` with the buttons **Discard** and **Keep editing**. Otherwise close at once. Refocus search afterwards.

## Delete confirmation

`dialog.NewConfirm`, titled `Delete snippet?`, with the message `Delete “<title>”? This cannot be undone.` Use `SetConfirmText("Delete")` and `SetDismissText("Cancel")`. On a failed save, show the error dialog `Could not delete: <err>`, and the library stays unchanged. On success, the status reads `Deleted “<title>”.`

## Menus

- **File**
  - **Import…** opens a file-open dialog filtered to `.json`. Then: decode errors give an error dialog `Import failed: <err>. Your library was not changed.`; zero snippets give an info dialog `The file contains no snippets.`; otherwise a confirm dialog `Import N snippets from <filename>? Importing the same file twice creates duplicates.` with buttons **Import** and **Cancel**. Afterwards the status reads `Imported N snippets.`, or on save failure an error dialog `Could not save the import: <err>. Your library was not changed.`
  - **Export…** opens a file-save dialog with the default name `sniper-export-<YYYY-MM-DD>.json` (from `Now`). Write with `store.Export` to the `URIWriteCloser`, then `Close`. Afterwards the status reads `Exported N snippets to <filename>.`, or on error shows `Export failed: <err>`.
  - **Open data folder** calls `OpenFolder(Store.Dir())`. On error it shows `Could not open <path>: <err>`.
  - Quit (Fyne adds it automatically).
- **Help**
  - **Keyboard shortcuts** shows an info dialog listing the keyboard map.
  - **About** shows `Sniper <version>`, the data folder path, and `Snippets and exports are stored as plain text on this computer. Sniper makes no network connections.`

Resize file dialogs to the window size minus a 40 px margin (Fyne's default is cramped).

## Recovery window

Title `Sniper – Recover library`. Size 520×320.

```text
Sniper could not open your snippet library.

<path to snippets.json>
<Err text>

Your data has not been changed.

[ Restore backup ]  [ Move damaged file aside and start empty ]
[ Open data folder ]  [ Quit ]
```

- **Restore backup** is disabled with the tooltip-like caption `No valid backup found.` when `BackupAvailable` is false. On success, the main window opens with the status `Restored from backup.`, and this window closes.
- **Move aside…** first confirms `Move snippets.json to <new name> and start with an empty library?` On success, the main window opens with the status `Moved damaged file to <name>.`
- Errors from either action show in an error dialog, and the window stays open.
- Open the main window **before** closing the recovery window, or the app will quit when its last window closes.

## Other windows

- **Already running:** `Sniper is already running. Switch to it from the taskbar.` with an **OK** button that calls `App.Quit()`.
- **Fatal:** a title plus selectable detail text (path and OS error) and a **Quit** button.

## Icon

`tools/genicon` draws a 512×512 PNG using only `image`, `image/color`, `image/draw`, and `image/png`: a rounded square in `#2B6CB0` with a white crosshair (a circle outline plus four ticks). The program is deterministic and writes `cmd/sniper/Icon.png`. Commit the output.
