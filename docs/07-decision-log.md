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

---

## Measurements

| Date | Metric | Value | Notes |
|---|---|---|---|
| 2026-10-01 | `BenchmarkSearch` (1,000 × 2 KB) | 3.2 ms/op (2.4 MB/op, 3,006 allocs) | M2, i7-13700K, 2-term query, worst case (every body lowercased) |
| 2026-10-01 | `dist/Sniper.exe` size | 23.5 MiB (24,591,872 bytes) | M1 spike, `fyne package -release` |
| | Cold start (human) | | |
| | Working set (human) | | |
