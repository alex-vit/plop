//go:build !darwin && !windows

package autostart

import "errors"

var errUnsupported = errors.New("autostart is only supported on macOS and Windows")

func MenuLabel() string {
	return ""
}

func IsEnabled(_ string) bool {
	return false
}

func Enable(_ string) error {
	return errUnsupported
}

func Disable(_ string) error {
	return errUnsupported
}
