# Plop

Dead-simple P2P file sync. Like Dropbox, but no cloud — files go directly between your machines over LAN or internet.

[Download v2.1.0](https://github.com/alex-vit/plop/releases/tag/v2.1.0) — **plop-setup.exe** (Windows installer) or **Plop.app** (macOS)

<img width="273" height="307" alt="Screenshot 2026-03-03 at 10 05 16" src="https://github.com/user-attachments/assets/8ded0522-d177-47a8-a9de-ed14d31e94d3" />

## How it works

Plop runs in your system tray. Pick a folder, share your ID with a friend, and files sync automatically. No accounts, no servers, no configuration.

- **Peer-to-peer** — connects directly over LAN or punches through NAT; falls back to relays when needed
- **System tray app** — lives in your menubar/taskbar, stays out of the way
- **Zero config** — just pair device IDs and go
- **Cross-platform** — macOS and Windows

Built on [Syncthing](https://syncthing.net/)'s battle-tested sync engine.

## Build

```bash
go build -tags noassets -o plop .
```
