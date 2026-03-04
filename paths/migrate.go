package paths

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
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

	if err := RobustRename(oldDir, newDir); err != nil {
		log.Printf("config dir migration: %v", err)
	}
}

// RobustRename renames oldPath to newPath. On Windows, it retries with
// exponential backoff (up to ~1.5s total) to handle transient "Access is denied"
// errors caused by open file handles from indexers, antivirus, or recently
// exited processes.
func RobustRename(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err == nil || runtime.GOOS != "windows" {
		return err
	}
	for _, d := range []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	} {
		log.Printf("rename %s → %s: retrying in %v (%v)", filepath.Base(oldPath), filepath.Base(newPath), d, err)
		time.Sleep(d)
		err = os.Rename(oldPath, newPath)
		if err == nil {
			return nil
		}
	}
	return err
}
