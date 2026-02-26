---
title: Internal Status Architecture (No Log Parsing)
created: 2026-02-26
updated: 2026-02-26
tags: [architecture, status, tray, syncthing]
status: proposed
---

# Problem

Tray and CLI status should not depend on parsing log output. Logs are non-contractual and can change format, buffering, and timing across versions.

# Goal

Expose sync health from Syncthing internals through stable in-process and local IPC interfaces.

# Proposed Design

1. Introduce an engine-native status service (`engine/status_service.go`) that computes and caches a typed `StatusSnapshot`.
2. Build snapshots from in-process primitives, not logs:
   - `app.Internals.FolderState(folderID)`
   - `app.Internals.DBSnapshot(folderID)` + `NeedSize(protocol.LocalDeviceID)`
   - `app.Internals.IsConnectedTo(deviceID)` for configured peers
   - `cfgWrapper.RawCopy()` for current folders/devices
3. Trigger refreshes event-driven via `evLogger.Subscribe(...)` and periodic safety polling.
4. Expose:
   - `Engine.StatusSnapshot() StatusSnapshot`
   - `Engine.StatusUpdates() <-chan StatusSnapshot`
5. Update tray to consume `Engine.StatusUpdates()` directly.
6. Add optional local IPC status endpoint for out-of-process clients (`plop status`), either:
   - Unix socket / named pipe, or
   - atomically written `status.json` heartbeat in `--home`.

# Data Model

`StatusSnapshot` fields:
- `State`: `starting|syncing|synced|waiting_peers|error|unavailable`
- `FolderID`
- `FolderState`
- `NeedTotalItems`
- `ConnectedPeers`
- `TotalPeers`
- `Error`
- `UpdatedAt`

# Rollout Plan

1. Remove all log parsing paths (done for tray monitor).
2. Add engine status service and wire tray to it.
3. Keep REST-based `plop status` as fallback for one release.
4. Add IPC-backed `plop status` and prefer it over REST.
5. Remove legacy REST polling in tray.

# First Steps (Implementation Plan)

## PR 1: Add engine-native status service (no behavior change yet)

1. Add `engine/status_snapshot.go`:
   - define `StatusState` constants (`starting`, `syncing`, `synced`, `waiting_peers`, `error`, `unavailable`)
   - define `StatusSnapshot` struct from this doc.
2. Add `engine/status_service.go`:
   - keep latest snapshot in memory (`sync.RWMutex`)
   - expose `Snapshot()` and `Updates()` (buffered channel, drop-old-on-slow-consumer)
   - compute snapshot using:
     - `cfgWrapper.RawCopy()`
     - `app.Internals.FolderState(folderID)`
     - `app.Internals.DBSnapshot(folderID)` + `NeedSize(protocol.LocalDeviceID)`
     - `app.Internals.IsConnectedTo(deviceID)`
   - run refresh loop with:
     - periodic polling (for example every 2-4s) as safety net
     - event-triggered refresh via `evLogger.Subscribe(...)`.
3. Wire into `engine/engine.go`:
   - store `statusSvc` on `Engine`
   - start it in `Start()` after `app.Start()`
   - stop it in `Stop()`
   - add `Engine.StatusSnapshot()` and `Engine.StatusUpdates()`.
4. Tests:
   - add table-driven mapping tests for state derivation (pure function).
   - add service test that verifies update publication and monotonic `UpdatedAt`.

## PR 2: Move tray from REST polling to engine updates

1. Change tray entrypoint to accept a status source from engine (instead of only `homeDir`).
2. Replace `tray/status_monitor.go` REST polling with snapshot subscription:
   - subscribe once on startup
   - map `StatusSnapshot` -> tray title/tooltip/icon.
3. Keep current strings/icons stable to minimize UX churn.
4. Regression test manually:
   - start app, edit `peers.txt`, disconnect/reconnect peer, verify tray transitions without REST failures.

## PR 3: Prepare `plop status` migration (keep fallback)

1. Keep current REST output path in `cmd/status.go`.
2. Add internal-status read path (IPC or heartbeat file) behind feature detection.
3. Prefer internal status when present; fall back to REST for one release.

## Scope Guardrails

1. Do not add IPC transport in PR 1.
2. Do not remove REST fallback for CLI in this iteration.
3. Prioritize tray correctness first, then CLI migration.

# Test Plan

1. Unit tests for status mapping:
   - idle + `NeedTotalItems=0` -> `synced`
   - idle + `NeedTotalItems>0` -> `syncing`
   - peers configured + none connected -> `waiting_peers`
   - folder error state -> `error`
2. Integration test:
   - start engine in temp home
   - verify status snapshot transitions from `starting` to stable state
3. Regression test:
   - status must work with empty or malformed `log.txt`

# Acceptance Criteria

1. No code path reads `log.txt` for status decisions.
2. Tray status continues to update through sleep/wake cycles.
3. `plop status` works even if logs are disabled or rotated.
