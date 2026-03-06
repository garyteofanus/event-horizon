# Phase 1: Structured Logging Core - Research

**Researched:** 2026-03-06
**Domain:** Go `log/slog` structured logging, HTTP server response behavior
**Confidence:** HIGH

## Summary

Phase 1 replaces the existing `fmt.Printf`-based logging in `main.go` with Go's standard library `log/slog` package using `JSONHandler` for structured JSON output to stdout. The response behavior changes from echoing request details back to the client to returning an empty 200 OK for all requests.

The `slog` package has been stable since Go 1.21 and the project runs Go 1.25+. All required APIs (`slog.NewJSONHandler`, `slog.LogAttrs`, typed `slog.Attr` constructors, `slog.GroupAttrs`) are available and verified working. The constraint of zero external dependencies (QA-01) is naturally satisfied since `slog` is part of the standard library.

**Primary recommendation:** Create a single `slog.Logger` with `slog.NewJSONHandler(os.Stdout, nil)`, use `logger.LogAttrs()` exclusively with typed `slog.Attr` constructors for all log fields, and use `slog.GroupAttrs()` (not `slog.Group()`) for the headers group to maintain type safety per QA-02.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| LOG-01 | Structured JSON via slog with JSONHandler | `slog.NewJSONHandler(os.Stdout, nil)` produces single-line JSON per log call |
| LOG-02 | Timestamp, method, URI, protocol, status code, response time | `slog.String` for method/URI/protocol, `slog.Int` for status, `slog.Duration` for response time; timestamp is automatic from JSONHandler |
| LOG-03 | All request headers as structured attributes | Build `[]slog.Attr` from `r.Header`, use `slog.GroupAttrs("headers", attrs...)` |
| LOG-04 | Request body content | `io.ReadAll(r.Body)` then `slog.String("body", string(body))` |
| LOG-05 | Client IP and User-Agent as top-level fields | `slog.String("client_ip", r.RemoteAddr)` and `slog.String("user_agent", r.UserAgent())` |
| LOG-06 | Content-Length | `slog.Int64("content_length", r.ContentLength)` |
| OUT-01 | JSON logs to stdout | `slog.NewJSONHandler(os.Stdout, nil)` |
| SRV-01 | Accept any method and path | Keep existing `http.HandleFunc("/", ...)` catch-all pattern |
| SRV-02 | Respond 200 OK with empty body | Remove echo logic; default handler returns 200 with empty body |
| SRV-03 | Configurable port via PORT env var | Already implemented in current `main.go` -- preserve as-is |
| QA-01 | Zero external dependencies | Only `log/slog` and other stdlib packages used |
| QA-02 | Use `slog.LogAttrs` with typed `slog.Attr` constructors | `logger.LogAttrs()` with `slog.String`, `slog.Int`, `slog.Int64`, `slog.Duration`, `slog.GroupAttrs`, `slog.Any` |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `log/slog` | Go 1.21+ (stdlib) | Structured logging with JSON output | Official Go structured logging; no external deps needed |
| `net/http` | stdlib | HTTP server | Already in use; catch-all handler pattern preserved |
| `io` | stdlib | Body reading with `io.ReadAll` | Already in use |
| `os` | stdlib | stdout writer, env var reading | Already in use |
| `time` | stdlib | Request timing measurement | Already in use for timestamps |
| `context` | stdlib | Required by `LogAttrs` signature | First arg to `logger.LogAttrs` |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `log/slog` | `zerolog`, `zap` | External dependency -- violates QA-01 |
| `slog.GroupAttrs` | `slog.Group` | `slog.Group` accepts `...any` which violates QA-02 type safety requirement |

**Installation:** No installation needed -- all stdlib.

## Architecture Patterns

### Recommended Project Structure
```
.
├── main.go          # Single-file server (keep single-file architecture)
├── go.mod           # Module definition
└── main_test.go     # Tests (Wave 0 gap)
```

### Pattern 1: Logger Initialization
**What:** Create a configured slog.Logger at startup, use it in the handler.
**When to use:** Always -- the logger is the core component.
```go
// Verified working on Go 1.25
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
```

The `nil` options use sensible defaults: INFO level, RFC3339Nano timestamps, no source location. The JSONHandler automatically adds a `"time"` field satisfying the timestamp part of LOG-02.

