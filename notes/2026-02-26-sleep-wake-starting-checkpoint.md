---
title: Sleep/Wake "Starting..." Investigation Checkpoint
created: 2026-02-26
updated: 2026-02-26
tags: [debug, tray, status, sleep-wake]
status: implemented
---

# Context

User report: after computer sleep/wake, tray status is stuck at `Status: Starting...` and appears to do nothing.

# What I Checked

1. Located tray status logic:
   - `tray/status_monitor.go` computes tray text/icon.
   - It falls back to `Starting...` when GUI address is missing.
   - GUI address resolution depends on parsing `log.txt` for marker: `GUI and API listening on `.

2. Checked runtime process state:
   - `plop` process is running (`PID 761`, uptime about 1h+ during investigation).
   - Listening sockets include:
     - `*:49185` (sync listener)
     - `127.0.0.1:49186` (GUI/API listener)

3. Checked runtime config and logs in home dir:
   - Home: `~/Library/Application Support/plop`
   - `config.xml` has GUI enabled with dynamic address:
     - `<address>127.0.0.1:0</address>`
     - API key present
   - `log.txt` exists but starts with many NUL bytes.
     - Size: 8458 bytes
     - Leading NUL bytes: 2433
     - Missing marker: `GUI and API listening on `

4. Checked live API health directly (using API key from config):
   - `GET /rest/system/ping` => `pong`
   - `GET /rest/db/status?folder=default` => `state: idle`, `needTotalItems: 0`
   - `GET /rest/system/connections` shows one peer connected (`L4ASN6X` true), one disconnected.

5. Checked log timeline around sleep/wake:
   - Saw expected network churn around sleep/wake (resets/no-route), then reconnections.
   - Examples:
     - `Lost primary connection ... no route to host`
     - `connection reset by peer`
     - `Established secure connection ...`
   - No panic/fatal crash evidence for active process.

# Current Conclusion

- Sync engine appears healthy and running.
- Tray `Starting...` is likely a status-detection bug:
  - runtime GUI is up and reachable,
  - but tray parser cannot resolve GUI address because marker line is absent in `log.txt`.

# Relevant Code Paths

- `tray/status_monitor.go`:
  - `computeTrayStatus`
  - `resolveGUIAddress`
  - `readRuntimeGUIAddress`
- `cmd/root.go`:
  - `setupLogFile` truncates and redirects output to `log.txt`
- `cmd/status.go`:
  - also cannot handle dynamic GUI address (`127.0.0.1:0`) and reports daemon not running.

# Suggested Fix Direction (original plan)

1. Make tray status robust to missing log marker:
   - infer GUI port from open localhost listener for current process, or
   - query Syncthing events/system endpoint via a more reliable runtime address source.
2. Optionally improve `plop status` command with the same runtime GUI address resolution fallback.
3. Consider trimming/sanitizing NUL-prefixed log reads before marker scan.

# Follow-up (2026-02-26)

Implemented an engine-side fix so the daemon no longer depends on log parsing to expose the GUI/API endpoint:

- `engine.New()` now calls `ensureRuntimeGUIAddress(&cfg)` before saving config.
- When GUI address is `host:0` (or empty), it allocates a concrete localhost port and persists it in `config.xml`.
- Result: tray status monitor and `plop status` can use `cfg.GUI.RawAddress` directly instead of requiring `log.txt` marker parsing.

Validation performed:
- Added tests in `engine/config_test.go` for:
  - dynamic `127.0.0.1:0` -> concrete port,
  - already-fixed address remains unchanged,
  - wildcard `0.0.0.0:0` normalizes to loopback host.
- `go test -tags noassets ./...` passes.
