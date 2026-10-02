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

**D4 — 2026-10-01 — CRLF conversion on Windows (provisional)**
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

---

## Measurements

| Date | Metric | Value | Notes |
|---|---|---|---|
| | `BenchmarkSearch` (1,000 × 2 KB) | | |
| | `dist/Sniper.exe` size | | |
| | Cold start (human) | | |
| | Working set (human) | | |
