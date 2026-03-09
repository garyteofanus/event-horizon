# Requirements: Blackhole Server

**Defined:** 2026-03-06
**Core Value:** Every request that hits the server is reliably captured and logged in structured JSON format

## v1.0 Requirements (Complete)

All v1.0 requirements shipped and validated.

### Logging

- [x] **LOG-01**: Server logs every request as structured JSON via Go's `slog` package with `JSONHandler`
- [x] **LOG-02**: Each log entry includes timestamp, method, URI, protocol, status code, and response time
- [x] **LOG-03**: Each log entry includes all request headers as structured attributes
- [x] **LOG-04**: Each log entry includes request body content
- [x] **LOG-05**: Each log entry includes client IP (from RemoteAddr) and User-Agent as top-level fields
- [x] **LOG-06**: Each log entry includes Content-Length

### Output

- [x] **OUT-01**: JSON logs are written to stdout
- [x] **OUT-02**: JSON logs are simultaneously written to a log file
- [x] **OUT-03**: Log file path is configurable via `LOG_FILE` env var (default: `requests.log`)

### Server

- [x] **SRV-01**: Server accepts any HTTP method and any path
- [x] **SRV-02**: Server responds with 200 OK and empty body for all requests
- [x] **SRV-03**: Server listens on configurable port via `PORT` env var (default: 8080)

### Quality

- [x] **QA-01**: Zero external dependencies — stdlib only
- [x] **QA-02**: Use `slog.LogAttrs` with typed `slog.Attr` constructors (no raw key-value pairs)

## v1.1 Requirements

Requirements for TUI Log Viewer milestone. Each maps to roadmap phases.

### TUI Core

- [x] **TUI-01**: Server launches a bubbletea v2 TUI on the main goroutine, HTTP server runs in background
- [x] **TUI-02**: HTTP handler sends request data to TUI via buffered channel (thread-safe bridge)
- [x] **TUI-03**: JSON structured logging writes to file only; TUI owns stdout with human-readable format

### Display

- [x] **DISP-01**: Each request displays as a compact one-line row: timestamp, method, path, status code
- [x] **DISP-02**: HTTP methods are color-coded (GET=green, POST=blue, DELETE=red, PUT=yellow, PATCH=cyan)
- [x] **DISP-03**: Status codes are color-coded (2xx=green, 4xx=yellow, 5xx=red)
- [x] **DISP-04**: Visual separation between log entries (borders, spacing, or alternating styles)

### Interaction

- [x] **INTR-01**: User can navigate entries with j/k or arrow keys
- [x] **INTR-02**: User can expand/collapse individual entries to see full detail (headers, body, client IP, response time)
- [x] **INTR-03**: User can clear all visible log entries with a keybind
- [x] **INTR-04**: Help footer displays available keybindings

### Robustness

- [x] **ROBU-01**: TUI adapts to terminal resize events

### Clipboard & Formatting

- [x] **COPY-01**: User can press c to copy the selected request's body to clipboard
- [x] **COPY-02**: User can press Shift+C to copy the full request to clipboard
- [x] **COPY-03**: "Copied!" flash feedback shown after successful copy
- [x] **COPY-04**: Empty body shows "No body to copy" flash message
- [x] **COPY-05**: Flash auto-dismisses after ~2 seconds
- [x] **FMT-01**: JSON bodies are auto-detected and displayed formatted with syntax highlighting
- [x] **FMT-02**: User can press f to toggle between formatted and raw body display
- [x] **FMT-03**: Body section label shows "(JSON)" or "(raw)" based on content and toggle state
- [x] **FMT-04**: Syntax highlighting uses distinct colors for keys, strings, numbers, and booleans
- [x] **KEY-01**: Clear is remapped to x only; c is now copy
- [x] **KEY-02**: Shift+C copies full request details

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Logging Enhancements

- **LOG-07**: Request ID generation (unique per request, also set as response header)
- **LOG-08**: Log level configuration via `LOG_LEVEL` env var
- **LOG-09**: Pretty-print text mode via `LOG_FORMAT=text` env var
- **LOG-10**: Query parameters parsed as structured key-value pairs

### Safety

- **SAFE-01**: Body size limit via `http.MaxBytesReader` (configurable, default 1MB)
- **SAFE-02**: Sensitive header redaction (Authorization, Cookie)

### Filtering

- **FILT-01**: User can filter requests by method, path, or status code
- **FILT-02**: User can search through request history

### Robustness (deferred)

- **ROBU-02**: Bounded memory with max entry cap to prevent unbounded growth
- **ROBU-03**: Graceful shutdown with context cancellation
- **ROBU-04**: Panic recovery to restore terminal on crash

### Compatibility

- **COMP-01**: `--no-tui` flag to restore v1.0 stdout JSON behavior for CI/piping

## Out of Scope

| Feature | Reason |
|---------|--------|
| Echo response body | Server is for passive logging, not request mirroring |
| Route-based behavior | All paths handled identically — this is a logger, not a mock |
| Authentication | Dev/debugging tool — run behind proxy if auth needed |
| TLS/HTTPS | Use a reverse proxy for TLS termination |
| Mouse support | Keyboard-first TUI; adds complexity without value |
| Split-pane layout | Inline expand is simpler and sufficient |
| Configurable themes | Keep it simple; hardcoded Charm colors |
| WebSocket streaming | HTTP-only tool |
| Request replay/export | Log file serves this purpose |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| LOG-01 | Phase 1 | Complete |
| LOG-02 | Phase 1 | Complete |
| LOG-03 | Phase 1 | Complete |
| LOG-04 | Phase 1 | Complete |
| LOG-05 | Phase 1 | Complete |
| LOG-06 | Phase 1 | Complete |
| OUT-01 | Phase 1 | Complete |
| OUT-02 | Phase 2 | Complete |
| OUT-03 | Phase 2 | Complete |
| SRV-01 | Phase 1 | Complete |
| SRV-02 | Phase 1 | Complete |
| SRV-03 | Phase 1 | Complete |
| QA-01 | Phase 1 | Complete |
| QA-02 | Phase 1 | Complete |
| TUI-01 | Phase 3 | Complete |
| TUI-02 | Phase 3 | Complete |
| TUI-03 | Phase 3 | Complete |
| DISP-01 | Phase 4 | Complete |
| DISP-02 | Phase 4 | Complete |
| DISP-03 | Phase 4 | Complete |
| DISP-04 | Phase 4 | Complete |
| INTR-01 | Phase 5 | Complete |
| INTR-02 | Phase 5 | Complete |
| INTR-03 | Phase 5 | Complete |
| INTR-04 | Phase 5 | Complete |
| ROBU-01 | Phase 5 | Complete |
| COPY-01 | Phase 6 | Complete |
| COPY-02 | Phase 6 | Complete |
| COPY-03 | Phase 6 | Complete |
| COPY-04 | Phase 6 | Complete |
| COPY-05 | Phase 6 | Complete |
| FMT-01 | Phase 6 | Complete |
| FMT-02 | Phase 6 | Complete |
| FMT-03 | Phase 6 | Complete |
| FMT-04 | Phase 6 | Complete |
| KEY-01 | Phase 6 | Complete |
| KEY-02 | Phase 6 | Complete |

**Coverage:**
- v1.0 requirements: 14 total (all complete)
- v1.1 requirements: 23 total (12 original + 11 Phase 6)
- Mapped to phases: 23
- Unmapped: 0

---
*Requirements defined: 2026-03-06*
*Last updated: 2026-03-09 after gap closure phase creation*
