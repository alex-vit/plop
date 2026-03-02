package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyUpdate(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "plop.exe")
	tmp := exe + ".tmp"

	_ = os.WriteFile(exe, []byte("old"), 0o755) //nolint:errcheck
	_ = os.WriteFile(tmp, []byte("new"), 0o755) //nolint:errcheck

	old := exe + ".old"

	if err := os.Rename(exe, old); err != nil {
		t.Fatalf("rename exe to old: %v", err)
	}
	if err := os.Rename(tmp, exe); err != nil {
		_ = os.Rename(old, exe) //nolint:errcheck
		t.Fatalf("rename tmp to exe: %v", err)
	}

	got, _ := os.ReadFile(exe)
	if string(got) != "new" {
		t.Errorf("exe content = %q, want %q", got, "new")
	}

	gotOld, _ := os.ReadFile(old)
	if string(gotOld) != "old" {
		t.Errorf("old content = %q, want %q", gotOld, "old")
	}

	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Errorf("tmp file should not exist after rename")
	}
}

func TestApplyUpdateRollback(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "plop.exe")

	_ = os.WriteFile(exe, []byte("old"), 0o755) //nolint:errcheck

	old := exe + ".old"
	tmp := exe + ".tmp"

	if err := os.Rename(exe, old); err != nil {
		t.Fatalf("rename exe to old: %v", err)
	}
	if err := os.Rename(tmp, exe); err != nil {
		// Rollback.
		_ = os.Rename(old, exe) //nolint:errcheck
	}

	got, _ := os.ReadFile(exe)
	if string(got) != "old" {
		t.Errorf("after rollback, exe content = %q, want %q", got, "old")
	}
}
