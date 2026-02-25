//go:build windows

package cmd

import (
	"os"
	"syscall"
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	setStdHandle = kernel32.NewProc("SetStdHandle")
)

const (
	stdOutputHandle = ^uintptr(0) - 11 + 1 // STD_OUTPUT_HANDLE = -11
	stdErrorHandle  = ^uintptr(0) - 12 + 1 // STD_ERROR_HANDLE  = -12
)

func redirectFd(f *os.File) {
	handle := uintptr(f.Fd())
	setStdHandle.Call(stdOutputHandle, handle)
	setStdHandle.Call(stdErrorHandle, handle)
}
