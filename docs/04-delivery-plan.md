# Delivery plan

These milestones are ordered for an autonomous agent. Do them in order. A milestone is done only when its **exit check** passes. After each milestone, commit, and update [07-decision-log.md](07-decision-log.md) with any decision made or measurement taken. Test details are in [06-testing.md](06-testing.md).

## M0: Toolchain and skeleton

1. Check the tools: `go version`, `gcc --version`, `git --version`. As of October 1, 2026, this machine has **Go 1.27.0** (`C:\Program Files\Go`) and **MSYS2 UCRT64 gcc 16.2.0** (`C:\msys64\ucrt64\bin`, on the user PATH), and a cgo test program builds. If a shell can't find them, refresh `PATH` as shown below. The install steps are kept for other machines.
   - Go: `winget install --id GoLang.Go -e --accept-source-agreements --accept-package-agreements`. Then refresh `PATH` in the session (`$env:Path = [Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [Environment]::GetEnvironmentVariable('Path','User')`).
   - gcc: `winget install --id MSYS2.MSYS2 -e`, then `C:\msys64\usr\bin\bash -lc "pacman -Syu --noconfirm; pacman -S --noconfirm mingw-w64-ucrt-x86_64-gcc"`. Prepend `C:\msys64\ucrt64\bin` to `PATH` for the session, and run `go env -w CGO_ENABLED=1`.
   - Verify: `go version`, `gcc --version`, and `go env CGO_ENABLED CC`.
   - If installation is not possible (no permission or no network), continue with the pure-Go milestones (M2, M3) using `CGO_ENABLED=0`, and report the missing tool as a blocker for M1 and M4–M7.
2. `git init`. Add a `.gitignore` with `dist/`, `*.exe`, `*.test`, `*.out`, `coverage*`, and `internal/ui/testdata/screens/`. Add a `.gitattributes` with `* text=auto eol=lf`, `*.png binary`, and `*.ps1 text eol=crlf`.
3. `go mod init sniper`. Set the `go` directive to the installed minor version (at least 1.24).
4. `go get fyne.io/fyne/v2@v2.8.1 github.com/gofrs/flock@latest`, then `go install fyne.io/tools/cmd/fyne@latest`.
5. Create the package directories from [03-architecture.md](03-architecture.md) with `doc.go` package comments.
6. Create `scripts/check.ps1` (spec in the testing doc) and `scripts/package.ps1`. `package.ps1` runs `go run ./tools/genicon` if `Icon.png` is missing, runs `fyne package -os windows -release` in `cmd/sniper`, moves the exe to `dist/`, and prints its size.

**Exit:** `go build ./...`, `go vet ./...`, and `scripts/check.ps1 -Quick` succeed. Commit `chore: project skeleton`.

## M1: De-risk the toolkit (spike)

Prove that Fyne builds, packages, and copies correctly before investing in features.

1. `internal/clipboard` (`Writer`, `New`, `ToPlatform` with build tags and tests).
2. `cmd/clipcheck` exactly as specified in the testing doc.
3. `tools/genicon`. Run it to produce `cmd/sniper/Icon.png`.
4. Temporary `cmd/sniper/main.go`: a window with a hard-coded list of 3 corpus snippets, a selectable preview, and a Copy button.
5. `cmd/sniper/FyneApp.toml`. Package with `fyne package -os windows -release` from `cmd/sniper`, and move the output to `dist/`.

**Exit:** `go test ./internal/clipboard/...` passes. `go run ./cmd/clipcheck` prints PASS for every corpus string, or the D4 fallback has been applied and it then passes. `dist/Sniper.exe` exists. Record the exe size in the decision log. Commit `feat: toolkit spike and clipboard check`.

## M2: Snippet core

Implement `internal/snippets` exactly per the API in the architecture doc, with all listed tests and the benchmark.

**Exit:** `CGO_ENABLED=0 go test -cover ./internal/snippets/...` passes with ≥ 95% coverage. The benchmark is < 10 ms/op and recorded. Commit `feat: snippet model, validation, search`.

## M3: Persistence and single instance

Implement `internal/store` (codec, Load, Save with backup, atomic write with Windows retry, RestoreBackup, QuarantinePrimary, Export) and `internal/instance`, with all listed tests including fault injection and the subprocess lock tests.

**Exit:** `CGO_ENABLED=0 go test -cover ./internal/store/... ./internal/instance/...` passes, with coverage ≥ 85% and ≥ 80% respectively. `TestRealReplaceOnWindows` passes on Windows. Commit `feat: JSON store, backups, instance lock`.

## M4: Main window

Replace the spike `main.go` with the real startup sequence. Implement `ShowMain`: search entry, list, preview, buttons, status, empty states, selection rules, keyboard map, shortcut routing, and Copy. Implement `ShowAlreadyRunning` and `ShowFatal`.

**Exit:** headless UI tests 1–3 and 13 pass, and the screens have been reviewed (testing doc, "Visual review"). The app builds and launches with `go run ./cmd/sniper`; if launching is not possible in the environment, the build alone is enough. Commit `feat: main window search and copy`.

## M5: Editing

Editor dialog (new and edit), delete confirmation, save-failure handling, and the discard confirmation.

**Exit:** headless UI tests 4–9 and 14 pass, and the screens have been reviewed. Commit `feat: create, edit, delete snippets`.

## M6: Recovery, import/export, menus

Recovery window, File and Help menus, Import, Export, Open data folder, and the About and Keyboard shortcuts dialogs.

**Exit:** headless UI tests 10–12 pass, and the screens have been reviewed. Commit `feat: recovery, import/export, menus`.

## M7: Package and document

1. `scripts/package.ps1` produces `dist/Sniper.exe`.
2. Generate `testdata/sample-import.json` from the corpus (a small test or `go run` helper that uses `store.Export`), and commit it.
3. Rewrite the README "Status" section and add **Run**, **Where your data lives**, **Back up and restore**, and **Build from source** sections. Keep them short and use the real commands.
4. Run `scripts/check.ps1` and `go run ./cmd/clipcheck`. Everything passes, `gofmt -l .` is empty, and `go vet` is clean.
5. Record the exe size and anything learned in the decision log.

**Exit:** all automated checks pass. Commit `docs: run, backup, build instructions`.

## Final report (what the agent hands back)

- Milestones completed and commits.
- Test summary (packages, pass/fail, coverage, benchmark).
- The `clipcheck` result.
- Decision-log entries added.
- **Pending human checks**: the list from the testing doc, with any notes.
- Any blockers (for example, a missing gcc) and exactly what the human must do to unblock them.

## Optional after real use (not in this plan)

Tray icon, global hotkey, bringing an existing instance to the front, and a "minimize after copy" setting. Automatic insertion, templates, sync, and browser extensions stay out.
