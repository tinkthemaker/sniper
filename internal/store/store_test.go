package store

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sniper/internal/snippets"
)

var corpus = []string{
	"plain ASCII",
	"line one\nline two\n\nline four",
	"tab\tseparated\tvalues",
	"trailing spaces   \n  leading spaces",
	"emoji \U0001F44B\U0001F3FD family \U0001F468\u200d\U0001F469\u200d\U0001F467 flag \U0001F1EF\U0001F1F5",
	"punctuation “quotes” ‘single’ — en–dash … © ® ™ € £ ¥",
	"日本語のテキスト / 中文文本 / 한국어 텍스트",
	"العربية مع English mixed — עברית",
	"Ελληνικά, Кириллица, हिन्दी, ไทย",
	"combining: e\u0301 n\u0303 (decomposed)",
	"zero width:\u200bjoiner\u200d end",
	`json-ish: {"a": "b\\n"} <tag attr='x'> & %s %d`,
}

func libOf(titles ...string) snippets.Library {
	lib := snippets.NewLibrary()
	for i, t := range titles {
		lib.Snippets = append(lib.Snippets, snippets.Snippet{ID: fmt.Sprintf("id%02d", i), Title: t, Body: "body of " + t})
	}
	return lib
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, dir, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func exists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func tmpFiles(t *testing.T, dir string) []string {
	t.Helper()
	m, err := filepath.Glob(filepath.Join(dir, "*.tmp"))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func mustEncode(t *testing.T, lib snippets.Library) string {
	t.Helper()
	b, err := Encode(lib)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	lib := snippets.NewLibrary()
	for i, s := range corpus {
		lib.Snippets = append(lib.Snippets, snippets.Snippet{
			ID: fmt.Sprintf("%032x", i), Title: strings.ReplaceAll(s, "\n", " "), Body: s,
		})
	}
	data := mustEncode(t, lib)
	if again := mustEncode(t, lib); again != data {
		t.Error("encoding is not stable")
	}
	if !strings.HasSuffix(data, "}\n") || strings.HasSuffix(data, "\n\n") {
		t.Errorf("want exactly one trailing newline, got tail %q", data[len(data)-4:])
	}
	if !strings.Contains(data, "\n  \"snippets\": [") {
		t.Errorf("not indented by 2 spaces:\n%s", data[:80])
	}
	if !strings.Contains(data, "<tag attr='x'> & %s") {
		t.Error("HTML characters were escaped")
	}
	got, err := Decode(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Snippets) != len(lib.Snippets) {
		t.Fatalf("got %d snippets", len(got.Snippets))
	}
	for i := range lib.Snippets {
		if got.Snippets[i] != lib.Snippets[i] {
			t.Errorf("snippet %d: got %+v want %+v", i, got.Snippets[i], lib.Snippets[i])
		}
	}
}

func TestEncodeNilSnippets(t *testing.T) {
	data := mustEncode(t, snippets.Library{Version: 1})
	if !strings.Contains(data, `"snippets": []`) {
		t.Errorf("nil slice should encode as []: %s", data)
	}
}

func TestDecodeRejects(t *testing.T) {
	rec := `{"id":"a","title":"T","body":"B"}`
	tests := []struct {
		name string
		in   string
		is   error
	}{
		{"malformed", `{"version":1,`, nil},
		{"trailing garbage", `{"version":1,"snippets":[]} xx`, nil},
		{"second object", `{"version":1,"snippets":[]}{"version":1}`, nil},
		{"unknown top-level field", `{"version":1,"snippets":[],"extra":1}`, nil},
		{"unknown snippet field", `{"version":1,"snippets":[{"id":"a","title":"T","body":"B","tags":[]}]}`, nil},
		{"missing version", `{"snippets":[]}`, ErrUnsupportedVersion},
		{"version 2", `{"version":2,"snippets":[]}`, ErrUnsupportedVersion},
		{"version wrong type", `{"version":"1","snippets":[]}`, nil},
		{"too large", `{"version":1,"snippets":[],"pad":"` + strings.Repeat("x", MaxFileBytes) + `"}`, ErrTooLarge},
		{"invalid record", `{"version":1,"snippets":[` + rec + `,{"id":"b","title":"","body":"B"}]}`, nil},
		{"duplicate id", `{"version":1,"snippets":[` + rec + `,` + rec + `]}`, nil},
		{"empty", ``, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decode(strings.NewReader(tt.in))
			if err == nil {
				t.Fatal("expected an error")
			}
			if tt.is != nil && !errors.Is(err, tt.is) {
				t.Errorf("error %v is not %v", err, tt.is)
			}
		})
	}
	if _, err := Decode(strings.NewReader(`{"version":2,"snippets":[]}`)); !strings.Contains(err.Error(), "version 2") {
		t.Errorf("version error should name the version: %v", err)
	}
}

func TestDecodeAcceptsBOMAndNull(t *testing.T) {
	lib, err := Decode(strings.NewReader("\xEF\xBB\xBF" + `{"version":1,"snippets":null}`))
	if err != nil {
		t.Fatal(err)
	}
	if lib.Snippets == nil || len(lib.Snippets) != 0 {
		t.Errorf("null snippets should become an empty slice: %#v", lib.Snippets)
	}
}

func TestDecodeImport(t *testing.T) {
	in := `{"version":1,"snippets":[
		{"title":"No ID","body":"a\r\nb"},
		{"id":"dup","title":"Dup1\r\nline","body":"x"},
		{"id":"dup","title":"Dup2","body":"y"}]}`
	items, err := DecodeImport(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 || items[0].Body != "a\nb" || items[1].Title != "Dup1  line" {
		t.Errorf("items = %+v", items)
	}
	for _, it := range items {
		if it.ID != "" {
			t.Errorf("IDs must be ignored, got %q", it.ID)
		}
	}

	bad := `{"version":1,"snippets":[
		{"title":"a","body":"b"},{"title":"c","body":"d"},{"title":" ","body":"e"}]}`
	if _, err := DecodeImport(strings.NewReader(bad)); err == nil || !strings.Contains(err.Error(), "record 3:") {
		t.Errorf("want error naming record 3, got %v", err)
	}
	if _, err := DecodeImport(strings.NewReader(`{"version":3,"snippets":[]}`)); !errors.Is(err, ErrUnsupportedVersion) {
		t.Errorf("got %v", err)
	}
}

func TestLoadStates(t *testing.T) {
	good := mustEncode(t, libOf("A"))
	goodBak := mustEncode(t, libOf("Old"))
	tests := []struct {
		name            string
		primary, backup *string
		state           LoadState
		backupAvailable bool
	}{
		{"no files", nil, nil, StateNew, false},
		{"valid primary", &good, nil, StateOK, false},
		{"valid primary and backup", &good, &goodBak, StateOK, false},
		{"corrupt primary no backup", ptr("xx"), nil, StateDamaged, false},
		{"corrupt primary valid backup", ptr("xx"), &goodBak, StateDamaged, true},
		{"missing primary valid backup", nil, &goodBak, StateDamaged, true},
		{"missing primary corrupt backup", nil, ptr("xx"), StateDamaged, false},
		{"corrupt both", ptr("xx"), ptr("yy"), StateDamaged, false},
		{"wrong version primary", ptr(`{"version":9,"snippets":[]}`), &goodBak, StateDamaged, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.primary != nil {
				write(t, dir, FileName, *tt.primary)
			}
			if tt.backup != nil {
				write(t, dir, BackupName, *tt.backup)
			}
			res := New(dir).Load()
			if res.State != tt.state || res.BackupAvailable != tt.backupAvailable {
				t.Fatalf("state=%v backup=%v err=%v; want state=%v backup=%v",
					res.State, res.BackupAvailable, res.Err, tt.state, tt.backupAvailable)
			}
			if tt.state == StateDamaged && res.Err == nil {
				t.Error("damaged result must carry an error")
			}
			if tt.state != StateDamaged && res.Library.Version != 1 {
				t.Errorf("library not usable: %+v", res.Library)
			}
			// Loading must never change what is on disk.
			if tt.primary != nil && read(t, dir, FileName) != *tt.primary {
				t.Error("primary changed by Load")
			}
			if tt.primary == nil && exists(dir, FileName) {
				t.Error("Load created a primary")
			}
			if tt.backup != nil && read(t, dir, BackupName) != *tt.backup {
				t.Error("backup changed by Load")
			}
		})
	}
}

func ptr(s string) *string { return &s }

func TestLoadTooLarge(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, FileName, strings.Repeat("x", MaxFileBytes+1))
	res := New(dir).Load()
	if res.State != StateDamaged || !errors.Is(res.Err, ErrTooLarge) {
		t.Errorf("got state=%v err=%v", res.State, res.Err)
	}
}

