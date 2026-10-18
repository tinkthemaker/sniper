# Decision log

The agent adds an entry whenever it makes a choice the docs did not dictate, changes a documented decision, or records a measurement. Keep entries short. Newest entries go at the bottom. If a decision changes the spec, update the spec in the same commit.

Format: `D<n> — <date> — <title>`, then **Decision**, **Why**, and optionally **Revisit if**.

---

**D1 — 2026-10-01 — Fyne v2.8.1, minimum v2.6**
Decision: pin `fyne.io/fyne/v2 v2.8.1`. Use `App.Clipboard()` and `Label.Selectable`, which both need v2.6 or later.
Why: v2.8.1 is the latest release. These APIs replace the older window-level clipboard and give a selectable preview.

**D2 — 2026-10-01 — No goroutines in the app**
Decision: all I/O runs synchronously on the UI goroutine. Only `cmd/clipcheck` uses `fyne.Do`/`DoAndWait`.
Why: library files are at most 16 MiB, and this removes a whole class of threading bugs.
Revisit if: saves become noticeably slow.

**D3 — 2026-10-01 — Backup is the previous good primary**
Decision: before replacing `snippets.json`, write the bytes last loaded or saved by this process to `snippets.bak.json` (atomically). If the backup write fails, the save fails.
Why: the in-memory bytes are known to be valid, so a damaged primary can never be copied over the backup.

**D4 — 2026-10-01 — CRLF conversion on Windows (confirmed)**
Decision: `clipboard.ToPlatform` converts LF to CRLF on Windows.
Why: CRLF is the Windows text convention. It is unconfirmed whether Fyne converts line endings itself.
Revisit if: `cmd/clipcheck` shows `\r\r\n`. Then make the conversion the identity and update this entry with the result.

**D5 — 2026-10-01 — Explicit shortcut routing**
Decision: custom entry widgets forward app shortcuts to a handler, in addition to `Canvas.AddShortcut`.
Why: a focused `widget.Entry` swallows unknown shortcuts.

**D6 — 2026-10-01 — Copy has no Ctrl+C binding; Delete is Ctrl+Shift+D**
Why: Ctrl+C must keep copying selected text, and Delete and Ctrl+Delete are text-editing keys inside entries.

**D7 — 2026-10-01 — Damaged libraries are moved aside, never overwritten**
Decision: the "start empty" option renames the damaged file to `snippets.damaged-<timestamp>.json` and requires confirmation.
Why: the spec forbids silently discarding data. Renaming keeps the file recoverable by hand.

**D8 — 2026-10-01 — Limits**
Decision: title ≤ 200 runes, body ≤ 65,536 bytes, ≤ 10,000 snippets, files ≤ 16 MiB.
Why: these limits are generous for a personal library and keep UI and file sizes bounded.

**D9 — 2026-10-01 — Screen-reader support is a known limitation**
Decision: support full keyboard operation, but not screen readers.
Why: Fyne draws its own widgets, and its screen-reader support on Windows is limited. Keyboard-only use (acceptance criterion 6) is fully supported.
Revisit if: a screen-reader user needs Sniper. Fyne releases may add support, or Walk would become the alternative.

**D10 — 2026-10-01 — Toolkit confirmed: Fyne over Walk**
Decision: keep Fyne.
Why: its headless `test` package lets the agent check UI behavior and screenshots without a human, and it is actively maintained. Walk would give a native look, a smaller exe, and no gcc, but it has no headless testing.

**D11 — 2026-10-01 — Modal guard for canvas shortcuts**
Decision: main-window shortcuts are ignored while a dialog is open.
Why: Fyne dialogs share the window canvas, so without the guard, Ctrl+N inside the editor would open a second editor.

**D12 — 2026-10-01 — Readable JSON**
Decision: `Encode` uses `SetEscapeHTML(false)`.
Why: snippets often contain `<`, `>`, and `&`, and backups should stay readable by hand.

**D13 — 2026-10-01 — Validation messages and Excerpt length**
Decision: validation messages are `Title is required.`, `Title is too long (N characters; limit 200).`, `Body is required.`, `Body is too long (N bytes; limit 65,536).`, `Title|Body contains characters that cannot be saved.` (invalid UTF-8 or NUL), and `The library is full (limit 10,000 snippets).` Record-level errors are prefixed `Snippet N: ` (stored library) or `record N: ` (append/import). `Excerpt(body, n)` returns at most n runes including the trailing `…`.
Why: the spec gave only two example strings; these follow the same pattern and keep rows within the stated width.

