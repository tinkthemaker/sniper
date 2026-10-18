# UI specification

Exact screens, behaviors, keys, and text. Put every user-facing string in `internal/ui/messages.go` so tests can assert on them. `“…”` are typographic quotes (U+201C and U+201D).

## Main window

Title `Sniper`. Initial size 520×600. Resizable. Closing the window quits the app. The look is an LCARS-style frame; the reference is `docs/design/lcars-mockup-v3.html` and its PNG.

```text
┌ File  Edit  Help ─────────────────────────────────────┐
│ ╭──────────────────────────────────────  SNIPER  ▮    │ ← top elbow: peach cap, bar, title, end cap
│ │ 12 IN LIB │ ╭ FIND ▮ ( Search snippets (Ctrl+F) ) ╮  │ ← block 1 (periwinkle): the count
│ ├───────────┤ │  01-0417  ADDRESS                    │ ← list rows: code, then title in caps
│ │ SEARCH    │ │ ▐02-1138  GREETING   ▌ orange pill   │ ← block 2 (gold)
│ ├───────────┤ │  03-2265  SIGN-OFF                   │
│ │           │ │ ▬▬▬▬▬▬ ▬ ▬▬▬ ▬▬▬ ▬▬▬                  │ ← fixed segmented divider
│ │ PREVIEW   │ │  PREVIEW · RECORD 02-1138            │ ← block 3 (lavender, takes the rest)
│ │           │ │  GREETING            (orange caps)   │
│ │           │ │  Hello,                (Jost body)   │
│ ├───────────┤ │ ╭──COPY──╮ NEW EDIT DELETE)          │
│ │ V0.1.0    │ ╰───────────────────────────────────── │ ← block 4 (red): the version
│ ╰─ Copied “Greeting”.      │CTRL+N NEW│CTRL+Z UNDO│▮ │ ← bottom elbow: red cap, status readout, hints
└────────────────────────────────────────────────────────┘
```

- **Left column** (96 px wide, blocks 4 px apart, labels black in the bottom-right corner of each block):
  1. periwinkle: the count, `<N> IN LIBRARY` with no query, `<M> OF <N>` while a query is active. A query of only spaces matches everything, so it counts as no query.
  2. gold: `SEARCH`.
  3. lavender, takes the rest of the column: `PREVIEW`.
  4. red: the version, `V<version>` in capitals, from the app metadata; `VDEV` when the build has none.
- **Search:** an orange `FIND` pill joined to an orange-outlined field of the same height, 4 px apart (the pill's left end and the field's right end are rounded). The field is the existing `searchEntry`, so its keys and shortcut routing are unchanged.
- **List rows** are 24 px high with no gap (the row pitch is `rowHeight`). A row is a record code (periwinkle, 13 px) and the title in capitals (Antonio, tan, 17 px, ellipsized). The selected row is an orange pill (24 px, fully rounded ends) with black text. Capitals are display only: stored titles, the preview body and search are untouched.
- **Record codes** are `NN-NNNN`, derived only from the snippet ID (FNV-1a of the ID), so a snippet keeps its code when the list is filtered, scrolled or reordered.
- **Double-click** a row selects it and copies it, exactly like Enter (same clipboard text, same status). A single click selects and does not copy. Double-clicks are ignored while a dialog is open.
- **Divider:** a fixed segmented bar (peach, gold, lavender, periwinkle, peach; 10 px, 4 px apart, outer ends rounded) between the list and the preview. It replaces the old resizable split; the list takes 55% of the room between the search and the buttons, and the preview the rest.
- **Preview:** a lavender label `PREVIEW · RECORD <code>` (just `PREVIEW` with no selection), the title in orange Antonio capitals, and the body in Jost, selectable and word-wrapped inside a scroll. With no selection the body shows `Select a snippet to preview it.`
- **Buttons** (all black Antonio capitals, 4 px apart, vertically centered on COPY): **COPY** is a large orange pill (120×40, left end rounded, starts on the pill line); **NEW** periwinkle; **EDIT** lavender; **DELETE** red with its right end rounded. COPY, EDIT and DELETE are disabled when nothing is selected; NEW is always enabled. A disabled pill drops to a dim version of its color and its label dims, so it stays clearly different from an enabled one. Delete asks for confirmation as before.
- **Status readout:** the status message sits in the red bottom bar, left-aligned and ellipsized, with `CTRL+N NEW` (gold) and `CTRL+Z UNDO` (lavender) hint segments to its right and an end cap. It keeps its message until the next action replaces it. No timers.
- Counted nouns agree with the number: `1 snippet`, `0 snippets`, `2 snippets` (in the import confirmation and the import and export statuses).
- Empty states replace the list content, with a centered label in the list area:
  - Library empty: `No snippets yet. Press Ctrl+N to create one.`
  - No matches: `No snippets match “<query>”.`

