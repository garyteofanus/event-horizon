# Codebase Concerns

**Analysis Date:** 2026-03-06

## Tech Debt

**Single-file monolith with inline handler:**
- Issue: The entire application logic (logging, header sorting, body reading, response writing) lives in a single anonymous closure inside `main()` in `main.go` (lines 19-63). The handler is not exported or named, making it impossible to unit test without starting the full HTTP server.
- Files: `main.go`
- Impact: Cannot write targeted unit tests for the echo handler. Any future feature additions will bloat `main()` further.
- Fix approach: Extract the anonymous handler into a named, exported function (`EchoHandler`) that returns an `http.HandlerFunc`. This enables direct testing via `httptest.NewRecorder`.

**No graceful shutdown:**
- Issue: The server uses `log.Fatal(http.ListenAndServe(...))` on line 66 with no signal handling. On SIGTERM/SIGINT the server terminates immediately, potentially dropping in-flight requests.
- Files: `main.go:66`
- Impact: Clients with active connections receive abrupt RST packets during deployment or restart. In containerized environments (Docker, Kubernetes), this means unclean shutdowns during rolling deploys.
- Fix approach: Use `http.Server` with `Shutdown(ctx)` and listen for `os.Signal` via `signal.NotifyContext` to drain connections before exit.

**No PORT validation:**
- Issue: The `PORT` environment variable (line 15) is used directly with no validation. A non-numeric or out-of-range value produces a cryptic error from `ListenAndServe` rather than a clear startup failure.
- Files: `main.go:14-17`
- Impact: Misconfiguration results in confusing error messages.
- Fix approach: Parse `PORT` with `strconv.Atoi` and validate range (1-65535) before attempting to listen.

## Known Bugs

No confirmed bugs detected. The codebase is minimal and functional for its stated purpose.

## Security Considerations

**Unbounded body read (denial of service):**
- Risk: `io.ReadAll(r.Body)` on line 42 reads the entire request body into memory with no size limit. A malicious client can send an arbitrarily large body and exhaust server memory.
- Files: `main.go:42`
- Current mitigation: None. Go's default `http.Server` has no `MaxBytesReader` applied.
- Recommendations: Wrap `r.Body` with `http.MaxBytesReader(w, r.Body, maxSize)` before reading, using a reasonable limit (e.g., 1 MB). Alternatively, set `ReadTimeout` and `MaxHeaderBytes` on a custom `http.Server`.

**No TLS support:**
- Risk: The server listens on plain HTTP only. Any sensitive data in echoed requests (auth headers, tokens, cookies) is transmitted in cleartext.
- Files: `main.go:66`
- Current mitigation: None. Assumed to run behind a reverse proxy or for local development only.
- Recommendations: For production use, add `ListenAndServeTLS` option or document that a TLS-terminating proxy is required.

**Request reflection enables header/body exfiltration:**
- Risk: By design, the server echoes all request details (headers, body) back to the caller. If exposed publicly, any forwarded requests (e.g., from SSRF attacks) will have their full contents reflected, including `Authorization`, `Cookie`, and other sensitive headers.
- Files: `main.go:54-62`
- Current mitigation: None. This is inherent to the event-horizon server design.
- Recommendations: If deployed beyond local development, add IP allowlisting or authentication. Consider redacting sensitive headers (`Authorization`, `Cookie`, `X-Api-Key`) from the echo response.

**No request timeout configuration:**
- Risk: Using the default `http.Server` via `http.ListenAndServe` means no `ReadTimeout`, `WriteTimeout`, or `IdleTimeout`. Slowloris-style attacks can hold connections open indefinitely, exhausting file descriptors.
- Files: `main.go:66`
- Current mitigation: None.
- Recommendations: Create an explicit `http.Server{}` with `ReadTimeout: 10s`, `WriteTimeout: 10s`, `IdleTimeout: 60s`.

## Performance Bottlenecks

**Stdout logging under high concurrency:**
- Problem: Every request writes multiple `fmt.Printf` calls to stdout (lines 22-50) without buffering. Under high concurrency, these calls contend on the stdout lock, and output from concurrent requests can interleave.
- Files: `main.go:22-50`
- Cause: `fmt.Printf` acquires a lock on `os.Stdout` per call. Multiple calls per request (7+ print statements) amplifies contention.
- Improvement path: Buffer the entire log entry into a `strings.Builder` or `bytes.Buffer`, then write it in a single `fmt.Print` call. Alternatively, use `log.Logger` with a mutex-protected writer for atomic multi-line output.

**Header sorting on every request:**
- Problem: Headers are sorted alphabetically on every request (lines 28-32). While cheap for typical request sizes, this is unnecessary work.
- Files: `main.go:28-32`
- Cause: `sort.Strings` on header keys for display purposes.
- Improvement path: Low priority. Only matters at extreme scale. Could skip sorting or use `maps.Keys` with `slices.Sort` for marginally cleaner code.

## Fragile Areas

**Duplicate echo logic:**
- Files: `main.go:22-50` (stdout logging), `main.go:54-62` (response writing)
- Why fragile: The request details are formatted twice -- once for stdout and once for the HTTP response -- using separate `fmt.Printf`/`fmt.Fprintf` calls. Changes to the output format require updating both locations, and they can drift out of sync.
- Safe modification: Extract a shared formatting function that writes to an `io.Writer`, then call it for both stdout and the response writer.
- Test coverage: Zero. No tests exist.

## Scaling Limits

**Single-process, in-memory only:**
- Current capacity: Handles concurrent requests via Go's goroutine-per-connection model. Adequate for development/debugging use.
- Limit: Memory-bound due to unbounded body reads. A single large request (e.g., 1 GB body) can crash the process.
- Scaling path: Add `MaxBytesReader`, request timeouts, and optionally rate limiting if public exposure is needed.

## Dependencies at Risk

No external dependencies. The project uses only Go standard library, which is a strength. The `go.mod` specifies `go 1.25.0`.

## Missing Critical Features

**No test suite:**
- Problem: Zero test files exist in the repository. The handler is an anonymous closure that cannot be tested in isolation.
- Blocks: Cannot verify behavior after changes. No regression protection.

**No health check endpoint:**
- Problem: No `/healthz` or `/readyz` endpoint for container orchestration readiness/liveness probes.
- Blocks: Kubernetes or Docker health checks must rely on TCP connect rather than application-level health.

**No structured logging:**
- Problem: Logging uses raw `fmt.Printf` to stdout with ad-hoc formatting. No structured format (JSON), no log levels, no request ID correlation.
- Blocks: Log aggregation and filtering in production environments.

## Test Coverage Gaps

**Entire codebase is untested:**
- What's not tested: All functionality -- request parsing, header echoing, body reading, response formatting, PORT configuration.
- Files: `main.go`
- Risk: Any modification could introduce regressions with no automated detection. The duplicate formatting logic (stdout vs response) is especially prone to silent drift.
- Priority: High. Extract the handler to a named function and add table-driven tests using `httptest.NewRequest` and `httptest.NewRecorder`.

---

*Concerns audit: 2026-03-06*
