# Architecture

## Stack

| Item | Choice |
|---|---|
| Language | Go. Use the newest stable release installed. The `go` directive in `go.mod` is the installed minor version, and must be at least `1.24`. |
| Module path | `github.com/tinkthemaker/sniper` |
| UI | `fyne.io/fyne/v2` pinned to `v2.8.1` (or a newer v2.8.x patch) |
| Lock | `github.com/gofrs/flock` (latest v0.x release at implementation time, pinned in `go.mod`) |
| Packaging tool | `fyne.io/tools/cmd/fyne` (dev tool, not a module dependency) |
| Concurrency | **None in the app.** All file I/O runs synchronously inside UI callbacks. Files are at most 16 MiB, so this is fine. No goroutines means no `fyne.Do`. |

No other runtime dependencies. `golang.org/x/sys` may come in indirectly.

## Repository layout

```text
go.mod / go.sum
CLAUDE.md                     agent operating instructions
README.md                     user-facing overview + run/backup instructions
docs/                         these documents
cmd/sniper/
  main.go                     startup wiring (see "Startup sequence")
  FyneApp.toml                packaging metadata
  Icon.png                    512×512 app icon (generated, committed)
cmd/clipcheck/main.go         Windows clipboard round-trip check (dev tool)
tools/genicon/main.go         generates cmd/sniper/Icon.png from code
internal/snippets/            model, normalization, validation, pure search, pure mutations
internal/store/               JSON encode/decode, load/save/backup/restore, import/export
internal/clipboard/           wrapper around fyne.Clipboard + platform line endings
internal/instance/            single-instance lock
internal/ui/                  windows, widgets, dialogs, shortcut routing, messages
dist/                         packaged output (git-ignored)
```

Dependency rules, which `go vet` will not catch, so review them by hand:

- `internal/snippets` imports only the standard library.
- `internal/store` imports `snippets` and the standard library. It never imports Fyne.
- `internal/instance` imports `gofrs/flock` and the standard library.
- `internal/clipboard` imports Fyne core (`fyne.io/fyne/v2`) only.
- `internal/ui` may import everything above. Nothing imports `ui` except `cmd/sniper`.

## Package APIs

The signatures below are the contract. Unexported helpers are free-form.

### `internal/snippets`

```go
const (
    SchemaVersion = 1
    MaxTitleRunes = 200
    MaxBodyBytes  = 65536
    MaxSnippets   = 10000
)

type Snippet struct {
    ID    string `json:"id"`
    Title string `json:"title"`
    Body  string `json:"body"`
}

type Library struct {
    Version  int       `json:"version"`
    Snippets []Snippet `json:"snippets"`
}

// ValidationError describes one rule violation. Field is "title", "body", "id" or "library".
type ValidationError struct {
    Field   string
    Message string // user-facing sentence, e.g. "Title is required."
}
func (e *ValidationError) Error() string

var ErrNotFound = errors.New("snippet not found")

func NewLibrary() Library                              // Version 1, empty non-nil slice
func NewID() (string, error)                           // 16 crypto/rand bytes, lowercase hex
func Normalize(title, body string) (string, string)    // rules in product spec
func ValidateFields(title, body string) error          // expects normalized input; *ValidationError
func ValidateLibrary(lib Library) error                // version, count, each record, non-empty unique IDs

// Pure mutations. They never modify lib's backing array: they return a new
// Library with a freshly allocated slice. Inputs are normalized and validated inside.
func Add(lib Library, title, body string) (Library, Snippet, error)
func Update(lib Library, id, title, body string) (Library, Snippet, error) // ErrNotFound
func Delete(lib Library, id string) (Library, error)                       // ErrNotFound
func Append(lib Library, items []Snippet) (Library, error)                 // fresh IDs, limit check

func Search(items []Snippet, query string) []Snippet // new slice; rules in product spec
```

### `internal/store`

