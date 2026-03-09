# Phase 3: TUI Scaffolding and HTTP Bridge - Research

**Researched:** 2026-03-06
**Domain:** Go TUI with bubbletea v2, channel-based HTTP-to-TUI bridge
**Confidence:** HIGH

## Summary

This phase introduces the first external dependency (bubbletea v2) and restructures the single-file server into a three-file architecture. The core challenge is wiring an HTTP server running in a background goroutine to a bubbletea TUI on the main goroutine via a buffered channel, while redirecting structured JSON logging from stdout to file-only.

Bubbletea v2 (released Feb 2026, current version v2.0.1) uses a new vanity import path (`charm.land/bubbletea/v2`) and has significant API changes from v1: `View()` returns `tea.View` instead of `string`, key events use `tea.KeyPressMsg` instead of `tea.KeyMsg`, and terminal features are declared on the View struct rather than via program options. The channel-based realtime pattern is well-established in bubbletea's official examples and maps directly to this phase's HTTP bridge requirement.

**Primary recommendation:** Use the official `listenForActivity`/`waitForActivity` channel pattern from bubbletea's realtime example, with the HTTP handler sending request structs to a buffered channel and the TUI model consuming them via a blocking receive command that re-registers after each message.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Request line format: `HH:MM:SS METHOD /path... STATUS TIMEms` (e.g., `14:32:05 GET /api/users?page=2&limit=1... 200 2ms`)
- URI truncated at ~40 chars with "..." suffix
- Response time in ms, "<1ms" for sub-millisecond
- Header bar: `blackhole :8080 -> requests.log` with separator below
- Status line: `3 requests . q to quit` with separator above
- New requests append at bottom, auto-scroll to latest
- Empty state: "Waiting for requests..." centered, replaced by list on first request
- TUI owns stdout; JSON logging to file only
- Log file path always shown in header
- Clean exit on q or ctrl+c: restore terminal, no summary, no stdout messages
- Three-file split: main.go (orchestration), handler.go (HTTP + channel), tui.go (bubbletea model)

### Claude's Discretion
- Channel buffer size for HTTP-to-TUI bridge
- Exact separator characters (box-drawing vs dashes)
- Internal file structure details within the three-file split
- How auto-scroll behaves when terminal is full

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| TUI-01 | Server launches a bubbletea v2 TUI on the main goroutine, HTTP server runs in background | bubbletea v2 Program.Run() blocks on main goroutine; HTTP server starts via goroutine in Init() or before Program creation |
| TUI-02 | HTTP handler sends request data to TUI via buffered channel (thread-safe bridge) | Official realtime example pattern: listenForActivity/waitForActivity with channel; handler sends struct, TUI receives via Cmd |
| TUI-03 | JSON structured logging writes to file only; TUI owns stdout with human-readable format | Remove os.Stdout from io.MultiWriter; bubbletea takes over stdout automatically when Program.Run() is called |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| charm.land/bubbletea/v2 | v2.0.1+ | TUI framework (Elm architecture) | Official Charm TUI framework, only serious Go TUI option with this architecture |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Go stdlib log/slog | (stdlib) | Structured JSON logging to file | Already in use, continues unchanged |
| Go stdlib net/http | (stdlib) | HTTP server | Already in use, continues unchanged |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| bubbletea v2 | tview | tview is widget-based not Elm-architecture; bubbletea is locked decision |
| channel bridge | Program.Send() | Program.Send() works but channel+Cmd pattern is idiomatic bubbletea |

**Installation:**
```bash
go get charm.land/bubbletea/v2@latest
```

## Architecture Patterns

### Recommended Project Structure
```
blackhole-server/
├── main.go       # Orchestration: config, channel creation, start HTTP goroutine, run TUI
├── handler.go    # HTTP handler: handleRequest() sends to channel + logs to file
├── tui.go        # Bubbletea model: receives from channel, renders request lines
├── go.mod
└── go.sum
```

