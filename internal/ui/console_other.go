//go:build !windows

package ui

// OwnConsole is only meaningful on Windows.
func OwnConsole() bool { return false }
