//go:build windows

package tui

import "syscall"

var (
	modUser32       = syscall.NewLazyDLL("user32.dll")
	procGetKeyState = modUser32.NewProc("GetKeyState")
)

const vkCapital = 0x14

// IsCapsLockOn returns true if the Caps Lock toggle state is currently ON on Windows.
func IsCapsLockOn() bool {
	ret, _, _ := procGetKeyState.Call(uintptr(vkCapital))
	return (ret & 0x0001) != 0
}
