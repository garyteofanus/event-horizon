# Requirements: Echo Server

**Defined:** 2026-03-06
**Core Value:** Every request that hits the server is reliably captured and logged in structured JSON format

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

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

## Out of Scope

| Feature | Reason |
|---------|--------|
| Echo response body | Server is for passive logging, not request mirroring |
| Route-based behavior | All paths handled identically — this is a logger, not a mock |
| Authentication | Dev/debugging tool — run behind proxy if auth needed |
| TLS/HTTPS | Use a reverse proxy for TLS termination |
| Web UI / dashboard | Logs are the interface — use `jq` or log viewers |
| Response customization | Fixed 200 OK — use httpbin/WireMock for mocking |
| External dependencies | Stdlib-only constraint is a feature, not a limitation |

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

**Coverage:**
- v1 requirements: 14 total
- Mapped to phases: 14
- Unmapped: 0

---
*Requirements defined: 2026-03-06*
*Last updated: 2026-03-06 after 01-01 plan completion*
