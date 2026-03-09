# Blackhole Server

## What This Is

A minimal Go HTTP server that accepts any incoming request, logs it with structured JSON output (via `slog`), and returns a `200 OK` with an empty body. Features a live TUI dashboard for inspecting incoming requests in real time. Designed as a lightweight debugging/inspection tool for HTTP traffic.

## Core Value

Every request that hits the server is reliably captured and logged in structured JSON format — nothing is lost, nothing is ambiguous.

## Current Milestone: v1.1 TUI Log Viewer

**Goal:** Replace raw stdout JSON with a live terminal UI for inspecting requests in real time.

**Target features:**
- Live TUI via bubbletea v2 with color-coded request display
- Compact one-line view per request (method, path, status) with expand-to-detail keypress
- Clear visible logs keybind
- Better visual separation between log entries
- JSON log file continues working alongside TUI

## Requirements

### Validated

- [x] Server accepts any HTTP method and path — v1.0
- [x] Server listens on configurable port (PORT env var, default 8080) — v1.0
- [x] Server captures method, URI, headers, body from every request — v1.0
- [x] Structured JSON logging via Go's `slog` package — v1.0
- [x] Log entries include: timestamp, method, path, status code, response time, all headers, request body, client IP, User-Agent — v1.0
- [x] Logs output to both stdout and a log file simultaneously — v1.0
- [x] Server responds with 200 OK and empty body — v1.0
- [x] Zero external dependencies (stdlib only) — v1.0

### Active

- [x] Live TUI dashboard using bubbletea v2
- [ ] Compact one-line view per request: method, path, status code
- [ ] Expand/collapse individual request detail with keypress (headers, body, client IP, response time)
- [ ] Color-coded output (methods, status codes, visual elements)
- [ ] Clear all visible logs with keybind
- [ ] Visual separation between log entries
- [x] JSON log file continues working alongside TUI

### Out of Scope

- Echo response body — replaced by empty 200 OK
- Routing or path-based behavior — all paths handled identically
- Authentication or authorization — this is a passive logging tool
- Database or persistent storage — log files are sufficient
- TLS/HTTPS — run behind a reverse proxy if needed
- Request filtering or search — keep it simple for v1.1
- Log export or replay — future consideration

## Context

- Brownfield: existing single-file Go server (`main.go`) with slog-based structured JSON logging
- v1.0 shipped: slog migration, dual output (stdout + file), renamed to event-horizon
- Go 1.25.0, Go modules
- Adding Charm ecosystem (bubbletea v2, lipgloss) as first external dependencies
- `bubbletea v2` is at `charm.land/bubbletea/v2`

## Constraints

- **TUI framework**: bubbletea v2 (Charm ecosystem)
- **Simplicity**: Minimal abstractions, keep the codebase small
- **Dual output**: TUI replaces stdout logging, but JSON file logging must continue

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use `slog` for logging | Standard library, structured JSON out of the box | ✓ Good |
| Empty 200 OK response | Server is for logging/inspection, not echoing | ✓ Good |
| Dual output (stdout + file) | Stdout for dev/container use, file for persistence | ✓ Good |
| Allow external deps (Charm) | TUI requires bubbletea; stdlib has no TUI support | ✓ Good |
| bubbletea v2 | Latest version, active development | ✓ Good |
| Compact + expand UI pattern | Minimal by default, detail on demand | — Pending |

---
*Last updated: 2026-03-06 after Phase 3 completion*
