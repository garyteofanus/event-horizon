# Feature Research: TUI Log Viewer

**Domain:** TUI request inspector / live log viewer for HTTP debugging tool
**Researched:** 2026-03-06
**Confidence:** HIGH

## Feature Landscape

### Table Stakes (Users Expect These)

Features every TUI log viewer is expected to have. Missing these and the tool feels broken or half-baked compared to even basic alternatives like `tail -f | jq`.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Live streaming display | Requests must appear in the TUI the moment they arrive. A log viewer that requires refresh is useless for debugging | MEDIUM | Requires goroutine-safe channel from HTTP handler to bubbletea model via `tea.Cmd` or `tea.Program.Send()` |
| Compact one-line-per-request list | Users need to scan many requests quickly. Every TUI log viewer (logtui, lazyjournal, otel-tui) uses a summary list as the primary view | MEDIUM | Show: method, path, status code, timestamp. Use bubbles `list` or custom viewport |
| Expand/collapse request detail | Users need to drill into a specific request's headers, body, timing without leaving the TUI. This is the equivalent of clicking a row in Postman's console | MEDIUM | Toggle with Enter key. Show full headers, body, client IP, response time, User-Agent in an expanded pane or inline |
| Keyboard navigation (up/down/scroll) | Arrow keys and j/k for moving through the request list is the absolute minimum. Every TUI tool supports this | LOW | j/k (vim-style) plus arrow keys. Page up/down for large lists. Home/End or g/G to jump |
| Color-coded HTTP methods | Visual differentiation of methods is universal in HTTP tools (Postman, Insomnia, HTTPie, browser devtools). Without color, the list is a wall of text | LOW | Convention from Postman/Swagger: GET=green, POST=blue, PUT=yellow/orange, DELETE=red, PATCH=cyan. Use lipgloss styles |
| Color-coded status indicators | Status codes need instant visual parsing. 2xx=green, 4xx=yellow, 5xx=red is a universal convention across devtools, browsers, and CLI tools | LOW | All responses are 200 currently, but build the pattern now for future flexibility |
| Quit keybind | q or Ctrl+C to exit cleanly. Every TUI app has this | LOW | Built into bubbletea's default key handling |
| Help indicator | Users need to know what keys are available. At minimum, a status bar showing key hints | LOW | Footer bar: "q quit | Enter expand | c clear | ? help" |
| JSON file logging continues | The TUI replaces stdout, but the file log must keep working. Users depend on the log file for post-mortem analysis | LOW | Redirect slog output to file-only writer; TUI becomes the "stdout" display. Existing `io.MultiWriter` drops stdout, keeps file |
| Visual separation between entries | Without borders or spacing, adjacent requests blur together in a list | LOW | Use lipgloss borders, alternating background colors, or simple separator lines |

### Differentiators (Competitive Advantage)

Features that make this tool feel polished beyond "I just piped JSON to a TUI." Not required, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Clear all visible logs keybind | During active debugging, old requests are noise. One keypress to start fresh without restarting the server is genuinely useful. Tools like logtui have pause/resume (s key) but clearing is more direct | LOW | c key to clear the in-memory request list. Does NOT clear the log file. Just resets the TUI display |
| Auto-scroll to newest with manual override | New requests should appear at bottom and auto-scroll, BUT if user scrolls up to inspect an old request, auto-scroll should pause. Resume when user returns to bottom | MEDIUM | Track whether user is "at bottom." If yes, auto-scroll. If no, show a "N new requests" indicator. Common pattern in chat apps and log viewers |
| Request body pretty-printing | JSON bodies displayed with indentation and syntax coloring are dramatically easier to read than raw strings. Most HTTP inspector tools do this | MEDIUM | Detect JSON content-type, parse and re-indent with `json.Indent`. Use lipgloss for key/value coloring. Fall back to raw string for non-JSON |
| Timestamp display (relative and absolute) | Show "2s ago" in compact view for quick scanning, full ISO timestamp in detail view for correlation with external systems | LOW | `time.Since(requestTime)` for relative; stored `time.Time` for absolute in detail pane |
| Request counter in status bar | "42 requests captured" gives immediate feedback that the server is working and how busy it has been | LOW | Atomic counter, displayed in header/footer bar |
| Responsive layout | TUI should adapt to terminal width. Narrow terminals should truncate paths gracefully rather than wrapping and breaking the layout | MEDIUM | Use lipgloss `Width()` and truncate long paths with ellipsis. Test with 80-column and 120-column terminals |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem appealing but conflict with the project's simplicity constraint or add disproportionate complexity.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Request filtering/search | Users with high traffic want to find specific requests | Adds significant UI complexity (search input, filter state, no-results handling). Explicitly out of scope for v1.1 per PROJECT.md | Use the JSON log file with `jq` or `grep` for filtering. Consider for v1.2 |
| Log export/replay | Users want to save or replay captured requests | Turns a viewer into a testing tool. Different product category (see: Hurl, Slumber, Posting) | Log file already captures everything in JSON. Export is already done |
| Mouse support | Click to expand, scroll with mouse wheel | Adds complexity and testing surface. Keyboard-first is the right approach for a developer tool. Bubbletea supports mouse but it is extra state to manage | Keyboard navigation covers all use cases. Mouse can be added later if demanded |
| Multiple panes/split view | Side-by-side list and detail view like logtui | Significantly increases layout complexity. For a simple request inspector, inline expand/collapse is sufficient and keeps the codebase small | Inline expand within the list is simpler, works at any terminal width, and is easier to implement |
| Configurable color themes | Users want Dracula, Nord, Monokai, etc. | Theme support is a rabbit hole. Multiple themes means testing multiple themes. One good default is better than five mediocre options | Pick one clean color scheme that works on both dark and light terminals. Use lipgloss adaptive colors |
| Persistent TUI state | Remember scroll position, expanded items across restarts | Adds file I/O, serialization, and edge cases for stale state. The tool is ephemeral by nature | Fresh state on every start. The log file provides persistence |
| WebSocket/SSE support | Capture non-HTTP protocols | Different protocol handling, different display needs. Scope creep | Stick to HTTP. Different protocols need different tools |

