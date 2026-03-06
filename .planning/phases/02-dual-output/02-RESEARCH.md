# Phase 2: Dual Output - Research

**Researched:** 2026-03-06
**Domain:** Go stdlib io.MultiWriter + slog dual-destination logging
**Confidence:** HIGH

## Summary

Phase 2 adds simultaneous log file output alongside existing stdout logging. The change is minimal: open a file with `os.OpenFile`, wrap it with `os.Stdout` via `io.MultiWriter`, and pass the combined writer to `slog.NewJSONHandler`. The `LOG_FILE` env var controls the file path (default: `requests.log`).

This is a well-understood Go pattern using only stdlib. The entire change touches only the `main()` function -- the `handleRequest` function and its logger usage remain untouched. The prior project decision (STATE.md) already locked `io.MultiWriter` as the approach.

**Primary recommendation:** Use `io.MultiWriter(os.Stdout, file)` as the writer for `slog.NewJSONHandler`. Open the log file with `os.O_APPEND|os.O_CREATE|os.O_WRONLY` and `0644` permissions.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| OUT-02 | JSON logs are simultaneously written to a log file | `io.MultiWriter` combines `os.Stdout` and the opened file into a single `io.Writer`; `slog.NewJSONHandler` accepts any `io.Writer` |
| OUT-03 | Log file path is configurable via `LOG_FILE` env var (default: `requests.log`) | `os.Getenv("LOG_FILE")` with fallback, same pattern already used for `PORT` |
</phase_requirements>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `io.MultiWriter` | stdlib | Combine multiple writers into one | Built-in, zero allocation overhead, exactly designed for this use case |
| `os.OpenFile` | stdlib | Open/create log file with append mode | Standard file creation with proper flags |
| `log/slog` | stdlib (Go 1.21+) | Structured JSON logging | Already in use from Phase 1 |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `io.MultiWriter` | Two separate `slog.Handler` instances with a custom multi-handler | Over-engineered; identical JSON output means single writer is correct. Multi-handler is not in stdlib -- would violate QA-01 |
| `os.OpenFile` | `os.Create` | `os.Create` truncates existing files -- wrong for log append |

## Architecture Patterns

### Current Structure (unchanged)

```
main.go    # single file, all code
```

No new files needed. The change is ~10 lines in `main()`.

### Pattern: MultiWriter Logger Setup

**What:** Create a single `io.Writer` that fans out to stdout and a log file, then pass it to the existing `slog.NewJSONHandler`.

**When to use:** When identical output must go to multiple destinations.

**Example:**

```go
// In main(), replace the current logger setup:

// Read log file path from env
logPath := "requests.log"
if lf := os.Getenv("LOG_FILE"); lf != "" {
    logPath = lf
}

// Open log file (create if not exists, append mode)
logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
if err != nil {
    // Fatal: can't start without log file
    slog.New(slog.NewJSONHandler(os.Stderr, nil)).LogAttrs(
        context.Background(), slog.LevelError, "failed to open log file",
        slog.String("path", logPath),
        slog.String("error", err.Error()),
    )
    os.Exit(1)
}
defer logFile.Close()

// Combine stdout + file
writer := io.MultiWriter(os.Stdout, logFile)
logger := slog.New(slog.NewJSONHandler(writer, nil))
```

### Anti-Patterns to Avoid

- **Two separate loggers:** Do not create two `slog.Logger` instances and log to each separately. This doubles the logging calls, risks inconsistency, and complicates the handler function signature.
- **Using `os.Create` instead of `os.OpenFile`:** `os.Create` truncates the file on each server restart. Use `os.OpenFile` with `os.O_APPEND` to preserve log history.
- **Ignoring the file open error:** A failed file open should be fatal at startup. Do not silently fall back to stdout-only.
- **Forgetting `defer logFile.Close()`:** The file handle must be closed on shutdown to flush buffered writes.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Writing to multiple destinations | Custom goroutine-based fan-out writer | `io.MultiWriter` | Stdlib handles synchronization and error propagation correctly |
| File creation with append | Manual file existence checks + create/open logic | `os.OpenFile` with `O_CREATE\|O_APPEND` flags | Single atomic call handles both cases |

## Common Pitfalls

### Pitfall 1: File Truncation on Restart

**What goes wrong:** Using `os.Create` or including `os.O_TRUNC` in flags wipes the log file every time the server starts.
**Why it happens:** `os.Create` is shorthand for `O_RDWR|O_CREATE|O_TRUNC`.
**How to avoid:** Use `os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)`.
**Warning signs:** Log file only contains entries from the current server session.

### Pitfall 2: MultiWriter Error Behavior

**What goes wrong:** If one writer in `io.MultiWriter` fails (e.g., disk full), the error is returned but the other writer may have already succeeded. The write is not atomic across destinations.
**Why it happens:** `io.MultiWriter` writes sequentially to each writer.
**How to avoid:** For this use case (logging), this is acceptable. Stdout rarely fails. If the file write fails, slog will surface the error. No special handling needed for a dev tool.

