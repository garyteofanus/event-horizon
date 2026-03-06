# Roadmap: Echo Server

## Overview

Migrate the existing echo server from unstructured `fmt.Printf` logging to structured JSON logging via `slog`, then add dual output (stdout + file). Two phases: first replace the logging core and response behavior, then layer on file output with configuration. The server's existing HTTP acceptance behavior is preserved throughout.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Structured Logging Core** - Replace fmt.Printf with slog JSONHandler, change response to empty 200 OK
- [ ] **Phase 2: Dual Output** - Add simultaneous file logging alongside stdout with configurable path

## Phase Details

### Phase 1: Structured Logging Core
**Goal**: Every request is logged as structured JSON to stdout with all required fields, and the server responds with empty 200 OK
**Depends on**: Nothing (first phase)
**Requirements**: LOG-01, LOG-02, LOG-03, LOG-04, LOG-05, LOG-06, OUT-01, SRV-01, SRV-02, SRV-03, QA-01, QA-02
**Success Criteria** (what must be TRUE):
  1. Sending any HTTP request to the server produces a single-line JSON log entry on stdout containing timestamp, method, URI, status code, and response time
  2. The JSON log entry includes all request headers as structured attributes, the request body content, client IP, User-Agent, and Content-Length as distinct fields
  3. The server responds with HTTP 200 and an empty body for every request regardless of method or path
  4. The server has zero external dependencies (only Go standard library imports)
  5. All log attributes use `slog.LogAttrs` with typed `slog.Attr` constructors (no raw key-value pairs)
**Plans:** 1 plan

Plans:
- [x] 01-01-PLAN.md — TDD: slog structured logging migration with tests

### Phase 2: Dual Output
**Goal**: JSON logs are written to both stdout and a configurable log file simultaneously
**Depends on**: Phase 1
**Requirements**: OUT-02, OUT-03
**Success Criteria** (what must be TRUE):
  1. After starting the server, JSON log entries appear on both stdout and in the log file with identical content
  2. Setting `LOG_FILE` env var changes the log file path; omitting it defaults to `requests.log`
**Plans**: TBD

Plans:
- [ ] 02-01: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Structured Logging Core | 1/1 | Complete | 2026-03-06 |
| 2. Dual Output | 0/? | Not started | - |
