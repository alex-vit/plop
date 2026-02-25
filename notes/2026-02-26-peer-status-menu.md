---
title: Peer Status in Tray Menu
created: 2026-02-26
updated: 2026-02-26
tags: [tray, ux, peers]
status: idea
sections:
  - Peer list in menu with dot indicators and last-seen text
---

# Peer Status in Tray Menu (Idea)

Show connected peers directly in the tray menu with at-a-glance status.

## Concept

A section in the menu listing each peer with a colored dot and status text:

```
Add or Edit Peers
──────────────
🟢 Alex's MacBook        Synced
🟡 Work PC               Syncing...
⚫ Dad's Laptop          Last seen 3d ago
```

Dot states:
- **Green** — online, fully synced
- **Yellow** — online, syncing in progress
- **Red** — online, error/conflict
- **Grey/black** — offline

Last seen text for offline peers: "Last seen 2h ago", "Last seen 3d ago", etc.

## Open Questions

- Where to get peer status? Syncthing events, REST API, or engine-level?
- Peer display names — Syncthing device names vs something from peers.txt?
- Menu items disabled (info only) or clickable (e.g. open shared folder, copy peer ID)?
- How often to refresh? On menu open, or live updates?
- Relationship with the status indicator idea — overall status should be derived from peer states
