# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

plop is a P2P file sync CLI — a "dumbed down Syncthing" that embeds Syncthing's Go libraries (`lib/`) directly as a library, not as a subprocess wrapper. LAN + WAN discovery (global announce, relay, NAT traversal), single sync folder, Dropbox-style UX.

## Build & Test

```bash
# Build a double-clickable macOS app bundle (.app)
./scripts/build-mac-app.sh

# Build plain CLI binary (noassets tag is REQUIRED — Syncthing GUI assets aren't in the Go module)
go build -tags noassets -o plop .

# E2E integration test
go test -tags noassets -v -count=1 -timeout 3m ./engine/
```

Requires **Go 1.25+**. Pinned to **Syncthing v1.30.0**.

## Architecture

**Two-layer design:**

- **`cmd/`** — Cobra CLI commands (`init`, `pair`, `run`, `status`, `id`). Each command is a thin shell that delegates to `engine/`.
  - `root.go` — Default (no subcommand): starts engine + system tray. Auto-inits on first run.
  - `run.go` — Headless/CLI mode (`plop run`): engine only, no tray.
- **`engine/`** — Wraps `syncthing.App` with three files:
  - `engine.go` — `Engine` struct: `New()` → `Start()` → `Stop()` lifecycle. `New()` auto-initializes (certs, config, sync folder) if not already set up.
  - `cert.go` — TLS cert generation/loading (`LoadOrGenerateCert`), device ID derivation.
  - `config.go` — Builds Syncthing XML config, `AddPeer()` for pairing.
- **`tray/`** — System tray UI (Open Plop Folder, Copy My ID, Add or Edit Peers, Open Config Folder, Exit). Double-click opens plop folder on Windows/Linux. `tray.Run()` blocks; `systray.Quit()` unblocks it.
- **`notes/`** — Development notes. "Notes" means these local project notes, not private notes.

**Data directory:** `~/Library/Application Support/plop` (overridable with `--home`), contains `cert.pem`, `key.pem`, `config.xml`, `db/`.

**Startup flow:** `engine.New()` auto-creates data dir, certs, and default config if missing → `engine.Start()` → tray blocks on `systray.Run()` → Exit/signal/engine-exit calls `systray.Quit()` → `defer eng.Stop()` cleans up.

## Critical Gotchas

- **Cert CN must be `"syncthing"`** — BEP protocol validates this during TLS handshake; any other CN causes connection failure.
- **Config wrapper + event logger must run as suture services BEFORE `App.Start()`** — Syncthing's `startup()` calls `cfg.Modify()` which deadlocks if the wrapper's `Serve()` loop isn't active.
- **GUI must be enabled** (bound to `127.0.0.1:0`) for the REST API that `status` uses.
- **Listen addresses use port 0** (`tcp://0.0.0.0:0`, `quic://0.0.0.0:0`, plus relay URL) — allows multiple instances; LAN discovery broadcasts actual addresses, WAN uses global announce + relays.
- **`FSWatcherDelayS: 10` must be set explicitly** — Go zero-value may not trigger Syncthing's default.
- Use `syncthing.LoadConfigAtStartup()` to load config (handles migrations and defaults).
- Use `backend.TuningAuto` for `backend.Open()`, not `config.TuningAuto`.
- Syncthing's `lib/` API is not officially stable — stay pinned to a release tag.
