//go:build windows

package cmd

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	setStdHandle    = kernel32.NewProc("SetStdHandle")
	createMutexW    = kernel32.NewProc("CreateMutexW")
)

const (
	stdOutputHandle = ^uintptr(0) - 11 + 1 // STD_OUTPUT_HANDLE = -11
	stdErrorHandle  = ^uintptr(0) - 12 + 1 // STD_ERROR_HANDLE  = -12
)

func init() {
	// Create a named mutex so the Inno Setup installer (AppMutex=PlopMutex)
	// can detect and close a running instance before overwriting the binary.
	name, _ := syscall.UTF16PtrFromString("PlopMutex")
	createMutexW.Call(0, 1, uintptr(unsafe.Pointer(name)))
}

func redirectFd(f *os.File) {
	handle := uintptr(f.Fd())
	setStdHandle.Call(stdOutputHandle, handle)
	setStdHandle.Call(stdErrorHandle, handle)
}
