# Echo Server

## What This Is

A minimal Go HTTP server that accepts any incoming request, logs it with structured JSON output (via `slog`), and returns a `200 OK` with an empty body. Designed as a lightweight debugging/inspection tool for HTTP traffic.

## Core Value

Every request that hits the server is reliably captured and logged in structured JSON format — nothing is lost, nothing is ambiguous.

## Requirements

### Validated

- [x] Server accepts any HTTP method and path — existing
- [x] Server listens on configurable port (PORT env var, default 8080) — existing
- [x] Server captures method, URI, headers, and body from every request — existing

### Active

- [ ] Structured JSON logging via Go's `slog` package
- [ ] Log entries include: timestamp, method, path, status code, response time, all headers, request body, client IP, User-Agent
- [ ] Logs output to both stdout and a log file simultaneously
- [ ] Server responds with 200 OK and empty body (replace current echo behavior)
- [ ] Zero external dependencies (stdlib only)

### Out of Scope

- Echo response body — replaced by empty 200 OK
- Routing or path-based behavior — all paths handled identically
- Authentication or authorization — this is a passive logging tool
- Database or persistent storage — log files are sufficient
- TLS/HTTPS — run behind a reverse proxy if needed

## Context

- Brownfield: existing single-file Go server (`main.go`) that currently echoes requests back as plain text
- Currently uses `fmt.Printf` for stdout logging — will be replaced by `slog`
- Go 1.25.0, no external dependencies, Go modules
- `slog` is part of the standard library (since Go 1.21), so zero-dep constraint is maintained

## Constraints

- **Tech stack**: Go standard library only — no external dependencies
- **Simplicity**: Single-file server preferred, minimal abstractions

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use `slog` for logging | Standard library, structured JSON out of the box | — Pending |
| Empty 200 OK response | Server is for logging/inspection, not echoing | — Pending |
| Dual output (stdout + file) | Stdout for dev/container use, file for persistence | — Pending |

---
*Last updated: 2026-03-06 after initialization*
