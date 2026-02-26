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
- `go build -tags noassets -o plop .` builds a plain CLI binary.
- `go run -tags noassets . run` runs headless engine mode for local development.
- `go test -tags noassets -v -count=1 -timeout 3m ./engine/` runs the end-to-end engine sync test.
- `go test -tags noassets ./...` runs all tests (currently centered in `engine/`).

## Coding Style & Naming Conventions
- Follow standard Go formatting: run `gofmt -w` on edited files.
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
