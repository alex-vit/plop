---
title: Status Indicator via Tray Icon
created: 2026-02-26
updated: 2026-02-26
tags: [tray, ux]
status: idea
sections:
  - Traffic light concept — green/yellow/red icon + status menu item
---

# Status Indicator (Idea)

Show sync status through the tray icon and a disabled menu item at the top.

## Concept

Traffic light approach:
- **Green** — idle, everything synced
- **Yellow** — syncing in progress
- **Red** — problem (disconnected, conflict, error)

Three surfaces:
1. **Tray icon** — swap icon/color to reflect current state
2. **Tooltip** — hover text on the tray icon, e.g. "plop — Synced", "plop — Syncing 3 files...", "plop — Not connected"
3. **Menu item** — disabled text below the version line with the same status text

## Open Questions

- Where to get status from? Syncthing events? Polling the REST API? Engine-level callback?
- Icon format — currently using `SetTemplateIcon` on macOS (monochrome, auto light/dark). Switch to `SetIcon` for color, like Zoom does. Colors need to contrast on both light and dark menu bars — traffic light colors (green/yellow/red) should work.
- How granular? Just the three states, or also "connecting", "scanning", etc.?
- Should the menu item show details (file count, peer count) or just the state?
