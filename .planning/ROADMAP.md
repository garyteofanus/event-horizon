# Roadmap: Blackhole Server

## Milestones

- v1.0 Structured Logging - Phases 1-2 (shipped 2026-03-06)
- v1.1 TUI Log Viewer - Phases 3-7 (in progress)

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

- [x] **Phase 3: TUI Scaffolding and HTTP Bridge** - Wire up bubbletea v2 TUI with channel-based request streaming from HTTP server
- [x] **Phase 4: Compact List with Styles** - Color-coded one-line-per-request display with visual separation
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
**Plans:** 2 plans

Plans:
- [x] 03-01-PLAN.md — Channel bridge, RequestData type, handler adaptation, file-only logging
- [x] 03-02-PLAN.md — Bubbletea v2 TUI model, main orchestration, end-to-end verification

### Phase 4: Compact List with Styles
**Goal**: Each request renders as a scannable, color-coded one-line row with clear visual separation between entries
**Depends on**: Phase 3
**Requirements**: DISP-01, DISP-02, DISP-03, DISP-04
**Success Criteria** (what must be TRUE):
  1. Each request row shows timestamp, HTTP method, path, and status code on a single line
  2. HTTP methods are visually distinct by color (GET=green, POST=blue, DELETE=red, PUT=yellow, PATCH=cyan)
  3. Status codes are visually distinct by color range (2xx=green, 4xx=yellow, 5xx=red)
  4. Adjacent log entries have visible separation (borders, spacing, or alternating styles) so the user can distinguish individual requests at a glance
**Plans**: 1 plan

Plans:
- [x] 04-01-PLAN.md — Styled compact request rows with Lip Gloss and render verification

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
**Plans**: 1 plan

Plans:
- [x] 05-01-PLAN.md — Interactive navigation, expansion, clear, footer, resize handling

## Progress

**Execution Order:**
Phases execute in numeric order: 3 -> 4 -> 5

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Structured Logging Core | v1.0 | 1/1 | Complete | 2026-03-06 |
| 2. Dual Output | v1.0 | 1/1 | Complete | 2026-03-06 |
| 3. TUI Scaffolding and HTTP Bridge | v1.1 | 2/2 | Complete | 2026-03-06 |
| 4. Compact List with Styles | v1.1 | 1/1 | Complete | 2026-03-06 |
| 5. Interactive Features and Polish | v1.1 | 0/? | Not started | - |
| 6. Copy & Format | v1.1 | 0/2 | Not started | - |
| 7. Retroactive Verification & Cleanup | v1.1 | 0/2 | Not started | - |

### Phase 6: Copy request body and format body in expanded view

**Goal:** Add clipboard copy for request body/full request and JSON pretty-printing with syntax highlighting in the expanded detail view
**Requirements**: COPY-01, COPY-02, COPY-03, COPY-04, COPY-05, FMT-01, FMT-02, FMT-03, FMT-04, KEY-01, KEY-02
**Depends on:** Phase 5
**Success Criteria** (what must be TRUE):
  1. User can press c to copy the selected request's body to clipboard, with "Copied!" flash feedback
  2. User can press Shift+C to copy the full request (headers, body, client IP, response time) to clipboard
  3. Empty body shows "No body to copy" flash; flash auto-dismisses after ~2 seconds
  4. JSON bodies are auto-detected and displayed formatted with syntax highlighting by default
  5. User can press f to toggle between formatted and raw body display
  6. Body section label shows "(JSON)" or "(raw)" based on content and toggle state
  7. Clear is remapped to x only; c is now copy
**Plans:** 2 plans

Plans:
- [ ] 06-01-PLAN.md — Clipboard copy with flash messages and keybinding remap
- [ ] 06-02-PLAN.md — JSON formatting, syntax highlighting, and format toggle

### Phase 7: Retroactive Verification and Requirements Cleanup
**Goal:** Close all documentation/process gaps identified by milestone audit — create missing VERIFICATION.md and SUMMARY.md files, update REQUIREMENTS.md with Phase 6 entries and correct checkboxes, remove dead code
**Depends on:** Phase 6
**Requirements**: TUI-01, TUI-02, TUI-03, DISP-01, DISP-02, DISP-03, DISP-04, INTR-01, INTR-02, INTR-03, INTR-04, ROBU-01, COPY-01, COPY-02, COPY-03, COPY-04, COPY-05, FMT-01, FMT-02, FMT-03, FMT-04, KEY-01, KEY-02
**Gap Closure:** Closes gaps from v1.1 milestone audit
**Success Criteria** (what must be TRUE):
  1. VERIFICATION.md exists for phases 3, 4, and 5 with requirement-level pass/fail results
  2. Phase 5 has a SUMMARY.md documenting implemented features
  3. REQUIREMENTS.md contains all Phase 6 requirement IDs (COPY-01–05, FMT-01–04, KEY-01–02) in both the requirements list and traceability table
  4. All implemented v1.1 requirement checkboxes are marked `[x]`
  5. formatRequestLine dead code is removed from handler.go
**Plans:** 2 plans

Plans:
- [ ] 07-01-PLAN.md — Retroactive VERIFICATION.md for phases 3, 4, 5 and Phase 5 SUMMARY.md
- [ ] 07-02-PLAN.md — Dead code removal (formatRequestLine)
