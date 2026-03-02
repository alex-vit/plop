# Auto-Update Implementation

Self-update via GitHub Releases with zero new dependencies. Ported from monibright's `update.go`.

## Design Decisions

### DIY vs library
Pure stdlib. Same reasoning as monibright — thin wrapper wins over heavy lib (e.g. `creativeprojects/go-selfupdate`, 30+ transitive deps) when usage is narrow.

### Platform support: Windows + macOS (Linux no-op)
Both platforms check GitHub Releases on startup, download the update silently, and apply it for next launch. Linux has no-op stubs since there's no standard app bundle format.

Implemented via build-constrained files: `_windows.go`, `_darwin.go`, `_linux.go`, plus `_shared.go` (no constraint) for common code.

### Windows approach
Asset: `plop-v1.2.3-windows-amd64.exe` (single binary). Windows can't overwrite a running exe, so uses a rename dance: `plop.exe` → `plop.exe.old`, then `plop.exe.tmp` → `plop.exe`. `cleanOldBinary()` removes the `.old` file on next startup.

### macOS approach
Asset: `plop-v1.2.3-macos-arm64-app.zip` (zip containing `Plop.app/`). Unix can overwrite running binaries, but we replace the whole `.app` bundle (binary + Info.plist + resources) for consistency. Steps:
1. Download zip to `Plop.app.update.zip`
2. Extract to temp dir (produces `Plop.app/`)
3. Rename `Plop.app` → `Plop.app.old`
4. Rename extracted `Plop.app` → `Plop.app`
5. Remove `Plop.app.old` (best-effort, `cleanOldBinary` retries on next startup)

Rollback: if step 4 fails, restores `Plop.app.old` → `Plop.app`.

`appBundleDir()` resolves the `.app` path by walking up from `os.Executable()` to find a component ending in `.app`. If not in a bundle (dev build, CLI `./plop`), update is silently skipped.

### Versioned asset names
Unlike monibright (`monibright.exe`), plop's release workflow produces versioned asset names. Each platform's `checkForUpdate` constructs the expected asset name from `rel.TagName`.

### Silent apply, no restart
Update is applied silently in the background. New version takes effect on next natural launch (reboot, autostart, or manual). No user prompt.

### Dev builds skip update
`isNewer()` returns false when `Version == ""` or `"dev"`, so dev/debug builds never trigger updates.

### Repo must be public
The `/releases/latest` GitHub API endpoint returns 404 for private repos without auth. Changed plop from private → public when implementing this (`gh repo edit --visibility public --accept-visibility-change-consequences`).

## Update Sequence

1. On startup: `cleanOldBinary()` removes any leftover `.old` file/bundle from a previous update.
2. In `rootCmd.RunE`, before `tray.Run()`: `go autoUpdate()` spawns a background goroutine.
3. `checkForUpdate()` calls `fetchLatestRelease()` (shared), compares semver, looks for platform-specific asset.
4. `downloadUpdate()` streams the asset to a temp file next to the current binary/bundle.
5. `applyUpdate()` does the platform-specific swap with rollback on failure.
6. New version takes effect on next launch.

## File Structure

- **`cmd/update_shared.go`** — types (`ghRelease`, `ghAsset`), `fetchLatestRelease`, `isNewer`, `parseSemver`, `versionDisplay`
- **`cmd/update_windows.go`** — Windows-specific: `autoUpdate`, `cleanOldBinary`, `checkForUpdate`, `downloadUpdate`, `applyUpdate`
- **`cmd/update_darwin.go`** — macOS-specific: same functions + `appBundleDir`, `appBundleDirFromPath`, `extractZip`
- **`cmd/update_linux.go`** — no-op stubs (`//go:build linux`)
- **`cmd/update_shared_test.go`** — `TestIsNewer` (cross-platform)
- **`cmd/update_windows_test.go`** — `TestApplyUpdate`, `TestApplyUpdateRollback`
- **`cmd/update_darwin_test.go`** — `TestAppBundleDir`, `TestApplyUpdateDarwin`, `TestApplyUpdateDarwinRollback`
- **`cmd/root.go`** — `cleanOldBinary()` after log setup; `go autoUpdate()` before `tray.Run()`

## Alternatives Considered

| Option | Decision |
|---|---|
| Immediate restart after update | Rejected — interrupts user; silent apply is better UX |
| Progress indicator in tray | Rejected — update is fast, not worth the UI complexity |
| Hash/signature verification | Not added — GitHub HTTPS + tag-based release is sufficient for personal tool |
| Binary-only swap on macOS | Rejected — replacing the full `.app` bundle is cleaner and handles Info.plist + resource changes |
