//go:build !windows

package paths

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigDir returns the plop configuration directory for macOS/Linux.
func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return filepath.Join(base, "plop"), nil
}
