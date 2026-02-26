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

## Future Polish: macOS Login Items Icon

- Current behavior: Login Items shows `plop` with generic `exec` icon and "Item from unidentified developer".
- What was tried:
  - Added `AssociatedBundleIdentifiers` to LaunchAgent plist (string, then array form).
  - Regenerated app icon as `.icns` and wired `CFBundleIconFile` to `AppIcon.icns`.
  - Fully removed and re-registered the launch agent entry.
- Result: entry still appears under `Unknown Developer` in `sfltool dumpbtm`, without associated bundle IDs.
- Likely reason: legacy LaunchAgent + unsigned ad-hoc bundle is not enough for Settings to attribute icon/developer identity.
- Follow-up options:
  - Sign/notarize app bundle (Developer ID) so Background Items can attribute to a real developer/app identity.
  - Evaluate moving from raw legacy LaunchAgent wiring to ServiceManagement (`SMAppService`) for modern Login Items integration.
