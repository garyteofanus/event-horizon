# Architecture

**Analysis Date:** 2026-03-06

## Pattern Overview

**Overall:** Single-file monolith / Minimal HTTP server

**Key Characteristics:**
- Zero external dependencies -- uses only the Go standard library
- Single catch-all HTTP handler registered on `/`
- No routing, middleware, or layering -- all logic lives in one anonymous handler function
- Dual output: logs to stdout AND echoes back to the HTTP response

## Layers

This codebase has no formal layers. All concerns (routing, logging, response writing) are handled inline within a single handler closure in `main.go`.

**Handler (inline in main):**
- Purpose: Receives every HTTP request, logs it, and echoes it back
- Location: `main.go:19-63`
- Contains: Request parsing, header sorting, body reading, stdout logging, response writing
- Depends on: Go standard library (`net/http`, `fmt`, `io`, `sort`, `time`, `os`, `log`)
- Used by: Go `net/http` default serve mux

## Data Flow

**Request Echo Flow:**

1. Client sends HTTP request to `:<PORT>` (default `8080`)
2. Go default mux routes all paths to the catch-all handler at `/` (`main.go:19`)
3. Handler captures timestamp, method, URI, protocol, host, remote address (`main.go:20-26`)
4. Headers are collected into a slice, sorted alphabetically (`main.go:28-38`)
5. Request body is read fully via `io.ReadAll` (`main.go:42-48`)
6. All captured data is printed to stdout with delimiters (`main.go:22-50`)
7. Same data is written to `http.ResponseWriter` as `text/plain` (`main.go:53-62`)

**State Management:**
- Stateless -- no persistence, no in-memory state between requests
- Each request is fully independent

## Key Abstractions

There are no custom abstractions, types, or interfaces. The entire server is a single `main` function with one anonymous `http.HandlerFunc`.

## Entry Points

**main():**
- Location: `main.go:13`
- Triggers: `go run main.go` or executing the compiled binary
- Responsibilities: Read PORT from environment, register handler, start HTTP server

## Error Handling

**Strategy:** Minimal / fail-fast

**Patterns:**
- Body read errors are logged to stdout but do not terminate the request (`main.go:43-44`)
- Server listen failure causes `log.Fatal` which exits the process (`main.go:66`)
- No error responses are sent to clients -- errors are only logged server-side

## Cross-Cutting Concerns

**Logging:** Direct `fmt.Printf` to stdout for request details; `log.Printf` / `log.Fatal` for server lifecycle messages
**Validation:** None -- all requests are accepted regardless of method, path, or content
**Authentication:** None
**Configuration:** Single `PORT` environment variable with default `8080` (`main.go:14-17`)

---

*Architecture analysis: 2026-03-06*
