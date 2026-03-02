//go:build darwin

package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestAppBundleDir(t *testing.T) {
	tests := []struct {
		name    string
		exe     string
		want    string
		wantErr bool
	}{
		{
			name: "standard app bundle",
			exe:  "/Applications/Plop.app/Contents/MacOS/plop",
			want: "/Applications/Plop.app",
		},
		{
			name: "nested in user Applications",
			exe:  "/Users/alice/Applications/Plop.app/Contents/MacOS/plop",
			want: "/Users/alice/Applications/Plop.app",
		},
		{
			name: "out directory bundle",
			exe:  "/Users/alice/code/plop/out/Plop.app/Contents/MacOS/plop",
			want: "/Users/alice/code/plop/out/Plop.app",
		},
		{
			name:    "not in app bundle",
			exe:     "/usr/local/bin/plop",
			wantErr: true,
		},
		{
			name:    "bare binary in home",
			exe:     "/Users/alice/plop",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := appBundleDirFromPath(tt.exe)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyUpdateDarwin(t *testing.T) {
	// Create a fake .app bundle.
	dir := t.TempDir()
	appDir := filepath.Join(dir, "Plop.app")
	macosDir := filepath.Join(appDir, "Contents", "MacOS")
	if err := os.MkdirAll(macosDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(macosDir, "plop"), []byte("old-binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	plist := filepath.Join(appDir, "Contents", "Info.plist")
	if err := os.WriteFile(plist, []byte("<plist>old</plist>"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a zip containing a new Plop.app bundle.
	zipPath := filepath.Join(dir, "update.zip")
	createTestZip(t, zipPath, map[string][]byte{
		"Plop.app/Contents/MacOS/plop":  []byte("new-binary"),
		"Plop.app/Contents/Info.plist":   []byte("<plist>new</plist>"),
		"Plop.app/Contents/Resources/x": []byte("resource"),
	})

	if err := applyUpdateToBundle(zipPath, appDir); err != nil {
		t.Fatalf("applyUpdate: %v", err)
	}

	// Verify new binary is in place.
	got, err := os.ReadFile(filepath.Join(appDir, "Contents", "MacOS", "plop"))
	if err != nil {
		t.Fatalf("read new binary: %v", err)
	}
	if string(got) != "new-binary" {
		t.Errorf("binary content = %q, want %q", got, "new-binary")
	}

	// Verify new plist.
	gotPlist, err := os.ReadFile(filepath.Join(appDir, "Contents", "Info.plist"))
	if err != nil {
		t.Fatalf("read new plist: %v", err)
	}
	if string(gotPlist) != "<plist>new</plist>" {
		t.Errorf("plist content = %q, want %q", gotPlist, "<plist>new</plist>")
	}

	// Verify new resource.
	gotRes, err := os.ReadFile(filepath.Join(appDir, "Contents", "Resources", "x"))
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if string(gotRes) != "resource" {
		t.Errorf("resource content = %q, want %q", gotRes, "resource")
	}

	// Verify old bundle is cleaned up.
	if _, err := os.Stat(appDir + ".old"); !os.IsNotExist(err) {
		t.Errorf(".old bundle should be removed, stat err = %v", err)
	}
}

func TestApplyUpdateDarwinRollback(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "Plop.app")
	macosDir := filepath.Join(appDir, "Contents", "MacOS")
	if err := os.MkdirAll(macosDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(macosDir, "plop"), []byte("old-binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a zip that does NOT contain the expected Plop.app directory.
	zipPath := filepath.Join(dir, "bad-update.zip")
	createTestZip(t, zipPath, map[string][]byte{
		"WrongApp.app/Contents/MacOS/plop": []byte("wrong"),
	})

	err := applyUpdateToBundle(zipPath, appDir)
	if err == nil {
		t.Fatal("expected error for mismatched bundle name")
	}

	// Verify original bundle is still intact (not renamed away).
	got, readErr := os.ReadFile(filepath.Join(appDir, "Contents", "MacOS", "plop"))
	if readErr != nil {
		t.Fatalf("original bundle should still exist: %v", readErr)
	}
	if string(got) != "old-binary" {
		t.Errorf("original binary content = %q, want %q", got, "old-binary")
	}
}

func createTestZip(t *testing.T, path string, files map[string][]byte) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
