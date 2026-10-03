# Sniper

Sniper is a compact Go desktop utility for Windows. Save reusable text, find it quickly, and copy it to the clipboard.

**Status:** version 0.1.0 is implemented. Its automated checks pass, and a hands-on run of the packaged exe passed (import, a byte-exact Notepad paste of all the test text, persistence, single instance, recovery, keyboard-only use). Still open: a paste into a browser, and a run on a clean Windows machine. See [the testing doc](docs/06-testing.md#human-verification). Agents should start with [CLAUDE.md](CLAUDE.md).

## What it does

Open a small window, type to search your snippets, press Enter, and paste into your browser or any other application. You can create, edit, and delete snippets in the same window. Data is kept in a local JSON file (`%AppData%\Sniper\snippets.json`) with an automatic last-good backup, and you can import and export it.

The stack is Go with Fyne 2.8. It needs no browser extension, account, or network access.

## Run

Build `dist\Sniper.exe` (see [Build from source](#build-from-source)), then double-click it or pin it to the taskbar. The search box has focus.

| Key | Action |
|---|---|
| Type | Filter by title or body |
| Down / Up, PageDown / PageUp | Move the selection |
| Enter | Copy the selected snippet |
| Escape | Clear the search |
| Ctrl+F | Focus the search box |
| Ctrl+N / Ctrl+E / Ctrl+Shift+D | New / edit / delete the selected snippet |
| Ctrl+S, Escape | Save, cancel (in the editor) |

Help → Keyboard shortcuts shows the same list. To run from source instead: `go run ./cmd/sniper`.

The exe is unsigned, so Windows SmartScreen may show "Windows protected your PC" the first time. Choose **More info → Run anyway**.

## Where your data lives

| File in `%AppData%\Sniper` | What it is |
|---|---|
| `snippets.json` | Your library |
| `snippets.bak.json` | The library as it was before the most recent save |
| `sniper.lock` | Held while Sniper runs, so a second copy cannot write the library |
| `snippets.damaged-<time>.json` | A damaged library you chose to move aside |

File → Open data folder opens it. The files are plain text. Moving or deleting `Sniper.exe` does not touch your data.

## Back up and restore

- **Back up:** File → Export… saves everything to a `.json` file (default name `sniper-export-<date>.json`). Or copy `snippets.json`.
- **Restore or move to another computer:** File → Import… reads an exported file and appends its snippets. Importing the same file twice creates duplicates, so Sniper asks first. Nothing is imported if any record is invalid.
- **If the library cannot be opened:** Sniper shows a recovery window and changes nothing until you choose. **Restore backup** uses `snippets.bak.json`; **Move damaged file aside and start empty** renames the bad file so you can inspect it.

To try it with awkward text (emoji, tabs, Arabic, Japanese, line breaks), import `testdata/sample-import.json`.

## Build from source

You need Go 1.27 or later and a C compiler, because Fyne uses CGO. An older Go (1.21+) will download 1.27 automatically. On Windows, install [MSYS2](https://www.msys2.org/) and its UCRT64 gcc, and put `C:\msys64\ucrt64\bin` on your `PATH`:

```bash
winget install --id GoLang.Go -e
winget install --id MSYS2.MSYS2 -e
C:\msys64\usr\bin\bash -lc "pacman -S --noconfirm mingw-w64-ucrt-x86_64-gcc"
```

Then, from this folder:

```bash
go run ./cmd/sniper                               # run
pwsh -NoProfile -File scripts/check.ps1           # gofmt, vet, staticcheck, all tests
pwsh -NoProfile -File scripts/check.ps1 -Quick    # gofmt, vet, tests only
go run ./cmd/clipcheck                            # real-clipboard check (needs a desktop session)
pwsh -NoProfile -File scripts/package.ps1         # builds dist\Sniper.exe
```

Or install it straight from GitHub (still needs gcc): `go install github.com/tinkthemaker/sniper/cmd/sniper@latest`.

`package.ps1` installs the `fyne` packaging tool if it is missing. To look at the UI without a display, run `SNIPER_SCREENS=1 go test ./internal/ui -run TestScreens` and open the PNGs in `internal/ui/testdata/screens/`.

## Documents

| Doc | Purpose |
|---|---|
| [CLAUDE.md](CLAUDE.md) | Operating instructions for the implementing agent |
| [01 Research](docs/01-research.md) | Rationale and verified toolkit facts |
| [02 Product spec](docs/02-product-spec.md) | Behavior, rules, limits, file format, acceptance criteria |
| [03 Architecture](docs/03-architecture.md) | Packages, APIs, persistence algorithm, startup |
| [04 Delivery plan](docs/04-delivery-plan.md) | Ordered milestones with exit checks |
| [05 UI spec](docs/05-ui-spec.md) | Layout, keyboard map, dialogs, exact strings |
| [06 Testing](docs/06-testing.md) | Automated tests, the clipboard check, human checks |
| [07 Decision log](docs/07-decision-log.md) | Decisions and measurements |

Direct insertion, sync, templates, tray, and global hotkeys are outside the MVP.

## License

[MIT](LICENSE)
