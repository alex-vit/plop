package paths

import (
	"log"
	"os"
	"path/filepath"
)

// MigrateConfigDir renames the old lowercase config directory ("plop") to the
// new capitalized name ("Plop"). On case-insensitive filesystems (macOS, Windows)
// this just changes the display name. On case-sensitive filesystems (Linux) it's
// a real rename. If anything goes wrong the old path is kept — the app still works.
func MigrateConfigDir() {
	newDir, err := ConfigDir()
	if err != nil {
		return
	}
	oldDir := filepath.Join(filepath.Dir(newDir), "plop")
	migrateConfigDir(oldDir, newDir)
}

func migrateConfigDir(oldDir, newDir string) {
	// Already the same path (shouldn't happen, but guard against it).
	if oldDir == newDir {
		return
	}

	oldInfo, err := os.Stat(oldDir)
	if err != nil || !oldInfo.IsDir() {
		return // Old dir doesn't exist — nothing to migrate.
	}

	// On case-sensitive filesystems, if both old and new exist as separate
	// directories, don't rename — avoid data loss.
	if newInfo, err := os.Stat(newDir); err == nil && newInfo.IsDir() && !os.SameFile(oldInfo, newInfo) {
		return
	}

	if err := os.Rename(oldDir, newDir); err != nil {
		log.Printf("config dir migration: %v", err)
	}
}
