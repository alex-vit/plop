# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

gosync is a P2P file sync CLI — a "dumbed down Syncthing" that embeds Syncthing's Go libraries (`lib/`) directly as a library, not as a subprocess wrapper. LAN-only discovery, single sync folder, Dropbox-style UX.

## Build & Test

```bash
# Build (noassets tag is REQUIRED — Syncthing GUI assets aren't in the Go module)
go build -tags noassets -o gosync .

# E2E integration test (builds, runs two instances, syncs a file)
bash test_sync.sh

# No unit tests yet — no *_test.go files
```

Requires **Go 1.25+**. Pinned to **Syncthing v1.30.0**.

## Architecture

**Two-layer design:**

- **`cmd/`** — Cobra CLI commands (`init`, `pair`, `run`, `status`, `id`). Each command is a thin shell that delegates to `engine/`.
- **`engine/`** — Wraps `syncthing.App` with three files:
  - `engine.go` — `Engine` struct: `New()` → `Start()` → `Stop()` lifecycle. Uses a `suture.Supervisor` for early services.
  - `cert.go` — TLS cert generation/loading, device ID derivation.
  - `config.go` — Builds Syncthing XML config, `AddPeer()` for pairing.

**Data directory:** `~/.gosync/` (overridable with `--home`), contains `cert.pem`, `key.pem`, `config.xml`, `db/`.

## Critical Gotchas

- **Cert CN must be `"syncthing"`** — BEP protocol validates this during TLS handshake; any other CN causes connection failure.
- **Config wrapper + event logger must run as suture services BEFORE `App.Start()`** — Syncthing's `startup()` calls `cfg.Modify()` which deadlocks if the wrapper's `Serve()` loop isn't active.
- **GUI must be enabled** (bound to `127.0.0.1:0`) for the REST API that `status` uses.
- **Listen addresses use port 0** (`tcp://0.0.0.0:0`, `quic://0.0.0.0:0`) — allows multiple instances; LAN discovery broadcasts actual addresses.
- **`FSWatcherDelayS: 10` must be set explicitly** — Go zero-value may not trigger Syncthing's default.
- Use `syncthing.LoadConfigAtStartup()` to load config (handles migrations and defaults).
- Use `backend.TuningAuto` for `backend.Open()`, not `config.TuningAuto`.
- Syncthing's `lib/` API is not officially stable — stay pinned to a release tag.