func TestSaveCreatesBackup(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	a, b, c := libOf("A"), libOf("B"), libOf("C")
	if err := s.Save(a); err != nil {
		t.Fatal(err)
	}
	if exists(dir, BackupName) {
		t.Error("first save on a new store must not write a backup")
	}
	if err := s.Save(b); err != nil {
		t.Fatal(err)
	}
	if read(t, dir, FileName) != mustEncode(t, b) || read(t, dir, BackupName) != mustEncode(t, a) {
		t.Error("after second save: primary should be B, backup A")
	}
	if err := s.Save(c); err != nil {
		t.Fatal(err)
	}
	if read(t, dir, FileName) != mustEncode(t, c) || read(t, dir, BackupName) != mustEncode(t, b) {
		t.Error("after third save: primary should be C, backup B")
	}
	if got := tmpFiles(t, dir); len(got) != 0 {
		t.Errorf("temp files left: %v", got)
	}
}

func TestSaveAfterLoadBacksUpLoadedBytes(t *testing.T) {
	dir := t.TempDir()
	old := mustEncode(t, libOf("Loaded"))
	write(t, dir, FileName, old)
	s := New(dir)
	if res := s.Load(); res.State != StateOK {
		t.Fatal(res.Err)
	}
	if err := s.Save(libOf("Next")); err != nil {
		t.Fatal(err)
	}
	if read(t, dir, BackupName) != old {
		t.Error("backup should hold the primary as loaded")
	}
}

