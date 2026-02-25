//go:build !windows

package cmd

import (
	"os"
	"syscall"
)

func redirectFd(f *os.File) {
	fd := int(f.Fd())
	syscall.Dup2(fd, 1)
	syscall.Dup2(fd, 2)
}