### Pitfall 3: Not Closing the File

**What goes wrong:** Buffered data may not be flushed if the file is not closed.
**Why it happens:** `os.File` may buffer writes at the OS level.
**How to avoid:** `defer logFile.Close()` immediately after successful `os.OpenFile`.

### Pitfall 4: Wrong File Permissions

**What goes wrong:** Using `0666` or `0777` creates world-writable files.
**Why it happens:** Copy-paste from examples.
**How to avoid:** Use `0644` (owner read/write, group/others read-only).

## Code Examples

### Complete main() Replacement

```go
func main() {
    port := "8080"
    if p := os.Getenv("PORT"); p != "" {
        port = p
    }

    logPath := "requests.log"
    if lf := os.Getenv("LOG_FILE"); lf != "" {
        logPath = lf
    }

    logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        slog.New(slog.NewJSONHandler(os.Stderr, nil)).LogAttrs(
            context.Background(), slog.LevelError, "failed to open log file",
            slog.String("path", logPath),
            slog.String("error", err.Error()),
        )
        os.Exit(1)
    }
    defer logFile.Close()

    writer := io.MultiWriter(os.Stdout, logFile)
    logger := slog.New(slog.NewJSONHandler(writer, nil))

    http.HandleFunc("/", handleRequest(logger))

    logger.LogAttrs(context.Background(), slog.LevelInfo, "server starting",
        slog.String("port", port),
        slog.String("log_file", logPath),
    )
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        logger.LogAttrs(context.Background(), slog.LevelError, "server failed",
            slog.String("error", err.Error()),
        )
        os.Exit(1)
    }
}
```

### Verification Commands

```bash
# Start server with default log file
go run main.go &

# Send a test request
curl http://localhost:8080/test

# Verify stdout had output (visual)
# Verify file has identical content
cat requests.log

# Compare: stdout and file should have matching JSON lines
# (stdout output must be captured separately for automated comparison)

# Test custom log file path
LOG_FILE=/tmp/custom.log go run main.go &
curl http://localhost:8080/test
cat /tmp/custom.log

# Kill background server
kill %1
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `log.SetOutput(io.MultiWriter(...))` | `slog.NewJSONHandler(io.MultiWriter(...), nil)` | Go 1.21 (Aug 2023) | slog accepts any `io.Writer`, same MultiWriter pattern works |

No deprecated APIs involved. `io.MultiWriter` and `os.OpenFile` are stable stdlib APIs unchanged since Go 1.0.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib, go1.25+) |
| Config file | None needed -- `go test` works out of the box |
| Quick run command | `go test ./... -v -count=1` |
| Full suite command | `go test ./... -v -count=1 -race` |

### Phase Requirements to Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| OUT-02 | JSON logs written to both stdout and log file simultaneously | integration | `go test -run TestDualOutput -v -count=1` | No -- Wave 0 |
| OUT-03 | LOG_FILE env var controls log file path, default requests.log | unit | `go test -run TestLogFilePath -v -count=1` | No -- Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./... -v -count=1`
- **Per wave merge:** `go test ./... -v -count=1 -race`
- **Phase gate:** Full suite green before verification

### Wave 0 Gaps

- [ ] `main_test.go` -- test file for dual output verification (OUT-02) and LOG_FILE config (OUT-03)
- [ ] Test helper: start server, send request, capture stdout + read file, compare JSON lines

Testing approach for OUT-02: Write JSON to an `io.MultiWriter` wrapping a `bytes.Buffer` (simulating stdout) and a temp file via `os.CreateTemp`. Verify both contain identical JSON. This avoids needing to actually start an HTTP server in tests.

## Open Questions

None. This is a straightforward stdlib pattern with no ambiguity.

## Sources

### Primary (HIGH confidence)

- [Go `io` package docs](https://pkg.go.dev/io) -- `MultiWriter` API
- [Go `os` package docs](https://pkg.go.dev/os) -- `OpenFile` flags and permissions
- [Go `log/slog` package docs](https://pkg.go.dev/log/slog) -- `NewJSONHandler(io.Writer, *HandlerOptions)`

### Secondary (MEDIUM confidence)

- [Better Stack slog guide](https://betterstack.com/community/guides/logging/logging-in-go/) -- MultiWriter pattern with slog
- [Go file append best practices](https://copyprogramming.com/howto/append-to-a-file-in-go) -- OpenFile flags

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- stdlib-only, well-documented APIs
- Architecture: HIGH -- pattern already decided in prior research (STATE.md), single function change
- Pitfalls: HIGH -- well-known Go file handling gotchas, verified against official docs

**Research date:** 2026-03-06
**Valid until:** 2026-04-06 (stable stdlib APIs, long validity)