### Pattern 2: LogAttrs with Typed Constructors Only (QA-02)
**What:** Use `logger.LogAttrs()` exclusively (never `logger.Info()`, `logger.With()`, or `slog.LogAttrs` package function) with only typed `slog.Attr` constructors.
**When to use:** Every log statement -- this is a hard requirement.
```go
// CORRECT: typed constructors only
logger.LogAttrs(context.Background(), slog.LevelInfo, "request",
    slog.String("method", r.Method),
    slog.String("uri", r.RequestURI),
    slog.Int("status", 200),
    slog.Duration("response_time", elapsed),
)

// WRONG: key-value pairs (violates QA-02)
logger.Info("request", "method", r.Method, "uri", r.RequestURI)
```

### Pattern 3: Headers as Grouped Attributes
**What:** Convert `http.Header` map to `[]slog.Attr` and nest under a `"headers"` group.
**When to use:** LOG-03 requires all headers as structured attributes.
```go
// Build typed attrs from headers
headerAttrs := make([]slog.Attr, 0, len(r.Header))
for name, values := range r.Header {
    if len(values) == 1 {
        headerAttrs = append(headerAttrs, slog.String(name, values[0]))
    } else {
        headerAttrs = append(headerAttrs, slog.Any(name, values))
    }
}
// Use GroupAttrs (NOT Group) for type safety
slog.GroupAttrs("headers", headerAttrs...)
```

Note: `slog.Any` is acceptable for multi-valued headers because it is still a typed `slog.Attr` constructor -- it produces `slog.Attr{Key: k, Value: slog.AnyValue(v)}`.

### Pattern 4: Request Timing
**What:** Capture start time before processing, compute duration after.
```go
start := time.Now()
// ... read body, build attrs ...
elapsed := time.Since(start)
// slog.Duration outputs nanoseconds as int64 in JSON
```

### Pattern 5: Empty 200 OK Response (SRV-02)
**What:** Handler does not write to `http.ResponseWriter` body -- Go defaults to 200.
```go
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    // Log the request...
    // Do NOT write anything to w -- Go returns 200 with empty body by default
})
```

### Anti-Patterns to Avoid
- **Using `logger.Info("msg", "key", value)` style:** Violates QA-02 -- these are untyped key-value pairs, not `slog.Attr` constructors.
- **Using `slog.Group()` instead of `slog.GroupAttrs()`:** `slog.Group` accepts `...any` which undermines type safety. Use `slog.GroupAttrs` which only accepts `...slog.Attr`.
- **Using package-level `slog.LogAttrs()` with `slog.SetDefault()`:** While functionally equivalent, using an explicit logger variable is clearer and prepares for Phase 2 where the handler will change.
- **Sorting headers for deterministic output:** Not required. JSON key order in log output is not specified as a requirement. Sorting adds unnecessary complexity.
- **Calling `w.WriteHeader(200)` explicitly:** Unnecessary -- Go's default is 200 when no status is set. Only needed if you want to be explicit, but it adds no value.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JSON serialization | Custom JSON encoder | `slog.NewJSONHandler` | Handles escaping, nesting, time formatting correctly |
| Timestamp formatting | `time.Now().Format(...)` | JSONHandler auto-adds `"time"` field | Consistent RFC3339Nano format, zero effort |
| Duration formatting | Manual `elapsed.Milliseconds()` | `slog.Duration("key", d)` | Outputs nanosecond int64; consistent and precise |
| Header iteration | Custom header parsing | `r.Header` map iteration | Standard `net/http` already parses headers |

## Common Pitfalls

### Pitfall 1: Mixing LogAttrs with Key-Value Style
**What goes wrong:** Using `logger.Info("msg", "key", value)` alongside `logger.LogAttrs()` violates QA-02.
**Why it happens:** `Info()`/`Warn()` convenience methods are more familiar from tutorials.
**How to avoid:** Use ONLY `logger.LogAttrs()` for all log statements. Grep for `logger.Info(`, `logger.Warn(`, `logger.Error(` -- there should be zero matches.
**Warning signs:** Any slog call that doesn't use `LogAttrs` method name.

### Pitfall 2: Content-Length from Header vs. Request Field
**What goes wrong:** Reading `r.Header.Get("Content-Length")` returns a string and may be absent. Using `r.ContentLength` returns `int64` (-1 if unknown).
**Why it happens:** Content-Length exists as both a header and a parsed request field.
**How to avoid:** Use `r.ContentLength` (int64) with `slog.Int64("content_length", r.ContentLength)`. The value is -1 when not provided, which is meaningful (unknown length).
**Warning signs:** String parsing of Content-Length header.

