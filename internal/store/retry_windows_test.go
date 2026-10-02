//go:build windows

package store

import (
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestRenameRetry(t *testing.T) {
	t.Run("succeeds after two sharing violations", func(t *testing.T) {
		dir := t.TempDir()
		s := New(dir)
		var sleeps []time.Duration
		s.fs.sleep = func(d time.Duration) { sleeps = append(sleeps, d) }
		real, calls := s.fs.rename, 0
		s.fs.rename = func(a, b string) error {
			calls++
			if calls <= 2 {
				return &os.LinkError{Op: "rename", Old: a, New: b, Err: syscall.Errno(32)}
			}
			return real(a, b)
		}
		if err := s.Save(libOf("A")); err != nil {
			t.Fatal(err)
		}
		if calls != 3 {
			t.Errorf("rename calls = %d, want 3", calls)
		}
		if len(sleeps) != 2 || sleeps[0] != 20*time.Millisecond || sleeps[1] != 40*time.Millisecond {
			t.Errorf("backoff = %v", sleeps)
		}
	})
	t.Run("access denied forever fails after 5 attempts", func(t *testing.T) {
		dir := t.TempDir()
		s := New(dir)
		s.fs.sleep = func(time.Duration) {}
		calls := 0
		s.fs.rename = func(a, b string) error {
			calls++
			return &os.LinkError{Op: "rename", Old: a, New: b, Err: syscall.Errno(5)}
		}
		err := s.Save(libOf("A"))
		if err == nil {
			t.Fatal("expected failure")
		}
		if calls != 5 {
			t.Errorf("rename calls = %d, want 5", calls)
		}
		var errno syscall.Errno
		if !errors.As(err, &errno) || errno != 5 {
			t.Errorf("error should wrap Errno 5: %v", err)
		}
		if got := tmpFiles(t, dir); len(got) != 0 {
			t.Errorf("temp files left: %v", got)
		}
	})
	t.Run("other errors fail immediately", func(t *testing.T) {
		s := New(t.TempDir())
		calls := 0
		s.fs.rename = func(a, b string) error { calls++; return &os.LinkError{Err: syscall.Errno(2)} }
		if err := s.Save(libOf("A")); err == nil || calls != 1 {
			t.Errorf("err=%v calls=%d", err, calls)
		}
	})
}

func TestRealReplaceOnWindows(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	for i := 0; i < 50; i++ {
		title := string(rune('A' + i%26))
		lib := libOf(title, "second")
		lib.Snippets[0].Body = "body " + title + string(rune('0'+i/26))
		if err := s.Save(lib); err != nil {
			t.Fatalf("save %d: %v", i, err)
		}
		res := New(dir).Load()
		if res.State != StateOK || res.Library.Snippets[0].Body != lib.Snippets[0].Body {
			t.Fatalf("after save %d: %+v", i, res)
		}
	}
}