```go
const (
    FileName     = "snippets.json"
    BackupName   = "snippets.bak.json"
    LockName     = "sniper.lock"
    MaxFileBytes = 16 << 20
)

var (
    ErrTooLarge           = errors.New("file is larger than 16 MiB")
    ErrUnsupportedVersion = errors.New("file was created by an unsupported version of Sniper")
)

func DefaultDir() (string, error) // os.UserConfigDir()/Sniper; does not create it

type LoadState int
const (
    StateOK      LoadState = iota // Library is valid
    StateNew                      // no primary, no backup: empty library
    StateDamaged                  // primary missing-with-backup, unreadable, or invalid
)

type LoadResult struct {
    State           LoadState
    Library         snippets.Library // valid when State is OK or New
    Err             error            // why the primary is damaged (StateDamaged)
    BackupAvailable bool             // backup exists AND validates
}

type Store struct { /* dir string; fs fileOps; lastGood []byte */ }

func New(dir string) *Store
func (s *Store) Dir() string
func (s *Store) Load() LoadResult
func (s *Store) Save(lib snippets.Library) error            // validate → backup → atomic replace
func (s *Store) RestoreBackup() (snippets.Library, error)   // validate backup → atomic copy to primary
func (s *Store) QuarantinePrimary(now time.Time) (string, error) // rename primary aside; returns new name

// Codec, shared by Load, RestoreBackup, Import and Export.
func Encode(lib snippets.Library) ([]byte, error)               // json.Encoder, SetIndent("", "  "), SetEscapeHTML(false) so <, >, & stay readable; Encoder adds the trailing "\n"
func Decode(r io.Reader) (snippets.Library, error)              // strict; validates library
func DecodeImport(r io.Reader) ([]snippets.Snippet, error)      // strict shape; IDs ignored; records normalized+validated
func Export(w io.Writer, lib snippets.Library) error
```

Decoding rules, shared by `Decode` and `DecodeImport`:

1. Read through `io.LimitReader(r, MaxFileBytes+1)`. More than `MaxFileBytes` bytes returns `ErrTooLarge`.
2. Strip a leading UTF-8 BOM (`EF BB BF`).
3. Lenient pre-pass: unmarshal into `struct{ Version *int }`. A missing version, or a version other than 1, returns `ErrUnsupportedVersion` (wrapped with the number found, if any).
4. Strict pass: `json.Decoder` with `DisallowUnknownFields()`. After the value, a second `Decode` must return `io.EOF`, because trailing data is invalid.
5. A `null` `snippets` array becomes empty.
6. `Decode` calls `snippets.ValidateLibrary`. Stored data is **not** normalized on load: a record that fails validation makes the file damaged. `DecodeImport` normalizes each record and validates title and body. Errors are wrapped as `record N: <message>`, with N 1-based.

`fileOps` is an unexported struct of function fields: `readFile`, `createTemp`, `rename`, `remove`, `stat`. `createTemp` returns an unexported `tempFile` interface (`Name() string`, `Write`, `Sync`, `Close`), which `*os.File` satisfies, so tests can fail individual writes. `New` fills them with `os` functions. Tests in package `store` replace them to inject failures. Do not add a public filesystem interface.

### `internal/clipboard`

```go
type Writer interface{ SetText(text string) }

func New(cb fyne.Clipboard) Writer   // wraps app.Clipboard()
func ToPlatform(text string) string  // windows: "\n" → "\r\n" (after collapsing existing "\r\n"); others: identity
```

`ToPlatform` lives in `lineendings_windows.go` and `lineendings_other.go` with build tags, and each has a table test.

### `internal/instance`

```go
var ErrAlreadyRunning = errors.New("another Sniper instance is running")

type Lock struct { /* *flock.Flock */ }
func Acquire(dir string) (*Lock, error) // TryLock on dir/sniper.lock; ErrAlreadyRunning if held
func (l *Lock) Release() error
```

### `internal/ui`

```go
// Storage is the subset of *store.Store the UI uses. It is an interface only so
// UI tests can wrap a real store and inject Save failures.
type Storage interface {
    Dir() string
    Save(lib snippets.Library) error
    RestoreBackup() (snippets.Library, error)
    QuarantinePrimary(now time.Time) (string, error)
}

type Deps struct {
    App       fyne.App
    Store     Storage
    Clipboard clipboard.Writer
    OpenFolder func(path string) error // injected so tests don't launch Explorer
    Now       func() time.Time         // injected for export filenames / quarantine names
}

func ShowMain(d Deps, lib snippets.Library) *MainWindow
func ShowRecovery(d Deps, res store.LoadResult)          // on success calls ShowMain and closes itself
func ShowAlreadyRunning(a fyne.App)
func ShowFatal(a fyne.App, title, detail string)

// MainWindow exposes just enough for headless tests.
type MainWindow struct { /* … */ }
func (m *MainWindow) Window() fyne.Window
func (m *MainWindow) Library() snippets.Library
func (m *MainWindow) Results() []snippets.Snippet
func (m *MainWindow) SelectedID() string
func (m *MainWindow) Status() string
```