func TestSaveRejectsInvalid(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	if err := s.Save(snippets.Library{Version: 7}); err == nil {
		t.Error("invalid library saved")
	}
	huge := snippets.NewLibrary()
	for i := 0; i < 300; i++ { // 300 × 64 KiB > 16 MiB
		huge.Snippets = append(huge.Snippets, snippets.Snippet{
			ID: fmt.Sprint(i), Title: "T", Body: strings.Repeat("a", snippets.MaxBodyBytes),
		})
	}
	if err := s.Save(huge); !errors.Is(err, ErrTooLarge) {
		t.Errorf("want ErrTooLarge, got %v", err)
	}
	if exists(dir, FileName) {
		t.Error("rejected save wrote a file")
	}
}

type failingFile struct {
	tempFile
	failWrite, failSync bool
}

func (f failingFile) Write(p []byte) (int, error) {
	if f.failWrite {
		return 0, errors.New("injected write failure")
	}
	return f.tempFile.Write(p)
}

func (f failingFile) Sync() error {
	if f.failSync {
		return errors.New("injected sync failure")
	}
	return f.tempFile.Sync()
}

func TestSaveFailurePreservesPrimary(t *testing.T) {
	// Each case sabotages a different step of the same third save, which writes
	// the backup first (call 1) and then the primary (call 2).
	type hook func(fo *fileOps, n *int)
	tests := []struct {
		name string
		hook hook
	}{
		{"createTemp", func(fo *fileOps, n *int) {
			fo.createTemp = func(string, string) (tempFile, error) { return nil, errors.New("injected createTemp failure") }
		}},
		{"write backup", func(fo *fileOps, n *int) {
			orig := fo.createTemp
			fo.createTemp = func(d, p string) (tempFile, error) {
				f, err := orig(d, p)
				return failingFile{tempFile: f, failWrite: true}, err
			}
		}},
		{"write primary", func(fo *fileOps, n *int) {
			orig := fo.createTemp
			fo.createTemp = func(d, p string) (tempFile, error) {
				f, err := orig(d, p)
				*n++
				return failingFile{tempFile: f, failWrite: *n == 2}, err
			}
		}},
		{"sync", func(fo *fileOps, n *int) {
			orig := fo.createTemp
			fo.createTemp = func(d, p string) (tempFile, error) {
				f, err := orig(d, p)
				return failingFile{tempFile: f, failSync: true}, err
			}
		}},
		{"rename backup", func(fo *fileOps, n *int) {
			fo.rename = func(string, string) error { return errors.New("injected rename failure") }
		}},
		{"rename primary", func(fo *fileOps, n *int) {
			orig := fo.rename
			fo.rename = func(a, b string) error {
				*n++
				if *n == 2 {
					return errors.New("injected rename failure")
				}
				return orig(a, b)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			s := New(dir)
			if err := s.Save(libOf("A")); err != nil {
				t.Fatal(err)
			}
			if err := s.Save(libOf("B")); err != nil {
				t.Fatal(err)
			}
			primaryBefore := read(t, dir, FileName)
			lastGoodBefore := append([]byte(nil), s.lastGood...)

			real := s.fs
			var n int
			tt.hook(&s.fs, &n)
			if err := s.Save(libOf("C")); err == nil {
				t.Fatal("expected save to fail")
			}
			s.fs = real

			if read(t, dir, FileName) != primaryBefore {
				t.Error("primary changed by failed save")
			}
			if !bytes.Equal(s.lastGood, lastGoodBefore) {
				t.Error("lastGood changed by failed save")
			}
			if got := tmpFiles(t, dir); len(got) != 0 {
				t.Errorf("temp files left: %v", got)
			}
			// A later successful save must still write the right backup.
			if err := s.Save(libOf("D")); err != nil {
				t.Fatal(err)
			}
			if read(t, dir, BackupName) != primaryBefore {
				t.Error("backup after recovery should be the last good primary")
			}
			if res := New(dir).Load(); res.State != StateOK || res.Library.Snippets[0].Title != "D" {
				t.Errorf("reload: %+v", res)
			}
		})
	}
}

func TestRetryNonRetryableFailsImmediately(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	calls := 0
	s.fs.rename = func(string, string) error { calls++; return errors.New("disk on fire") }
	if err := s.Save(libOf("A")); err == nil {
		t.Fatal("expected failure")
	}
	if calls != 1 {
		t.Errorf("rename called %d times, want 1", calls)
	}
}

func TestRestoreBackup(t *testing.T) {
	dir := t.TempDir()
	bak := mustEncode(t, libOf("Backed up"))
	write(t, dir, FileName, "xx")
	write(t, dir, BackupName, bak)
	s := New(dir)
	lib, err := s.RestoreBackup()
	if err != nil {
		t.Fatal(err)
	}
	if lib.Snippets[0].Title != "Backed up" {
		t.Errorf("lib = %+v", lib)
	}
	res := New(dir).Load()
	if res.State != StateOK || res.Library.Snippets[0].Title != "Backed up" {
		t.Errorf("Load after restore: %+v", res)
	}
	// The restored bytes are the new last-good, so the next save backs them up.
	if err := s.Save(libOf("Next")); err != nil {
		t.Fatal(err)
	}
	if read(t, dir, BackupName) != bak {
		t.Error("restored library should become the backup on next save")
	}
}

func TestRestoreBackupInvalid(t *testing.T) {
	for name, bak := range map[string]*string{"corrupt": ptr("not json"), "missing": nil} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			write(t, dir, FileName, "xx")
			if bak != nil {
				write(t, dir, BackupName, *bak)
			}
			if _, err := New(dir).RestoreBackup(); err == nil {
				t.Fatal("expected an error")
			}
			if read(t, dir, FileName) != "xx" {
				t.Error("primary was touched")
			}
		})
	}
}

