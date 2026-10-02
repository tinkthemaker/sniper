//go:build windows

package ui

import "os/exec"

// DefaultOpenFolder opens path in Explorer. Explorer exits non-zero even on
// success, so only a failure to start it is an error.
func DefaultOpenFolder(path string) error {
	return exec.Command("explorer", path).Start()
}
