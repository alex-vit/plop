package engine

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	stconfig "github.com/syncthing/syncthing/lib/config"
)

func TestEnsureRuntimeGUIAddressAssignsPort(t *testing.T) {
	cfg := stconfig.Configuration{}
	cfg.GUI.Enabled = true
	cfg.GUI.RawAddress = "127.0.0.1:0" //nolint:goconst

	if err := ensureRuntimeGUIAddress(&cfg); err != nil {
		t.Fatalf("ensureRuntimeGUIAddress: %v", err)
	}

	host, port, err := net.SplitHostPort(cfg.GUI.RawAddress)
	if err != nil {
		t.Fatalf("split host/port: %v", err)
	}
	if host != "127.0.0.1" { //nolint:goconst
		t.Fatalf("host = %q, want 127.0.0.1", host)
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		t.Fatalf("port parse: %v", err)
	}
	if portNum <= 0 {
		t.Fatalf("port = %d, want > 0", portNum)
	}
}

func TestEnsureRuntimeGUIAddressKeepsConfiguredPort(t *testing.T) {
	cfg := stconfig.Configuration{}
	cfg.GUI.Enabled = true
	cfg.GUI.RawAddress = "127.0.0.1:8384"

	if err := ensureRuntimeGUIAddress(&cfg); err != nil {
		t.Fatalf("ensureRuntimeGUIAddress: %v", err)
	}
	if cfg.GUI.RawAddress != "127.0.0.1:8384" {
		t.Fatalf("raw address = %q, want unchanged", cfg.GUI.RawAddress)
	}
}

func TestEnsureRuntimeGUIAddressNormalizesWildcardHost(t *testing.T) {
	cfg := stconfig.Configuration{}
	cfg.GUI.Enabled = true
	cfg.GUI.RawAddress = "0.0.0.0:0"

	if err := ensureRuntimeGUIAddress(&cfg); err != nil {
		t.Fatalf("ensureRuntimeGUIAddress: %v", err)
	}

	host, _, err := net.SplitHostPort(cfg.GUI.RawAddress)
	if err != nil {
		t.Fatalf("split host/port: %v", err)
	}
	if host != "127.0.0.1" {
		t.Fatalf("host = %q, want 127.0.0.1", host)
	}
}

func TestMigrateFolderNameRenamesExistingDir(t *testing.T) {
	tmp := t.TempDir()
	oldDir := filepath.Join(tmp, "plop")
	newDir := filepath.Join(tmp, "Plop")
	if err := os.Mkdir(oldDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write a marker file so we can verify contents survived.
	if err := os.WriteFile(filepath.Join(oldDir, "test.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := stconfig.Configuration{
		Folders: []stconfig.FolderConfiguration{{Path: oldDir}},
	}
	changed := migrateFolderName(&cfg)
	if !changed {
		t.Fatal("expected migrateFolderName to return true")
	}
	if cfg.Folders[0].Path != newDir {
		t.Fatalf("path = %q, want %q", cfg.Folders[0].Path, newDir)
	}
	// Old dir should no longer exist (or be the same inode on case-insensitive FS).
	if _, err := os.Stat(newDir); err != nil {
		t.Fatalf("new dir does not exist: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(newDir, "test.txt"))
	if err != nil {
		t.Fatalf("reading marker file: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("marker file content = %q, want %q", data, "hello")
	}
}

func TestMigrateFolderNameOpenFileHandle(t *testing.T) {
	tmp := t.TempDir()
	oldDir := filepath.Join(tmp, "plop")
	if err := os.Mkdir(oldDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Hold an open file handle inside the directory (simulates Syncthing or
	// another process having a file open in the sync folder).
	f, err := os.Create(filepath.Join(oldDir, "lockfile.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		// On Windows, open handles block directory renames.
		// Close after a delay so the retry loop can succeed.
		go func() {
			time.Sleep(150 * time.Millisecond)
			f.Close()
		}()
	} else {
		defer f.Close()
	}

	cfg := stconfig.Configuration{
		Folders: []stconfig.FolderConfiguration{{Path: oldDir}},
	}
	changed := migrateFolderName(&cfg)
	if !changed {
		t.Fatal("expected migrateFolderName to return true")
	}

	// Verify actual directory name on disk (not just case-insensitive access).
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatal(err)
	}
	var found string
	for _, e := range entries {
		if strings.EqualFold(e.Name(), "plop") {
			found = e.Name()
		}
	}
	if found != "Plop" {
		t.Fatalf("directory name on disk = %q, want %q", found, "Plop")
	}
}

func TestMigrateFolderNameNoDirOnDisk(t *testing.T) {
	tmp := t.TempDir()
	oldDir := filepath.Join(tmp, "plop")
	newDir := filepath.Join(tmp, "Plop")

	cfg := stconfig.Configuration{
		Folders: []stconfig.FolderConfiguration{{Path: oldDir}},
	}
	changed := migrateFolderName(&cfg)
	if !changed {
		t.Fatal("expected migrateFolderName to return true")
	}
	if cfg.Folders[0].Path != newDir {
		t.Fatalf("path = %q, want %q", cfg.Folders[0].Path, newDir)
	}
}

func TestMigrateFolderNameAlreadyCapitalized(t *testing.T) {
	tmp := t.TempDir()
	cfg := stconfig.Configuration{
		Folders: []stconfig.FolderConfiguration{{Path: filepath.Join(tmp, "Plop")}},
	}
	if migrateFolderName(&cfg) {
		t.Fatal("expected no change for already-capitalized path")
	}
}

func TestMigrateFolderNameCustomPath(t *testing.T) {
	tmp := t.TempDir()
	cfg := stconfig.Configuration{
		Folders: []stconfig.FolderConfiguration{{Path: filepath.Join(tmp, "my-sync")}},
	}
	if migrateFolderName(&cfg) {
		t.Fatal("expected no change for custom path")
	}
}

func TestMigrateFolderNameNoFolders(t *testing.T) {
	cfg := stconfig.Configuration{}
	if migrateFolderName(&cfg) {
		t.Fatal("expected no change for empty folders")
	}
}
