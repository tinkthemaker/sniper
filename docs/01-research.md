# Research and rationale

Revised October 1, 2026. Based on official documentation; Grammarly was not tested hands-on. Fyne facts below were checked against the Fyne source and pkg.go.dev on October 1, 2026.

## Reference behavior

Grammarly's snippets offer searchable reusable phrases and paragraphs, keyboard selection, and library management. Sniper borrows the quick-retrieval idea and delivers text through ordinary clipboard copy and paste. [Grammarly: Create snippets](https://support.grammarly.com/hc/en-us/articles/4403077145485-Create-snippets)

## Recommendation

A small Go desktop utility fits the user's chosen direction. It copies plain text, and the user pastes it into their browser or another application. It needs no browser integration, page access, or editor-specific insertion logic.

Use Fyne for a Go UI built from standard widgets. Store snippets in a versioned JSON file. A personal library does not need SQLite. Keep file persistence in its own package so the storage format can change later if usage requires it.

## Verified toolkit facts (Fyne v2.8.x)

| Fact | Consequence for Sniper |
|---|---|
| Latest release is `fyne.io/fyne/v2` **v2.8.1** (Aug 22, 2026). No v3 module exists. | Pin `v2.8.1` or a newer v2.8.x patch release. Minimum supported is v2.6.0, because of the next two rows. |
| `fyne.App.Clipboard() fyne.Clipboard` exists **since 2.6**. `Window.Clipboard()` is the older API. | Use `app.Clipboard()`. Do not use the window-level clipboard. |
| `fyne.Clipboard` is `{ Content() string; SetContent(string) }`. Neither method returns an error. | The app cannot detect a failed copy. `cmd/clipcheck` checks real behavior instead (see [testing](06-testing.md)). |
| `widget.Label` has `Selectable bool` **since 2.6**, plus `Wrapping` and `Truncation`. | The preview is a selectable label, so users can still select and copy text manually. |
| `fyne.Do(func())` and `fyne.DoAndWait(func())` must wrap UI changes made from goroutines. | Sniper does all of its work on the UI goroutine and starts no goroutines, so it never needs them. The only exception is `cmd/clipcheck`. |
| The packaging tool is installed with `go install fyne.io/tools/cmd/fyne@latest` and run as `fyne package -os windows -release`. It reads `FyneApp.toml` (`[Details] Icon, Name, ID, Version, Build`) and defaults to `Icon.png`. | `cmd/sniper/` holds `FyneApp.toml` and `Icon.png`. |
| Fyne desktop builds need CGO and a C compiler. On Windows this is MSYS2 MinGW-w64 gcc. | Toolchain setup is milestone 0. Pure-Go packages are still testable with `CGO_ENABLED=0`. |
| Widget shortcuts go to the **focused** widget first. A focused `widget.Entry` swallows custom shortcuts it does not recognise, so they never reach `Canvas.AddShortcut`. | App shortcuts must be routed explicitly from custom entry widgets (see [UI spec](05-ui-spec.md#shortcut-routing)). |

Sources: [Fyne: Clipboard](https://docs.fyne.io/api/v2/fyne/clipboard/), [Fyne: Desktop packaging](https://docs.fyne.io/started/packaging/), [Fyne: Metadata](https://docs.fyne.io/started/metadata/), [Fyne: Quick start](https://docs.fyne.io/started/quick/), [pkg.go.dev versions](https://pkg.go.dev/fyne.io/fyne/v2?tab=versions).

## Other dependency choices

- **Single-instance lock:** `github.com/gofrs/flock`. It uses `LockFileEx` on Windows and `flock` on Unix. The OS releases the lock when the process exits or crashes, which avoids stale sentinel files.
- **IDs:** `crypto/rand` from the standard library.
- **Atomic replacement:** `os.Rename`. On Windows, Go implements it with `MoveFileEx(MOVEFILE_REPLACE_EXISTING)`, which replaces an existing destination. Antivirus or indexer handles can make it fail briefly with a sharing or access error, so the store retries (see [architecture](03-architecture.md#persistence-algorithm)).

No other third-party dependencies are allowed without a decision-log entry ([07-decision-log.md](07-decision-log.md)).

## Scope

Windows is the first and only tested target, based on this workspace. Code must still compile for Linux and macOS, with OS-specific code behind build tags, so cross-platform support stays possible. "Compact" means a small, focused interface and simple distribution. Binary size, startup time, and memory use are measured and recorded, not targeted.

Out of scope: direct insertion, global keyboard hooks, an always-running agent, tray icon, accounts, cloud sync, templates or variables, rich text, and browser extensions.
