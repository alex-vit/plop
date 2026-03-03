//go:build darwin || windows

package autostart

import "path/filepath"

func cleanHomeDir(homeDir string) string {
	if homeDir == "" {
		return ""
	}
	abs, err := filepath.Abs(homeDir)
	if err != nil {
		return filepath.Clean(homeDir)
	}
	return abs
}
