//go:build !windows

package main

import (
	"os"
	"syscall"
)

func redirectFd(f *os.File) {
	fd := int(f.Fd())
	_ = syscall.Dup2(fd, 1)
	_ = syscall.Dup2(fd, 2)
}
