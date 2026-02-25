# Plan: System Tray GUI (MVP)

## Context

plop is CLI-only. We want a system tray presence so the app can run as a background daemon with a menu bar icon. MVP scope: tray icon, disabled version label, Exit menu item. No engine/sync integration yet.

## Library

**`github.com/energye/systray`** — same library used in [monibright](https://github.com/alex-vit/monibright). Cross-platform (macOS/Windows/Linux), removed GTK dependency. Callback-based API (`item.Click(fn)`) rather than channel-based. Requires CGo (on by default for native builds).

## Patterns borrowed from monibright

- **Version**: `var version = ""` in `cmd` package + `displayVersion()` returning `"dev"` when empty; injected via `-X github.com/alex-vit/plop/cmd.Version=v0.1.0` ldflags
- **Icon embedding**: dedicated `icon/` package with `//go:embed` and a `go:generate` generator script
- **Disabled title item**: `mTitle := systray.AddMenuItem("plop "+version, ""); mTitle.Disable()`
- **Callback click handlers**: `item.Click(func() { ... })` instead of channels
- **Quit helper**: `addQuit()` function for the Exit menu item

## Files to create

### `icon/gen_icon.go` — icon generator (run via `go generate`)

Standalone `main` in a `//go:build ignore` file. Draws a simple sync-arrows symbol:
- 22x22 PNG, black on transparent (macOS template icon convention)
- Outputs `icon/icon.png`
- Run once: `go generate ./icon`

### `icon/icon.go` — embedded icon bytes

```go
//go:generate go run gen_icon.go
//go:embed icon.png
var Data []byte
```

### `tray/tray.go` — tray lifecycle

```go
func Run(version string)      // blocks on main goroutine; calls systray.Run
func onReady(version string)  // sets icon, adds disabled version label + Exit
func onExit()                 // cleanup stub
```

- `SetTemplateIcon(icon.Data, icon.Data)` — macOS auto-adapts to light/dark; other platforms use as-is
- `SetTooltip("plop")` — no `SetTitle` (avoids text next to icon on macOS)
- Disabled `"plop " + displayVersion(version)` menu item
- Separator
- `"Exit"` item with `item.Click(func() { systray.Quit() })`

### `cmd/gui.go` — cobra command

Thin shell: `tray.Run(Version)`. Follows existing `cmd/` → package delegation pattern.

## Files to modify

### `cmd/root.go` — add version variable

```go
var Version = ""
```

Overridable via `-X github.com/alex-vit/plop/cmd.Version=v0.1.0`. Empty means "dev" (same pattern as monibright).

### `go.mod` — add dependency

`go get github.com/energye/systray@latest`

## What does NOT change

- `engine/` — no sync integration in this MVP
- `main.go` — already calls `cmd.Execute()`
- Existing commands (init, pair, run, status, id)
- Build tag `noassets` — unrelated to systray

## Verification

1. `go generate ./icon` — generates `icon/icon.png`
2. `go build -tags noassets -o plop .` — compiles
3. `./plop gui --help` — shows description
4. `./plop gui` — icon appears in menu bar/tray; menu shows disabled "plop dev", separator, Exit; clicking Exit quits
5. `go test -tags noassets -timeout 3m ./engine/` — existing E2E test still passes
6. Version override: `go build -tags noassets -ldflags "-X github.com/alex-vit/plop/cmd.Version=v0.1.0" -o plop .` → menu shows "plop v0.1.0"
