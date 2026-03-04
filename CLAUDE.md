# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

plop is a P2P file sync CLI — a "dumbed down Syncthing" that embeds Syncthing's Go libraries (`lib/`) directly as a library, not as a subprocess wrapper. LAN + WAN discovery (global announce, relay, NAT traversal), single sync folder, Dropbox-style UX.

## Build & Test

```bash
# Build plain CLI binary (noassets tag is REQUIRED — Syncthing GUI assets aren't in the Go module)
go build -tags noassets -o plop .

# macOS app bundle → out/Plop.app
./scripts/build-mac-app.sh

# Windows portable exe + Inno Setup installer → out/Plop.exe + out/plop-setup.exe
pwsh ./scripts/build-windows-release.ps1 -Version vX.Y.Z

# All tests (unit + E2E)
go test -tags noassets ./...

# Unit tests only (skip slow E2E sync test)
go test -tags noassets -short ./...

# E2E integration test only (~30s, two engines sync a file over LAN)
go test -tags noassets -v -count=1 -timeout 3m ./engine/

# Regenerate icon assets (icon.png + icon.ico)
go generate ./icon
```

Requires **Go 1.25+**. Pinned to **Syncthing v1.30.0**.

Version is injected via `-ldflags "-X main.version=<tag>"` (build scripts do this automatically). Without it, displays `"dev"`.

## Architecture

**Two-layer design:**

- **Root package (`main`)** — CLI dispatcher + commands (`init`, `pair`, `run`, `status`, `id`). Each command is a thin shell that delegates to `engine/`. Uses stdlib `flag` for parsing.
  - `cli.go` — `run()` dispatcher, `--home` global flag, default command (starts engine + system tray), `setupLogFile`, `printUsage`.
  - `cmd_run.go` — Headless/CLI mode (`plop run [folder]`): engine only, no tray. Optional `--peer` flags. When a folder arg is given, uses a per-folder isolated instance (SHA-256 hash → `instances/<4hex>/`).
  - `cmd_init.go`, `cmd_pair.go`, `cmd_status.go`, `cmd_id.go` — Subcommand handlers.
  - `update_*.go` — Auto-update (platform-specific). `redirect_*.go` — fd/handle redirection.
- **`engine/`** — Wraps `syncthing.App`:
  - `engine.go` — `Engine` struct: `New()` → `Start()` → `Stop()` lifecycle. `New()` auto-initializes (certs, config, sync folder) if not already set up.
  - `cert.go` — TLS cert generation/loading (`LoadOrGenerateCert`), device ID derivation.
  - `config.go` — Builds Syncthing XML config, `AddPeer()` / `RemovePeer()` for pairing, `ensureRuntimeGUIAddress()` pre-allocates a concrete port before Syncthing starts.
  - `peers.go` — `peers.txt` file parsing/writing. `watchPeers()` uses `syncthing/notify` to live-watch the file; edits take effect without restart via `syncPeers()` reconciliation (adds missing, removes stale).
  - `status_service.go` — Event-driven + polling status computation, publishes `StatusSnapshot` on an in-process channel for tray.
  - `status_file.go` — Atomic JSON heartbeat to `status.json` every 3s for `plop status` CLI to read out-of-process. Deleted on clean stop; rejected as stale if >15s old.
  - `status_snapshot.go` — `StatusState` constants + `StatusSnapshot` struct.
- **`tray/`** — System tray UI (Open Plop Folder, Copy My ID, Add or Edit Peers, Open Config Folder, Exit). Double-click opens plop folder on Windows/Linux. `tray.Run()` blocks; `systray.Quit()` unblocks it.
  - `status_monitor.go` — Reads `StatusSnapshot` channel, updates tray icon (traffic-light) and tooltip.
- **`paths/`** — Platform-specific config directory: macOS `~/Library/Application Support/Plop`, Windows `%LocalAppData%\Plop` (not `%AppData%`), Linux `~/.config/Plop`. `RobustRename` helper for Windows retry logic.
- **`autostart/`** — macOS LaunchAgent plist, Windows registry (`HKCU\...\Run`). Linux stub (unsupported, menu item hidden).
- **`icon/`** — Embedded `icon.png`/`icon.ico` via `//go:embed`. `status_icon.go` generates traffic-light status icons at runtime (binary-opaque pixels, no anti-aliasing).
- **`notes/`** — Development notes. "Notes" means these local project notes, not private notes.

**Data directory** (overridable with `--home`): macOS `~/Library/Application Support/Plop`, Windows `%LocalAppData%\Plop`, Linux `~/.config/Plop`. Contains `cert.pem`, `key.pem`, `config.xml`, `peers.txt`, `status.json`, `log.txt` (GUI mode), `db/`.

**Startup flow:** `engine.New()` auto-creates data dir, certs, and default config if missing → early suture supervisor starts config wrapper + event logger → `engine.Start()` → tray blocks on `systray.Run()` → Exit/signal/engine-exit calls `systray.Quit()` → `defer eng.Stop()` cleans up.

**Log redirection (GUI mode only):** Unix uses `syscall.Dup2` to redirect fd 1+2; Windows uses `kernel32.dll SetStdHandle`. This captures output from libraries that grabbed `os.Stderr` at init time.

## Release

```bash
# Create + push annotated git tag, optionally watch GH Actions
./scripts/release-tag.sh v1.0.0 [--watch]

# Rebuild + restart macOS app for local iteration
./scripts/restart-mac-app.sh [--no-build]
```

CI (`.github/workflows/release.yml`): tag push triggers parallel Windows + macOS builds → GitHub Release with artifacts.

## Critical Gotchas

- **Cert CN must be `"syncthing"`** — BEP protocol validates this during TLS handshake; any other CN causes connection failure.
- **Config wrapper + event logger must run as suture services BEFORE `App.Start()`** — Syncthing's `startup()` calls `cfg.Modify()` which deadlocks if the wrapper's `Serve()` loop isn't active.
- **GUI must be enabled** (bound to `127.0.0.1:0`) for the internal REST API. `ensureRuntimeGUIAddress()` pre-allocates a concrete port to avoid stale-address bugs after sleep/wake.
- **Listen addresses use port 0** (`tcp://0.0.0.0:0`, `quic://0.0.0.0:0`, plus relay URL) — allows multiple instances; LAN discovery broadcasts actual addresses, WAN uses global announce + relays.
- **`FSWatcherDelayS: 10` must be set explicitly** — Go zero-value may not trigger Syncthing's default.
- Use `syncthing.LoadConfigAtStartup()` to load config (handles migrations and defaults).
- Use `backend.TuningAuto` for `backend.Open()`, not `config.TuningAuto`.
- Syncthing's `lib/` API is not officially stable — stay pinned to a release tag.
- **Windows installer uses `%LocalAppData%\Plop`** — must match the `paths/` package; `%AppData%` (roaming) is wrong.
- **Windows directory renames can fail** with "Access is denied" when file handles are open — `paths.RobustRename` retries with backoff. `Engine.Stop()` retries the sync folder rename after Syncthing releases handles.
