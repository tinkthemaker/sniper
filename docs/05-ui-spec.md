# UI specification

Exact screens, behaviors, keys, and text. Put every user-facing string in `internal/ui/messages.go` so tests can assert on them. `“…”` are typographic quotes (U+201C and U+201D).

## Main window

Title `Sniper`. Initial size 520×600. Resizable. Closing the window quits the app.

```text
┌ File  Help ─────────────────────────────────────┐
│ SNIPER  SNIPPETS                    12 IN LIBRARY│  ← header strip, 32 px, full width
│─────────────────────────────────────────────────│  ← metal edge (1 px), full width
│  [ Search snippets (Ctrl+F)                  ]  │  ← searchEntry, 12 px above and below
│ ▌Greeting                          (bold)       │  ← widget.List. Row = the title only (ellipsis),
│  Sign-off                                       │    24 px high; 3 px blood-red bar overlaid on
│  Zebra notes                                    │    the row edge when selected; 8 px clear
│                                                 │    above the divider
│──────────────── VSplit (0.55) ──────────────────│  ← divider in metal edge, dim grip
│  PREVIEW                           (microlabel) │
│  Greeting                          (bold)       │  ← preview title
│  Hello,                                         │  ← preview body: Label, Selectable,
│  Thanks for reaching out.                       │    TextWrapWord, inside VScroll
│─────────────────────────────────────────────────│
│  ┌────────┐                                     │
│  │  Copy  │ [New] [Edit] [Delete]  Copied “Greeting”.│ ← Copy 112×40; the rest 8 px apart, centered
│  └────────┘                                     │    on it; status right-aligned, centered
│ Enter copy  Ctrl+N new  Ctrl+E edit  Ctrl+Shift+D delete │ ← key strip, 32 px
└─────────────────────────────────────────────────┘
```

- Rows show only the title, bold and truncated with an ellipsis. Duplicate titles are allowed; the preview shows the selected snippet's body, which tells them apart.
- **Double-click** a row selects it and copies it, exactly like Enter (same clipboard text, same status). A single click selects and does not copy. Double-clicks are ignored while a dialog is open.
- Header count: `<N> IN LIBRARY` with no query, `<M> OF <N>` while a query is active.
- **Copy**, **Edit**, and **Delete** are disabled when nothing is selected. **New** is always enabled. Delete is an ordinary secondary button; red appears only on the confirmation (see Visual design).
- Empty states replace the list content, with a centered label in the list area:
  - Library empty: `No snippets yet. Press Ctrl+N to create one.`
  - No matches: `No snippets match “<query>”.`
- The preview with no selection shows `Select a snippet to preview it.`
- The status sits on the button row, right-aligned, because it reports on those actions. It keeps its message until the next action replaces it. No timers.

## Visual design (Slayer finish)

The look is the **Slayer** finish from the Screech tool: blood red on black iron, with ember as the hot end. It is a Fyne theme (`internal/ui/theme.go`) plus a few chrome helpers (`chrome.go`). It is always dark, whatever the OS setting.

| Role | Value | Used for |
|---|---|---|
| Canvas | `#0A0A0A` | window background |
| Panel | `#100908` | inputs, menus |
| Surface | `#211211` | header and key strips, dialogs |
| Raised | `#2E1816` | secondary buttons (one step above the surface) |
| Selection | `#351A19` | selected row |
| Hover | `#F5EAEA` at 6% | translucent veil over hovered rows and buttons |
| Metal edge | `#510A09` | input borders, rules between regions |
| Text bright / mid / dim | `#F5EAEA` / `#C0AEAC` / `#A38A89` | body, secondary, microlabels |
| Blood red | `#C21A16` | structure only: selection bar, danger |
| Red text | `#E8433C` | key chips, error text, the Delete confirm button |
| Ember | `#FF6A18` | action and focus: Copy button, focus rings, links |

Rules:

- **Blood red never colors small text.** It is 3.3:1 on the canvas. Red that must be read uses red text; actions use ember. `TestThemeContrastFloors` keeps every text color at 4.5:1 or better on the surfaces it can land on.
- **One primary per view.** Ember fills at most one button (Copy). Other buttons use the quiet raised fill `#2E1816`, one step above the surface, so they still read as buttons, including on dialogs, whose background is the surface color. Hover is a translucent light veil (`#F5EAEA` at 6%) blended over whatever it covers.
- **Type:** Jost, a free Futura-spirited geometric sans (SIL Open Font License), embedded in the exe so every PC renders the same: Medium (500) for text and Semi (600) for bold, each with its italic (`internal/ui/fonts/`). Symbol and monospace text keep Fyne's fonts. Scripts Jost lacks (Greek, Devanagari, Thai, CJK, Arabic) fall back to the system fonts. 14 px text, 12 px captions, no other sizes. Microlabels (`SNIPER`, `PREVIEW`, the count) are uppercase captions; everything else is sentence case.
- **Shape:** 2 px radii, 1 px edges, no gradients, glow, or shadows. Dialogs use the surface color so they stand off the window.
- **Density:** list rows are 24 px high with no gap between them, so the row pitch is 24 px (it was 28.4: a 24.4 px row plus the list's 4 px gap). The title is centered vertically and sits on the text line. The list drops its gap through `denseTheme`; the 4 px inner padding stays, so row text stays on the text line.
- **Copy button:** the primary action, at least 112×40 with 16 px bold text, set by `buttonRowLayout` and a text-size theme override on that button only. It starts on the box line; New, Edit and Delete keep their size, sit 8 px right of it and are centered on it. The status is centered on the same line. The row keeps 12 px above and below.
- **Grid.** Every margin is on a 4 px grid, and two vertical lines run down every window and dialog (constants in `chrome.go`):
  - **Box line, 12 px in from each side:** search box, list (and its selection), buttons, editor fields.
  - **Text line, 16 px in:** all text. Widgets inset their text by the theme's 4 px inner padding, so a widget on the box line has its text on the text line. Windows asks for 16 px between a surface edge and its text.
  - **Edge to edge:** strips, rules and the split divider. The main and recovery windows turn off Fyne's window padding so these reach the edges.
- **Spacing:** 12 px between regions (header → search → list; rule → buttons → strip); 8 px between buttons and between a header and its content; 4 px between a field label and its field (8 px with the label's inner padding). Related things sit closer than unrelated ones.
- **Strips:** header and key strips are exactly 32 px high, with their caption text centered vertically.
- **Dialogs:** custom dialog content (the editor) is placed on the same grid measured from the dialog's edges. `onDialogGrid` cancels Fyne's fixed dialog padding, and the dialog's own heading is left empty and drawn inside the content, because Fyne pins that heading to the dialog edge. Fyne's standard dialogs (confirmations, About, errors) keep Fyne's layout.
- **Not ours to align:** the menu bar (`File  Help`) is window chrome drawn by Fyne with its own padding, so its text sits 8 px in rather than on the text line.
- **Key strip:** the shortcuts are shown at the point of use, in red; their actions are dim.

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
Sniper could not open your snippet library.     (bold)
<path to snippets.json>
<Err text>
Your data has not been changed.                 ← directly under the error it qualifies

[ Restore backup                              ]  ← primary (ember)
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
