# Product specification: compact Go utility

This document defines **what** Sniper does. [03-architecture.md](03-architecture.md) defines how it is built, and [05-ui-spec.md](05-ui-spec.md) defines exact screens, keys, and text. If they disagree, fix the docs and log the change in [07-decision-log.md](07-decision-log.md). Do not guess.

## Goal

Save text once, find it quickly, copy it, and paste it wherever needed.

## Main flow

1. Open Sniper from its executable or a taskbar shortcut. The search box has focus.
2. Type to filter snippets by title or body. The first result is selected automatically.
3. Press Enter, or click **Copy**, to put the selected snippet's body on the clipboard.
4. Switch to the browser or other application and paste normally.

Everything happens in one compact, resizable window: a search box, a result list, a body preview, action buttons, a status line, and a menu bar. New and Edit open a modal form with a title and a multiline body. There is no separate management application.

## Snippet rules

A snippet has an `ID`, a `Title`, and a `Body`. All text is UTF-8.

| Rule | Value |
|---|---|
| Title normalization | Replace every CR or LF with a space, then trim leading and trailing whitespace. |
| Title validity | Non-empty after normalization. At most **200 runes**. Valid UTF-8. No NUL (`\x00`). |
| Body normalization | Convert `\r\n` and lone `\r` to `\n`. Otherwise store exactly as typed: no trimming, and tabs and trailing spaces are kept. |
| Body validity | Not blank (`strings.TrimSpace(body) != ""`). At most **65,536 bytes**. Valid UTF-8. No NUL. |
| ID | 32 lowercase hex characters from 16 `crypto/rand` bytes. Unique within the library. Never shown to users. |
| Library size | At most **10,000** snippets. |
| Duplicate titles | Allowed. Rows show only the title, so the preview pane tells duplicates apart: it shows the selected snippet's body. |

Validation errors name the field and the rule, for example "Title is required." or "Body is too long (70,112 bytes; limit 65,536)." The exact strings are in the UI spec.

## Search

- The query is lowercased with `strings.ToLower` and split into terms with `strings.Fields`.
- A snippet **matches** when every term appears as a substring of the lowercased title or the lowercased body. Each term may match in either field.
- **Ranking:** tier 0 holds snippets where every term appears in the title. Tier 1 holds the rest. Within a tier, sort by lowercased title (byte order), then by position in the library file, so the order is stable.
- An empty or whitespace-only query returns every snippet, sorted by lowercased title and then file position.
- Search runs on every keystroke, synchronously, with no debounce. It must take under 10 ms for 1,000 snippets with 2 KB bodies (benchmark in [06-testing.md](06-testing.md)).

## Included features

- Create, edit, and delete snippets. Delete asks for confirmation.
- Clipboard copy that preserves line breaks, tabs, emoji, punctuation, and non-Latin scripts.
- Local JSON persistence across restarts.
- JSON export and import for backup and transfer.
- **File** menu: Import…, Export…, Open data folder. **Help** menu: Keyboard shortcuts, About.
- Full keyboard operation (see the UI spec).

## Clipboard behavior

Copy deliberately replaces the current clipboard contents. Sniper does not restore the previous clipboard and does not paste automatically. After a copy, the status line reads `Copied “<title>”.` The toolkit cannot confirm success, so Sniper never claims more than that. The preview is a selectable label, so manual select-and-copy (Ctrl+C) also works.

On Windows the clipboard wrapper converts `\n` to `\r\n` before writing. `cmd/clipcheck` confirms the result. See decision D4 in the decision log.

## Local data

| Item | Location |
|---|---|
| Data directory | `filepath.Join(os.UserConfigDir(), "Sniper")`. On Windows this is `%AppData%\Sniper`. Created with mode `0o700` if missing. |
| Library | `snippets.json` |
| Last-good backup | `snippets.bak.json`: the library as it was before the most recent successful save. |
| Daily backups | `backups/snippets-YYYY-MM-DD.json` (folder mode `0o700`): the library as it was before the first save of that local calendar day. The newest 7 are kept. |
| Lock file | `sniper.lock` (held by the running instance) |
| Quarantined damaged files | `snippets.damaged-YYYYMMDD-HHMMSS.json` |

Data lives outside the executable folder, so moving or upgrading the executable does not move the library. Removing the executable does not remove user data.

### File format (schema version 1)

```json
{
  "version": 1,
  "snippets": [
    { "id": "3f2a…", "title": "Greeting", "body": "Hello,\nThanks for reaching out." }
  ]
}
```

