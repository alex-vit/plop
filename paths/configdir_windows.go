//go:build windows

package paths

import (
	"errors"
	"os"
	"path/filepath"
)

// ConfigDir returns the plop configuration directory for Windows.
// Uses %LocalAppData% to match InnoSetup's {localappdata}.
func ConfigDir() (string, error) {
	dir := os.Getenv("LocalAppData")
	if dir == "" {
		return "", errors.New("config dir: %LocalAppData% is not set")
	}
	return filepath.Join(dir, "Plop"), nil
}
