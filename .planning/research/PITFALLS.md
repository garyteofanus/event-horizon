# Pitfalls Research

**Domain:** Adding bubbletea v2 TUI to existing Go HTTP blackhole server
**Researched:** 2026-03-06
**Confidence:** HIGH

## Critical Pitfalls

### Pitfall 1: Stdout Logging Corrupts TUI Display

**What goes wrong:**
The existing `main.go` uses `io.MultiWriter(os.Stdout, logFile)` for slog output. Once bubbletea starts, it puts the terminal in raw mode and takes exclusive control of stdout for rendering. Any slog output written to stdout (via the MultiWriter) injects raw JSON text into the middle of the TUI, causing garbled display, cursor misplacement, and rendering artifacts.

**Why it happens:**
Bubbletea's renderer manages the terminal buffer cell-by-cell. Writing directly to stdout outside of the `View()` method bypasses this renderer entirely. The renderer has no way to account for text it did not produce, so the display becomes corrupted.

**How to avoid:**
Before starting `tea.Program`, remove `os.Stdout` from the slog MultiWriter. Slog should write ONLY to the log file while the TUI is active. The TUI's `View()` method becomes the sole channel for displaying request data on screen. Bubbletea v2 supports `Program.Println()` / `Program.Printf()` for unmanaged output, but for this project the View-based approach is cleaner.

```go
// WRONG: keep stdout in MultiWriter while TUI runs
writer := io.MultiWriter(os.Stdout, logFile)

// RIGHT: file-only logging when TUI is active
logger := slog.New(slog.NewJSONHandler(logFile, nil))
```

**Warning signs:**
- Garbled terminal output when requests arrive
- TUI "jumps" or flickers on every incoming request
- JSON text appearing between TUI elements

**Phase to address:**
Phase 1 (logging/TUI integration setup) -- this must be resolved before any TUI code renders.

---

### Pitfall 2: Race Conditions Sharing State Between HTTP Handlers and TUI Model

**What goes wrong:**
HTTP handlers run in their own goroutines (one per request). The bubbletea event loop runs in a single goroutine processing messages sequentially. If the HTTP handler directly mutates shared state (e.g., a request log slice) that the TUI model also reads in `View()`, you get a data race. Go's race detector will catch this, but in production it causes corrupted reads, panics, or silently wrong data.

**Why it happens:**
Developers familiar with Go's `http.Handler` concurrency model forget that bubbletea's Update/View cycle is single-threaded by design. The temptation is to append to a shared `[]RequestEntry` from the handler and read it from `View()`. This is a textbook race.

**How to avoid:**
Never share mutable state between HTTP handlers and the bubbletea model. Use `Program.Send(msg)` to inject a custom message from the HTTP handler into bubbletea's event loop. The model's `Update()` method then safely appends to its internal state. `Program.Send()` is goroutine-safe -- it writes to a channel that the event loop reads.

```go
// In HTTP handler:
p.Send(RequestReceivedMsg{Method: r.Method, URI: r.RequestURI, ...})

// In model.Update():
case RequestReceivedMsg:
    m.requests = append(m.requests, msg)
    return m, nil
```

**Warning signs:**
- `-race` detector fires during testing
- Intermittent panics on `index out of range` or `slice header corruption`
- Request data appears partially populated in the TUI

**Phase to address:**
Phase 1 (core integration) -- the message-passing pattern must be the foundation, not bolted on later.

---

### Pitfall 3: Bubbletea v2 API Differences from v1 Examples

**What goes wrong:**
Most bubbletea tutorials, Stack Overflow answers, and blog posts target v1. Using v1 patterns with v2 causes compile errors or subtle behavioral differences. The biggest change: `View()` returns `tea.View` (a struct), not `string`. Program options like `tea.WithAltScreen()` no longer exist -- they are View struct fields. Key matching constants like `tea.KeyCtrlC` are gone.

**Why it happens:**
Bubbletea v2 shipped in late 2025. The v1 ecosystem has years of accumulated content. Developers (and LLMs trained on older data) copy-paste v1 examples and hit confusing compile errors.

**How to avoid:**
Use only `charm.land/bubbletea/v2` import path (not `github.com/charmbracelet/bubbletea`). Reference the official upgrade guide and the v2 pkg.go.dev docs. Key changes to internalize:

| v1 Pattern | v2 Replacement |
|------------|----------------|
| `View() string` | `View() tea.View` |
| `tea.WithAltScreen()` | `view.AltScreen = true` |
| `tea.KeyCtrlC` constant | `msg.String() == "ctrl+c"` |
| `tea.KeyMsg` for all keys | `tea.KeyPressMsg` for presses |
| `tea.Sequentially()` | `tea.Sequence()` |
| `tea.WindowSize()` cmd | `tea.RequestWindowSize` msg |
| `case " ":` for space | `case "space":` |
| `tea.WithMouseCellMotion()` | `view.MouseMode = tea.MouseModeCellMotion` |
| `tea.EnterAltScreen` cmd | Set `view.AltScreen = true` in View() |

**Warning signs:**
- Compile errors about wrong return types on `View()`
- Key handlers that never fire (wrong message type)
- Alt screen not activating despite "setting it up"

**Phase to address:**
Phase 1 (initial bubbletea setup) -- get the v2 boilerplate right from the start.

---

### Pitfall 4: Uncoordinated Lifecycle Between HTTP Server and TUI

**What goes wrong:**
The HTTP server continues running after the user quits the TUI (presses 'q' or ctrl+c). If the handler still calls `p.Send(msg)` after the tea.Program has exited, `Send()` is a no-op (documented as safe), BUT the bigger problem is: the HTTP server keeps accepting requests that nobody sees, and the process hangs or does not exit.

**Why it happens:**
`tea.Program.Run()` blocks until the user quits. `http.ListenAndServe()` also blocks. These two blocking calls must run concurrently, and their lifecycle must be coordinated. Developers often forget to shut down the HTTP server when the TUI exits, or vice versa.

**How to avoid:**
Use a shared `context.Context` with cancellation. When bubbletea's `Run()` returns, cancel the context. Use `http.Server.Shutdown(ctx)` for graceful HTTP shutdown. Run both in goroutines and coordinate.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

srv := &http.Server{Addr: ":8080", Handler: mux}