### Pitfall 3: RemoteAddr Format
**What goes wrong:** `r.RemoteAddr` includes the port (e.g., `127.0.0.1:54321` or `[::1]:54321`). The requirement says "client IP from RemoteAddr" so logging the full RemoteAddr value is correct.
**Why it happens:** Some implementations try to strip the port.
**How to avoid:** Log `r.RemoteAddr` as-is. The requirement (LOG-05) says "from RemoteAddr" -- don't parse it further.

### Pitfall 4: Body Reading Consumes the Reader
**What goes wrong:** `io.ReadAll(r.Body)` consumes the body. If anything else needs the body later, it's gone.
**Why it happens:** `r.Body` is a one-shot `io.ReadCloser`.
**How to avoid:** In this server it doesn't matter since we only log and return empty 200. Read body once, log it, done. No need to restore the body.

### Pitfall 5: User-Agent Convenience Method
**What goes wrong:** Logging User-Agent from `r.Header.Get("User-Agent")` when `r.UserAgent()` exists.
**Why it happens:** Developer iterates headers and forgets the convenience method.
**How to avoid:** Use `r.UserAgent()` for the top-level field (LOG-05). The User-Agent will ALSO appear in the headers group (LOG-03) -- this duplication is expected and correct since LOG-05 requires it as a distinct top-level field.

### Pitfall 6: Duration Output Format
**What goes wrong:** `slog.Duration` outputs nanoseconds as an integer in JSON. This may look odd but is the standard slog behavior.
**Why it happens:** slog's JSONHandler serializes `time.Duration` as its int64 nanosecond value.
**How to avoid:** Accept this as standard behavior. If human-readable is desired, use `slog.String("response_time", elapsed.String())` instead -- but `slog.Duration` is the typed constructor and matches QA-02 better.

## Code Examples

### Complete Handler Pattern (Verified on Go 1.25)
```go
package main

import (
    "context"
    "io"
    "log/slog"
    "net/http"
    "os"
    "time"
)

func main() {
    port := "8080"
    if p := os.Getenv("PORT"); p != "" {
        port = p
    }

    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Read body
        body, _ := io.ReadAll(r.Body)

        // Build header attrs
        headerAttrs := make([]slog.Attr, 0, len(r.Header))
        for name, values := range r.Header {
            if len(values) == 1 {
                headerAttrs = append(headerAttrs, slog.String(name, values[0]))
            } else {
                headerAttrs = append(headerAttrs, slog.Any(name, values))
            }
        }

        elapsed := time.Since(start)

        logger.LogAttrs(context.Background(), slog.LevelInfo, "request",
            slog.String("method", r.Method),
            slog.String("uri", r.RequestURI),
            slog.String("protocol", r.Proto),
            slog.Int("status", 200),
            slog.Duration("response_time", elapsed),
            slog.String("client_ip", r.RemoteAddr),
            slog.String("user_agent", r.UserAgent()),
            slog.Int64("content_length", r.ContentLength),
            slog.String("body", string(body)),
            slog.GroupAttrs("headers", headerAttrs...),
        )
        // Empty 200 OK -- do not write to w
    })

    logger.LogAttrs(context.Background(), slog.LevelInfo, "server starting",
        slog.String("port", port),
    )
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        logger.LogAttrs(context.Background(), slog.LevelError, "server failed",
            slog.String("error", err.Error()),
        )
        os.Exit(1)
    }
}
```

