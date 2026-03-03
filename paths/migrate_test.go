package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMigrateConfigDir_OldExists(t *testing.T) {
	tmp := t.TempDir()
	oldDir := filepath.Join(tmp, "plop")
	newDir := filepath.Join(tmp, "Plop")

	if err := os.Mkdir(oldDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldDir, "config.xml"), []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	migrateConfigDir(oldDir, newDir)

	if _, err := os.Stat(filepath.Join(newDir, "config.xml")); err != nil {
		t.Fatalf("expected config.xml in new dir: %v", err)
	}
}

func TestMigrateConfigDir_OldNotExists(t *testing.T) {
	tmp := t.TempDir()
	oldDir := filepath.Join(tmp, "plop")
	newDir := filepath.Join(tmp, "Plop")

	// Should be a no-op, no panic or error.
	migrateConfigDir(oldDir, newDir)

	if _, err := os.Stat(newDir); !os.IsNotExist(err) {
		t.Fatalf("expected new dir to not exist, got: %v", err)
	}
}

func TestMigrateConfigDir_BothExist(t *testing.T) {
	tmp := t.TempDir()
	oldDir := filepath.Join(tmp, "plop")
	newDir := filepath.Join(tmp, "Plop")

	if err := os.Mkdir(oldDir, 0o700); err != nil {
		t.Fatal(err)
	}
	// On case-insensitive FS, creating both fails — they're the same dir.
	if err := os.Mkdir(newDir, 0o700); err != nil {
		t.Skip("case-insensitive filesystem: cannot create both plop and Plop")
	}

	if err := os.WriteFile(filepath.Join(oldDir, "old.txt"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "new.txt"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	migrateConfigDir(oldDir, newDir)

	// Both should still exist with original contents — no data loss.
	if _, err := os.Stat(filepath.Join(oldDir, "old.txt")); err != nil {
		t.Fatalf("old dir contents should be preserved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(newDir, "new.txt")); err != nil {
		t.Fatalf("new dir contents should be preserved: %v", err)
	}
}
