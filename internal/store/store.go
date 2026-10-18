package store

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/tinkthemaker/sniper/internal/snippets"
)

const (
	renameAttempts = 5
	renameBackoff  = 20 * time.Millisecond

	// DailyBackupsKept is how many dated backups the backups folder keeps.
	DailyBackupsKept = 7
	dailyLayout      = "2006-01-02"
)

// dailyName matches only the files the store itself writes; anything else in
// the backups folder is the user's and is never deleted.
var dailyName = regexp.MustCompile(`^snippets-(\d{4}-\d{2}-\d{2})\.json$`)

// DefaultDir returns os.UserConfigDir()/Sniper without creating it.
func DefaultDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find config directory: %w", err)
	}
	return filepath.Join(base, "Sniper"), nil
}

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

// Store reads and writes the library files in one directory.
type Store struct {
	dir      string
	fs       fileOps
	lastGood []byte // primary's bytes as last loaded or saved by this process
	now      func() time.Time
}

func New(dir string) *Store { return &Store{dir: dir, fs: osFileOps(), now: time.Now} }

func (s *Store) Dir() string { return s.dir }

func (s *Store) path(name string) string { return filepath.Join(s.dir, name) }

// Load reads the primary file. It never writes or removes anything.
func (s *Store) Load() LoadResult {
	raw, err := s.readLimited(FileName)
	if errors.Is(err, fs.ErrNotExist) {
		return s.loadMissingPrimary()
	}
	if err != nil {
		return s.damaged(fmt.Errorf("read %s: %w", FileName, err))
	}
	lib, err := Decode(bytes.NewReader(raw))
	if err != nil {
		return s.damaged(fmt.Errorf("%s is not valid: %w", FileName, err))
	}
	s.lastGood = raw
	return LoadResult{State: StateOK, Library: lib}
}

func (s *Store) loadMissingPrimary() LoadResult {
	if _, err := s.fs.stat(s.path(BackupName)); errors.Is(err, fs.ErrNotExist) {
		return LoadResult{State: StateNew, Library: snippets.NewLibrary()}
	}
	return s.damaged(fmt.Errorf("%s is missing", FileName))
}

func (s *Store) damaged(err error) LoadResult {
	_, backupErr := s.readBackup()
	return LoadResult{State: StateDamaged, Err: err, BackupAvailable: backupErr == nil}
}

// readLimited reads a file in the data directory, rejecting oversized files unread.
func (s *Store) readLimited(name string) ([]byte, error) {
	p := s.path(name)
	info, err := s.fs.stat(p)
	if err != nil {
		return nil, err
	}
	if info.Size() > MaxFileBytes {
		return nil, ErrTooLarge
	}
	return s.fs.readFile(p)
}

// readBackup returns the backup's bytes if it exists and validates.
func (s *Store) readBackup() ([]byte, error) {
	raw, err := s.readLimited(BackupName)
	if err != nil {
		return nil, err
	}
	if _, err := Decode(bytes.NewReader(raw)); err != nil {
		return nil, err
	}
	return raw, nil
}

// Save validates lib, backs up the last good primary, then atomically replaces
// the primary. A failure leaves the primary and the in-memory state untouched.
func (s *Store) Save(lib snippets.Library) error {
	if err := snippets.ValidateLibrary(lib); err != nil {
		return err
	}
	data, err := Encode(lib)
	if err != nil {
		return err
	}
	if len(data) > MaxFileBytes {
		return fmt.Errorf("save library: %w", ErrTooLarge)
	}
	if s.lastGood != nil {
		if err := s.writeDailyBackup(); err != nil {
			return fmt.Errorf("write daily backup: %w", err)
		}
		if err := s.writeAtomic(BackupName, s.lastGood); err != nil {
			return fmt.Errorf("write backup: %w", err)
		}
	}
	if err := s.writeAtomic(FileName, data); err != nil {
		return fmt.Errorf("write %s: %w", FileName, err)
	}
	s.lastGood = data
	s.pruneDailyBackups()
	return nil
}

