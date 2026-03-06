# Architecture Research

**Domain:** TUI log viewer integration with existing HTTP server
**Researched:** 2026-03-06
**Confidence:** HIGH

## System Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     main() orchestration                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐    chan RequestEntry    ┌──────────────┐  │
│  │  HTTP Server  │ ─────────────────────> │  TUI Program  │  │
│  │  (goroutine)  │                        │  (tea.Run)    │  │
│  └──────┬───────┘                        └──────┬───────┘  │
│         │                                       │           │
│         │ slog.Handler                          │ View()    │
│         v                                       v           │
│  ┌──────────────┐                        ┌──────────────┐  │
│  │  File Logger  │                        │   Terminal    │  │
│  │  (JSON file)  │                        │   Output      │  │
│  └──────────────┘                        └──────────────┘  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### How This Differs From Current Architecture

**Current (v1.0):**
- `main()` creates `io.MultiWriter(os.Stdout, logFile)` and passes it to slog
- `handleRequest(logger)` logs to both stdout and file via that single writer
- Everything is synchronous from the handler's perspective
- `http.ListenAndServe` blocks the main goroutine

**New (v1.1):**
- stdout is no longer a log destination -- TUI owns the terminal
- File logger continues writing JSON via slog (unchanged format)
- HTTP handler sends structured data to TUI via a Go channel
- `tea.Program.Run()` blocks the main goroutine; HTTP server runs in a background goroutine

### Component Responsibilities

| Component | Responsibility | New vs Modified |
|-----------|----------------|-----------------|
| `main()` | Create channel, start HTTP server goroutine, run TUI | **Modified** |
| `handleRequest()` | Log to file, send request data to channel | **Modified** |
| `RequestEntry` | Typed struct for request data passed to TUI | **New** |
| TUI `Model` | Hold request list, selection state, expanded items | **New** |
| TUI `Update()` | Handle keypresses, receive new requests via messages | **New** |
| TUI `View()` | Render scrollable list with expand/collapse | **New** |
| File logger | slog.JSONHandler writing to log file only | **Modified** (remove stdout from MultiWriter) |

## Recommended Project Structure

```
blackhole-server/
├── main.go              # Orchestration: channel, HTTP goroutine, TUI run
├── handler.go           # handleRequest() + RequestEntry type
├── tui.go               # Model, Update, View, message types
└── requests.log         # JSON log output (unchanged)
```

### Structure Rationale

- **3 files, no packages:** This is a small tool. Splitting into `handler.go` and `tui.go` separates HTTP concerns from TUI concerns without over-engineering with subdirectories.
- **main.go stays as orchestrator:** Wires the channel, starts the HTTP server in a goroutine, runs `tea.Program`.
- **No `internal/` or `pkg/`:** Unnecessary for a single-binary tool with 3 source files.

## Architectural Patterns

### Pattern 1: Channel Bridge (HTTP goroutine to TUI event loop)

**What:** A Go channel carries `RequestEntry` values from HTTP handler goroutines into the bubbletea event loop. The TUI model reads from this channel via a `tea.Cmd` that blocks on channel receive.

**When to use:** Whenever external goroutines need to inject data into a bubbletea program. This is the idiomatic bubbletea approach -- use commands (which run in managed goroutines), not `Program.Send()`.

**Why not `Program.Send()`:** While `Program.Send()` works, it creates coupling between the HTTP handler and the `tea.Program` instance. The channel + command pattern is more decoupled and avoids potential deadlock issues if the program exits while a handler is trying to send.

**Trade-offs:** Channel adds a layer of indirection, but is safe, idiomatic Go, and plays well with bubbletea's Elm architecture. The blocking receive in the command naturally waits for the next request without polling.

**Example:**

```go
// handler.go

type RequestEntry struct {
    Timestamp   time.Time
    Method      string
    URI         string
    Status      int
    Duration    time.Duration
    ClientIP    string
    UserAgent   string
    Headers     http.Header
    Body        string
    ContentLen  int64
}

// handleRequest logs to file and sends to TUI channel
func handleRequest(logger *slog.Logger, entries chan<- RequestEntry) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        body, _ := io.ReadAll(r.Body)
        elapsed := time.Since(start)

        // Log to file via slog (unchanged logic)
        logger.LogAttrs(context.Background(), slog.LevelInfo, "request",
            slog.String("method", r.Method),
            // ... same attrs as current code ...
        )

        // Send to TUI (non-blocking with select to avoid backpressure)
        entry := RequestEntry{
            Timestamp:  time.Now(),
            Method:     r.Method,
            URI:        r.RequestURI,
            Status:     200,
            Duration:   elapsed,
            ClientIP:   r.RemoteAddr,
            UserAgent:  r.UserAgent(),
            Headers:    r.Header,
            Body:       string(body),
            ContentLen: r.ContentLength,
        }
        select {
        case entries <- entry:
        default:
            // TUI buffer full, request still logged to file
        }
    }
}
```

