# Security

Seven holes from the reference implementation are closed in this codebase.

## Hole 1 — Path traversal via filename
**Fixed in** `internal/api/handler.go` `validateFilename()`
Filenames are rejected if they contain `/` or `\`, start with `.`, or exceed 255 characters.

## Hole 2 — Shell-style directory commands
**Fixed in** `internal/runner/runner.go` `RunJob()`
All directory creation and deletion uses `os.MkdirTemp` and `os.RemoveAll`. No shell commands anywhere.

## Hole 3 — Compiler flag injection
**Fixed in** `internal/api/handler.go` `isFlagAllowed()`
Each flag is checked against a per-language allowlist. Any flag not on the list returns HTTP 400.

## Hole 4 — No request size limits
**Fixed in** `internal/api/handler.go` `RunHandler()`
`http.MaxBytesReader` caps the request body at 256 KiB before decoding.

## Hole 5 — UID collisions under load
**Fixed in** `internal/runner/runner.go` `RunJob()`
`os.MkdirTemp` creates a directory with a random suffix. Directories are never reused.

## Hole 6 — Unbounded child output
**Fixed in** `internal/runner/runner.go` `runCmd()`
Captured stdout is capped at 4 MiB. Output beyond that is truncated with a marker.

## Hole 7 — Stale jail directories
**Fixed in** `internal/runner/runner.go` `RunJob()`
`defer os.RemoveAll(jailDir)` runs on every exit path including panics.