// writeDailyBackup writes lastGood to backups/snippets-YYYY-MM-DD.json for
// today (local time) unless that file already exists, so each dated file holds
// the library as it was when that day's first save began.
func (s *Store) writeDailyBackup() error {
	dir := filepath.Join(s.dir, BackupDir)
	name := "snippets-" + s.now().Format(dailyLayout) + ".json"
	_, err := s.fs.stat(filepath.Join(dir, name))
	if err == nil {
		return nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := s.fs.mkdirAll(dir, 0o700); err != nil {
		return err
	}
	return s.writeAtomicIn(dir, name, s.lastGood)
}

// pruneDailyBackups keeps the newest DailyBackupsKept dated files. Failures are
// ignored: an old backup left behind must never fail a save.
func (s *Store) pruneDailyBackups() {
	dir := filepath.Join(s.dir, BackupDir)
	entries, err := s.fs.readDir(dir)
	if err != nil {
		return
	}
	var names []string
	for _, e := range entries {
		m := dailyName.FindStringSubmatch(e.Name())
		if m == nil || e.IsDir() {
			continue
		}
		if _, err := time.Parse(dailyLayout, m[1]); err == nil {
			names = append(names, e.Name())
		}
	}
	if len(names) <= DailyBackupsKept {
		return
	}
	sort.Strings(names) // the date sorts lexically
	for _, n := range names[:len(names)-DailyBackupsKept] {
		_ = s.fs.remove(filepath.Join(dir, n))
	}
}

// RestoreBackup validates the backup and atomically copies it over the primary.
func (s *Store) RestoreBackup() (snippets.Library, error) {
	raw, err := s.readBackup()
	if err != nil {
		return snippets.Library{}, fmt.Errorf("backup cannot be restored: %w", err)
	}
	lib, err := Decode(bytes.NewReader(raw))
	if err != nil {
		return snippets.Library{}, fmt.Errorf("backup cannot be restored: %w", err)
	}
	if err := s.writeAtomic(FileName, raw); err != nil {
		return snippets.Library{}, fmt.Errorf("restore backup: %w", err)
	}
	s.lastGood = raw
	return lib, nil
}

// QuarantinePrimary renames the primary to snippets.damaged-<timestamp>.json and
// returns the new name. It returns "" if there is no primary. It never overwrites
// an existing file.
func (s *Store) QuarantinePrimary(now time.Time) (string, error) {
	if _, err := s.fs.stat(s.path(FileName)); errors.Is(err, fs.ErrNotExist) {
		s.lastGood = nil
		return "", nil
	}
	stamp := now.Format("20060102-150405")
	name := fmt.Sprintf("snippets.damaged-%s.json", stamp)
	for n := 2; ; n++ {
		if _, err := s.fs.stat(s.path(name)); errors.Is(err, fs.ErrNotExist) {
			break
		}
		name = fmt.Sprintf("snippets.damaged-%s-%d.json", stamp, n)
	}
	if err := s.renameRetry(s.path(FileName), s.path(name)); err != nil {
		return "", fmt.Errorf("move %s aside: %w", FileName, err)
	}
	s.lastGood = nil
	return name, nil
}

// writeAtomic writes data to a temp file in the same directory, then renames it
// over name. The destination is never deleted or truncated first.
func (s *Store) writeAtomic(name string, data []byte) error {
	return s.writeAtomicIn(s.dir, name, data)
}

func (s *Store) writeAtomicIn(dir, name string, data []byte) (err error) {
	f, err := s.fs.createTemp(dir, ".sniper-*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer func() {
		if err != nil {
			_ = s.fs.remove(tmp)
		}
	}()
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return s.renameRetry(tmp, filepath.Join(dir, name))
}

// renameRetry retries only the transient Windows sharing errors.
func (s *Store) renameRetry(from, to string) error {
	var err error
	for attempt := 1; attempt <= renameAttempts; attempt++ {
		if err = s.fs.rename(from, to); err == nil || !retryableRename(err) {
			return err
		}
		if attempt < renameAttempts {
			s.fs.sleep(renameBackoff * time.Duration(attempt))
		}
	}
	return err
}
