# OS-standard paths and simplified init

Config directory now uses `os.UserConfigDir()` instead of `~/.plop`:
- macOS: `~/Library/Application Support/plop`
- Linux: `~/.config/plop`
- Windows: `%AppData%/plop`

`plop init` is now zero-arg with a `--folder` flag (default `~/Sync`):
- `plop init` — uses `~/Sync` and OS config dir
- `plop init --folder /path/to/folder` — custom sync folder
- `--home` still overrides the config directory