**D14 — 2026-10-01 — Quarantine details**
Decision: `QuarantinePrimary` returns `""` and no error when there is no primary file (nothing to move), and never overwrites an existing damaged file: if `snippets.damaged-<ts>.json` exists it uses `snippets.damaged-<ts>-2.json`, `-3`, and so on.
Why: Windows `rename` replaces existing files, and two quarantines in the same second would otherwise destroy the earlier damaged copy.

**D15 — 2026-10-01 — Editor details**
Decision: (1) Tab inside the multiline body inserts a tab character, because Fyne's `Entry.AcceptsTab()` is true for multiline entries; Shift+Tab does nothing there. Keyboard users leave the body with Ctrl+S or Escape, both of which work. Tab in the single-line title moves focus as normal, and Enter in the title moves focus to the body. (2) The modal guard counts open dialogs (`MainWindow.dialogs`, `modalOpen()`) instead of a single `modalOpen bool`, because the discard confirmation opens on top of the editor and a bool would clear when the inner dialog closed. (3) The discard confirmation uses the title `Discard changes?` and the message `Your changes to this snippet will be lost.`; the delete confirm button uses danger styling.
Why: the UI spec asked for the Tab behavior to be recorded and did not give a discard message.

**D16 — 2026-10-01 — Recovery, import and menu details**
Decision: (1) When the primary file is missing (a backup exists), "Move damaged file aside…" has nothing to rename: the confirmation reads `snippets.json is missing, so there is nothing to move. Start with an empty library?` and the main window opens with the status `Started with an empty library.` The backup is left alone. (2) Templates in the UI spec of the form `<err>. …` strip a trailing period from `<err>`, because validation messages are full sentences and would otherwise print `..`. (3) The recovery window is not the app's master window; the main window is. (4) File and Help menu actions do nothing while a dialog is open, like the keyboard shortcuts. (5) Messages say `N snippets` even when N is 1, to match the UI spec strings exactly. (6) `FileDialog.Resize` panics before `Show`, so file dialogs are resized after they are shown. (7) Import checks the 10,000-snippet limit before asking for confirmation, so the user hears about it early; the check is repeated when applying.
Why: the docs left these cases open; each choice is the smallest that keeps every acceptance criterion.

**D17 — 2026-10-01 — Newer transitive dependencies**
Decision: after `go get fyne.io/fyne/v2@v2.8.1`, also raise the indirect modules `golang.org/x/image` (0.24.0 → 0.46.0), `golang.org/x/net` (0.35.0 → 0.59.0), `golang.org/x/crypto` (0.33.0 → 0.57.0), plus the `x/sys` and `x/text` that came with them. No new direct dependency was added; only Fyne and gofrs/flock are direct.
Why: `govulncheck` reported four vulnerabilities in the older `x/image`. Sniper decodes no untrusted images, but the fix is a one-line bump and `govulncheck ./...` now reports none.
Revisit if: a Fyne release pins an older version that breaks the build.

**D18 — 2026-10-01 — Spec text vs. code details found while building**
Decision: (1) `go.mod` says `go 1.27.0` (the installed toolchain), not just the minor version. (2) The sample import file is checked by a store test (`TestSampleImportFile`) and regenerated with `SNIPER_WRITE_SAMPLE=1 go test ./internal/store -run TestSampleImportFile`. (3) Corpus strings in Go source write zero-width and combining characters as `\u` escapes, because staticcheck (ST1018) rejects raw format characters and editors can silently change them.
Why: small gaps between the docs and what the tools actually require.

**D19 — 2026-10-01 — Prepared for public release**
Decision: module path `github.com/tinkthemaker/sniper`, MIT license, commits authored with the GitHub no-reply address, and test screenshots git-ignored (a malformed `.gitignore` line had let them be committed).
Why: the repository is published at https://github.com/tinkthemaker/sniper. The module path makes `go install …@latest` work.

