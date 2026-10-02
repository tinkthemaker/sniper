# Agent instructions: Sniper

You are building Sniper, a compact Windows desktop snippet utility in Go and Fyne, from these docs and without human help. The docs are the contract.

## Read first, in order

1. [docs/02-product-spec.md](docs/02-product-spec.md): behavior, rules, limits, file format
2. [docs/03-architecture.md](docs/03-architecture.md): packages, exact APIs, algorithms, startup
3. [docs/05-ui-spec.md](docs/05-ui-spec.md): screens, keys, strings, shortcut routing
4. [docs/06-testing.md](docs/06-testing.md): required tests and commands
5. [docs/04-delivery-plan.md](docs/04-delivery-plan.md): milestone order and exit checks
6. [docs/07-decision-log.md](docs/07-decision-log.md): decisions already made
7. [docs/01-research.md](docs/01-research.md): background and verified toolkit facts

## Rules

- **Follow the delivery plan in order.** Do not start a milestone until the previous exit check passes.
- **Do not expand scope.** Out of scope: tray, global hotkeys, auto-paste, templates, sync, accounts, network, rich text, installers, and any setting not in the spec.
- **Dependencies:** only `fyne.io/fyne/v2` and `github.com/gofrs/flock` (plus their transitive deps). Anything else needs a decision-log entry explaining why the standard library is insufficient.
- **Respect the package boundaries** in the architecture doc. `snippets` and `store` never import Fyne.
- **No goroutines in app code** (D2). Never update widgets from another goroutine.
- **Never weaken data safety** to make a test pass: no truncating or deleting the primary before the rename, no starting empty over a damaged file, and no publishing an unsaved library to the UI.
- **All user-facing strings** go in `internal/ui/messages.go` and match the UI spec exactly.
- **When the docs are ambiguous or wrong**, choose the simplest option that keeps every acceptance criterion, record it as a new `D<n>` entry, and update the affected doc in the same commit. Do not stop to ask unless the choice would contradict the product spec.
- **When a step needs a human** (for example a clean-machine run or a paste into a browser), do not fake it. Add it to the "pending human checks" in the final report.
- Write idiomatic Go: `gofmt`, small functions, errors wrapped with `%w` and context, and comments only where intent is not obvious. Write no speculative abstractions.

## Commands

The one command to verify any change (after M0 creates it):

```bash
pwsh -NoProfile -File scripts/check.ps1
```

Use `-Quick` while iterating, and the full run before each milestone commit.

Other commands:

```bash
go run ./cmd/clipcheck
go run ./cmd/sniper
go run ./tools/genicon
pwsh -NoProfile -File scripts/package.ps1
```

To look at the UI without a display, run `SNIPER_SCREENS=1 go test ./internal/ui -run TestScreens`, then open the PNGs in `internal/ui/testdata/screens/`. Do this after every UI milestone.

Fyne needs CGO and gcc. Both are installed on this machine: Go 1.27.0, and MSYS2 UCRT64 gcc at `C:\msys64\ucrt64\bin` on the user PATH. If a shell started before the install can't find them, refresh PATH with `$env:Path = [Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [Environment]::GetEnvironmentVariable('Path','User')`. On another machine, M0 has the install commands. Ignore the old Go 1.23.6 copies in the Daemonbane projects' `.toolchain` folders; they are too old.

If desktop-control (computer-use) tools are available and the user approves, use them to run the human checks 2–5 from the testing doc on this machine. Without them, list those checks as pending.

## Git

The workspace is not a git repo yet, so M0 runs `git init`. Commit once per milestone with the messages given in the delivery plan. Never push. There is no remote.

## Done means

Every milestone exit check passes, `go test ./...` is green, `gofmt -l .` is empty, `dist/Sniper.exe` exists, the README has run, backup, and build instructions, the decision log has its measurements, and the final report follows the delivery plan's format.
