---
title: Cross-platform Build Tooling (No .sh-only workflows)
created: 2026-02-26
updated: 2026-02-26
tags: [build, tooling, windows, release]
status: idea
sections:
  - Replace shell-only build entrypoints with cross-platform commands
  - Add repeatable cross-compile targets
---

# Cross-platform Build Tooling

## Goal

Avoid `.sh` as the primary interface for common build tasks, so local dev and release workflows feel first-class on macOS, Windows, and Linux.

## Why now

- Current app bundle build entrypoint is `scripts/build-mac-app.sh` (mac-only, shell-first).
- We want GUI-first product usage, but still need predictable developer tooling for all hosts.
- Windows contributors should not need Git Bash as a prerequisite for normal tasks.

## Sibling repo findings

From sibling Go repos, `monibright` is the clearest reference implementation:

- Windows-only tray app, release pipeline runs on `windows-2022`.
- Uses plain `go build` for app exe + InnoSetup (`installer.iss`) for installer packaging.
- Release workflow uploads both portable exe and installer.
- Installer and app share the same HKCU autostart registry key pattern.

Other sibling Go repos in this folder did not expose reusable cross-platform build orchestration files (no `Makefile` / `Taskfile` / release workflows like `monibright`).

## Options

1. **Makefile**
   - Pros: familiar, short aliases (`make test`, `make build`).
   - Cons: native Windows requires extra setup (GNU Make / MSYS / WSL), so this is not truly zero-friction on Windows.

2. **Taskfile (`go-task`)**
   - Pros: cross-platform command runner, cleaner Windows support than GNU Make.
   - Cons: extra dependency/tool install.

3. **Go-native build tool (`go run ./tools/build ...`)** (recommended)
   - Pros: no extra toolchain beyond Go; easiest to make OS-aware and testable; best Windows story.
   - Cons: more code than Makefile.

## Cross-compile reality check

CLI binary cross-compilation works today from macOS host:

- `GOOS=windows GOARCH=amd64 go build -tags noassets -o /tmp/plop-windows-amd64.exe .`
- `GOOS=linux GOARCH=amd64 go build -tags noassets -o /tmp/plop-linux-amd64 .`

Both succeeded on 2026-02-26.

Note: `monibright`'s vendored systray README says cgo is required. We should still keep a CI matrix build to catch host-specific cross-compile regressions early.

## Proposed MVP

1. Add a single cross-platform entrypoint (`make` or Go tool) for:
   - `build-cli`
   - `test`
   - `build-mac-app`
   - `cross-cli` (windows/linux/mac matrix)
2. Keep existing `.sh` temporarily as compatibility shim.
3. Add CI matrix for cross-compile outputs.

## Open Questions

- Do we prefer GNU Make ergonomics despite Windows setup friction?
- Should Windows packaging target just `plop.exe` first, or installer too?
- Should tray GUI and headless server binaries split into separate targets?
