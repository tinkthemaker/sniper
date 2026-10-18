package store

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

var day1 = time.Date(2026, 10, 1, 9, 0, 0, 0, time.Local)

// onDay points the store's clock at the given day offset from day1.
func onDay(s *Store, n int) {
	s.now = func() time.Time { return day1.AddDate(0, 0, n).Add(3 * time.Hour) }
}

func dated(n int) string {
	return filepath.Join(BackupDir, "snippets-"+day1.AddDate(0, 0, n).Format("2006-01-02")+".json")
}

func datedFiles(t *testing.T, dir string) []string {
	t.Helper()
	m, err := filepath.Glob(filepath.Join(dir, BackupDir, "snippets-*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for i := range m {
		m[i] = filepath.Base(m[i])
	}
	sort.Strings(m)
	return m
}

func TestDailyBackupFirstSaveOfDay(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	onDay(s, 0)
	a, b, c, d := libOf("A"), libOf("B"), libOf("C"), libOf("D")
	if err := s.Save(a); err != nil {
		t.Fatal(err)
	}
	if exists(dir, BackupDir) {
		t.Error("first save on a new install must not create backups/")
	}
	if err := s.Save(b); err != nil {
		t.Fatal(err)
	}
	if read(t, dir, dated(0)) != mustEncode(t, a) {
		t.Error("day 1 file should hold the pre-save library A")
	}
	if err := s.Save(c); err != nil {
		t.Fatal(err)
	}
	if read(t, dir, dated(0)) != mustEncode(t, a) {
		t.Error("a second save the same day must not touch the day's file")
	}
	onDay(s, 1)
	if err := s.Save(d); err != nil {
		t.Fatal(err)
	}
	if read(t, dir, dated(0)) != mustEncode(t, a) || read(t, dir, dated(1)) != mustEncode(t, c) {
		t.Error("next day should add a file holding C and leave day 1 alone")
	}
	if read(t, dir, BackupName) != mustEncode(t, c) {
		t.Error("snippets.bak.json behavior must be unchanged")
	}
	if got := tmpFiles(t, filepath.Join(dir, BackupDir)); len(got) != 0 {
		t.Errorf("temp files left: %v", got)
	}
}

func TestDailyBackupAfterLoadUsesLoadedBytes(t *testing.T) {
	dir := t.TempDir()
	old := mustEncode(t, libOf("Loaded"))
	write(t, dir, FileName, old)
	s := New(dir)
	onDay(s, 0)
	if res := s.Load(); res.State != StateOK {
		t.Fatal(res.Err)
	}
	if err := s.Save(libOf("Next")); err != nil {
		t.Fatal(err)
	}
	if read(t, dir, dated(0)) != old {
		t.Error("dated backup should hold the primary as loaded")
	}
}

func TestDailyBackupPrunesToSeven(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	if err := os.MkdirAll(filepath.Join(dir, BackupDir, "snippets-2020-01-01.json.d"), 0o700); err != nil {
		t.Fatal(err)
	}
	keep := []string{"notes.txt", "snippets-latest.json", "snippets-2020-01-01.json.bak", "snippets-2026-13-45.json"}
	for _, k := range keep {
		write(t, filepath.Join(dir, BackupDir), k, "mine")
	}
	for n := 0; n < 9; n++ { // day 0's save has no lastGood, so days 1..8 make eight files
		onDay(s, n)
		if err := s.Save(libOf(fmt.Sprint("v", n))); err != nil {
			t.Fatal(err)
		}
	}
	var matching []string
	for _, f := range datedFiles(t, dir) {
		if dailyName.MatchString(f) && f != "snippets-2026-13-45.json" {
			matching = append(matching, f)
		}
	}
	if len(matching) != DailyBackupsKept {
		t.Fatalf("want %d dated files, got %v", DailyBackupsKept, matching)
	}
	if last := "snippets-" + day1.AddDate(0, 0, 8).Format("2006-01-02") + ".json"; matching[len(matching)-1] != last {
		t.Errorf("newest = %s, want %s", matching[len(matching)-1], last)
	}
	if exists(dir, dated(1)) {
		t.Error("oldest dated file should have been pruned")
	}
	for _, k := range keep {
		if read(t, filepath.Join(dir, BackupDir), k) != "mine" {
			t.Errorf("%s should survive pruning", k)
		}
	}
	if fi, err := os.Stat(filepath.Join(dir, BackupDir, "snippets-2020-01-01.json.d")); err != nil || !fi.IsDir() {
		t.Error("foreign directory should survive pruning")
	}
}

func TestDailyBackupPrunesOldDatedDirNamedLikeFile(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	// A directory that matches the pattern is not a backup and is never removed.
	if err := os.MkdirAll(filepath.Join(dir, BackupDir, "snippets-2019-01-01.json"), 0o700); err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 10; n++ {
		onDay(s, n)
		if err := s.Save(libOf(fmt.Sprint("v", n))); err != nil {
			t.Fatal(err)
		}
	}
	if fi, err := os.Stat(filepath.Join(dir, BackupDir, "snippets-2019-01-01.json")); err != nil || !fi.IsDir() {
		t.Error("matching directory should survive")
	}
}

func TestDailyBackupWriteFailureFailsSave(t *testing.T) {
	cases := map[string]func(fo *fileOps){
		"mkdir": func(fo *fileOps) {
			fo.mkdirAll = func(string, os.FileMode) error { return errors.New("injected mkdir failure") }
		},
		"write": func(fo *fileOps) {
			orig := fo.createTemp
			fo.createTemp = func(d, p string) (tempFile, error) {
				f, err := orig(d, p)
				if strings.HasSuffix(d, BackupDir) {
					return failingFile{tempFile: f, failWrite: true}, err
				}
				return f, err
			}
		},
		"rename": func(fo *fileOps) {
			orig := fo.rename
			fo.rename = func(a, b string) error {
				if strings.Contains(b, BackupDir) {
					return errors.New("injected rename failure")
				}
				return orig(a, b)
			}
		},
	}
	for name, hook := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			s := New(dir)
			onDay(s, 0)
			if err := s.Save(libOf("A")); err != nil {
				t.Fatal(err)
			}
			if err := s.Save(libOf("B")); err != nil {
				t.Fatal(err)
			}
			onDay(s, 1) // a new day, so a dated file is due
			primary, bak := read(t, dir, FileName), read(t, dir, BackupName)
			lastGood := append([]byte(nil), s.lastGood...)

			real := s.fs
			hook(&s.fs)
			if err := s.Save(libOf("C")); err == nil {
				t.Fatal("expected save to fail")
			}
			s.fs = real

			if read(t, dir, FileName) != primary || read(t, dir, BackupName) != bak {
				t.Error("primary and .bak must be unchanged by the failed save")
			}
			if !bytes.Equal(s.lastGood, lastGood) {
				t.Error("lastGood changed by failed save")
			}
			if exists(dir, dated(1)) {
				t.Error("failed write left a dated file")
			}
			if got := tmpFiles(t, filepath.Join(dir, BackupDir)); len(got) != 0 {
				t.Errorf("temp files left: %v", got)
			}
			if err := s.Save(libOf("D")); err != nil {
				t.Fatal(err)
			}
			if read(t, dir, dated(1)) != primary {
				t.Error("retry should write the dated backup from the last good primary")
			}
		})
	}
}

func TestDailyBackupPruneFailureDoesNotFailSave(t *testing.T) {
	for name, hook := range map[string]func(fo *fileOps){
		"remove": func(fo *fileOps) { fo.remove = func(string) error { return errors.New("injected remove failure") } },
		"readDir": func(fo *fileOps) {
			fo.readDir = func(string) ([]os.DirEntry, error) { return nil, errors.New("injected") }
		},
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			s := New(dir)
			for n := 0; n < 9; n++ {
				onDay(s, n)
				if err := s.Save(libOf(fmt.Sprint("v", n))); err != nil {
					t.Fatal(err)
				}
			}
			real := s.fs
			hook(&s.fs)
			onDay(s, 9)
			err := s.Save(libOf("last"))
			s.fs = real
			if err != nil {
				t.Fatalf("prune failure failed the save: %v", err)
			}
			if res := New(dir).Load(); res.State != StateOK || res.Library.Snippets[0].Title != "last" {
				t.Errorf("reload: %+v", res)
			}
		})
	}
}

func TestDailyBackupNewInstallWritesNone(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	if err := s.Save(libOf("A")); err != nil {
		t.Fatal(err)
	}
	if exists(dir, BackupDir) {
		t.Error("no dated backup on a new install")
	}
}
