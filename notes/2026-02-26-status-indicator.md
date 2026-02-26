---
title: Status Indicator via Tray Icon
created: 2026-02-26
updated: 2026-02-26
tags: [tray, ux]
status: implemented
sections:
  - Blob + anime expression icons — green/yellow/red with ^_^, o_o, >_< eyes
  - 16x16 pixel art source, nearest-neighbor scaled to 22px and 32px
---

# Status Indicator (Implemented)

Show sync status through the tray icon, tooltip, and a disabled menu item.

## Implementation

Blob character with anime expression eyes, three states:
- **Green ^_^** `#22C55E` / `#15803D` outline — idle, synced (`StatusLightSynced`)
- **Yellow o_o** `#F59E0B` / `#B45309` outline — syncing or starting (`StatusLightSyncing`)
- **Red >_<** `#EF4444` / `#DC2626` outline — problem: error, disconnected (`StatusLightAttention`)

Three surfaces, all implemented:
1. **Tray icon** — blob with expression on both platforms (Windows ICO 16+32px, macOS PNG 22px)
2. **Tooltip** — hover text, e.g. "plop - Synced (1/2 peers connected)", "plop - Syncing..."
3. **Menu item** — disabled "Status: ..." text below the version line

### Icon rendering (`icon/status_icon.go`)

- Runtime-generated at startup, no embedded status icon files
- **Source of truth:** 16x16 pixel art grids (`.`=transparent, `O`=outline, `B`=body, `E`=eye), defined as string arrays in Go. Design explored in `notes/plop-icon-pixel-art.html`.
- **Scaling:** Nearest-neighbor from 16x16 to target sizes (22px macOS, 32px Windows). No anti-aliasing — hard pixel edges at all sizes.
- **Windows:** ICO format (16+32px). **macOS:** PNG (22px). Uses `SetIcon` (not `SetTemplateIcon`) so colors render as-is.
- **Static app icon** (`gen_icon.go`): green ^_^ blob, same bitmap approach. Generated via `go generate ./icon`.

#### Design evolution
1. Monochrome traffic light (3 lamps in rectangle) — too detailed, muddy at 16px
2. Colored filled circles — clean but generic, no personality
3. **Blob + anime eyes** (current) — readable at 16px, expressive, memorable brand identity

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
