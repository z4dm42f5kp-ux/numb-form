//go:build windows

package manual

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// Console modes (wincon.h).
const (
	enableProcessedInput  = 0x0001
	enableLineInput       = 0x0002
	enableEchoInput       = 0x0004
	enableVTInput         = 0x0200
	enableVTProcessing    = 0x0004
	enableProcessedOutput = 0x0001
)

func kernel32() *syscall.LazyDLL {
	root := os.Getenv("SystemRoot")
	if root == "" {
		root = `C:\Windows`
	}
	return syscall.NewLazyDLL(filepath.Join(root, "System32", "kernel32.dll"))
}

// RawTerminal switches the console to single-key input with escape
// sequences for the arrow and page keys (Windows 10 and later), and
// returns a function that restores it, and the window size.
func RawTerminal() (restore func(), rows, cols int, err error) {
	k := kernel32()
	get, set, info := k.NewProc("GetConsoleMode"), k.NewProc("SetConsoleMode"), k.NewProc("GetConsoleScreenBufferInfo")
	in, out := os.Stdin.Fd(), os.Stdout.Fd()
	var inMode, outMode uint32
	if r, _, _ := get.Call(in, uintptr(unsafe.Pointer(&inMode))); r == 0 { // #nosec G103 -- Win32 call.
		return nil, 0, 0, errors.New("not a console")
	}
	if r, _, _ := get.Call(out, uintptr(unsafe.Pointer(&outMode))); r == 0 { // #nosec G103 -- Win32 call.
		return nil, 0, 0, errors.New("not a console")
	}
	newIn := inMode&^(enableLineInput|enableEchoInput|enableProcessedInput) | enableVTInput
	if r, _, _ := set.Call(in, uintptr(newIn)); r == 0 {
		return nil, 0, 0, errors.New("the console has no single-key input")
	}
	if r, _, _ := set.Call(out, uintptr(outMode|enableVTProcessing|enableProcessedOutput)); r == 0 {
		_, _, _ = set.Call(in, uintptr(inMode))
		return nil, 0, 0, errors.New("the console cannot redraw the screen")
	}
	var csbi struct {
		size, cursor        [2]int16
		attributes          uint16
		left, top, right, b int16
		max                 [2]int16
	}
	rows, cols = 24, 80
	if r, _, _ := info.Call(out, uintptr(unsafe.Pointer(&csbi))); r != 0 { // #nosec G103 -- Win32 call.
		rows, cols = int(csbi.b-csbi.top)+1, int(csbi.right-csbi.left)+1
	}
	return func() {
		_, _, _ = set.Call(in, uintptr(inMode))
		_, _, _ = set.Call(out, uintptr(outMode))
	}, rows, cols, nil
}
