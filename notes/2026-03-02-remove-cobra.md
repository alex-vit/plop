# Remove Cobra dependency, flatten cmd/ into root package

Date: 2026-03-02

## What changed

Replaced the `cobra` CLI framework with stdlib `flag` + a switch-based dispatcher. Removed 3 transitive dependencies: `cobra`, `pflag`, `mousetrap`.

Then flattened the `cmd/` package into the root `main` package since `cmd/` was a Cobra convention with no remaining purpose as a separate package.

## File layout (before → after)

- `cmd/root.go` → `cli.go` — dispatcher (`run()`), `--home` global flag, default command (tray mode), `setupLogFile`, `printUsage`
- `cmd/init.go` → `cmd_init.go`
- `cmd/pair.go` → `cmd_pair.go`
- `cmd/run.go` → `cmd_run.go` — includes `stringSlice` type for repeatable `--peer` flag
- `cmd/status.go` → `cmd_status.go`
- `cmd/id.go` → `cmd_id.go`
- `cmd/redirect_*.go` → `redirect_*.go`
- `cmd/update_*.go` → `update_*.go`
- Tests follow their source files

## Dispatch pattern

`cli.go` exports nothing. `main.go` calls `run(os.Args[1:])`.

```go
func run(args []string) error {
    // Parse --home via global flag.FlagSet (stops at first non-flag)
    // Switch on subcommand string, pass remaining args to runX(args)
    // Handle flag.ErrHelp centrally
}
```

Each subcommand: `func runX(args []string) error` — creates a `flag.FlagSet`, parses, validates arg count, runs body.

## `--home` placement

Must come before the subcommand: `plop --home /foo init`. This is a minor behavior change from Cobra which allowed it anywhere. Acceptable for this tool.

## Version injection

Changed from `github.com/alex-vit/plop/cmd.Version` to `main.version` (unexported, same package as main). Build scripts updated. `-X` linker flag works with unexported vars.

## Dependencies removed

- `github.com/spf13/cobra`
- `github.com/spf13/pflag`
- `github.com/inconshreveable/mousetrap`
