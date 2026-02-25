---
title: Start on Login
created: 2026-02-26
updated: 2026-02-26
tags: [tray, ux, startup]
status: idea
sections:
  - Tray menu checkbox to toggle launch at login, per-platform mechanisms
---

# Start on Login (Idea)

Add a toggleable "Start on Login" / "Start with Windows" menu item in the tray, like MonitorBright does.

## Concept

A checkable menu item in the tray menu. When enabled, plop launches automatically on login/boot.

## Per-platform mechanisms

- **macOS** — Launch Agent plist in `~/Library/LaunchAgents/`
- **Windows** — Registry key in `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, or shortcut in Startup folder
- **Linux** — `.desktop` file in `~/.config/autostart/`

## Open Questions

- Where to persist the setting? Detect from OS state (plist/registry exists) or store in plop config?
- Label: "Start on Login" (macOS), "Start with Windows" (Windows), or just "Start at Login" everywhere?
- Need to know the exe path at runtime to register — `os.Executable()` should work
