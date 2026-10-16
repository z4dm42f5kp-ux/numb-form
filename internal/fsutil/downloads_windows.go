//go:build windows

package fsutil

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// folderIDDownloads is FOLDERID_Downloads {374DE290-123F-4565-9164-39C4925E467B}.
var folderIDDownloads = syscall.GUID{
	Data1: 0x374DE290, Data2: 0x123F, Data3: 0x4565,
	Data4: [8]byte{0x91, 0x64, 0x39, 0xC4, 0x92, 0x5E, 0x46, 0x7B},
}

func systemDLL(name string) *syscall.LazyDLL {
	// Load by absolute path from System32 to avoid DLL search-order
	// hijacking. Both DLLs are also KnownDLLs.
	root := os.Getenv("SystemRoot")
	if root == "" {
		root = `C:\Windows`
	}
	return syscall.NewLazyDLL(filepath.Join(root, "System32", name))
}

// downloadsDir asks the shell for the Downloads known folder, which
// follows redirection (for example to OneDrive or a network drive).
func downloadsDir() (string, error) {
	shell32 := systemDLL("shell32.dll")
	ole32 := systemDLL("ole32.dll")
	get := shell32.NewProc("SHGetKnownFolderPath")
	free := ole32.NewProc("CoTaskMemFree")
	if err := get.Find(); err == nil {
		var p *uint16
		// #nosec G103 -- required to call the Win32 API.
		r, _, _ := get.Call(uintptr(unsafe.Pointer(&folderIDDownloads)), 0, 0, uintptr(unsafe.Pointer(&p)))
		if r == 0 && p != nil {
			defer free.Call(uintptr(unsafe.Pointer(p))) // #nosec G103
			n := 0
			for ptr := unsafe.Pointer(p); *(*uint16)(ptr) != 0; n++ { // #nosec G103
				ptr = unsafe.Add(ptr, 2)
			}
			return syscall.UTF16ToString(unsafe.Slice(p, n)), nil // #nosec G103
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Downloads"), nil
}