## Feature Dependencies

```
HTTP Server (existing)
    |
    +-- Request Channel -----> TUI Model (new)
    |   (goroutine-safe)           |
    |                              +-- Compact List View
    |                              |       |
    |                              |       +-- Keyboard Navigation
    |                              |       |
    |                              |       +-- Color-Coded Methods
    |                              |       |
    |                              |       +-- Color-Coded Status
    |                              |
    |                              +-- Expand/Collapse Detail
    |                              |       |
    |                              |       +-- Body Pretty-Print (enhances)
    |                              |
    |                              +-- Clear Logs Keybind
    |                              |
    |                              +-- Auto-Scroll Logic
    |                              |
    |                              +-- Status Bar (help + counter)
    |
    +-- File Logger (existing, modified)
        slog writes to file only; TUI replaces stdout
```

### Dependency Notes

- **Request Channel requires HTTP server modification:** The handler must send captured request data to the TUI via a channel or `tea.Program.Send()`. This is the key integration point between existing code and new TUI code.
- **Compact List View is the foundation:** Every other visual feature (colors, expand/collapse, navigation) depends on the list being rendered first.
- **Expand/Collapse requires Compact List:** Cannot expand what does not exist in a list. The data model must track per-item expanded state.
- **Body Pretty-Print enhances Expand/Collapse:** Only visible when a request is expanded. Not a dependency, but only useful in that context.
- **File Logger modification is independent:** Can be done in parallel with TUI work. Just change `io.MultiWriter(os.Stdout, file)` to write only to `file`.

## MVP Definition

### Launch With (v1.1)

Minimum to replace raw stdout JSON with a useful TUI. Matches PROJECT.md requirements exactly.

- [ ] **Live request streaming to TUI** -- without this, the TUI is static and useless
- [ ] **Compact one-line list view** -- method (colored), path, status, timestamp per row
- [ ] **Expand/collapse with Enter** -- show headers, body, client IP, response time in expanded view
- [ ] **Color-coded HTTP methods** -- GET=green, POST=blue, PUT=orange, DELETE=red, PATCH=cyan, others=white
- [ ] **Color-coded status codes** -- 2xx=green, 3xx=cyan, 4xx=yellow, 5xx=red (future-proof even though all are 200 now)
- [ ] **Keyboard navigation** -- j/k or arrows for up/down, Enter to toggle expand, q to quit
- [ ] **Clear visible logs** -- c key to reset the display (not the file)
- [ ] **Visual separation** -- alternating styles or borders between list entries
- [ ] **Status bar** -- show keybind hints and request count
- [ ] **JSON file logging preserved** -- slog writes to file; TUI replaces stdout

### Add After Validation (v1.x)

Features to add once the core TUI is stable and usable.

- [ ] **Auto-scroll with manual override** -- add when users complain about losing their scroll position during high traffic
- [ ] **JSON body pretty-printing** -- add when users report difficulty reading request bodies in the detail view
- [ ] **Relative timestamps** -- add "2s ago" display when users request quicker temporal scanning
- [ ] **Responsive truncation** -- add when users report layout issues on narrow terminals
- [ ] **Request filtering (/ key)** -- add when users have enough traffic that scanning the list is slow. This is the natural next feature after v1.1