```go
// tui.go

// Message type for new requests arriving
type requestMsg RequestEntry

// Command that waits for the next request from the channel
func waitForRequest(entries <-chan RequestEntry) tea.Cmd {
    return func() tea.Msg {
        return requestMsg(<-entries)
    }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case requestMsg:
        m.requests = append(m.requests, RequestEntry(msg))
        // Immediately start waiting for the next request
        return m, waitForRequest(m.entries)
    // ... key handling ...
    }
    return m, nil
}
```

**Confidence:** HIGH -- `tea.Cmd` as a blocking function returning `tea.Msg` is the documented bubbletea pattern for async I/O. Channel receive in a Cmd is shown in official examples and community guides.

### Pattern 2: Elm Architecture with Expand/Collapse State

**What:** The TUI model maintains a flat list of requests plus a `map[int]bool` tracking which indices are expanded. View renders compact or detailed based on this map. Cursor position tracks which request is selected.

**When to use:** Any list UI with detail-on-demand. Simpler than nested models for this use case.

**Trade-offs:** Simple and fast. The map lookup is O(1). For thousands of requests, the slice will grow unbounded -- add a max capacity with oldest-first eviction.

**Example:**

```go
// tui.go

type model struct {
    entries    <-chan RequestEntry  // receive channel
    requests   []RequestEntry      // all received requests
    expanded   map[int]bool        // which indices are expanded
    cursor     int                 // selected index
    offset     int                 // scroll offset for viewport
    height     int                 // terminal height (from WindowSizeMsg)
    width      int                 // terminal width (from WindowSizeMsg)
}

func initialModel(entries <-chan RequestEntry) model {
    return model{
        entries:  entries,
        requests: make([]RequestEntry, 0, 256),
        expanded: make(map[int]bool),
    }
}

func (m model) Init() tea.Cmd {
    return waitForRequest(m.entries)
}
```

**Confidence:** HIGH -- standard Elm architecture pattern. The expand/collapse via map is a common bubbletea community pattern.

### Pattern 3: File-Only slog (TUI replaces stdout)

**What:** Remove `os.Stdout` from the `io.MultiWriter`. The slog handler writes only to the log file. The TUI exclusively owns terminal output via bubbletea's renderer.

**When to use:** Always, when running with TUI. You cannot write to stdout while bubbletea controls the terminal -- it will corrupt the display.

**Trade-offs:** Lose ability to pipe stdout to other tools. Acceptable because the JSON file serves that purpose. Could add a `--no-tui` flag later that falls back to v1.0 stdout behavior.

**Example:**

```go
// main.go

func main() {
    // ... port, logPath setup unchanged ...

    logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil { /* ... */ }
    defer logFile.Close()

    // File-only logger (no stdout -- TUI owns the terminal)
    logger := slog.New(slog.NewJSONHandler(logFile, nil))

    // Channel bridges HTTP handlers to TUI
    entries := make(chan RequestEntry, 64)

    // Start HTTP server in background goroutine
    http.HandleFunc("/", handleRequest(logger, entries))
    go func() {
        if err := http.ListenAndServe(":"+port, nil); err != nil {
            logger.LogAttrs(context.Background(), slog.LevelError, "server failed",
                slog.String("error", err.Error()),
            )
            os.Exit(1)
        }
    }()

    // TUI runs on main goroutine (blocks until quit)
    p := tea.NewProgram(initialModel(entries))
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
        os.Exit(1)
    }
}
```

**Confidence:** HIGH -- bubbletea documentation explicitly states you cannot write to stdout while the program runs. `Program.Println()` exists for printing above the TUI but is for debug use only.

### Pattern 4: View Rendering with tea.View (v2 API)

**What:** In bubbletea v2, `View()` returns a `tea.View` struct instead of a string. Use `tea.NewView(content)` to create it, then set fields for terminal behavior.