**D20 — 2026-10-02 — Editor Tab order fixed; hands-on verification run**
Decision: the editor lays out its children in visual order (title, body, buttons) with `container.New(layout.NewBorderLayout(...), top, bodyRow, bottom)` instead of `container.NewBorder`, which puts the center (body) first. `TestEditorTabOrderIsTitleBodyButtons` guards it.
Why: Fyne's Tab order follows child order, so Tab from Title went to Cancel and Save before Body. This was found during the hands-on run below. The headless tests missed it because they focus fields directly, and Enter in the title already jumped to the body.
Hands-on run (dev machine, `dist\Sniper.exe`, driven through computer use): checks 2–5 in [06-testing.md](06-testing.md#human-verification) pass, except the browser paste.
- **Import:** the confirmation text matches the spec, and all 12 snippets were imported.
- **Notepad:** all 12 corpus snippets copied with the keyboard and pasted into Notepad match byte for byte.
- **Persistence:** create, save and restart keep the data, and the backup is written on the second save.
- **Second copy:** shows "already running"; OK exits it; the library is untouched.
- **Recovery:** after `xx` is appended to the library, the recovery window appears; Restore makes the primary byte-identical to the backup.
Still pending: the browser paste (no automated route from this environment: the built-in browser's Ctrl+V is simulated, computer use cannot type into browsers, and Claude in Chrome was not connected) and the clean-machine run (Windows Sandbox not installed).

**D21 — 2026-10-03 — Slayer finish and a denser main window**
Decision: the UI adopts the Slayer finish from the Screech tool (blood red on black iron, ember as the action color; palette, rules, and contrast floors in the UI spec's Visual design section). It ships as a Fyne theme (`theme.go`), always dark. Text is Consolas loaded from `%SystemRoot%\Fonts`, with Fyne's fonts as the fallback; nothing is bundled. The main window gains a header strip with a count, a preview microlabel, dense two-line rows with a selection bar, and a key strip showing the shortcuts. Copy is the only filled button; New, Edit, and Delete are flat, so red appears only on the delete confirmation and on errors.
Why: research on launcher and Windows apps (dense rows, one accent used sparingly, shortcuts shown where they apply) plus a house style the user already uses. Blood red is 3.3:1 on the canvas, so it is structure only; small red text uses a lifted red, and ember carries actions.
Revisit if: Consolas or the fallback fonts mis-render a script in the real app (headless screenshots with the corpus look correct), or a light theme is wanted.

**D22 — 2026-10-03 — Layout grid: one box line, one text line**
Decision: every window and the editor dialog sit on a 4 px grid with a 12 px box line (boxed controls) and a 16 px text line (all text). The theme's inner padding is 4, so any widget on the box line has its text on the text line. Strips, rules and the divider run edge to edge (window padding off), and strips are 32 px. Regions are 12 px apart, buttons and header-to-content 8 px. The status moved onto the button row, right-aligned. Secondary buttons use a raised fill (`#2E1816`) instead of being flat, and hover is a translucent veil. The editor puts labels above fields and draws its own heading so it lines up. The recovery window moved "Your data has not been changed" under the error. The split divider is recolored to match the rules, and the list keeps 8 px clear of it.
Why: measured from headless renders, the old layout had text starting at seven different x positions (8–22 px) and boxes at 4, 8 and 12. Sources (human-written):
- [Microsoft: content layout and spacing](https://learn.microsoft.com/en-us/windows/apps/design/basics/content-basics): 16 px from a surface edge to its text, 12 between content areas, 8 between buttons and between a header and its control.
- [Microsoft: typography in Windows](https://learn.microsoft.com/en-us/windows/apps/design/signature-experiences/typography): flush-left alignment, 14/20 and 12/16 type, ellipsis truncation.
- [NN/g: proximity](https://www.nngroup.com/articles/gestalt-proximity/): related items closer than unrelated ones; equal spacing everywhere destroys grouping. That is why the status moved beside the buttons and the reassurance under the error.
- [NN/g: principles of visual design](https://www.nngroup.com/articles/principles-visual-design/): few sizes, and contrast to signal function.
Flat secondary buttons also blended into dialogs, which are drawn in the button's old color.
Left as is: the menu bar is window chrome drawn by Fyne from the global inner padding, so its text sits 8 px in. Fyne's standard dialogs keep their own layout. Microsoft recommends sentence case and a proportional UI font; the uppercase microlabels and Consolas are kept as part of the D21 house style (Consolas caps are monospaced, so they don't need the extra letterspacing that all-caps text otherwise needs, per [Butterick](https://practicaltypography.com/letterspacing.html)).

**D23 — 2026-10-09 — Title-only rows, double-click to copy, tighter rows, bigger Copy**
Decision:
- **Rows show only the title.** The excerpt line is gone, so `snippets.Excerpt` and `TestExcerpt` are removed (no callers; older entries that mention them are left as written). Duplicate titles stay allowed; the preview, which shows the selected snippet's body, tells them apart.
- **Row pitch 28.4 px → 24 px.** Before: with the excerpt dropped, a row was 24.4 px (one 14 px bold line plus 4 px inner padding above and below) and the list added a 4 px gap, 28.4 px per row. After: rows are a fixed 24 px (`rowHeight`, on the 4 px grid and the minimum allowed) and the list adds no gap (`denseTheme` now wraps the list instead of each row), so the pitch is 24 px, 4.4 px less. The title is centered vertically in a row a hair shorter than the label's natural height, which only trims transparent padding. The selection fill and bar cover the whole row; `ScrollTo` uses the same height, and a test checks that a clicked row ends up fully visible. Measured from `TestScreens` PNGs: boxes start at x = 12, text at x = 16, rows 24 px apart.
- **Double-click copies.** Fyne 2.8.1's `widget.List` item (`listItem`) is only `Tappable`. The GLFW driver holds `Tapped` back for the double-tap delay (300 ms) on any hit object that is also `DoubleTappable`, so making a row `DoubleTappable` would delay every single click. Instead each row is a small `listRow` widget that is `Tappable` only and sits on top of the list's own item, so it receives the click. A tap selects the row at once (no delay) and records the snippet ID and the time from `Deps.Now`; a second tap on the same ID within 500 ms (Windows' default double-click time; `doubleClickWindow`) also runs `copySelected`. No goroutines and no timers. The row is found by the ID it currently shows, never an index, so filtering and scrolling cannot misdirect it. `tapRow` returns early while a dialog is open (`modalOpen`), and focuses the list as Fyne's own item did.
- **Trade-off:** the double-click is recognized by us rather than the driver. Single-click selection stays immediate, but the window is ours (500 ms, not Fyne's 300 ms), and two quick clicks on one row count as a double-click even if the pointer moved a little between them. Hover and the selection fill still come from Fyne's item underneath. A press-drag on a row is not special-cased.
- **Copy button:** at least 112×40, 16 px bold, through `buttonRowLayout` (the first button is enlarged, the others keep their size and are centered on it) and a text-size theme override around that button only; global sizes are unchanged. It starts on the box line, the status is vertically centered on the same line, and the row keeps 12 px above and below. Ember fill, disabled look, focus ring and contrast floors are unchanged.
- Help → Keyboard shortcuts gains a Mouse section; the key strip is unchanged.
Why: rows with one line are denser and scan faster; double-click is the usual way to act on a list item; Copy is the primary action and should look it.
Revisit if: 500 ms feels too loose, or press-drag on a row needs handling.

**D24 — 2026-10-11 — Jost replaces Consolas**
Decision: the UI face is Jost, a free Futura-spirited geometric sans (SIL Open Font License 1.1, © The Jost Project Authors), embedded in the exe from `internal/ui/fonts/` with its license. Medium (500) is the text weight and Semi (600) the bold, each with its italic; symbol and monospace text keep Fyne's fonts. Consolas is no longer read from `%SystemRoot%\Fonts`.
Why: the user wanted a less ubiquitous face, something 60s designers loved. Futura was everywhere in 60s advertising, on the Apollo 11 plaque and in *2001: A Space Odyssey*; Jost is a free face in its spirit (Futura itself is proprietary). Embedding it means the work PC renders the same without installing anything, and it adds about 425 KB to the exe. Weights were chosen from side-by-side renders of the corpus: Book (400) looked thin on the dark canvas, so Medium carries text; Semi matches Futura's classic "Demi" weight and Microsoft's advice to emphasize with semibold. Jost covers Latin and Cyrillic; Greek, Devanagari, Thai, CJK and Arabic fall back to system fonts with no missing glyphs (checked headlessly with the corpus).
Effects: text heights became fractional, which exposed two over-strict layout assumptions. The button row truncated its vertical offset (now rounded, so a button is never more than half a pixel off center), and a row-visibility test compared float edges exactly (now with a 0.01 px tolerance; the measured overshoot was 0.0002 px).
Supersedes: the D21 font choice, and the note in D22 that Consolas caps need no extra letterspacing. Jost's uppercase microlabels are proportional, and Fyne cannot letterspace text; Futura's wide geometric capitals hold up untracked at 12 px, so they are left as is.
Revisit if: the microlabels look cramped in the real app, or a script renders badly through fallback.

**D25 — 2026-10-11 — Main window Tab order fixed; Ctrl+S works from the editor's buttons**
Decision: the main window's top-level layout lists its children top to bottom (`container.New(layout.NewBorderLayout(top, bottom, nil, nil), top, divided, bottom)`), the same fix as D20. A canvas-level Ctrl+S handler (`saveOpenEditor`) saves the editor when it is the top dialog, so the shortcut also works while Cancel or Save has focus.
Why: a full functional QA pass found that Tab from the search box skipped the list and preview and landed on Copy (the center region is first in `NewBorder`'s child order), and that Ctrl+S did nothing after Tab moved focus onto an editor button, because only the editor's entries route it. `TestMainTabOrderFollowsTheSpec` and `TestCtrlSSavesWhileAnEditorButtonHasFocus` guard them.
Not fixed: Escape does nothing while an editor button has focus, because the driver sends Escape to the focused button's `TypedKey`, not to the canvas.

**D26 — 2026-10-11 — Singular counts; spaces-only search counts as no query**
Decision: the import confirmation and the import and export statuses say `1 snippet` for one (`snippetCount` in messages.go), and the header shows `<N> IN LIBRARY` when the query is only spaces. Raised by the D25 QA pass (spec issues 2 and 3); the user chose to fix both.
Why: "Imported 1 snippets." is wrong English, and a spaces-only query matches everything (search splits on whitespace), so `4 OF 4` implied a filter that was not there.

**D27 — 2026-10-18 — Undo and daily rolling backups**
Problem: there was one `snippets.bak.json`, overwritten by every save, and no undo, so a delete followed by two more saves was unrecoverable.
Decision:
- **Undo (UI only).** `MainWindow` keeps a stack of up to 20 `snippets.Library` values, pushed after each successful save of a create, edit, delete or import (libraries are immutable, so snapshots are cheap). Ctrl+Z and Edit → Undo pop one, save it through `Store.Save` like any change, then publish it; a failed save shows `Could not undo: <err>` and changes nothing. No redo, and an undo is not itself pushed. The stack lives only for the session; the store knows nothing about it.
- **Routing.** Fyne delivers Ctrl+Z as `*fyne.ShortcutUndo`. It is handled like the other app shortcuts (`shortcutAction` now holds any `fyne.Shortcut`). **Trade-off:** the search box gives up its own text undo, because a focused `widget.Entry` would otherwise consume Ctrl+Z and the box has no history worth keeping beyond a few characters. The editor's entries keep text undo, and the modal guard drops app undo while any dialog is open. The menu item has no `Shortcut` field, because Fyne runs a menu item's shortcut before the focused widget, which would take Ctrl+Z from the editor.
- **Daily backups (store).** On the first successful save of a local calendar day, before `snippets.bak.json` and the primary are written, `lastGood` is also written atomically to `backups/snippets-YYYY-MM-DD.json` (folder `0o700`). An existing file for the day is left alone, so it holds the state from the start of that day's work. The newest 7 matching files are kept, and pruning ignores its own errors and never touches other names. A failure to write the dated file fails the save with the primary and `.bak` untouched (the same rule as D3). A new install has no `lastGood` and writes none. `Store` has an unexported `now func() time.Time` (default `time.Now`) so tests can pick the day; `store.New` is unchanged.
- **Not done:** recovery does not list dated backups, because the recovery window's job is one obvious restore; the README explains the manual copy. Importing a dated file would append duplicates, so the README says to copy it over `snippets.json` with Sniper closed.
Why: seven days of start-of-day snapshots survive any number of same-day saves, and an in-session undo covers the common mistake, with no new dependency, goroutine or setting.

---

## Measurements

| Date | Metric | Value | Notes |
|---|---|---|---|
| 2026-10-01 | `BenchmarkSearch` (1,000 × 2 KB) | 3.2 ms/op (2.4 MB/op, 3,006 allocs) | M2, i7-13700K, 2-term query, worst case (every body lowercased) |
| 2026-10-01 | `dist/Sniper.exe` size | 24.0 MiB (25,129,984 bytes) | M7, `fyne package -release`; the M1 spike was 23.5 MiB |
| 2026-10-01 | Cold start | ~120 ms to a visible window (median of 5; first launch 361 ms) | **Agent-measured** with a polling script on the dev machine (i7-13700K); the human stopwatch run is still pending |
| 2026-10-01 | Working set | ~85 MB after 2 s (84–85 MB on warm runs; 139 MB on the first) | **Agent-measured** with `Process.WorkingSet64`; Task Manager check by a human is still pending |
| 2026-10-02 | Cold start (packaged exe, 12-snippet library) | 152 ms to a visible window | D20 hands-on run, measured from `Start-Process` until the main window exists |
| 2026-10-02 | Memory (packaged exe, 12-snippet library) | 87.4 MB working set, 114.2 MB private, after 2 s | D20 hands-on run; 129.5 MB working set on the very first launch of the session |
