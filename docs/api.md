 # API

## GET /healthz
Liveness check. Returns 200 if the process is up.

## GET /readyz
Readiness check. Returns 200 if all language runtimes are available. Returns 503 with a breakdown if any are missing.

## GET /info
Returns build info, registered languages, server limits, and runtime stats.

## POST /run
Runs untrusted code and returns per-test results.

Request:
- language: required, must match a configured language id
- source: required, UTF-8, max 256 KiB
- source_filename: optional, single path component only
- artifact_filename: optional, single path component only
- build.flags: optional, filtered against per-language allowlist
- tests: required, at least one entry

Response status values:
- accepted
- wrong_output
- output_whitespace_mismatch
- time_exceeded
- memory_exceeded
- runtime_error
- build_failed
- internal_error
