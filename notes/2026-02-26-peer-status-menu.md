---
title: Peer Status in Tray Menu
created: 2026-02-26
updated: 2026-02-28
tags: [tray, ux, peers]
status: implemented
sections:
  - Peer names from peers.txt
  - ✓/✗ connection indicators
---

# Peer Status in Tray Menu

Show connected peers directly in the tray menu with at-a-glance status.

## Implemented (2026-02-28)

Each peer appears as a menu item with a `✓`/`✗` prefix and a display name:

```
✓ Mac
✗ Poco
```

**Why `✓`/`✗`:** `●`/`○` were used originally but are visually similar (differ only in fill).
`[✓]`/`[ ]` bracket style was considered but `[` renders as a broken char in PowerShell terminals
(though Win32 tray menus are Unicode-safe). `✓`/`✗` is maximally distinct and renders correctly
in both tray menus and common terminals.

**Alternatives considered:**
- `●`/`○` — original; too visually similar at menu font size
- `[✓]`/`[ ]` — clear, but `[` breaks in some terminals
- `(online)`/`(offline)` suffix — unambiguous but verbose
- `+`/`-` prefix — pure ASCII, works everywhere; chose `✓`/`✗` for clarity

## Peer Names

Names come from `peers.txt` via two supported formats:

```
# MacBook
NVUIHRB-CAIDSJU-...

L4ASN6X-XR7BHYQ-... Poco
```

- **Comment-before-ID**: a `# Name` line directly above a device ID line sets its name.
  A blank line between them discards the comment (plain comment).
- **Inline after ID**: `DEVICE_ID  Name` on the same line; inline takes precedence over a comment.

Names are stored in `config.DeviceConfiguration.Name` (Syncthing's own field) and propagated
through `PeerStatus.Name`. The tray falls back to `ShortID` when `Name` is empty.

`syncPeersConfig` updates the name in config whenever peers.txt changes (live reload).

## Open / Future Ideas

- Last-seen text for offline peers: "Last seen 2h ago"
- Per-peer sync state (synced vs syncing vs error) — needs Syncthing per-device folder state
- Clickable peer items (open shared folder, copy peer ID)
