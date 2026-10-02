//go:build !windows

package clipboard

// ToPlatform returns text unchanged; other platforms use LF.
func ToPlatform(text string) string { return text }
