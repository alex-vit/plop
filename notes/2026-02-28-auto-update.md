# Auto-Update Implementation

Self-update via GitHub Releases with zero new dependencies. Ported from monibright's `update.go`.

## Design Decisions

### DIY vs library
Pure stdlib (~160 lines in `cmd/update_windows.go`). Same reasoning as monibright — thin wrapper wins over heavy lib (e.g. `creativeprojects/go-selfupdate`, 30+ transitive deps) when usage is narrow.

### Windows-only (for now)
macOS `.app` bundles are directories, not single files. The rename-based exe replacement approach is Windows-specific. macOS update would require downloading a zip, extracting, replacing the bundle, and relaunching — significantly more complex. Marked for future work if needed.

Implemented via `_windows.go` / `_other.go` filename split so macOS and Linux builds compile cleanly with no-op stubs.

### Versioned asset names
Unlike monibright (`monibright.exe`), plop's release workflow produces versioned asset names: `plop-v1.2.3-windows-amd64.exe`. `checkForUpdate` constructs `"plop-" + rel.TagName + "-windows-amd64.exe"` to find the right asset.

### Silent apply, no restart
Same approach as monibright: update is applied silently in the background. New exe takes effect on next natural launch (reboot, autostart, or manual). No user prompt.

### Dev builds skip update
`isNewer()` returns false when `Version == ""` or `"dev"`, so dev/debug builds never trigger updates.

### Repo must be public
The `/releases/latest` GitHub API endpoint returns 404 for private repos without auth. Changed plop from private → public when implementing this (`gh repo edit --visibility public --accept-visibility-change-consequences`).

## Update Sequence

1. On startup: `cleanOldBinary()` removes any leftover `.old` file from a previous update.
2. In `rootCmd.RunE`, before `tray.Run()`: `go autoUpdate()` spawns a background goroutine.
3. `checkForUpdate()` calls the GitHub API, compares semver, returns download URL if newer.
4. `downloadUpdate()` streams the asset to `plop.exe.tmp`.
5. `applyUpdate()` does the Windows rename dance:
   - `plop.exe` → `plop.exe.old`
   - `plop.exe.tmp` → `plop.exe`
   - If step 2 fails, restore `.old` → `plop.exe`
6. New version takes effect on next launch.

## Files Changed

- **`cmd/update_windows.go`** (new) — `cleanOldBinary`, `checkForUpdate`, `downloadUpdate`, `applyUpdate`, `isNewer`, `parseSemver`, `versionDisplay`
- **`cmd/update_other.go`** (new) — no-op stubs for non-Windows (`//go:build !windows`)
- **`cmd/update_windows_test.go`** (new) — `TestIsNewer`, `TestApplyUpdate`, `TestApplyUpdateRollback`
- **`cmd/root.go`** — `cleanOldBinary()` after log setup; `go autoUpdate()` before `tray.Run()`

## Alternatives Considered

| Option | Decision |
|---|---|
| macOS update support | Deferred — .app bundles need zip extract + bundle replace, not just rename |
| Immediate restart after update | Rejected — interrupts user; silent apply is better UX |
| Progress indicator in tray | Rejected — update is fast, not worth the UI complexity |
| Hash/signature verification | Not added — GitHub HTTPS + tag-based release is sufficient for personal tool |