func TestQuarantine(t *testing.T) {
	now := time.Date(2026, 10, 1, 15, 30, 0, 0, time.UTC)
	t.Run("without backup", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, FileName, "xx")
		s := New(dir)
		name, err := s.QuarantinePrimary(now)
		if err != nil {
			t.Fatal(err)
		}
		if name != "snippets.damaged-20261001-153000.json" {
			t.Errorf("name = %q", name)
		}
		if exists(dir, FileName) || read(t, dir, name) != "xx" {
			t.Error("primary should have moved aside intact")
		}
		if res := s.Load(); res.State != StateNew {
			t.Errorf("state = %v", res.State)
		}
	})
	t.Run("with backup", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, FileName, "xx")
		write(t, dir, BackupName, mustEncode(t, libOf("A")))
		s := New(dir)
		if _, err := s.QuarantinePrimary(now); err != nil {
			t.Fatal(err)
		}
		if res := s.Load(); res.State != StateDamaged || !res.BackupAvailable {
			t.Errorf("result = %+v", res)
		}
	})
	t.Run("never overwrites", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, "snippets.damaged-20261001-153000.json", "older")
		write(t, dir, FileName, "newer")
		name, err := New(dir).QuarantinePrimary(now)
		if err != nil {
			t.Fatal(err)
		}
		if name != "snippets.damaged-20261001-153000-2.json" {
			t.Errorf("name = %q", name)
		}
		if read(t, dir, "snippets.damaged-20261001-153000.json") != "older" || read(t, dir, name) != "newer" {
			t.Error("a damaged file was overwritten")
		}
	})
	t.Run("no primary", func(t *testing.T) {
		name, err := New(t.TempDir()).QuarantinePrimary(now)
		if err != nil || name != "" {
			t.Errorf("name=%q err=%v", name, err)
		}
	})
	t.Run("rename failure", func(t *testing.T) {
		dir := t.TempDir()
		write(t, dir, FileName, "xx")
		s := New(dir)
		s.fs.rename = func(string, string) error { return errors.New("injected") }
		if _, err := s.QuarantinePrimary(now); err == nil {
			t.Fatal("expected an error")
		}
		if read(t, dir, FileName) != "xx" {
			t.Error("primary lost")
		}
	})
}