Files are UTF-8 without a BOM, indented 2 spaces, and end with a newline. Readers ignore a leading UTF-8 BOM. Snippet order in the file is creation order, and new snippets are appended. Unknown fields, a missing `version`, or any version other than `1` make the file invalid. Files larger than **16 MiB** are rejected without being parsed.

### Save guarantees

- A change is written to disk **before** the UI shows it or reports `Saved`.
- A failed save leaves `snippets.json` exactly as it was, keeps the in-memory library unchanged, and keeps the edit form open with an error so the user can retry.
- `snippets.bak.json` always contains the previously saved good library (after the second save onward).
- **Daily backups.** On the first successful save of a local calendar day, before the primary is replaced, the same bytes that go to `snippets.bak.json` are also written atomically to `backups/snippets-YYYY-MM-DD.json`. An existing file for today is never touched, so each dated file holds the library as it stood when that day's work began. After a successful save only the newest 7 files matching `snippets-YYYY-MM-DD.json` (by the date in the name) are kept; failing to delete an old one is ignored, and files in the folder that do not match the pattern are never deleted. If the dated backup cannot be written, the save fails and the primary and `snippets.bak.json` are untouched, like any backup failure. The first save of a new install has nothing to back up and writes none. To restore one, quit Sniper and copy the dated file over `snippets.json`; the recovery window does not list them, and importing one would append duplicates.

### Startup states

| Situation | Behavior |
|---|---|
| No `snippets.json` and no backup | First run. Start with an empty library. Nothing is written until the first save. |
| `snippets.json` valid | Load it and show the main window. |
| `snippets.json` missing, unreadable, too large, invalid JSON, wrong version, or invalid records | Show the **recovery window** (UI spec). Never start empty over the top of it. |
| Another instance holds the lock | Show "Sniper is already running…" with an OK button, then exit. Do not load data. |
| Data directory cannot be created or the lock cannot be created | Show an error window with the path and the OS error, plus a Quit button. |

Recovery window options:

1. **Restore backup.** Enabled only if `snippets.bak.json` loads and validates. Copies the backup over `snippets.json` atomically, then opens the main window.
2. **Move damaged file aside and start empty.** Renames `snippets.json` (if present) to `snippets.damaged-<timestamp>.json`, then opens an empty main window. Requires a confirmation dialog.
3. **Open data folder.**
4. **Quit.**

## Import and export

- **Export** writes the full library in the schema-v1 format to a path the user picks. The default name is `sniper-export-YYYY-MM-DD.json`. Afterwards the status reads `Exported N snippet(s) to <filename>.`
- **Import** reads a schema-v1 file. The whole file is decoded and every record is normalized and validated before anything changes. IDs in the file are ignored (they may be missing or duplicated), and every imported snippet gets a fresh ID. Imported snippets are appended in file order.
  - If any record is invalid, nothing is imported, and the error names the 1-based record number and the problem.
  - If the result would exceed 10,000 snippets, nothing is imported.
  - If the file contains zero snippets, show "The file contains no snippets." and stop.
  - Otherwise confirm: "Import N snippet(s) from <filename>? Importing the same file twice creates duplicates." On confirm, append them and save once. If the save fails, the library stays unchanged.

## Privacy and network

Sniper makes no network requests. There is no account and no telemetry. Library and export files are plaintext; the About dialog says so.

## Acceptance criteria

Each criterion maps to the tests in [06-testing.md](06-testing.md). **(A)** means automated; **(H)** means it needs a human on Windows.

1. Create, edit, restart, and reopen preserve content exactly. (A store round-trip, H restart)
2. Search finds title and body text, case-insensitively, with title matches ranked first. (A)
3. Copy and paste preserve multiline text, tabs, emoji, punctuation, and non-Latin text. Windows clipboard bytes are checked by `clipcheck` (A on Windows); pasting into Edge or Chrome and into Notepad is (H).
4. Export then import round-trips content. Invalid imports leave the library unchanged. (A)
5. Save failures and corrupt files never silently discard data. (A with fault injection, H recovery window)
6. Keyboard-only users can search, select, copy, create, edit, and delete. (A headless UI tests, H)
7. A second instance cannot write the library. (A lock test, H launch twice)
8. The packaged Windows executable runs without Go or developer tools installed. (H)

Tray access and a global shortcut may follow if reopening from the taskbar feels slow. They are not launch requirements.