### Future Consideration (v2+)

Features to defer until the TUI is proven and users are actively requesting them.

- [ ] **Mouse support** -- defer until keyboard-first UX is polished
- [ ] **Split pane view** -- defer; inline expand is sufficient for the request volume this tool handles
- [ ] **Color theme configuration** -- defer; one good default is enough
- [ ] **Request diffing** -- compare two requests side-by-side; useful but niche
- [ ] **Export selected requests** -- copy a request as curl command; cool but scope creep for now

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Live request streaming | HIGH | MEDIUM | P1 |
| Compact list view | HIGH | MEDIUM | P1 |
| Expand/collapse detail | HIGH | MEDIUM | P1 |
| Color-coded methods | HIGH | LOW | P1 |
| Keyboard navigation (j/k/arrows) | HIGH | LOW | P1 |
| Clear logs keybind | MEDIUM | LOW | P1 |
| Status bar with hints | MEDIUM | LOW | P1 |
| Visual separation | MEDIUM | LOW | P1 |
| Color-coded status codes | MEDIUM | LOW | P1 |
| JSON file logging preserved | HIGH | LOW | P1 |
| Auto-scroll with override | MEDIUM | MEDIUM | P2 |
| JSON body pretty-printing | MEDIUM | MEDIUM | P2 |
| Relative timestamps | LOW | LOW | P2 |
| Responsive layout | MEDIUM | MEDIUM | P2 |
| Request filtering | HIGH | HIGH | P3 |
| Mouse support | LOW | MEDIUM | P3 |

**Priority key:**
- P1: Must have for v1.1 launch (matches PROJECT.md requirements)
- P2: Should have, add in v1.x when core is stable
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | logtui | lazyjournal | otel-tui | Browser DevTools | Our Approach |
|---------|--------|-------------|----------|-----------------|--------------|
| Live streaming | Yes (stdin/file) | Yes (journald/docker) | Yes (OTLP receiver) | Yes (network tab) | Yes (internal channel from HTTP handler) |
| Compact + detail | List + detail pane (side by side) | Table + log output | Trace list + span detail | Request list + detail panel | Inline expand/collapse (simpler, no split pane) |
| Filtering | Regex (/ key) | Fuzzy + regex + priority | Search (/ key) | Text filter + type filter | Deferred to v1.2. Use log file + jq for now |
| Color coding | 11+ themes | 6 color groups | Minimal | Full (method + status) | Single clean theme, method + status colors |
| Keyboard nav | Vim-style (h/j/k/l) | Vim-style + F-keys | Tab + letter keys | Mouse-primary | Vim-style (j/k) + arrows |
| Clear/reset | Ctrl+L (screen) | No | No | Clear button | c key (clear request list) |
| Pause intake | s key | No | No | No | Not in v1.1; consider later |
| Column control | c key (toggle/reorder) | No | No | Column picker | Not needed; fixed compact format |

**Key insight from competitors:** logtui is the closest analog -- a keyboard-first TUI for streaming JSON logs with a details pane. But logtui is a general-purpose JSON log viewer. Our tool is purpose-built for HTTP request inspection, which means we can optimize the compact view specifically for method/path/status rather than arbitrary JSON fields. This specialization is our advantage.

## Sources

- [logtui - Keyboard-first TUI for streaming JSON logs](https://github.com/jnatten/logtui) -- closest competitor; list + detail + regex filtering
- [lazyjournal - TUI for journald/Docker/K8s logs](https://github.com/Lifailon/lazyjournal) -- color highlighting patterns, keyboard conventions
- [otel-tui - Terminal OpenTelemetry viewer](https://github.com/ymtdzzz/otel-tui) -- foldable spans pattern for expand/collapse
- [Postman HTTP method colors](https://github.com/postmanlabs/postman-app-support/issues/1337) -- GET=green, POST=orange, DELETE=red convention
- [Mozilla DevTools status code colors](https://bugzilla.mozilla.org/show_bug.cgi?id=1417805) -- 2xx=green, 4xx=red, 5xx=pink convention
- [HTTPie colors and formatting docs](https://httpie.io/docs/cli/colors-and-formatting) -- terminal color scheme patterns
- [Posting - modern API client TUI](https://github.com/darrenburns/posting) -- keyboard-centric workflow patterns
- [Bubble Tea framework](https://github.com/charmbracelet/bubbletea) -- TUI framework, Elm architecture
- [Bubbles component library](https://github.com/charmbracelet/bubbles) -- list, viewport, spinner components

---
*Feature research for: TUI request inspector for event-horizon v1.1*
*Researched: 2026-03-06*