Suggested files: `main_window.go`, `search_entry.go`, `editor.go`, `recovery.go`, `dialogs.go`, `shortcuts.go`, `messages.go` (all user-facing strings as constants or format funcs), `openfolder_windows.go`, and `openfolder_other.go`.

`OpenFolder` default is `exec.Command("explorer", path).Start()` on Windows (ignore Explorer's exit code). On Linux it is `xdg-open`; on macOS it is `open`.

## State model

`MainWindow` owns a single `lib snippets.Library`, plus `query string`, `results []Snippet`, and `selectedID string`. Every mutation follows the same pattern:

```go
next, item, err := snippets.Add(m.lib, title, body) // pure, returns a copy
if err != nil { showInline(err); return }
if err := m.deps.Store.Save(next); err != nil { showInline(saveFailed(err)); return } // m.lib untouched
m.lib = next
m.refresh(selectID(item.ID))
```

`refresh` recomputes `results = snippets.Search(m.lib.Snippets, m.query)`, then reselects by ID following the rules in the UI spec, then refreshes the list and preview widgets.

## Persistence algorithm

`Save(lib)`:

1. `snippets.ValidateLibrary(lib)`. Then `data := Encode(lib)`, and reject it if it exceeds `MaxFileBytes`.
2. **Backup:** if `s.lastGood != nil` (bytes of the primary as last loaded or saved by this process), `writeAtomic(BackupName, s.lastGood)`. If this fails, return the error and do **not** touch the primary.
3. **Primary:** `writeAtomic(FileName, data)`.
4. On success, set `s.lastGood = data`.

`writeAtomic(name, data)`:

1. `f := createTemp(dir, ".sniper-*.tmp")`. The temp file must be in the same directory, so the rename stays on one volume.
2. `f.Write(data)`, then `f.Sync()`, then `f.Close()`. On any error, `remove(tmp)` and return the error.
3. `rename(tmp, dir/name)`. On Windows, retry up to **5 attempts** with a sleep of 20 ms × attempt between them, but only when the error is `ERROR_ACCESS_DENIED` (5) or `ERROR_SHARING_VIOLATION` (32). Check with `errors.As` against `syscall.Errno`, in a `_windows.go` helper. After final failure, `remove(tmp)` and return the error.
4. Never delete or truncate the destination before the rename.

`Load()`:

1. Read the primary. If it does not exist, check the backup: no backup gives `StateNew`; a valid backup gives `StateDamaged` with `Err` set to "snippets.json is missing" and `BackupAvailable=true`. If the backup exists but is invalid, also use `StateDamaged`, so it never gets overwritten silently.
2. `Decode` the primary. On error, return `StateDamaged` with `Err`, and set `BackupAvailable` by decoding the backup.
3. On success, set `s.lastGood` to the raw bytes and return `StateOK`.

`RestoreBackup()` decodes the backup, then `writeAtomic(FileName, backupBytes)`, then sets `lastGood`. `QuarantinePrimary(now)` renames the primary to, for example, `snippets.damaged-20261001-153000.json` (Go layout `20060102-150405`), using the same retry policy, and sets `lastGood = nil`.

## Startup sequence (`cmd/sniper/main.go`)

```text
a := app.NewWithID("app.sniper.desktop"); a.SetIcon(embedded Icon.png via //go:embed)
dir := store.DefaultDir(); os.MkdirAll(dir, 0o700)    → on error ShowFatal
lock := instance.Acquire(dir)                          → ErrAlreadyRunning: ShowAlreadyRunning; other: ShowFatal
defer lock.Release()
st := store.New(dir); res := st.Load()
deps := ui.Deps{App: a, Store: st, Clipboard: clipboard.New(a.Clipboard()), OpenFolder: ui.DefaultOpenFolder, Now: time.Now}
StateOK / StateNew → ui.ShowMain(deps, res.Library)
StateDamaged       → ui.ShowRecovery(deps, res)
a.Run()
```

Use the version string from `a.Metadata().Version` in About. When built with plain `go build` it is empty, so show `dev`.

## Distribution

`cmd/sniper/FyneApp.toml`:

```toml
[Details]
Icon = "Icon.png"
Name = "Sniper"
ID = "app.sniper.desktop"
Version = "0.1.0"
Build = 1
```

Package from `cmd/sniper` with `fyne package -os windows -release`, then move `Sniper.exe` to `dist/`. The executable embeds the icon and metadata. Personal data stays in `%AppData%\Sniper`. Installer, auto-updater, tray, global shortcut, and code signing are out of scope.