### Pattern 1: Channel Bridge (HTTP to TUI)
**What:** A buffered channel carries request data structs from the HTTP handler goroutine to the bubbletea model's Update loop via a blocking receive command.
**When to use:** Whenever external goroutines need to push data into a bubbletea program.
**Example:**
```go
// Source: bubbletea realtime example (adapted for this project)

// Shared request data struct
type RequestData struct {
    Timestamp    time.Time
    Method       string
    URI          string
    Status       int
    ResponseTime time.Duration
}

// Message type for TUI
type requestMsg RequestData

// Command that blocks until a request arrives on the channel
func waitForRequest(ch <-chan RequestData) tea.Cmd {
    return func() tea.Msg {
        return requestMsg(<-ch)
    }
}

// In HTTP handler: send to channel (non-blocking with select to avoid blocking HTTP responses)
func handleRequest(logger *slog.Logger, reqCh chan<- RequestData) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        // ... existing logging logic ...
        elapsed := time.Since(start)

        // Send to TUI (non-blocking: drop if channel full)
        select {
        case reqCh <- RequestData{
            Timestamp:    start,
            Method:       r.Method,
            URI:          r.RequestURI,
            Status:       200,
            ResponseTime: elapsed,
        }:
        default:
            // Channel full, skip TUI update (log still goes to file)
        }
    }
}
```

### Pattern 2: Bubbletea v2 Model with Channel
**What:** The TUI model holds the channel reference and re-registers the wait command after each received message.
**When to use:** Standard pattern for all bubbletea channel consumers.
**Example:**
```go
// Source: bubbletea v2 API + realtime example pattern

type model struct {
    reqCh    <-chan RequestData
    requests []RequestData
    port     string
    logPath  string
}

func (m model) Init() tea.Cmd {
    return waitForRequest(m.reqCh)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    case requestMsg:
        m.requests = append(m.requests, RequestData(msg))
        return m, waitForRequest(m.reqCh) // Re-register for next message
    }
    return m, nil
}

func (m model) View() tea.View {
    // Build view string with header, request lines, status bar
    var s strings.Builder
    // ... render header, separator, requests, separator, status ...
    return tea.NewView(s.String())
}
```

### Pattern 3: Main Orchestration
**What:** main() creates the channel, starts HTTP server in a goroutine, and runs the TUI on the main goroutine.
**When to use:** The single entry point pattern.
**Example:**
```go
func main() {
    port := "8080"
    if p := os.Getenv("PORT"); p != "" {
        port = p
    }
    logPath := "requests.log"
    if lf := os.Getenv("LOG_FILE"); lf != "" {
        logPath = lf
    }

    // Open log file
    logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
        os.Exit(1)
    }
    defer logFile.Close()

    // File-only logger (TUI owns stdout)
    logger := slog.New(slog.NewJSONHandler(logFile, nil))

    // Channel bridge
    reqCh := make(chan RequestData, 64)

    // HTTP handler
    http.HandleFunc("/", handleRequest(logger, reqCh))

    // Start HTTP server in background
    go func() {
        if err := http.ListenAndServe(":"+port, nil); err != nil {
            fmt.Fprintf(os.Stderr, "server error: %v\n", err)
            os.Exit(1)
        }
    }()

    // Run TUI on main goroutine
    p := tea.NewProgram(model{
        reqCh:   reqCh,
        port:    port,
        logPath: logPath,
    })
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
        os.Exit(1)
    }
}
```

### Anti-Patterns to Avoid
- **Using goroutines inside Update():** Never spawn goroutines from Update. Use tea.Cmd functions which bubbletea runs in managed goroutines.
- **Blocking channel send in HTTP handler:** Use `select` with `default` to avoid blocking HTTP responses when the TUI channel is full.
- **Writing to stdout while TUI is active:** bubbletea owns stdout via its renderer. Use `tea.Printf`/`tea.Println` commands if you must print, or log to file.
- **Using v1 API patterns:** Do not use `tea.WithAltScreen()` option, `tea.KeyMsg` type, or `View() string` signature -- these are all v2 breaking changes.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Terminal raw mode / alt screen | Custom terminal handling | bubbletea Program (handles it automatically) | Terminal state management is platform-specific and error-prone |
| Elm architecture event loop | Custom event dispatcher | bubbletea Init/Update/View cycle | Battle-tested, handles concurrency, signal handling |
| Terminal resize detection | SIGWINCH handler | bubbletea tea.WindowSizeMsg | Bubbletea sends this automatically |
| Terminal restoration on exit | defer/signal cleanup | bubbletea handles on Program exit | Bubbletea restores terminal even on panic (with default config) |

