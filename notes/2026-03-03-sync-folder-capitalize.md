# Capitalize default names: plop → Plop

**Date:** 2026-03-03

## What changed

The default sync folder was renamed from `~/plop` to `~/Plop` so it looks better in file managers (Finder, Explorer).

## Migration

Existing installs have their folder path stored as an absolute path in `config.xml`. On startup, `engine.New()` calls `migrateFolderName()` which:

1. Checks if the folder path's basename is `"plop"` (lowercase).
2. If so, renames the directory on disk via `os.Rename` (old → new).
3. Updates the config path to use `"Plop"`.

On case-insensitive filesystems (macOS HFS+/APFS default, Windows NTFS), `os.Rename` just changes the display name. On case-sensitive filesystems (Linux ext4) it's a real rename. If the rename fails for any reason, the old path is kept — the app still works fine.

The folder ID (`"default"`) is unchanged, so peer connectivity is unaffected.

## Config directory rename

The config/data directory was also renamed from `plop` to `Plop` (e.g. `~/Library/Application Support/Plop` on macOS). `paths.MigrateConfigDir()` handles existing installs:

1. Derives the old path by replacing the basename of `ConfigDir()` with `"plop"`.
2. If the old dir exists, renames it via `os.Rename`. On case-insensitive FS (macOS, Windows) this just changes the display name. On case-sensitive FS (Linux) it's a real rename.
3. If both old and new exist as separate directories (case-sensitive FS edge case), skips the rename to avoid data loss.

Called from `cli.go`'s `run()` only when using the default home path — custom `--home` paths are never touched.

The Windows installer (`installer.iss`) already uses `{localappdata}\Plop`, so this makes the code match.

## Files changed

Sync folder:
- `engine/config.go` — added `migrateFolderName()` function
- `engine/engine.go` — changed default from `"plop"` to `"Plop"`, calls migration after loading config
- `cmd_init.go` — default flag value `"plop"` → `"Plop"`
- `cmd_pair.go` — fallback path `"plop"` → `"Plop"`
- `engine/config_test.go` — migration tests (rename existing dir, no dir on disk, already capitalized, custom path, no folders)

Config directory:
- `paths/configdir_unix.go` — `"plop"` → `"Plop"`
- `paths/configdir_windows.go` — `"plop"` → `"Plop"`
- `paths/migrate.go` — new `MigrateConfigDir()` + `migrateConfigDir()` helper
- `paths/migrate_test.go` — tests: old exists → renamed, old missing → no-op, both exist → skip
- `cli.go` — calls `paths.MigrateConfigDir()` after flag parsing when using default home