## Visual design (LCARS finish)

The look follows `docs/design/lcars-mockup-v3.html`: a black canvas, pastel blocks, and Antonio capitals. It is a Fyne theme (`internal/ui/theme.go`), a frame laid out in `internal/ui/lcars.go`, and the rows in `internal/ui/rows.go`. It is always dark, whatever the OS setting. The name and marks of the show it echoes are not used anywhere in the app.

| Role | Value | Used for |
|---|---|---|
| Canvas | `#000000` | window background, inputs, menus |
| Orange | `#FF9900` | action and focus: FIND, the selected row, COPY, the preview title, outlines |
| Peach | `#FF9966` | the top elbow and bar, divider ends |
| Gold | `#FFCC66` | block 2, the `CTRL+N` hint |
| Tan | `#FFCC99` | list titles, the search text |
| Lavender | `#CC99CC` | block 3, EDIT, the preview label, the `CTRL+Z` hint |
| Periwinkle | `#9999FF` | block 1, NEW, record codes |
| Red | `#CC6666` | block 4, DELETE, the bottom elbow, error text |
| Ink | `#F1E3D3` | running text (preview body, dialogs) |
| Dim | `#9A8A9A` | placeholders |
| Dialog surface / raised | `#1B151B` / `#332736` | dialogs, and the quiet buttons on them |

Rules:

- **Text on a colored block is black.** Every text color is at 4.5:1 or better on every surface it lands on, and black is at 4.5:1 or better on every block color, also under the hover and focus veils (`TestThemeContrastFloors`).
- **Type:** two faces, both under the SIL Open Font License and embedded in the exe (`internal/ui/fonts/`): **Antonio** SemiBold, a tall condensed gothic, for everything bold (titles, labels, buttons, headings), and **Jost** Medium for running text, with Medium Italic and Semi Italic for the italic styles. Antonio is used in its one weight, so every Antonio text is `TextStyle{Bold: true}`. Symbol and monospace text keep Fyne's fonts; scripts the faces lack (Greek, Cyrillic, Devanagari, Thai, CJK, Arabic) fall back to the system fonts. Sizes: 14 text and 12 captions; in the main window 13 (codes, labels), 17 (rows), 22 (preview title), 34 (SNIPER) and 12 (hints).
- **Shape:** the frame is flat rectangles with a radius per corner (Fyne 2.8's `canvas.Rectangle`): no gradients, glow or shadows. Buttons in dialogs are full pills; fields are 14 px rounded.
- **The frame.** The top elbow is a peach cap (96×56, 44 px top-left radius) joined to a 24 px bar by a neck; a black rectangle with one 24 px rounded corner, starting at the bar's lower edge, is the panel's inner corner, so the curve meets the bar tangentially. Same-colored shapes overlap by 8 px instead of abutting, because the GL painter feathers every edge and abutting shapes show a seam; the black corner extends past the neck onto the black canvas so its square edges cannot show. `SNIPER` (orange, 34 px) breaks the bar on the right, closed by a 28 px end cap. The bottom elbow mirrors it in red (96×40, 34 px bottom-left radius) with a 22 px bar holding the readout, the two hints and an end cap. `frame.Layout` computes every position from the window size, so the frame stretches with the window: the column's third block, the readout and the panel grow.
- **Two lines** run down the panel (constants in `lcars.go`). **Pill line, x = 120** (8 margin + 96 column + 16): pills and highlight bars start here, so the FIND pill, the selected row, the divider, the list and COPY share one left edge. **Text line, 12 px further in (x = 132):** all text starts here (FIND, record codes, the preview label, title and body). Widgets inset their text by the theme's 4 px inner padding, so a text widget sits at pill line + 8. The right edge is 12 px from the window edge (x = 508) for the search field, list and divider. `TestMainWindowTwoLines` checks the layout and the rendered pixels.
- **Centering:** Antonio's capitals sit about 2 px below the middle of their line box, so text centered in a bar is lifted by `inkLift` (2 px). Measured from renders, ink is within 1 px of the center of its pill or bar.
- **Dialogs, the editor, the recovery window and menus** get the palette and the pill buttons through the theme (a large button radius, Antonio bold labels) and keep their layouts; there are no elbows there. They stay on the 4 px grid with the older two lines, measured from their own edges (box line 12 px, text line 16 px; `onDialogGrid`).
- **Not ours to align:** the menu bar (`File  Edit  Help`) is window chrome drawn by Fyne with its own padding.
- **Hover and focus:** a translucent white veil (`#FFFFFF` at 11% and 22%) blended over the pill, so it shows on every pill color. The list's own hover fill is a pill too (the app theme's selection radius is 12).

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
| Double-click | A list row | Selects the row and copies it, like Enter. Ignored while a dialog is open |
| Ctrl+F | Anywhere in main window | Focuses search and selects all of its text |
| Ctrl+N | Anywhere in main window | New snippet |
| Ctrl+E | Anywhere in main window | Edit selected snippet |
| Ctrl+Shift+D | Anywhere in main window | Delete selected snippet (with confirmation) |
| Ctrl+Z | Anywhere in main window | Undo the most recent library change made this session (create, edit, delete, import); up to 20 steps, no redo. Not in the editor, where it stays the entry's own text undo |
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
6. **Modal guard.** Fyne dialogs are overlays on the *same* window canvas, so canvas shortcuts still fire while a dialog is open. `MainWindow` tracks `modalOpen bool`, set when any dialog opens and cleared in its close callback. While it is true, `handleShortcut` ignores main-window actions (Ctrl+F/N/E/Z/Shift+D), so Ctrl+N inside the editor cannot open a second editor.
7. Ctrl+Z arrives as `*fyne.ShortcutUndo`, not a custom shortcut. It is listed in `shortcutActions` beside the others, so it is registered on the canvas and `handleShortcut` matches it by `ShortcutName()` (`Undo`). The search box therefore runs app undo and loses its own text undo (D27). The editor's entries never pass it to `handleShortcut`, so they keep Fyne's text undo, and a canvas Ctrl+Z with the editor open is dropped by the modal guard. The Edit → Undo menu item has no `Shortcut` set, because Fyne runs a menu item's shortcut before the focused widget and would take Ctrl+Z away from the editor's entries.

Headless tests must prove that Ctrl+N works while search has focus, that Ctrl+S works while the body entry has focus, and that Ctrl+N does nothing while the editor is open.

**Editor Tab order:** Title → Body → Cancel → Save. Fyne's Tab order follows the order of a container's children, and `container.NewBorder` puts the center object first. So build the editor with `container.New(layout.NewBorderLayout(top, bottom, nil, nil), top, bodyRow, bottom)`, listing the children in visual order (D20).

**Tab inside the multiline body:** keep Fyne's native behavior, whether it moves focus or inserts a tab. Tabs in pasted text are always preserved. Record which one Fyne does in the decision log. Ctrl+S and Escape must work in either case.

## Editor dialog

A modal custom dialog without built-in buttons (`dialog.NewCustomWithoutButtons`), titled `New snippet` or `Edit snippet`. Size: the window size minus a 40 px margin, and at least 420×380.

```text
New snippet                                  (bold heading, drawn in the content)
Title
[ single-line editorEntry                         ]
Body
[ multiline editorEntry, TextWrapWord,            ]
[ fills remaining height                          ]
<error label, DangerImportance, hidden when empty>
                                    [ Cancel ] [ Save ]
```

Labels sit above their fields, as in Windows text boxes, so both fields share one left edge and the full width. Everything is on the grid from the dialog's edges (see Visual design → Dialogs). Cancel is a secondary button and Save is the primary; they are right-aligned, 8 px apart.

- Focus starts on Title for New and on Body for Edit.
- **Save** (or Ctrl+S): normalize, then validate. A validation error is shown in the error label, and focus moves to the offending field. Then mutate and call `Store.Save`. If the save fails, show `Could not save: <err>. Your changes are still here; try again.` and keep the dialog open. On success, close the dialog, set the status to `Saved “<title>”.`, and refocus search.
- **Cancel** (or Escape): if the title or body differs from its initial value, confirm `Discard changes?` with the buttons **Discard** and **Keep editing**. Otherwise close at once. Refocus search afterwards.