**Key insight:** Bubbletea manages the entire terminal lifecycle. The only custom code needed is the channel bridge and rendering logic.

## Common Pitfalls

### Pitfall 1: Forgetting to Re-register Channel Wait
**What goes wrong:** TUI receives the first request but stops updating after that.
**Why it happens:** The `waitForRequest` command completes after receiving one message. If Update doesn't return a new `waitForRequest` command, no more messages are received.
**How to avoid:** Always return `waitForRequest(m.reqCh)` as the command when handling `requestMsg`.
**Warning signs:** First request appears, subsequent requests only go to log file.

### Pitfall 2: Blocking HTTP Handler on Full Channel
**What goes wrong:** HTTP responses slow down or hang when TUI is busy or channel is full.
**Why it happens:** Unbuffered or small-buffer channel with blocking send.
**How to avoid:** Use a buffered channel (recommend 64-256) and non-blocking `select`/`default` send in the handler. Dropping a TUI update is acceptable since the log file always captures it.
**Warning signs:** HTTP response times increase under load.

### Pitfall 3: Using v1 API in v2
**What goes wrong:** Compilation errors or unexpected behavior.
**Why it happens:** Most tutorials and examples online are for v1. The v2 API has breaking changes.
**How to avoid:** Use `tea.KeyPressMsg` (not `tea.KeyMsg`), `tea.NewView()` (not returning string), `charm.land/bubbletea/v2` import path.
**Warning signs:** "cannot use ... as type" compiler errors.

### Pitfall 4: Writing to stdout Before TUI Starts
**What goes wrong:** "server starting" log message corrupts TUI display.
**Why it happens:** Logger writes to stdout before bubbletea takes over the terminal.
**How to avoid:** Switch logger to file-only BEFORE starting the TUI. Don't log the "server starting" message to stdout at all.
**Warning signs:** Garbled first line of TUI display.

### Pitfall 5: Unbounded Request Slice Growth
**What goes wrong:** Memory grows without limit as requests accumulate.
**Why it happens:** Appending to `[]RequestData` forever.
**How to avoid:** For Phase 3, this is acceptable (ROBU-02 is deferred to v2). But be aware of it. A reasonable approach: keep a max of 10,000 entries in memory for now.
**Warning signs:** Memory usage climbing steadily over time.

## Code Examples

### Request Line Formatting (per user spec)
```go
// Format: "14:32:05 GET /api/users?page=2&limit=1... 200 2ms"
func formatRequestLine(r RequestData) string {
    timestamp := r.Timestamp.Format("15:04:05")

    uri := r.URI
    if len(uri) > 40 {
        uri = uri[:37] + "..."
    }

    var responseTime string
    if r.ResponseTime < time.Millisecond {
        responseTime = "<1ms"
    } else {
        responseTime = fmt.Sprintf("%dms", r.ResponseTime.Milliseconds())
    }

    return fmt.Sprintf("%s %s %s %d %s", timestamp, r.Method, uri, r.Status, responseTime)
}
```

### View Rendering (per user spec)
```go
func (m model) View() tea.View {
    var s strings.Builder

    // Header bar
    s.WriteString(fmt.Sprintf("blackhole :%s -> %s\n", m.port, m.logPath))

    // Separator
    s.WriteString("────────────────────────────────────────\n")

    // Content area
    if len(m.requests) == 0 {
        // Empty state: centered "Waiting for requests..."
        s.WriteString("\n          Waiting for requests...\n\n")
    } else {
        for _, r := range m.requests {
            s.WriteString(formatRequestLine(r))
            s.WriteString("\n")
        }
    }

    // Separator
    s.WriteString("────────────────────────────────────────\n")

    // Status line
    s.WriteString(fmt.Sprintf("%d requests . q to quit", len(m.requests)))

    return tea.NewView(s.String())
}
```

