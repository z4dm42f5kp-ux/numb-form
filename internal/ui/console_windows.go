//go:build windows

package ui

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// OwnConsole reports whether this process is the only one attached to its
// console, which is the case when it was started by double-clicking it in
// Explorer. The window would then close as soon as the program exits.
func OwnConsole() bool {
	root := os.Getenv("SystemRoot")
	if root == "" {
		root = `C:\Windows`
	}
	k32 := syscall.NewLazyDLL(filepath.Join(root, "System32", "kernel32.dll"))
	proc := k32.NewProc("GetConsoleProcessList")
	if proc.Find() != nil {
		return false
	}
	var ids [4]uint32
	n, _, _ := proc.Call(uintptr(unsafe.Pointer(&ids[0])), uintptr(len(ids))) // #nosec G103 -- Win32 call.
	return n == 1
}
