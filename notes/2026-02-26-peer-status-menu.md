---
title: Peer Status in Tray Menu
created: 2026-02-26
updated: 2026-03-02
tags: [tray, ux, peers]
status: implemented
sections:
  - Peer names from peers.txt
  - ✓/✗ connection indicators
  - Syncthing-style per-peer status labels
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

## Per-Peer Status Labels (2026-03-02)

Replaced the old `synced`/`syncing`/`online`/`offline`/`last seen ...` labels with
Syncthing-style statuses using per-device `NeedSize()` data:

```
✓ Mac - Up to Date
✓ Poco - Syncing
✗ Poco - Disconnected
✗ Poco - Disconnected (42 MB pending)
```

**How it works:** `NeedSize(folderID, deviceID)` returns what a remote device still needs
(queried from local DB, works even when disconnected). `NeedBytes > 0` → Syncing or pending;
`NeedBytes == 0` → Up to Date or plain Disconnected.

**Byte formatting:** KB (min 1) / MB / X.X GB — binary 1024-based, no decimals below GB.

**Reference:** Modeled after Syncthing Android app (`DevicesAdapter.java`) and web GUI
(`syncthingController.js`). Android shows "Up to Date" / "Syncing (X%)" / "Disconnected".
Web GUI adds per-device `needBytes` from `/rest/db/completion`. We use `NeedSize` which
provides the same data from the Go internals.

## Open / Future Ideas

- Clickable peer items (open shared folder, copy peer ID)