## Delete confirmation

`dialog.NewConfirm`, titled `Delete snippet?`, with the message `Delete “<title>”? This cannot be undone.` Use `SetConfirmText("Delete")` and `SetDismissText("Cancel")`. On a failed save, show the error dialog `Could not delete: <err>`, and the library stays unchanged. On success, the status reads `Deleted “<title>”.`

## Menus

- **File**
  - **Import…** opens a file-open dialog filtered to `.json`. Then: decode errors give an error dialog `Import failed: <err>. Your library was not changed.`; zero snippets give an info dialog `The file contains no snippets.`; otherwise a confirm dialog `Import N snippet(s) from <filename>? Importing the same file twice creates duplicates.` with buttons **Import** and **Cancel**. Afterwards the status reads `Imported N snippet(s).`, or on save failure an error dialog `Could not save the import: <err>. Your library was not changed.`
  - **Export…** opens a file-save dialog with the default name `sniper-export-<YYYY-MM-DD>.json` (from `Now`). Write with `store.Export` to the `URIWriteCloser`, then `Close`. Afterwards the status reads `Exported N snippet(s) to <filename>.`, or on error shows `Export failed: <err>`.
  - **Open data folder** calls `OpenFolder(Store.Dir())`. On error it shows `Could not open <path>: <err>`.
  - Quit (Fyne adds it automatically).
- **Edit**
  - **Undo** undoes the most recent library change made in this session: a created, edited or deleted snippet, or an import. `MainWindow` keeps a stack of the libraries from before each saved change (at most 20; the oldest is dropped), pushed only after the save succeeded. Undo saves the previous library through `Store.Save` like any change, and only then publishes it. On a save failure it shows the error dialog `Could not undo: <err>` and leaves the stack and library as they were. There is no redo. The item is disabled while the stack is empty. Statuses: `Undid new snippet “<title>”.`, `Undid edit of “<title>”.`, `Undid delete of “<title>”.`, `Undid import of N snippet(s).`, and with nothing to undo `Nothing to undo.`
  - After undoing an edit or a delete, the snippet it restored is selected (the query is cleared if it would hide it). After undoing a create or an import, the current selection stays if it still exists; otherwise the delete selection rules apply at its old position.
- **Help**
  - **Keyboard shortcuts** shows an info dialog listing the keyboard map.
  - **About** shows `Sniper <version>`, the data folder path, and `Snippets and exports are stored as plain text on this computer. Sniper makes no network connections.`

Resize file dialogs to the window size minus a 40 px margin (Fyne's default is cramped).

## Recovery window

Title `Sniper – Recover library`. Size 520×320.

```text
Sniper could not open your snippet library.     (bold)
<path to snippets.json>
<Err text>
Your data has not been changed.                 ← directly under the error it qualifies

[ Restore backup                              ]  ← primary (orange)
[ Move damaged file aside and start empty     ]
[ Open data folder     ]  [ Quit              ]  ← equal halves, 8 px apart
```

On the same grid as the main window: buttons on the 12 px box line, 8 px apart; text on the 16 px text line.

- **Restore backup** is disabled with the tooltip-like caption `No valid backup found.` when `BackupAvailable` is false. On success, the main window opens with the status `Restored from backup.`, and this window closes.
- **Move aside…** first confirms `Move snippets.json to <new name> and start with an empty library?` On success, the main window opens with the status `Moved damaged file to <name>.`
- Errors from either action show in an error dialog, and the window stays open.
- Open the main window **before** closing the recovery window, or the app will quit when its last window closes.

## Other windows

- **Already running:** `Sniper is already running. Switch to it from the taskbar.` with an **OK** button that calls `App.Quit()`.
- **Fatal:** a title plus selectable detail text (path and OS error) and a **Quit** button.

## Icon

`tools/genicon` draws a 512×512 PNG using only `image`, `image/color`, `image/draw`, and `image/png`: a rounded square in `#2B6CB0` with a white crosshair (a circle outline plus four ticks). The program is deterministic and writes `cmd/sniper/Icon.png`. Commit the output.