**When to use:** All v2 bubbletea programs.

**Example:**

```go
func (m model) View() tea.View {
    var b strings.Builder

    // Render header
    b.WriteString(headerStyle.Render("Blackhole Server - :8080"))
    b.WriteString("\n\n")

    // Render visible request rows
    for i := m.offset; i < len(m.requests) && i < m.offset+m.viewportHeight(); i++ {
        req := m.requests[i]
        if i == m.cursor {
            b.WriteString(selectedStyle.Render(compactLine(req)))
        } else {
            b.WriteString(normalStyle.Render(compactLine(req)))
        }
        b.WriteString("\n")

        if m.expanded[i] {
            b.WriteString(detailView(req))
            b.WriteString("\n")
        }
    }

    // Render footer with keybinds
    b.WriteString(footerStyle.Render("j/k: navigate  enter: expand  c: clear  q: quit"))

    return tea.NewView(b.String())
}
```

**Confidence:** HIGH -- `tea.NewView()` and `tea.View` struct verified in official v2 package docs at pkg.go.dev.

## Data Flow

### Request Flow (v1.1)

```
HTTP Request
    |
    v
handleRequest()
    |
    +---> slog.Logger ---> logFile (JSON, unchanged format)
    |
    +---> entries channel ---> tea.Cmd (waitForRequest blocks on receive)
                                  |
                                  v
                            Update(requestMsg)
                                  |
                                  v
                            model.requests = append(...)
                                  |
                                  v
                            View() renders list
                                  |
                                  v
                            Terminal (bubbletea renderer)
```

### Key Press Flow

```
Keyboard Input
    |
    v
tea.KeyPressMsg
    |
    v
Update() switch on msg.String():
    "j" / "down"   -> cursor++, adjust scroll offset
    "k" / "up"     -> cursor--, adjust scroll offset
    "enter"        -> toggle expanded[cursor]
    "c"            -> clear requests slice + expanded map
    "q" / "ctrl+c" -> return model, tea.Quit()
```

### Key Data Flows

1. **Request ingestion:** HTTP handler goroutine -> buffered channel (cap 64) -> `tea.Cmd` blocks on receive -> `Update` appends to model slice -> `View` re-renders.
2. **User interaction:** Keyboard -> `tea.KeyPressMsg` -> `Update` mutates cursor/expanded state -> `View` re-renders.
3. **File logging:** HTTP handler -> `slog.JSONHandler` -> `os.File`. Completely independent of TUI. Never touches the channel.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 0-10 req/sec | No issues. Channel buffer of 64 is plenty. |
| 10-100 req/sec | Add max request cap (e.g., 1000) with FIFO eviction to prevent unbounded memory growth. |
| 100+ req/sec | TUI rendering becomes the bottleneck. Consider rate-limiting View updates with tick-based batching (render at most 30fps). |

### Scaling Priorities

1. **First bottleneck: Memory from unbounded request slice.** Fix: cap at N entries (1000), drop oldest when full. Simple slice rotation with index adjustment.
2. **Second bottleneck: View rendering with many expanded items.** Fix: only compute visible viewport rows in `View()`, skip off-screen items entirely.

## Anti-Patterns

### Anti-Pattern 1: Using Program.Send() from HTTP Handlers

**What people do:** Pass `*tea.Program` to the HTTP handler and call `p.Send(msg)` directly.
**Why it's wrong:** Creates tight coupling between HTTP and TUI layers. If the TUI exits before all in-flight requests complete, `Send()` can block or deadlock. Multiple handler goroutines all holding a reference to the Program is unnecessary.
**Do this instead:** Use a buffered channel. Handler writes to channel (non-blocking with `select/default`). A `tea.Cmd` reads from the channel inside the event loop.

### Anti-Pattern 2: Writing to stdout While TUI is Running

**What people do:** Keep `os.Stdout` in the MultiWriter or use `fmt.Println` for debugging.
**Why it's wrong:** Bubbletea owns the terminal. Any writes to stdout outside of `View()` will corrupt the TUI display, producing garbled output.
**Do this instead:** Log to file only. Use `p.Println()` for debug messages that need to appear in the terminal (these print above the TUI).

### Anti-Pattern 3: Mutating Model State Outside Update()

