package instance

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"
)

const helperEnv = "SNIPER_LOCK_HELPER_DIR"

// TestHelperHold runs only as a child process: it takes the lock, announces it,
// and then waits to be killed.
func TestHelperHold(t *testing.T) {
	dir := os.Getenv(helperEnv)
	if dir == "" {
		t.Skip("helper process only")
	}
	if _, err := Acquire(dir); err != nil {
		os.Stdout.WriteString("error: " + err.Error() + "\n")
		os.Exit(2)
	}
	os.Stdout.WriteString("locked\n")
	time.Sleep(time.Minute)
}

func TestAcquireTwiceAndRelease(t *testing.T) {
	dir := t.TempDir()
	l, err := Acquire(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Acquire(dir); !errors.Is(err, ErrAlreadyRunning) {
		t.Errorf("second Acquire: got %v, want ErrAlreadyRunning", err)
	}
	if err := l.Release(); err != nil {
		t.Fatal(err)
	}
	l2, err := Acquire(dir)
	if err != nil {
		t.Fatalf("Acquire after Release: %v", err)
	}
	_ = l2.Release()
}

func TestAcquireBadDir(t *testing.T) {
	if _, err := Acquire(t.TempDir() + "/does/not/exist"); err == nil || errors.Is(err, ErrAlreadyRunning) {
		t.Errorf("want a path error, got %v", err)
	}
}

func TestLockReleasedWhenHolderIsKilled(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperHold$")
	cmd.Env = append(os.Environ(), helperEnv+"="+dir)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	killed := false
	defer func() {
		if !killed {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()

	line, err := bufio.NewReader(out).ReadString('\n')
	if err != nil || line != "locked\n" {
		t.Fatalf("helper did not take the lock: %q %v", line, err)
	}
	if _, err := Acquire(dir); !errors.Is(err, ErrAlreadyRunning) {
		t.Fatalf("while child holds the lock: got %v", err)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = cmd.Wait()
	killed = true

	var l *Lock
	for i := 0; i < 50; i++ { // the OS may take a moment to close the dead handle
		if l, err = Acquire(dir); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("lock not released after the holder was killed: %v", err)
	}
	_ = l.Release()
}
