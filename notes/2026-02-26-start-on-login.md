---
title: Start on Login
created: 2026-02-26
updated: 2026-02-26
tags: [tray, ux, startup, installer]
status: implemented
sections:
  - Tray menu checkbox to toggle launch at login, per-platform mechanisms
  - Windows installer autostart option and clean uninstall
---

# Start on Login (Implemented)

Toggleable "Start on Login" / "Start with Windows" menu item in the tray, plus installer integration on Windows.

## Implementation

- **macOS** — Launch Agent plist in `~/Library/LaunchAgents/com.alexvit.plop.plist`. Tray menu label: "Start on Login".
- **Windows** — Registry key `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, value name `Plop`. Tray menu label: "Start with Windows".
- **Linux** — Not supported (stub returns unsupported, menu item hidden).
- Setting is detected from OS state (plist/registry exists), not stored in plop config.
- Exe path resolved at runtime via `os.Executable()`.

## Windows Installer Integration

The Inno Setup installer (`installer.iss`) mirrors the app's autostart:

- **"Start with Windows" checkbox** — `[Tasks]` section offers the option during install.
- **Same registry key** — `[Registry]` section writes the same `HKCU\...\Run\Plop` value the tray menu uses. `uninsdeletevalue` flag cleans it up on uninstall.
- **Clean uninstall** — `[UninstallDelete]` removes the entire `%LocalAppData%\Plop` directory (certs, config, db, logs are all regenerable; the sync folder at `~/plop` is separate and untouched).

The installer and the app's tray toggle stay in sync since they read/write the same registry entry.

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
