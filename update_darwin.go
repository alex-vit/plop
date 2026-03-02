//go:build darwin

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func autoUpdate() {
	latestVer, url, err := checkForUpdate()
	if err != nil {
		log.Printf("update check failed: %v", err)
		return
	}
	if url == "" {
		log.Printf("no update available (current=%s)", versionDisplay())
		return
	}
	log.Printf("update available: v%s", latestVer)
	zipPath, err := downloadUpdate(url)
	if err != nil {
		log.Printf("update download failed: %v", err)
		return
	}
	defer os.Remove(zipPath)
	if err := applyUpdate(zipPath); err != nil {
		log.Printf("update apply failed: %v", err)
	}
}

// cleanOldBinary removes a leftover .app.old bundle from a previous update.
func cleanOldBinary() {
	appDir, err := appBundleDir()
	if err != nil {
		return
	}
	old := appDir + ".old"
	if _, err := os.Stat(old); err != nil {
		return // nothing to clean
	}
	if err := os.RemoveAll(old); err != nil {
		log.Printf("failed to remove old bundle: %v", err)
		return
	}
	log.Printf("removed old bundle: %s", old)
}

// checkForUpdate queries the GitHub releases API and returns the latest
// version and asset download URL if it's newer than the current version.
func checkForUpdate() (latestVer, downloadURL string, err error) {
	rel, err := fetchLatestRelease()
	if err != nil {
		return "", "", err
	}

	latestVer = strings.TrimPrefix(rel.TagName, "v")
	if !isNewer(latestVer, version) {
		return "", "", nil // up to date
	}

	// Asset names: plop-v1.2.3-macos-arm64-app.zip
	wantAsset := "plop-" + rel.TagName + "-macos-arm64-app.zip"
	for _, a := range rel.Assets {
		if strings.EqualFold(a.Name, wantAsset) {
			return latestVer, a.BrowserDownloadURL, nil
		}
	}
	return "", "", fmt.Errorf("no %s asset in release %s", wantAsset, rel.TagName)
}

// downloadUpdate downloads the zip to a temp file next to the .app bundle.
func downloadUpdate(url string) (zipPath string, err error) {
	appDir, err := appBundleDir()
	if err != nil {
		return "", err
	}

	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned %d", resp.StatusCode)
	}

	zipPath = appDir + ".update.zip"
	f, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		_ = os.Remove(zipPath)
		return "", err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(zipPath)
		return "", err
	}

	log.Printf("downloaded update to %s", zipPath)
	return zipPath, nil
}

// applyUpdate extracts the zip (containing Plop.app/) into a temp directory,
// then swaps it into place.
func applyUpdate(zipPath string) error {
	appDir, err := appBundleDir()
	if err != nil {
		return err
	}
	return applyUpdateToBundle(zipPath, appDir)
}

// applyUpdateToBundle replaces the .app bundle at appDir with the contents
// of the zip. Uses a rename dance for atomicity:
//
//	Plop.app → Plop.app.old
//	Plop.app.new (extracted) → Plop.app
//	Remove Plop.app.old
func applyUpdateToBundle(zipPath, appDir string) error {
	parentDir := filepath.Dir(appDir)
	appName := filepath.Base(appDir) // e.g. "Plop.app"

	// Extract zip to a temp directory next to the .app.
	extractDir, err := os.MkdirTemp(parentDir, "plop-update-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(extractDir)

	if err := extractZip(zipPath, extractDir); err != nil {
		return fmt.Errorf("extract zip: %w", err)
	}

	// The zip contains Plop.app/ at the top level.
	newApp := filepath.Join(extractDir, appName)
	if _, err := os.Stat(newApp); err != nil {
		return fmt.Errorf("extracted bundle not found at %s: %w", newApp, err)
	}

	oldApp := appDir + ".old"

	// Rename current → .old
	if err := os.Rename(appDir, oldApp); err != nil {
		return fmt.Errorf("rename current to .old: %w", err)
	}

	// Rename extracted → current
	if err := os.Rename(newApp, appDir); err != nil {
		// Rollback: restore old bundle.
		_ = os.Rename(oldApp, appDir)
		return fmt.Errorf("rename new to current: %w", err)
	}

	// Clean up old bundle (best-effort).
	_ = os.RemoveAll(oldApp)

	log.Printf("applied update, new version ready on next launch")
	return nil
}

// appBundleDir resolves the .app bundle directory from the running binary.
func appBundleDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	// Resolve symlinks to get the real path.
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}

	return appBundleDirFromPath(exe)
}

// appBundleDirFromPath walks up from a binary path to find the .app bundle root.
// The binary is at e.g. /Applications/Plop.app/Contents/MacOS/plop.
func appBundleDirFromPath(exe string) (string, error) {
	dir := exe
	for {
		if strings.HasSuffix(dir, ".app") {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}
	return "", fmt.Errorf("not running from a .app bundle: %s", exe)
}

// extractZip extracts a zip archive to the given directory.
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		// Prevent zip slip: ensure extracted path stays within destDir.
		target := filepath.Join(destDir, f.Name) //nolint:gosec
		if !strings.HasPrefix(filepath.Clean(target)+string(os.PathSeparator), filepath.Clean(destDir)+string(os.PathSeparator)) {
			if filepath.Clean(target) != filepath.Clean(destDir) {
				return fmt.Errorf("zip entry escapes target dir: %s", f.Name)
			}
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		if err := extractZipFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

func extractZipFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, rc); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
