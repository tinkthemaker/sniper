//go:build windows

package ui

import (
	"syscall"
	"unsafe"
)

// spiGetClientAreaAnimation is SPI_GETCLIENTAREAANIMATION: the "Animation
// effects" switch in Settings → Accessibility → Visual effects.
const spiGetClientAreaAnimation = 0x1042

// systemAnimations reports whether Windows has animation effects on. If the
// call fails it says yes, the platform default.
func systemAnimations() bool {
	proc := syscall.NewLazyDLL("user32.dll").NewProc("SystemParametersInfoW")
	var on int32 = 1
	r, _, _ := proc.Call(spiGetClientAreaAnimation, 0, uintptr(unsafe.Pointer(&on)), 0)
	if r == 0 {
		return true
	}
	return on != 0
}
