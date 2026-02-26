---
title: Status Indicator via Tray Icon
created: 2026-02-26
updated: 2026-02-26
tags: [tray, ux]
status: implemented
sections:
  - Traffic light concept — green/yellow/red icon + status menu item
  - Colorful circle icons on both Windows (ICO 16+32px) and macOS (PNG 22px)
---

# Status Indicator (Implemented)

Show sync status through the tray icon, tooltip, and a disabled menu item.

## Implementation

Traffic light approach with three states:
- **Green** `#22C55E` — idle, everything synced (`StatusLightSynced`)
- **Yellow** `#EAB308` — syncing in progress or starting (`StatusLightSyncing`)
- **Red** `#EF4444` — problem: disconnected, conflict, error (`StatusLightAttention`)

Three surfaces, all implemented:
1. **Tray icon** — colored filled circle on both platforms (Windows ICO 16+32px, macOS PNG 22px)
2. **Tooltip** — hover text, e.g. "plop - Synced (1/2 peers connected)", "plop - Syncing..."
3. **Menu item** — disabled "Status: ..." text below the version line

### Icon rendering (`icon/status_icon.go`)

- Runtime-generated at startup, no embedded status icon files
- **Windows:** Colored filled circles in ICO format (16+32px). Hard pixel edges, no anti-aliasing (clean at tray scale). Inspired by monibright's `icon/gen.go` approach.
- **macOS:** Colored filled circles as PNG (22px per menu bar convention). Uses `SetIcon` (not `SetTemplateIcon`) so colors render as-is, like Zoom.app.
- Previous approach (monochrome traffic light with 3 lamps in a rectangle) was too detailed for 16/32px and looked muddy.

### Status pipeline (`engine/status_service.go` → `tray/status_monitor.go`)

- Engine computes `StatusSnapshot` via event-driven + 3s polling hybrid
- Snapshots flow via channel to tray's status monitor goroutine
- `trayStatusFromSnapshot()` maps `StatusState` → title + tooltip + `StatusLight`
- `setTrayIcon()` picks PNG or ICO per platform

## Resolved Questions

- **Status source:** Syncthing event subscription + polling fallback (`status_service.go`)
- **Icon format:** `SetIcon` with colored icons on both platforms — ICO (16+32px) on Windows, PNG (22px) on macOS. Not `SetTemplateIcon` (that strips color).
- **Granularity:** Three broad states. Starting maps to yellow (syncing). Waiting-for-peers maps to red (attention).
- **Menu item details:** Tooltip includes peer counts; menu item shows state only.
