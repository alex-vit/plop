# OS-standard paths and simplified init

Config directory now uses `os.UserConfigDir()` instead of `~/.gosync`:
- macOS: `~/Library/Application Support/gosync`
- Linux: `~/.config/gosync`
- Windows: `%AppData%/gosync`

`gosync init` is now zero-arg with a `--folder` flag (default `~/Sync`):
- `gosync init` — uses `~/Sync` and OS config dir
- `gosync init --folder /path/to/folder` — custom sync folder
- `--home` still overrides the config directory
