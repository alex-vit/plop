# .stfolder marker not created on init

**Status:** fixed

## Problem

First Windows install showed red "Unavailable" status. `status.json` reported:

```
"error": "folder marker missing (this indicates potential data loss, ...)"
```

The sync folder (`C:\Users\alex\plop`) existed and contained `.stignore`, but `.stfolder` was missing. Syncthing requires this marker directory to consider a folder valid.

## Root cause

`engine.New()` created the sync folder and wrote `.stignore` but never created the `.stfolder` marker:

```go
for _, folder := range cfg.Folders {
    os.MkdirAll(folder.Path, 0o755)
    writeDefaultStignore(folder.Path)  // ← no .stfolder
}
```

On macOS this went unnoticed because Syncthing creates the marker itself on first scan — but on Windows the folder entered an error state before that could happen.

## Fix

Added `os.MkdirAll(filepath.Join(folder.Path, ".stfolder"), 0o755)` in the init loop. `MkdirAll` is idempotent so existing installs are unaffected.