### Auto-scroll Consideration
When the terminal fills up, bubbletea renders from the top of the content string. To achieve auto-scroll (latest at bottom), only render the last N lines that fit in the terminal height. Use `tea.WindowSizeMsg` to track terminal height.

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    return m, nil
```

Then in View(), calculate available lines:
```go
availableLines := m.height - 4 // header(1) + sep(1) + sep(1) + status(1)
startIdx := 0
if len(m.requests) > availableLines {
    startIdx = len(m.requests) - availableLines
}
for _, r := range m.requests[startIdx:] {
    s.WriteString(formatRequestLine(r))
    s.WriteString("\n")
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `github.com/charmbracelet/bubbletea` | `charm.land/bubbletea/v2` | Feb 2026 (v2.0.0) | Must use new import path |
| `View() string` | `View() tea.View` | v2.0.0 | Return `tea.NewView(s)` |
| `tea.KeyMsg` | `tea.KeyPressMsg` | v2.0.0 | Different type switch |
| `tea.WithAltScreen()` option | `view.AltScreen = true` | v2.0.0 | Declarative in View |
| `p.Start()` | `p.Run()` | v2.0.0 | Method renamed |
| `tea.Sequentially()` | `tea.Sequence()` | v2.0.0 | Function renamed |
| Space as `" "` | Space as `"space"` | v2.0.0 | String matching changed |

## Open Questions

1. **Channel closure on TUI exit**
   - What we know: When `tea.Quit` is returned, `Program.Run()` returns. The HTTP server goroutine is still running.
   - What's unclear: Best practice for cleanly shutting down the HTTP server when TUI exits.
   - Recommendation: Use `context.WithCancel` passed to `http.Server.Shutdown()`. The main goroutine cancels context after `p.Run()` returns. For Phase 3, simply exiting the process (which kills all goroutines) is acceptable.

2. **Alt screen usage**
   - What we know: Alt screen provides a clean full-terminal view that restores the previous terminal content on exit.
   - What's unclear: Whether to use alt screen for this "log tail" style TUI.
   - Recommendation: Do NOT use alt screen. A log-tail style TUI should behave like `tail -f` -- stay in the main screen buffer so the user can scroll back in their terminal after exit. This matches the "no summary output, no stdout messages" exit behavior.

3. **Channel buffer size**
   - What we know: This is at Claude's discretion per CONTEXT.md.
   - Recommendation: Use buffer size 256. This handles burst traffic without dropping, and a single `RequestData` struct is ~200 bytes so 256 entries is ~50KB -- negligible memory.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none -- Go testing needs no config file |
| Quick run command | `go test ./...` |
| Full suite command | `go test -v ./...` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TUI-01 | TUI launches on main goroutine, HTTP in background | integration/manual | Manual: `go run .` then verify TUI appears | No -- manual verification |
| TUI-02 | HTTP handler sends to channel, TUI receives | unit | `go test -run TestRequestChannel -v` | No -- Wave 0 |
| TUI-03 | JSON logging to file only, not stdout | unit | `go test -run TestFileOnlyLogging -v` | No -- Wave 0 |

### Sampling Rate
- **Per task commit:** `go build ./...` (compilation check)
- **Per wave merge:** `go test -v ./...`
- **Phase gate:** Full suite green + manual TUI verification

### Wave 0 Gaps
- [ ] `handler_test.go` -- test that handleRequest sends RequestData to channel and logs to file
- [ ] `tui_test.go` -- test model Update handles requestMsg correctly, test formatRequestLine output
- [ ] No framework install needed -- Go testing is built in

## Sources

### Primary (HIGH confidence)
- [bubbletea v2 pkg.go.dev](https://pkg.go.dev/charm.land/bubbletea/v2) - Full API reference, v2.0.1
- [bubbletea v2 upgrade guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md) - All breaking changes from v1 to v2
- [bubbletea realtime example](https://github.com/charmbracelet/bubbletea/blob/main/examples/realtime/main.go) - Channel bridge pattern

### Secondary (MEDIUM confidence)
- [Bubble Tea v2: What's New discussion](https://github.com/charmbracelet/bubbletea/discussions/1374) - Feature overview
- [bubbletea releases](https://github.com/charmbracelet/bubbletea/releases) - Version history

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - verified via pkg.go.dev, official releases, v2.0.1 confirmed
- Architecture: HIGH - channel bridge pattern from official example, three-file split from user decisions
- Pitfalls: HIGH - well-documented in bubbletea community (channel re-registration, blocking sends)
- v2 API: HIGH - verified against official upgrade guide and pkg.go.dev

**Research date:** 2026-03-06
**Valid until:** 2026-04-06 (stable -- bubbletea v2 just released, unlikely to break soon)
