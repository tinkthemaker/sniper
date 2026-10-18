# Testing

Everything that can be automated is automated. An agent can run every command in this document except those in [Human verification](#human-verification).

## Commands

```bash
gofmt -l .                      # must print nothing
go vet ./...
go test ./...                   # needs CGO + gcc for internal/ui and cmd/*
CGO_ENABLED=0 go test ./internal/snippets/... ./internal/store/... ./internal/instance/...
go test -race ./internal/snippets/... ./internal/store/...   # race needs CGO
go test -bench . -run ^$ ./internal/snippets/
go run ./cmd/clipcheck          # Windows only; needs an interactive desktop session
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

`scripts/check.ps1` (created in M0) runs the whole list except `clipcheck`, in this order: gofmt, vet, staticcheck, tests with coverage for the pure packages, then all tests. It stops at the first failure and exits non-zero. It is the single command an agent uses to verify any change: `pwsh -NoProfile -File scripts/check.ps1`. A `-Quick` switch runs only gofmt, vet, and `go test ./...`.

Coverage goals (`go test -cover`): `snippets` ≥ 95%, `store` ≥ 85%, `instance` ≥ 80%. `ui` has no number; test the behaviors listed below instead.

## Unicode corpus

Define this once, in `internal/snippets/testdata_test.go` (and copy it into `cmd/clipcheck`), and use it in round-trip tests:

```go
var corpus = []string{
    "plain ASCII",
    "line one\nline two\n\nline four",
    "tab\tseparated\tvalues",
    "trailing spaces   \n  leading spaces",
    "emoji 👋🏽 family 👨‍👩‍👧 flag 🇯🇵",
    "punctuation “quotes” ‘single’ — en–dash … © ® ™ € £ ¥",
    "日本語のテキスト / 中文文本 / 한국어 텍스트",
    "العربية مع English mixed — עברית",
    "Ελληνικά, Кириллица, हिन्दी, ไทย",
    "combining: é ñ (decomposed)",
    "zero width:​joiner‍ end",
    `json-ish: {"a": "b\\n"} <tag attr='x'> & %s %d`,
}
```

## `internal/snippets`

| Test | Cases |
|---|---|
| `TestNormalize` | CRLF and lone CR in the body become LF. A title with a newline becomes a space and is trimmed. The body is not trimmed. |
| `TestValidateFields` | Empty or whitespace title; title of 200 runes is OK and 201 fails (count runes, not bytes, using a multi-byte title); blank body; body of 65,536 bytes is OK and 65,537 fails; invalid UTF-8; NUL in either field. Check `ValidationError.Field` for each. |
| `TestValidateLibrary` | Wrong version; 10,001 snippets; empty ID; duplicate ID; one invalid record. |
| `TestNewID` | 32 lowercase hex characters; 10,000 calls produce no duplicates. |
| `TestAddUpdateDelete` | Returned libraries are correct; the **input library is unchanged** (compare with a deep copy taken beforehand, and check that the slice backing array is not shared by appending to the result); ErrNotFound. |
| `TestAppend` | Fresh IDs differing from the input IDs; order preserved; limit exceeded leaves the input unchanged. |
| `TestSearch` | Empty query returns all, sorted; case-insensitive; multiple terms with AND across fields; title-tier before body-tier; equal titles keep file order; non-Latin query (`"日本"`); emoji query; no matches gives an empty non-nil slice. |
| `BenchmarkSearch` | 1,000 snippets with 2 KB bodies and a 2-term query. Must be < 10 ms/op. Record the result in the decision log. |

## `internal/store`

Use `t.TempDir()` for every test.

| Test | Cases |
|---|---|
| `TestEncodeDecodeRoundTrip` | Every corpus string as title (newline-free variants) and body. Bytes are stable: encoding twice gives identical output, with a trailing newline and 2-space indentation. |
| `TestDecodeRejects` | Malformed JSON; trailing garbage after the object; an unknown top-level field; an unknown snippet field; missing version; version 2 (`errors.Is(err, ErrUnsupportedVersion)`); 16 MiB + 1 byte (`ErrTooLarge`); an invalid record. |
| `TestDecodeAcceptsBOM` | A UTF-8 BOM prefix decodes. |
| `TestDecodeImport` | Missing and duplicate IDs are accepted; records are normalized (CRLF); an invalid record gives an error containing `record 3:`. |
| `TestLoadStates` | No files gives StateNew; valid primary gives StateOK; corrupt primary with no backup gives StateDamaged and BackupAvailable=false; corrupt primary with a valid backup gives BackupAvailable=true; missing primary with a valid backup gives StateDamaged; corrupt primary with a corrupt backup gives StateDamaged and false. **After each, the primary bytes on disk are unchanged.** |
| `TestSaveCreatesBackup` | Save A, then save B: the primary is B and the backup is A. Save C: the backup is B. The first save on a new store writes no backup. |
| `TestSaveFailurePreservesPrimary` | Inject failures via `fileOps` into `createTemp`, the temp-file write (have `createTemp` return a file wrapper whose `Write` fails; this means `createTemp` returns a small unexported interface rather than `*os.File`), `rename`, and the backup write. Each time the primary bytes are unchanged, no `.tmp` files are left, and `lastGood` is unchanged (a later successful save still writes the correct backup). |
| `TestRenameRetry` (windows build tag) | `rename` fails twice with `syscall.Errno(32)` and then succeeds: Save succeeds. It always fails with Errno 5: Save errors after 5 attempts. A non-retryable error fails immediately. |
| `TestRealReplaceOnWindows` | No injection: save 50 times in a loop over an existing file. Each time the file decodes and holds the latest content. |
| `TestDailyBackup*` (`daily_test.go`, with `s.now` set to chosen days) | The first save of a day writes `backups/snippets-<day>.json` holding the pre-save bytes (also when those bytes were loaded, not saved); a second save the same day leaves it alone; the next day adds a new file. An 8th dated file prunes the oldest so 7 remain; files with other names, an invalid date, or a directory with a matching name survive. A failure creating the folder, writing the file or renaming it fails the save and leaves the primary, `snippets.bak.json` and `lastGood` unchanged, with no stray temp files, and a retry succeeds. A failing `readDir` or `remove` does not fail the save. A new install writes no dated backup. |
| `TestRestoreBackup` | After restoring, Load gives StateOK with the backup's content. An invalid backup gives an error and the primary is untouched. |
| `TestQuarantine` | The primary is renamed to the timestamped name; a later Load gives StateNew (or Damaged if a backup exists). |
| `TestExportImportRoundTrip` | Export a library, then `DecodeImport` the output: titles and bodies are equal and in order. |

## `internal/instance`

- Acquire in a temp dir succeeds. A second Acquire on the same dir in the same process returns `ErrAlreadyRunning`, because flock locks are per handle on Windows. If this turns out to be platform-dependent, test it with a helper subprocess (`os/exec` re-running the test binary with `-test.run=TestHelperHold` and an environment flag).
- After Release, Acquire succeeds again.
- Subprocess test: a child holds the lock and gets killed with `Process.Kill`; the parent can then Acquire. This proves the OS releases the lock when a process crashes.

## `internal/clipboard`

- `ToPlatform` table tests: `"a\nb"` becomes `"a\r\nb"` on Windows; `"a\r\nb"` also becomes `"a\r\nb"` (no doubling); no newline is unchanged; other OSes are the identity.
- `New(cb).SetText` with a fake `fyne.Clipboard` records the platform-converted text.

## `internal/ui` (headless)

Use `fyne.io/fyne/v2/test`: `test.NewTempApp(t)` (or `test.NewApp()` with `defer a.Quit()` if unavailable), `test.Type`, `test.TypeOnCanvas`, `w.Canvas().Focused()`, and `test.Tap`. Use a real `store.Store` on `t.TempDir()`, a fake `clipboard.Writer` that records calls, a fake `OpenFolder`, and a fixed `Now`.

Practical notes:

- Write the tests in `package ui` (not `ui_test`), so they can reach unexported fields: the current editor, its entries, the error label, the open dialog, and `modalOpen`. Don't add exported test hooks.
- **Simulate keys** by sending them to the focused widget: `w.Canvas().Focused().(fyne.Focusable).TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})`.
- **Simulate shortcuts** the same way: `.(fyne.Shortcutable).TypedShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierShortcutDefault})`. This exercises the real routing path.
- Import and export are implemented as `m.importFrom(r io.Reader, name string)` and `m.exportTo(w io.Writer, name string)`. The file-dialog callbacks only open the URI and call these.

Required behaviors:

1. Startup focus is on search. Typing filters `Results()`, and the first result is selected.
2. Down and Up move the selection and clamp at the ends. Enter copies the selected body (the fake clipboard receives it) and the status equals `Copied “<title>”.`
3. Escape clears the query.
4. Ctrl+N with search focused opens the editor (the shortcut-routing regression test). Type a title and body, press Ctrl+S with the body focused: the snippet is saved on disk (re-`Load` the store), selected, and the status is set.
5. Validation: an empty title shows `Title is required.` and the dialog stays open; nothing is written.
6. Save failure: wrap the real store in a test `ui.Storage` whose `Save` returns an error. (Do not rely on read-only directories; Windows ignores the read-only attribute on folders.) The dialog stays open with the error text, `Library()` is unchanged, and the file on disk is unchanged.
7. Edit via Ctrl+E changes the body; the change is visible in the preview and saved.
8. Delete via Ctrl+Shift+D, then confirm: the snippet is removed, and the selection moves per the rules.
9. Cancel with changes shows `Discard changes?`.
10. Import: valid file, then confirm, then the snippets are appended. Invalid file: the error dialog appears and the library is unchanged. Exercise these through internal functions that take an `io.Reader`, because the file dialog itself cannot be driven headlessly.
11. Export writes decodable JSON to a buffer.
12. Recovery: with a damaged primary and a valid backup, Restore opens the main window with the backup content. Move aside, then confirm, creates the damaged file and opens an empty main window.
13. Empty-state labels appear for an empty library and for no matches. Copy, Edit, and Delete are disabled with no selection.
14. Ctrl+N while the editor is open does nothing (modal guard).

Undo (D27; `undo_test.go`). Ctrl+Z is delivered as `&fyne.ShortcutUndo{}` to the focused widget if it is `Shortcutable`, otherwise to the canvas, as the driver does:

- Undoing a create, an edit, a delete and an import each restore the library and the file on disk (re-`Load` the store), with the right status and selection; an undone delete or edit under a query that hides it clears the query.
- The stack is capped at 20 steps; after 20 undos, `Nothing to undo.` and no save happens.
- A failing `Save` during undo shows `Could not undo: <err>` and leaves the stack, library, file and status unchanged; a retry works.
- Ctrl+Z works from the search box (which keeps its query), the list and a button, and through Edit → Undo.
- Inside the editor, Ctrl+Z is the entry's text undo and the library and stack do not change. While any dialog is open, Ctrl+Z and the menu item do nothing.
- The Edit menu sits between File and Help with one item, **Undo**, disabled while the stack is empty and enabled after a change.

Theme and chrome (added with the Slayer finish; see the UI spec's Visual design section):

15. `TestThemeContrastFloors`: every text color is at least 4.5:1 on each surface it can land on, and on both filled buttons.
16. `TestThemeIsAlwaysDark` and `TestThemeUsesEmbeddedJost`: the theme ignores the OS variant; each text style gets the right embedded Jost file (Medium, Medium Italic, Semi, Semi Italic), the files are real TrueType fonts, and symbol and monospace text keep Fyne's fonts.
17. `TestHeaderCountFollowsTheQuery` and `TestRowShowsSelectionBar`: the header count and the selection bar follow the query and the selection.
18. Editor Tab order is Title → Body → Cancel → Save, and Shift+Tab reverses it. Walk `Canvas().FocusNext()` and `FocusPrevious()`, which is what the real driver calls for Tab. Focusing fields directly would hide an order bug (D20).

Rows, double-click and the Copy button (D23). Fixtures live in `internal/ui/testdata/fixtures/`: `duplicates.json` (three snippets titled `Greeting`) and `long-titles.json` (200 runes, CJK, an emoji ZWJ sequence, RTL), loaded with `store.DecodeImport` and `snippets.Append`; the 12-item `testdata/sample-import.json` corpus is reused, and a 300-snippet library is generated in the test (`bigLibrary`). The tests tap the real row objects with `test.Tap`:

19. Rows hold only a bar and a title, truncated with an ellipsis.
20. Double-click on an unselected row selects it, copies its body, and sets the status. A single click selects without copying; two slow clicks, or clicks on different rows, are not a double-click.
21. After filtering, a double-click copies the row that was clicked.
22. In the 300-snippet library, after scrolling, a double-click copies the right snippet and leaves its row fully visible.
23. Double-click does nothing while the editor or a confirmation is open.
24. With duplicate titles, selecting each shows its own body in the preview and a double-click copies the right one.
25. Every row is `rowHeight` tall, and consecutive rows are `rowHeight` apart, over the corpus, long-titles, duplicates and large libraries.
26. The Copy button is at least 112×40 with 16 px text, taller than New, Edit and Delete, starts at x = 12, with the others 8 px apart and centered on it.
27. In a 520×600 window the search box, list and Copy start at x = 12, the search box and list end at width − 12, and the header and key strips span the full width.

### Visual review (screenshots without a display)

`TestScreens` renders key states with the software test driver and writes PNGs to `internal/ui/testdata/screens/` (git-ignored), but only when `SNIPER_SCREENS=1` is set. The states are: empty library, populated list with a selection, no matches, the editor with a validation error, the delete confirmation, the recovery window (with and without a backup), About, the import confirmation, and three screens from the Unicode corpus (emoji, CJK, right-to-left) to check font fallback under Jost, duplicate titles, long titles, and the 300-snippet library scrolled with a selection. Measure alignment from the PNGs instead of judging by eye: print, per horizontal band, the first and last non-background column (boxes start at 12, text at 16). Use `w.Resize(fyne.NewSize(520, 600))`, then `img := w.Canvas().Capture()`, then `png.Encode`.

After M4, M5, and M6, the agent runs `SNIPER_SCREENS=1 go test ./internal/ui -run TestScreens` and **opens the PNGs** (an image-capable Read tool works) to check the layout against the UI spec: nothing clipped, empty states centered, buttons enabled or disabled correctly. Fix anything wrong before the exit check. These screenshots are for review, not golden-image assertions, because font rendering differs between machines.

## `cmd/clipcheck` (Windows clipboard check)

A tiny Fyne program:

1. `a := app.New()`, then create and show a 1×1 window. After `a.Run()` starts (use `w.SetOnClosed`/`a.Lifecycle().SetOnStarted`), run the check in a goroutine. All Fyne calls inside the goroutine are wrapped in `fyne.DoAndWait`.
2. For each corpus string: `clipboard.New(a.Clipboard()).SetText(s)`. Then run `powershell -NoProfile -Command "$t = Get-Clipboard -Raw; [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($t))"`, base64-decode the output, and compare with `clipboard.ToPlatform(s)`.
3. Print `PASS <n>` or `FAIL <n>: want %q got %q` for each string, then `fyne.Do(a.Quit)`. Exit code 1 on any failure.

How to read the result:

- If every string passes, the CRLF conversion is correct. Record this in the decision log under D4.
- If the output shows `\r\r\n`, the toolkit already converts line endings. Make `ToPlatform` the identity on Windows, update the tests and D4, and re-run.
- If PowerShell cannot read the clipboard (for example, no interactive session), report that as a human-verification item. Do not mark it as passing.

## Human verification

An agent cannot do these unattended. List them in the final report as **pending human check** with exact steps. Items 2–5 can be done on the dev machine by an agent that has desktop-control (computer-use) tools and the user's approval. Item 1 always needs a clean machine.

To help, the agent generates `testdata/sample-import.json` from the corpus (one snippet per corpus string, with titles like `Corpus 01 – emoji`). The human imports it instead of typing the snippets.

1. Run `dist\Sniper.exe` on a Windows machine without Go or gcc installed (or a fresh Windows Sandbox). It launches with the icon visible in the taskbar. The exe is unsigned, so expect a SmartScreen "Windows protected your PC" prompt when it was downloaded; choose **More info → Run anyway** and note the prompt in the report.
2. Import `testdata/sample-import.json`. Copy each one and paste it into Notepad and into a browser `<textarea>` (for example a new GitHub issue or any web form). The text must match exactly, including line breaks and emoji.
3. Restart the app: the data persists. Launch a second copy: the "already running" message appears, and the first copy keeps working.
4. Corrupt `%AppData%\Sniper\snippets.json` (add `xx` at the end) and launch: the recovery window appears. Restore works.
5. Do a keyboard-only pass of the whole main flow using the keyboard map.
6. Record the exe size, cold-start time (stopwatch), and working-set memory (Task Manager) in the decision log.

### Status (2026-10-02)

| Item | Status |
|---|---|
| 1. Clean machine | **Pending**: needs Windows Sandbox or a second PC. |
| 2. Import + paste | Import and **Notepad: passed** (12/12 byte-exact). **Browser `<textarea>`: pending.** |
| 3. Persistence, second copy | **Passed** |
| 4. Corrupt + recover | **Passed** |
| 5. Keyboard-only pass | **Passed** after fixing the editor Tab order (D20). |
| 6. Measurements | Agent-measured on the packaged exe (decision log). Task Manager and stopwatch confirmation pending. |

Details are in decision log entry D20.

### Notes for agents running these checks with computer use

- **Blank screenshots:** on some Windows setups, computer-use screenshots come back as a solid `#1E1E1E` frame even for permitted apps ([anthropics/claude-code#91079](https://github.com/anthropics/claude-code/issues/91079)). Clicks and typing still work. Work around it by capturing the target window yourself (`System.Drawing.Graphics.CopyFromScreen` on the window's rectangle from `GetWindowRect`, after `SetProcessDPIAware`) and viewing the PNG.
- **Windows stay buried:** `open_application` may not raise a window above maximized apps. `SwitchToThisWindow` followed by `SetForegroundWindow` on the process's main window handle does.
- **Run a launched app by path:** a packaged exe outside the Start menu must be granted by its full path (for example `E:\...\dist\Sniper.exe`), not by name.
- **Paste checks:** paste into Notepad, save to a file, and compare bytes. Don't trust a screenshot. Separate items with something you can split on even if typed keystrokes arrive out of order (`(\r\n)*@@@(\r\n)?`), and check the total length.
- **Browser paste:** this needs a real paste into a real browser. The built-in browser pane's Ctrl+V is simulated, and computer use can't type into browsers.
- **Escape can't be tested through computer use.** Computer use reserves Esc as its global stop key and consumes the key press ([docs](https://code.claude.com/docs/en/computer-use): "the `Esc` key aborts computer use from anywhere, and the key press is consumed"), so an Escape sent by the tool never reaches the app and looks like a bug. Escape behavior is covered headlessly ("Escape clears the query", "Cancel with changes"); a real-app check needs a human pressing Esc.
- **Double-click works through computer use.** `double_click` on a row, then read the result with `Get-Clipboard -Raw`. Expect the body with CRLF line endings on Windows.
