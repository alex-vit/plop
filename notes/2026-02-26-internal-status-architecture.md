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
