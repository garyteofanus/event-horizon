# Phase 3: TUI Scaffolding and HTTP Bridge - Context

**Gathered:** 2026-03-06
**Status:** Ready for planning

<domain>
## Phase Boundary

Wire up bubbletea v2 TUI with channel-based request streaming from the HTTP server. TUI takes over stdout with plain-text request lines as they arrive. JSON structured logging continues to file only. No colors, no interactivity beyond quit — those are Phase 4 and 5.

</domain>

<decisions>
## Implementation Decisions

### Request line content
- Full summary format: timestamp + method + URI + status + response time
- Timestamp format: HH:MM:SS (time only, no date, no milliseconds)
- Response time: show milliseconds, use "<1ms" for sub-millisecond times
- URI: show full path + query string, truncated at ~40 chars with "..." if too long (full URI visible in Phase 5 expand)
- Example: `14:32:05 GET /api/users?page=2&limit=1... 200 2ms`

### TUI chrome
- Header bar: one line showing server name + port + log file path (always shown, even for default)
- Format: `blackhole :8080 -> requests.log`
- Horizontal separator line below header
- Bottom status line with request count + quit hint: `3 requests . q to quit`
- Horizontal separator line above status line

### Request ordering
- New requests append at the bottom (chronological, like a terminal log)
- Auto-scroll to latest request

### Empty/startup state
- Show "Waiting for requests..." centered in the content area before any requests arrive
- Message disappears and is replaced by the request list once the first request arrives
- Status line shows "0 requests . q to quit" during empty state

### Stdout transition
- TUI owns stdout completely; JSON logging goes to file only
- Log file path always shown in header bar so user knows where structured logs go
- Clean exit on q or ctrl+c: restore terminal, no summary output, no stdout messages

### Claude's Discretion
- Channel buffer size for HTTP-to-TUI bridge
- Exact separator characters (box-drawing vs dashes)
- Internal file structure details within the three-file split (main.go, handler.go, tui.go)
- How auto-scroll behaves when terminal is full

</decisions>

<specifics>
## Specific Ideas

- Header format: `blackhole :8080 -> requests.log` — arrow indicates where logs flow
- Status line format: `3 requests . q to quit` — dot separator between count and hint
- The TUI should feel like a live log tail, not a dashboard — new entries flow in at the bottom

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- `handleRequest(logger)` in `main.go:14-47`: Current HTTP handler with slog logging — needs to be adapted to also send request data to a channel
- `slog.NewJSONHandler` + `io.MultiWriter` pattern in `main.go:71-72`: Will be simplified to file-only writer when TUI owns stdout

### Established Patterns
- Environment variable config with defaults (`PORT`, `LOG_FILE`): Continue this pattern for any new config
- `slog.LogAttrs` with typed constructors: Keep for file logging
- Single `main()` orchestration: Will expand to start both HTTP server (goroutine) and TUI (main goroutine)

### Integration Points
- `main.go` splits into three files: `main.go` (orchestration + channel creation), `handler.go` (HTTP handler + channel send), `tui.go` (bubbletea model + channel receive)
- HTTP handler needs to send a request struct through a buffered channel to TUI
- Logger changes from `MultiWriter(stdout, file)` to file-only writer
- bubbletea v2 is a new external dependency (first non-stdlib dep)

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 03-tui-scaffolding-and-http-bridge*
*Context gathered: 2026-03-06*