func TestExportImportRoundTrip(t *testing.T) {
	lib := snippets.NewLibrary()
	for i, s := range corpus {
		lib.Snippets = append(lib.Snippets, snippets.Snippet{
			ID: fmt.Sprintf("%032x", i), Title: strings.ReplaceAll(s, "\n", " "), Body: s,
		})
	}
	var buf bytes.Buffer
	if err := Export(&buf, lib); err != nil {
		t.Fatal(err)
	}
	items, err := DecodeImport(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != len(lib.Snippets) {
		t.Fatalf("got %d items", len(items))
	}
	for i, it := range items {
		if it.Title != strings.TrimSpace(lib.Snippets[i].Title) || it.Body != lib.Snippets[i].Body {
			t.Errorf("item %d changed: %+v", i, it)
		}
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

func TestExportWriteError(t *testing.T) {
	if err := Export(errWriter{}, libOf("A")); err == nil {
		t.Error("expected write error")
	}
}

func TestDefaultDir(t *testing.T) {
	dir, err := DefaultDir()
	if err != nil {
		t.Skipf("no config dir in this environment: %v", err)
	}
	if filepath.Base(dir) != "Sniper" {
		t.Errorf("dir = %q", dir)
	}
	if New(dir).Dir() != dir {
		t.Error("Dir() mismatch")
	}
}

func TestLoadReadError(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, FileName, mustEncode(t, libOf("A")))
	s := New(dir)
	s.fs.readFile = func(string) ([]byte, error) { return nil, errors.New("injected read error") }
	if res := s.Load(); res.State != StateDamaged {
		t.Errorf("state = %v", res.State)
	}
}
