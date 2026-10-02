//go:build windows

package clipboard

import "strings"

// ToPlatform converts line breaks to CRLF, the Windows text convention.
func ToPlatform(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.ReplaceAll(text, "\n", "\r\n")
}
