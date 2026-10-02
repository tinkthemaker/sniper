//go:build !windows

package store

// retryableRename is always false: rename on Unix replaces atomically without transient locks.
func retryableRename(error) bool { return false }
