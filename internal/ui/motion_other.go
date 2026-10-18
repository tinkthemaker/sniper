//go:build !windows

package ui

// systemAnimations is always true off Windows, where there is no setting to read.
func systemAnimations() bool { return true }
