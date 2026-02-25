# plop - Dev Notes

## e2e test (2025-02-25)

Replaced `test_sync.sh` with `engine/sync_test.go`. Two engines run in-process -- no binary build or subprocess management needed.

Key details:
- Peers configured upfront via `NewConfig` (no separate pair step)
- LAN discovery needs ~2s warmup before writing the test file
- Full sync completes in ~26s
- `t.Cleanup(eng.Stop)` for reliable shutdown (runs even on `t.Fatal`)
- Run: `go test -tags noassets -v -count=1 -timeout 3m ./engine/`
