# Roadmap: Blackhole Server

## Milestones

- v1.0 Structured Logging - Phases 1-2 (shipped 2026-03-06)
- v1.1 TUI Log Viewer - Phases 3-5 (in progress)

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

<details>
<summary>v1.0 Structured Logging (Phases 1-2) - SHIPPED 2026-03-06</summary>

- [x] **Phase 1: Structured Logging Core** - Replace fmt.Printf with slog JSONHandler, change response to empty 200 OK
- [x] **Phase 2: Dual Output** - Add simultaneous file logging alongside stdout with configurable path

</details>

### v1.1 TUI Log Viewer (In Progress)

**Milestone Goal:** Replace raw stdout JSON with a live terminal UI for inspecting requests in real time.

- [ ] **Phase 3: TUI Scaffolding and HTTP Bridge** - Wire up bubbletea v2 TUI with channel-based request streaming from HTTP server
- [ ] **Phase 4: Compact List with Styles** - Color-coded one-line-per-request display with visual separation
- [ ] **Phase 5: Interactive Features and Polish** - Navigate, expand/collapse, clear logs, help footer, resize handling

## Phase Details

<details>
<summary>v1.0 Structured Logging (Phases 1-2) - SHIPPED 2026-03-06</summary>

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
**Plans:** 1 plan

Plans:
- [x] 02-01-PLAN.md — TDD: io.MultiWriter dual output with LOG_FILE configuration

</details>

### Phase 3: TUI Scaffolding and HTTP Bridge
**Goal**: A working bubbletea v2 TUI launches on startup, receives live request data from the HTTP server via channel, and displays plain-text request lines as they arrive
**Depends on**: Phase 2 (existing server with file logging)
**Requirements**: TUI-01, TUI-02, TUI-03
**Success Criteria** (what must be TRUE):
  1. Running the server opens a terminal UI that takes over stdout; the user sees a TUI screen instead of raw JSON output
  2. Sending an HTTP request to the server causes a new text line to appear in the TUI within one second
  3. JSON structured log entries continue to be written to the log file while the TUI is active
  4. Pressing q or ctrl+c exits the TUI and shuts down the HTTP server cleanly (no hung processes)
**Plans**: TBD

### Phase 4: Compact List with Styles
**Goal**: Each request renders as a scannable, color-coded one-line row with clear visual separation between entries
**Depends on**: Phase 3
**Requirements**: DISP-01, DISP-02, DISP-03, DISP-04
**Success Criteria** (what must be TRUE):
  1. Each request row shows timestamp, HTTP method, path, and status code on a single line
  2. HTTP methods are visually distinct by color (GET=green, POST=blue, DELETE=red, PUT=yellow, PATCH=cyan)
  3. Status codes are visually distinct by color range (2xx=green, 4xx=yellow, 5xx=red)
  4. Adjacent log entries have visible separation (borders, spacing, or alternating styles) so the user can distinguish individual requests at a glance
**Plans**: TBD

### Phase 5: Interactive Features and Polish
**Goal**: Users can navigate the request list, inspect individual request details on demand, manage visible entries, and see available keybindings -- all adapting to terminal size changes
**Depends on**: Phase 4
**Requirements**: INTR-01, INTR-02, INTR-03, INTR-04, ROBU-01
**Success Criteria** (what must be TRUE):
  1. User can move a visible cursor through the request list using j/k or arrow keys
  2. User can press a key on a selected request to expand it, revealing headers, body, client IP, and response time; pressing again collapses it
  3. User can press a keybind to clear all visible log entries from the TUI (file log is unaffected)
  4. A footer displays available keybindings so the user does not need to memorize controls
  5. Resizing the terminal window re-renders the TUI correctly without crashing or corrupting the display
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 3 -> 4 -> 5

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Structured Logging Core | v1.0 | 1/1 | Complete | 2026-03-06 |
| 2. Dual Output | v1.0 | 1/1 | Complete | 2026-03-06 |
| 3. TUI Scaffolding and HTTP Bridge | v1.1 | 0/? | Not started | - |
| 4. Compact List with Styles | v1.1 | 0/? | Not started | - |
| 5. Interactive Features and Polish | v1.1 | 0/? | Not started | - |
