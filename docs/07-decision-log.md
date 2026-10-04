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