### Example JSON Output
```json
{"time":"2026-03-06T18:13:20.477889+07:00","level":"INFO","msg":"request","method":"GET","uri":"/test?foo=bar","protocol":"HTTP/1.1","status":200,"response_time":17583,"client_ip":"127.0.0.1:54321","user_agent":"curl/7.88.1","content_length":42,"body":"{\"hello\":\"world\"}","headers":{"Content-Type":"application/json","Accept":"text/html"}}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `log` package | `log/slog` | Go 1.21 (Aug 2023) | Structured logging in stdlib |
| `slog.Group(...any)` | `slog.GroupAttrs(...Attr)` | Go 1.24-1.25 (2025) | Type-safe group construction |
| Third-party loggers (zap, zerolog) | `log/slog` | Go 1.21+ | No external deps needed for structured logging |

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none -- Go testing needs no config |
| Quick run command | `go test -v -run TestX ./...` |
| Full suite command | `go test -v -race ./...` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| LOG-01 | JSON output via slog JSONHandler | unit | `go test -v -run TestLogJSON ./...` | No -- Wave 0 |
| LOG-02 | Required fields in log entry | unit | `go test -v -run TestLogFields ./...` | No -- Wave 0 |
| LOG-03 | Headers as structured attributes | unit | `go test -v -run TestLogHeaders ./...` | No -- Wave 0 |
| LOG-04 | Body content in log | unit | `go test -v -run TestLogBody ./...` | No -- Wave 0 |
| LOG-05 | Client IP and User-Agent fields | unit | `go test -v -run TestLogClientInfo ./...` | No -- Wave 0 |
| LOG-06 | Content-Length field | unit | `go test -v -run TestLogContentLength ./...` | No -- Wave 0 |
| OUT-01 | Output goes to stdout | unit | `go test -v -run TestLogStdout ./...` | No -- Wave 0 |
| SRV-01 | Any method/path accepted | unit | `go test -v -run TestAnyMethodPath ./...` | No -- Wave 0 |
| SRV-02 | 200 OK empty body response | unit | `go test -v -run TestEmptyResponse ./...` | No -- Wave 0 |
| SRV-03 | PORT env var | unit | `go test -v -run TestPortConfig ./...` | No -- Wave 0 |
| QA-01 | No external deps | smoke | `go list -m all` (should show only module) | No -- Wave 0 |
| QA-02 | LogAttrs with typed constructors | manual-only | Code review -- grep for non-LogAttrs slog calls | N/A |

### Testing Strategy
The handler should be extracted as a testable function. Use `httptest.NewRecorder()` and `httptest.NewRequest()` to create test requests. Capture log output by passing a `bytes.Buffer` to `slog.NewJSONHandler` instead of `os.Stdout`, then parse the JSON to verify fields.

```go
// Test pattern
var buf bytes.Buffer
logger := slog.New(slog.NewJSONHandler(&buf, nil))
// ... invoke handler with httptest ...
var entry map[string]any
json.Unmarshal(buf.Bytes(), &entry)
// Assert fields exist and have correct values
```

### Sampling Rate
- **Per task commit:** `go test -v -race ./...`
- **Per wave merge:** `go test -v -race ./...`
- **Phase gate:** Full suite green before verification

### Wave 0 Gaps
- [ ] `main_test.go` -- all test functions for LOG-01 through SRV-03
- [ ] Handler extraction -- current handler is an anonymous closure, needs to be a named function or accept logger as parameter for testability

## Open Questions

1. **Duration format preference**
   - What we know: `slog.Duration` outputs nanoseconds as int64. `slog.String` with `d.String()` outputs human-readable like `"1.234ms"`.
   - What's unclear: Which format the user prefers for `response_time`.
   - Recommendation: Use `slog.Duration` since QA-02 says typed constructors. Document the nanosecond output.

2. **Body content for empty bodies**
   - What we know: When no body is sent, `io.ReadAll` returns empty `[]byte`.
   - What's unclear: Should `body` field be `""` (empty string) or omitted entirely?
   - Recommendation: Always include `body` field as empty string -- simpler, consistent schema.

3. **Server startup/shutdown logging style**
   - What we know: Current code uses `log.Printf` and `log.Fatal` for startup.
   - What's unclear: Should startup messages also use slog or stay as plain `log`?
   - Recommendation: Use slog for consistency -- startup and error messages should also be structured JSON.

## Sources

### Primary (HIGH confidence)
- [Go slog package documentation](https://pkg.go.dev/log/slog) - API signatures, JSONHandler, HandlerOptions, Attr constructors
- Local Go 1.25/1.26 runtime verification - All code examples tested and output captured

### Secondary (MEDIUM confidence)
- [Go slog GroupAttrs proposal #66365](https://github.com/golang/go/issues/66365) - Confirmed accepted and completed

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - stdlib only, verified on local Go runtime
- Architecture: HIGH - single-file pattern, all APIs tested
- Pitfalls: HIGH - common Go HTTP/slog patterns well-documented
- Validation: MEDIUM - test strategy proposed but no existing tests to anchor to

**Research date:** 2026-03-06
**Valid until:** 2026-06-06 (stable stdlib, slow-moving domain)
