# Repository Guidelines

## Project Structure & Module Organization
- `main.go` starts the CLI entrypoint.
- `cmd/` contains Cobra commands (`init`, `pair`, `run`, `status`, `id`) and root command wiring.
- `engine/` contains core sync logic and Syncthing integration (`engine.go`, `config.go`, `cert.go`).
- `tray/` contains system tray behavior for desktop mode.
- `paths/` contains OS-specific config directory handling (`*_unix.go`, `*_windows.go`).
- `icon/` stores app icon assets and generation helpers.
- `notes/` stores design and implementation notes; not runtime code.

## Build, Test, and Development Commands
- `./scripts/build-mac-app.sh` builds a double-clickable macOS bundle at `out/Plop.app`.
- `pwsh ./scripts/build-windows-release.ps1 -Version vX.Y.Z` builds Windows `out/plop.exe` and `out/plop-setup.exe` (Inno Setup required).
- `./scripts/release-tag.sh vX.Y.Z --watch` creates/pushes the release tag and optionally waits for the GitHub release workflow.
- `go build -tags noassets -o plop .` builds a plain CLI binary.
- `go run -tags noassets . run` runs headless engine mode for local development.
- `go test -tags noassets -v -count=1 -timeout 3m ./engine/` runs the end-to-end engine sync test.
- `go test -tags noassets ./...` runs all tests (currently centered in `engine/`).

## Coding Style & Naming Conventions
- Prefer `goimports -w` on edited Go files (it also runs `gofmt` behavior while fixing imports). Use `gofmt -w` only if `goimports` is unavailable.
- Keep packages focused by domain (`cmd`, `engine`, `tray`, `paths`).
- Use short, descriptive lowercase package names; exported identifiers in `CamelCase`, unexported in `camelCase`.
- Name platform-specific files with Go suffixes (`*_unix.go`, `*_windows.go`).

## Testing Guidelines
- Place tests next to implementation files as `*_test.go`.
- Prefer black-box behavior tests for sync lifecycle and peer interactions.
- Keep tests deterministic: set explicit timeouts and clean up resources (`t.Cleanup(...)`).
- Before opening a PR, run: `go test -tags noassets ./...`.

## Commit & Pull Request Guidelines
- Use concise, imperative commit subjects (examples from history: `Add --peer flag to plop run`, `Improve tray menu UX`).
- Keep commits scoped to one behavior change when possible.
- PRs should include:
  - clear summary of user-visible behavior changes,
  - test evidence (exact command/output summary),
  - screenshots or short recordings for tray/UI changes,
  - linked issue/task when applicable.

## Configuration & Safety Notes
- Requires Go `1.25+` and currently targets Syncthing `v1.30.0`.
- Default config/data path is OS-specific; use `--home` to override during testing.
- Do not commit local runtime artifacts (binaries, zips, logs, or generated config/state files).
- Create temporary artifacts (for example screenshots, debug captures, and ad-hoc exports) in `/tmp` by default on macOS (for readability), and otherwise use the OS temp directory; do not place them in the repository tree.

## Architecture Snapshot
- `plop` is a P2P file sync CLI that embeds Syncthing `lib/` as a library (not a subprocess wrapper).
- `cmd/` is a thin Cobra shell:
  - default/root command starts engine + tray,
  - `run` starts headless engine mode,
  - other subcommands delegate into `engine/`.
- `engine/` wraps `syncthing.App` lifecycle:
  - `New()` auto-initializes data dir, certs, and default config when missing,
  - `Start()` launches Syncthing services,
  - `Stop()` handles cleanup.
- `tray/` blocks on `systray.Run()` until exit, then `systray.Quit()` unblocks and shutdown continues.
- Default data directory is `~/Library/Application Support/plop` unless overridden by `--home`.

## Critical Syncthing Gotchas
- Certificate CN must be exactly `"syncthing"` or BEP TLS handshake fails.
- Config wrapper and event logger services must be running before `App.Start()` to avoid startup deadlock on config modification.
- GUI must be enabled (even on `127.0.0.1:0`) because `status` depends on Syncthing REST API wiring.
- Listen addresses should use port `0` (`tcp://0.0.0.0:0`, `quic://0.0.0.0:0`, relay URL) to avoid collisions and let discovery advertise actual addresses.
- Set `FSWatcherDelayS: 10` explicitly; do not rely on zero-value/default inference.
- Prefer `syncthing.LoadConfigAtStartup()` for config loading to preserve migrations/default behavior.
- Use `backend.TuningAuto` for backend open tuning.

## Claude Code -> Codex Translation
- `CLAUDE.md` guidance should be mirrored in `AGENTS.md`; Codex prioritizes `AGENTS.md`.
- Claude slash commands map to Codex skills and normal-language requests:
  - `/bisect` -> `bisect` skill,
  - `/always-allow` -> `rule-forge` skill flow,
  - `/sentry ...` style tasks -> `sentry` skill.
- Claude `Bash(...)` permission patterns map to Codex `prefix_rule(...)` entries in `~/.codex/rules/default.rules`.
- Use `python3 ~/.codex/skills/rule-forge/scripts/rule_forge.py inspect` to review over-specific rules and `from-session --write` to promote/prune.
- Treat persistent memory as files in repo (`notes/`) plus Codex config/session artifacts; do not assume hidden long-term memory.
- In this repo, references to "CLI Hub" or "clihub" mean the project at `https://github.com/thellimist/clihub`.
