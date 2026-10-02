//go:build !windows

package ui

import (
	"os/exec"
	"runtime"
)

// DefaultOpenFolder opens path in the platform file manager.
func DefaultOpenFolder(path string) error {
	opener := "xdg-open"
	if runtime.GOOS == "darwin" {
		opener = "open"
	}
	return exec.Command(opener, path).Start()
}
