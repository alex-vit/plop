# OS-standard paths and simplified init

Config directory uses platform-specific app-data paths:
- macOS: `~/Library/Application Support/plop` (via `os.UserConfigDir()`)
- Linux: `~/.config/plop` (via `os.UserConfigDir()`)
- Windows: `%LocalAppData%\plop` (matches InnoSetup `{localappdata}`)

Default sync folder is `~/plop`.

`plop init` is zero-arg with a `--folder` flag (default `~/plop`):
- `plop init` — uses `~/plop` and OS config dir
- `plop init --folder /path/to/folder` — custom sync folder
- `--home` still overrides the config directory