**What people do:** Share a pointer to the model and mutate it from the HTTP handler goroutine.
**Why it's wrong:** Race condition. Bubbletea's Elm architecture assumes only `Update()` mutates state. Concurrent mutation from handler goroutines will cause data races that `go test -race` will catch.
**Do this instead:** All state changes go through the message loop. Handler sends data via channel, `Update()` receives it as a `requestMsg` and mutates the model.

### Anti-Pattern 4: Using the Bubbles List Component for This Use Case

**What people do:** Reach for `charm.land/bubbles/v2/list` for any scrollable list.
**Why it's wrong:** The list bubble includes filtering, pagination, status bar, title, and delegate rendering -- features we do not need. It adds complexity and fights against our compact+expand pattern.
**Do this instead:** Build a simple cursor+offset model. The scrolling logic is ~20 lines. The rendering is custom anyway (color-coded methods, expand/collapse). A hand-rolled list is simpler and gives full control.

## Integration Points

### Existing Code Changes

| Current Code | Change Required | Reason |
|-------------|-----------------|--------|
| `io.MultiWriter(os.Stdout, logFile)` | Remove `os.Stdout`, keep only `logFile` | TUI owns terminal |
| `handleRequest(logger)` | Add `entries chan<- RequestEntry` parameter | Bridge to TUI |
| `http.ListenAndServe` in main goroutine | Move to `go func()` | `tea.Run()` needs main goroutine |
| slog log attributes | Unchanged | File logging is independent of TUI |

### New Components

| Component | File | Dependencies |
|-----------|------|-------------|
| `RequestEntry` struct | `handler.go` | None (stdlib types only) |
| `model` struct | `tui.go` | `charm.land/bubbletea/v2` |
| `waitForRequest` cmd | `tui.go` | Channel from `handler.go` |
| `View()` rendering + styles | `tui.go` | `charm.land/lipgloss/v2` |
| `requestMsg` type | `tui.go` | None |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| HTTP handler <-> TUI | Buffered channel (`chan RequestEntry`, cap 64) | Non-blocking send with `select/default` to avoid handler stalls |
| HTTP handler <-> File logger | Direct slog call (synchronous) | Unchanged from v1.0 |
| TUI Model <-> Renderer | `View()` return value (`tea.View`) | Bubbletea manages all terminal I/O |

## Build Order (Dependency-Aware)

Each step produces a working program that can be tested independently:

1. **RequestEntry struct + modified handler** -- Extract `handleRequest` to `handler.go`, define `RequestEntry`, add channel parameter with non-blocking send. File logging still works. Can test with `go run` and curl.

2. **Basic TUI model with channel receive** -- Wire up `tea.NewProgram`, receive `requestMsg` via the channel bridge, display a plain text list of "METHOD /path" lines. Proves the channel bridge works end-to-end. Move HTTP server to background goroutine. Remove stdout from MultiWriter.

3. **Compact one-line rendering with lipgloss** -- Color-coded method badges (GET=green, POST=blue, etc.), formatted path and status. Pure rendering changes, no new state.

4. **Expand/collapse with cursor navigation** -- Add cursor movement (j/k/up/down), expanded map, enter to toggle, detail view showing headers/body/timing. The core interactive feature.

5. **Clear logs keybind** -- Trivial: reset `requests` slice and `expanded` map in `Update()` on "c" keypress.

6. **Polish: terminal resize, visual separation, help footer** -- Handle `tea.WindowSizeMsg` for dynamic viewport sizing. Add horizontal rules between entries. Add footer bar showing available keybinds.

## Sources

- [Bubble Tea v2 package docs (charm.land)](https://pkg.go.dev/charm.land/bubbletea/v2) -- HIGH confidence
- [Bubble Tea v2: What's New discussion](https://github.com/charmbracelet/bubbletea/discussions/1374) -- HIGH confidence
- [Tips for building Bubble Tea programs (leg100)](https://leg100.github.io/en/posts/building-bubbletea-programs/) -- MEDIUM confidence, community patterns
- [Injecting messages from outside the program loop (issue #25)](https://github.com/charmbracelet/bubbletea/issues/25) -- HIGH confidence, official repo
- [Lip Gloss v2 package docs](https://pkg.go.dev/charm.land/lipgloss/v2) -- HIGH confidence
- [Bubbles v2 list package docs](https://pkg.go.dev/charm.land/bubbles/v2/list) -- HIGH confidence

---
*Architecture research for: TUI log viewer integration with blackhole-server*
*Researched: 2026-03-06*
