# Windows Explorer Launch Blocked — Investigation

**Date:** 2026-02-27
**Status:** Resolved

## Symptom

Double-clicking `plop.exe` from Windows Explorer (File Explorer, Start Menu shortcut, or directly in `%LocalAppData%\plop\`) shows a brief loading cursor (1-2 seconds) then nothing happens. The process never starts. The autostart registry entry (`HKCU\...\Run`) also fails at boot for the same reason — Explorer processes Run entries.

## What Works

| Launch method               | Result |
|-----------------------------|--------|
| Bash / terminal direct      | Works  |
| `cmd /c start "" "path"`   | Works  |
| PowerShell `Start-Process`  | Works  |
| PowerShell `ShellExecute` P/Invoke | Works |
| `.bat` wrapper double-click | Works  |
| Explorer double-click       | **Fails silently** |
| Start Menu shortcut         | **Fails silently** |
| Right-click → Run as Admin  | **Fails silently** |

## What Was Ruled Out

| Mechanism | How verified |
|-----------|-------------|
| PE subsystem | Both plop (GUI/2) and monibright (GUI/2) confirmed via PE header |
| Digital signature | Both unsigned — monibright works, plop doesn't |
| Zone.Identifier (MOTW) | No ADS on plop.exe |
| Windows Defender detections | No threats reported; `Get-MpThreat` empty |
| Defender exclusions | Added path + process exclusions (confirmed in list) — still blocked |
| SmartScreen | UAC is disabled → SmartScreen is effectively off. Registry value `SmartScreenEnabled` is null |
| AppLocker | No rules configured |
| Software Restriction Policies | No policies |
| Image File Execution Options | No IFEO entry for plop.exe |
| Attack Surface Reduction rules | None configured |
| Exploit Protection per-app | No overrides for plop.exe |
| PCA (Program Compatibility Assistant) | Deleted all plop entries from Store — no effect |
| Third-party AV | Only Windows Defender installed |
| `.rsrc` PE section from rsrc.exe | Rebuilt without rsrc (16 sections, no .rsrc) — still blocked |

## Key Observations

- **log.txt is NOT modified** after failed Explorer launch → process never reaches `main()` / Go runtime never initializes
- **monibright.exe** (9 MB, unsigned Go GUI exe) launches fine from double-click
- **minimal-test.exe** (~2 MB, trivial Go GUI exe) launches fine from double-click
- **plop.exe** (~26 MB, embeds Syncthing libraries) does NOT launch from double-click
- A `.bat` wrapper that runs `start "" "plop.exe"` works fine from double-click → Explorer blocks the `.exe` specifically, not the process creation

## Environment

- Windows 10 Home 10.0.19045
- UAC disabled
- Go 1.25, built with `-tags noassets -ldflags "-H=windowsgui"`
- Syncthing v1.30.0 embedded as library

## Root Cause

**Cobra's mousetrap feature.** Cobra includes `github.com/inconshreveable/mousetrap` which calls `mousetrap.StartedByExplorer()` to detect if the parent process is `explorer.exe`. If so, it prints "This is a command line tool. You need to open cmd.exe and run it from there." and exits after 5 seconds (`MousetrapDisplayDuration`).

For a `-H=windowsgui` app, there's no console — the message is invisible and the app just silently exits. This also blocks the registry Run key autostart, because `explorer.exe` processes those entries too.

MoniBright worked because it doesn't use cobra.

## Fix

```go
func init() {
    cobra.MousetrapHelpText = "" // Allow launching from Explorer (GUI app).
}
```

## Key Lesson

Building a console version (`go build` without `-H=windowsgui`) immediately revealed the error message. One web search with the exact message text in quotes found the answer. The hours spent checking SmartScreen, Defender, AppLocker, ASR, etc. were unnecessary.
