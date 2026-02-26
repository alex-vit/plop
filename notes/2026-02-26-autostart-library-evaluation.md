---
title: Autostart Library Evaluation
created: 2026-02-26
updated: 2026-02-26
tags: [tray, startup, architecture]
status: decided
sections:
  - Evaluate Go autostart libraries vs current in-repo implementation
---

# Autostart Library Evaluation (Decision)

## Context

Revisit the hand-rolled `autostart/` package and decide whether to replace it with a library.

Constraint: avoid CGO unless there is a strong reason.

## Investigated

- `github.com/emersion/go-autostart` (latest pseudo-version timestamp: 2025-04-03)
  - Windows backend uses CGO (`autostart_windows.c`) and Startup shortcut creation.
  - Open upstream reports around Windows cross-compile friction and argument handling.
- `github.com/ProtonMail/go-autostart` (latest pseudo-version timestamp: 2026-02-10)
  - Active fork; includes a quoting-side-effect fix vs emersion.
  - Still uses CGO on Windows and Startup shortcut approach.
- `github.com/spiretechnology/go-autostart/v2` (`v2.0.0`, 2023-08-09)
  - Broader scope (system mode, service helpers, log redirection), heavier than needed.
  - Windows path still depends on CGO shortcut creation.
- `github.com/jimbertools/go-autostart` (fork snapshot timestamp: 2023-12-20)
  - Uses pure-Go Windows registry (`HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run`), no CGO.
  - Low-activity fork and still basic Windows argument joining/quoting.

## Decision

Keep the in-repo autostart implementation for now:

- `autostart/autostart_darwin.go` (LaunchAgent plist)
- `autostart/autostart_windows.go` (HKCU Run registry value)
- `autostart/autostart_other.go` (explicit unsupported behavior outside macOS/Windows)

## Why

- Meets the no-CGO preference.
- Already matches current behavior needs, including `--home` argument support.
- No external library clearly improves correctness/maintenance enough to justify migration risk.

## Revisit Triggers

Re-evaluate this decision if:

- a well-maintained no-CGO library with robust Windows argument quoting emerges,
- Linux autostart support is added and shared abstraction becomes valuable,
- startup scope expands to system service management.