go func() {
    if err := srv.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()

// Run TUI (blocks until quit)
if _, err := p.Run(); err != nil {
    log.Fatal(err)
}

// TUI exited, shut down HTTP server
srv.Shutdown(context.Background())
```

**Warning signs:**
- Process does not exit after pressing 'q' in the TUI
- Terminal stays in raw mode after quitting (cursor invisible, input echoed wrong)
- Orphaned goroutines from the HTTP server

**Phase to address:**
Phase 1 (lifecycle management) -- must be designed upfront, extremely painful to retrofit.

---

### Pitfall 5: Panic in Update/View Leaves Terminal in Raw Mode

**What goes wrong:**
If any panic occurs inside `Update()` or `View()` (e.g., index out of range on the request slice, nil pointer on an empty model), bubbletea cannot run its cleanup code. The terminal is left in raw mode: no cursor, no echo, line editing broken. The user must manually run `reset` or `stty sane` to recover.

**Why it happens:**
Bubbletea puts the terminal in raw mode at startup and restores it on clean exit. A panic bypasses the deferred cleanup. This is especially dangerous during early development when the model's state management is still being debugged.

**How to avoid:**
Add a `recover()` wrapper around the entire `main()` function that restores terminal state. Also write defensive `View()` code: check slice bounds, handle the zero-state model gracefully.

```go
func main() {
    defer func() {
        if r := recover(); r != nil {
            os.Stdout.WriteString("\033[?25h") // show cursor
            os.Stdout.WriteString("\033c")     // full reset
            fmt.Fprintf(os.Stderr, "panic: %v\n", r)
            os.Exit(1)
        }
    }()
    // ... rest of main
}
```

**Warning signs:**
- Terminal becomes unusable after a crash during development
- Developers having to repeatedly run `reset` in another terminal tab

**Phase to address:**
Phase 1 (boilerplate setup) -- add the recovery wrapper on day one.

---

### Pitfall 6: Unbounded Request Body Read (DoS Vector)

**What goes wrong:**
The current code uses `io.ReadAll(r.Body)` with no size limit. A malicious client can send a multi-gigabyte body, exhausting server memory. This pre-existing issue becomes worse with a TUI because the giant body also has to be stored in the model's request list for display.

**Why it happens:**
Developers think "I just need to log the body" and reach for `io.ReadAll` without considering adversarial inputs.

**How to avoid:**
Wrap `r.Body` with `http.MaxBytesReader` before reading. Choose a sensible limit (e.g., 1 MB).

```go
const maxBodySize = 1 << 20 // 1 MB
r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
body, err := io.ReadAll(r.Body)
```

**Warning signs:**
- `io.ReadAll(r.Body)` anywhere without a preceding `MaxBytesReader`
- Server memory grows under load testing

**Phase to address:**
Phase 1 (handler rewrite) -- fix before exposing to any traffic.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Storing all requests in memory with no cap | Simple implementation | Unbounded memory growth, eventual OOM | MVP only -- add a ring buffer / max cap in Phase 2 |
| Using `fmt.Sprintf` for all View rendering | Quick to write | Slow rendering at high volumes, no color support | Never -- use lipgloss from the start since it is already a dependency |
| Passing `*tea.Program` as global variable | Easy access from HTTP handler | Tight coupling, hard to test | Never -- pass via closure or struct field |
| Skipping `-race` flag in tests | Faster test runs | Race conditions ship to production | Never -- always test with `-race` for this project |
| Monolithic model (one giant Update function) | Fast initial development | Unmaintainable as features grow (expand/collapse, scrolling, clear) | Phase 1 only -- extract sub-models in Phase 2 if needed |
| `io.ReadAll` without `MaxBytesReader` | Simpler code | DoS vulnerability | Never -- always cap body reads |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| slog + bubbletea | Writing slog to stdout while TUI owns the terminal | Redirect slog to file-only; TUI View() is the display channel |
| HTTP handler + tea.Program | Sharing mutable state (slices, maps) across goroutines | Use `p.Send(customMsg)` to inject into the event loop |
| Graceful shutdown | Not coordinating HTTP server and TUI lifecycles | Shared context.Context, cancel on TUI exit, srv.Shutdown() |
| lipgloss v2 + bubbletea v2 | Using lipgloss v1 (incompatible renderer model) | Use `charm.land/lipgloss/v2` -- v2 lipgloss is "pure", bubbletea manages I/O |
| bubbles v2 components | Importing v1 bubbles (viewport, list) with v2 bubbletea | Use `charm.land/bubbles/v2` -- must match bubbletea major version |
| Message ordering | Assuming requests appear in arrival order | Messages from concurrent HTTP handlers arrive non-deterministically; add timestamps and sort in View if order matters |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Unbounded request list in memory | Memory usage grows linearly, eventual OOM | Ring buffer or capped slice (e.g., keep last 1000 requests) | ~10K+ requests in a long-running session |
| Re-rendering entire list on every message | UI lag, dropped keystrokes | Only re-render visible portion; use viewport component for scrolling | >100 visible items being re-rendered per frame |
| Allocating strings in View() on every frame | GC pressure, frame drops | Cache rendered strings, only rebuild on state change | >50 requests/second sustained |
| Sending one message per request field separately | Message queue floods, UI falls behind | Send a single composite message per request containing all data | >10 requests/second |
| Blocking in Update() (e.g., file I/O, parsing) | Entire TUI freezes, keypresses unresponsive | All I/O happens in the HTTP handler before `p.Send()`; Update only processes pre-built messages | Any blocking call in Update freezes everything |
| Reading full body into memory per request | OOM under load, high GC pressure | `MaxBytesReader` with sensible limit | Large bodies under sustained traffic |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Displaying raw request bodies in TUI without size limits | Malicious large body causes terminal flood / memory spike | Truncate displayed body to reasonable limit (e.g., 4KB); log full body to file only |
| Logging sensitive headers (Authorization, Cookie) to file | Credential leakage in log files | Document that the tool logs everything; do not use with production credentials |
| No bind-address restriction | Server listens on all interfaces by default (`:8080`) | Acceptable for a debugging tool, but document the exposure |
| Log file with permissive permissions (0644) | Other users can read captured credentials | Use 0600 permissions on log files |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| No visual feedback when requests arrive | User unsure if server is working | Flash/highlight new entries briefly, show a request counter |
| Expanded detail view shows raw JSON blob | Hard to read, defeats purpose of TUI | Format detail view with labeled fields, syntax highlighting for body |
| No indication of scroll position | User lost in long lists | Show scroll position indicator or "X of Y requests" counter |
| Clear-all has no confirmation | Accidental data loss (display only, file log preserved) | Either add confirmation, or make clear obviously recoverable since file log persists |
| TUI does not show listening port on startup | User must check another source to know where to send requests | Display "Listening on :8080" in a status bar |
| No zero-state message | Blank screen on launch, user wonders if it is broken | Show "Waiting for requests on :8080..." when no requests received yet |

## "Looks Done But Isn't" Checklist

- [ ] **TUI renders requests:** But does it handle the zero-state (no requests yet) gracefully with a helpful message?
- [ ] **Quit works:** But does the HTTP server also shut down, and does the terminal fully restore (cursor, echo, line discipline)?
- [ ] **Requests display:** But are they thread-safe? Run with `go run -race main.go` under concurrent load.
- [ ] **Expand/collapse works:** But does it work correctly when new requests arrive while detail is expanded (list shifting under cursor)?
- [ ] **Color-coded methods:** But do colors degrade gracefully on terminals without true color (e.g., basic 16-color terminals)?
- [ ] **Log file works alongside TUI:** But is the slog handler writing to file-only, not also to the now-TUI-controlled stdout?
- [ ] **Clear visible logs works:** But does it only clear the display, preserving the file log? Is this obvious to the user?
- [ ] **Body displayed in detail view:** But is it truncated to prevent terminal flood from large bodies?
- [ ] **View() returns tea.View:** Not `string` -- v2 signature, not v1.

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Stdout logging corrupts TUI | LOW | Remove stdout from MultiWriter, restart. No data lost (file log intact). |
| Race condition in shared state | MEDIUM | Refactor to message-passing via `p.Send()`. Requires restructuring handler-to-TUI communication. |
| v1 API used instead of v2 | LOW | Mechanical replacement following upgrade guide. Compiler errors guide you. |
| Terminal stuck in raw mode | LOW | Run `reset` or `stty sane`. Add panic recovery wrapper to prevent recurrence. |
| Unbounded memory growth | MEDIUM | Implement ring buffer / capped slice. Requires deciding on retention policy and updating View logic. |
| Lifecycle coordination missing | HIGH | Requires restructuring main() with context cancellation and goroutine coordination. Much harder to retrofit than to design upfront. |
| Unbounded body read | LOW | Add `MaxBytesReader` wrapper; single function change. |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Stdout/TUI conflict | Phase 1: Core TUI setup | slog writes to file only; no garbled output under load |
| Race conditions | Phase 1: HTTP-to-TUI bridge | `go run -race main.go` under concurrent requests shows zero races |
| v2 API misuse | Phase 1: Initial boilerplate | Code compiles with `charm.land/bubbletea/v2`; View returns `tea.View` |
| Lifecycle coordination | Phase 1: Main loop structure | Process exits cleanly when user quits TUI; no orphaned goroutines |
| Panic recovery | Phase 1: Main wrapper | Deliberately trigger panic; terminal recovers automatically |
| Unbounded body read | Phase 1: Handler hardening | `MaxBytesReader` present; test with oversized body |
| Unbounded memory | Phase 2: Hardening | Run with 10K+ requests; memory usage stays bounded |
| View performance | Phase 2: Polish | 100 req/s sustained; TUI stays responsive, no dropped keystrokes |
| UX gaps (zero-state, scroll) | Phase 2: UI refinement | Manual testing of edge cases: empty state, full list, mid-scroll |

## Sources

- [Bubbletea v2 Upgrade Guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md) - Official v1-to-v2 migration reference
- [Bubbletea v2 What's New (Discussion #1374)](https://github.com/charmbracelet/bubbletea/discussions/1374) - Overview of v2 changes
- [Tips for building Bubble Tea programs](https://leg100.github.io/en/posts/building-bubbletea-programs/) - Community best practices on concurrency, message ordering, architecture
- [Injecting messages from outside the program loop (Issue #25)](https://github.com/charmbracelet/bubbletea/issues/25) - Official guidance on `Program.Send()` from external goroutines
- [Race condition fix PR #330](https://github.com/charmbracelet/bubbletea/pull/330) - Historical race condition in renderer
- [Bubbletea v2 pkg.go.dev](https://pkg.go.dev/charm.land/bubbletea/v2) - Official v2 API reference
- [Bubbletea Logging and Debugging (DeepWiki)](https://deepwiki.com/charmbracelet/bubbletea/5.6-logging-and-debugging) - Logging patterns with bubbletea
- [Charm v2 blog post](https://charm.land/blog/v2/) - Official v2 announcement

---
*Pitfalls research for: bubbletea v2 TUI integration with Go HTTP blackhole server*
*Researched: 2026-03-06*
