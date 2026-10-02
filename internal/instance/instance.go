package instance

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/gofrs/flock"
)

// LockName is the lock file inside the data directory.
const LockName = "sniper.lock"

var ErrAlreadyRunning = errors.New("another Sniper instance is running")

// Lock is the held single-instance lock. The OS drops it if the process dies.
type Lock struct{ f *flock.Flock }

// Acquire takes the lock on dir/sniper.lock without blocking.
func Acquire(dir string) (*Lock, error) {
	f := flock.New(filepath.Join(dir, LockName))
	ok, err := f.TryLock()
	if err != nil {
		return nil, fmt.Errorf("lock %s: %w", f.Path(), err)
	}
	if !ok {
		return nil, ErrAlreadyRunning
	}
	return &Lock{f: f}, nil
}

// Release drops the lock. The lock file itself is left in place.
func (l *Lock) Release() error {
	if err := l.f.Unlock(); err != nil {
		return fmt.Errorf("release lock: %w", err)
	}
	return nil
}
