//go:build windows

package store

import (
	"errors"
	"syscall"
)

const (
	errorAccessDenied     syscall.Errno = 5
	errorSharingViolation syscall.Errno = 32
)

// retryableRename reports whether a rename failure is likely a short-lived lock
// held by antivirus or an indexer.
func retryableRename(err error) bool {
	var errno syscall.Errno
	return errors.As(err, &errno) && (errno == errorAccessDenied || errno == errorSharingViolation)
}
