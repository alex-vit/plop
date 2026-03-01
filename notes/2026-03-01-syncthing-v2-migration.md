# Syncthing v2 Migration Scope

## Summary

Migrating plop from Syncthing v1.30.0 to v2.0.14. This fixes the quic-go v0.52.0 panic
(`crypto/tls bug: where's my session ticket?`) that crashes plop when built with Go 1.26,
and unblocks the Go 1.26 upgrade.

## Motivation

- **quic-go panic**: Go 1.26 changed `crypto/tls` session ticket behavior; quic-go v0.52.0
  (pinned by syncthing v1.30.0) panics. v2.0.14 uses quic-go v0.56.0 (fix was in v0.53.0).
- **Go 1.26 unblocked**: With v2, can drop the `GOTOOLCHAIN=go1.25.0` workaround.
- **Syncthing v1 is EOL**: v1.30.0 was the last v1 release (July 2025). v2.0 released August 2025.

## Import Method

Syncthing v2 does **not** follow Go module versioning (no `/v2` suffix on module path).
Must import via pseudo-version pointing to a commit hash:

```
github.com/syncthing/syncthing v1.30.0-rc.1.0.20260202043224-b40f2acdad80
```

This resolves to the v2.0.14 tag commit (`b40f2ac`). Forum thread:
https://forum.syncthing.net/t/go-module-path-for-syncthing-lib-remains-at-v1-unable-to-import-v2/25532

## Code Changes Required (all verified — compiles and tests pass)

### 1. `engine/engine.go` — database + config API

**`lib/db/backend` removed.** The entire package is gone (DB moved to `internal/db/sqlite`).
Replace `backend.Open(path, backend.TuningAuto)` → `syncthing.OpenDatabase(path, 0)`.
The second arg is delete-retention duration; 0 uses the default (15 months).

**`LoadConfigAtStartup` lost a parameter.** The `noDefaultFolder` bool (5th arg) was removed.
6 args → 5 args: `LoadConfigAtStartup(path, cert, evLogger, allowNewerConfig, skipPortProbing)`.

### 2. `engine/cert.go` — new `compatible` parameter

`tlsutil.NewCertificate` gained a 5th bool `compatible`. Pass `false` for Ed25519 certs
(standard for sync connections). `true` would generate RSA for compatibility with older peers.

### 3. `engine/status_service.go` — DBSnapshot removed

`Internals.DBSnapshot(folderID)` is gone. Use `Internals.NeedSize(folder, device)` directly.
Returns `(Counts, error)` with `.TotalItems()` method. Simpler — no snapshot lifecycle to manage.

### 4. `cmd/root.go` — logger package deleted

`lib/logger` is entirely removed in v2. Syncthing v2 uses `log/slog` (Go stdlib).
Replace `stlogger.DefaultLogger.AddHandler(...)` with `slog.SetDefault(...)` using a
`slog.TextHandler` writing to the log file.

### 5. `go.mod` — dependency changes

- syncthing: v1.30.0 → pseudo-version (v2.0.14 commit)
- quic-go: v0.52.0 → v0.56.0 (indirect, pulled by syncthing)
- New transitive deps: modernc.org/sqlite, jmoiron/sqlx (SQLite DB backend)
- CGO not required — syncthing v2 has `db_open_nocgo.go` using modernc.org/sqlite (pure Go)

## Files Changed

| File | Change |
|------|--------|
| `engine/engine.go` | Remove `lib/db/backend` import; use `syncthing.OpenDatabase()`; fix `LoadConfigAtStartup` args |
| `engine/cert.go` | Add `false` arg to `NewCertificate` |
| `engine/status_service.go` | Replace `DBSnapshot` + `snap.NeedSize()` with `Internals.NeedSize()` |
| `cmd/root.go` | Replace `lib/logger` with `log/slog` |
| `go.mod` / `go.sum` | Updated dependencies |

## Runtime Considerations

### Database Migration (LevelDB → SQLite)

Syncthing v2 switched from LevelDB to SQLite. On first start with a v1 `db/` directory,
syncthing will automatically migrate. This may be slow for large databases.

`syncthing.TryMigrateDatabase()` exists as a public function if we want explicit control,
but the default auto-migration on `OpenDatabase()` should work.

### Protocol Compatibility

v2 peers use multiple connections by default (3 per peer). v2 is protocol-compatible with v1,
but some users have reported edge cases. The Mac peer running plop v1 should still work during
rollout — just upgrade both sides.

### Go Version

After migration, the `go 1.25.0` directive can be bumped to `go 1.26.0` and the
`GOTOOLCHAIN=go1.25.0` workarounds in build scripts can be removed. This also unblocks
`go fix ./...` (the `auto.Assets` blocker from the v1 dependency is gone).

## Alternatives Considered

| Alternative | Pros | Cons |
|-------------|------|------|
| **Pin GOTOOLCHAIN=go1.25.0** (current workaround) | Zero code changes | Fragile; v1 is EOL; doesn't fix root cause |
| **Patch quic-go via replace directive** | Targeted fix | quic-go v0.53+ has breaking API; can't replace cleanly |
| **Upgrade to Syncthing v2** (this plan) | Fixes root cause; unblocks Go 1.26; future-proof | Migration effort; pseudo-version import is ugly |

## Effort

~30 minutes of code changes (already done in worktree `syncthing-v2-migration`).
Main risk is runtime behavior — needs manual testing of:
- Fresh init (new user)
- Existing user with LevelDB database (migration path)
- Peer connectivity (both plop↔plop and plop↔Syncthing Android)
