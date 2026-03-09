# Project Research Summary

**Project:** event-horizon v1.1 — TUI Log Viewer
**Domain:** TUI integration with existing Go HTTP server
**Researched:** 2026-03-06
**Confidence:** HIGH

## Executive Summary

This project adds a real-time TUI request inspector to an existing Go HTTP event-horizon server using the Charm ecosystem (Bubble Tea v2, Lip Gloss v2, Bubbles v2). The domain is well-understood: streaming log viewers with expand/collapse detail are a solved pattern in the Go TUI space, with logtui and lazyjournal as close analogs. The Charm v2 ecosystem, released February 2026, is the only viable choice for interactive Go TUIs and provides all needed primitives (viewport scrolling, styled rendering, keyboard handling) out of the box.

The recommended approach is a channel-bridged architecture: the HTTP server runs in a background goroutine, sends structured request data through a buffered Go channel, and a Bubble Tea program on the main goroutine consumes it via blocking `tea.Cmd` functions. This keeps the Elm architecture intact, avoids race conditions, and cleanly separates HTTP concerns from TUI concerns. The existing slog file logger continues unchanged; only stdout is redirected away from slog to give the TUI exclusive terminal ownership.

The primary risks are all Phase 1 concerns that must be addressed in the initial scaffolding: stdout/TUI conflict (slog must stop writing to stdout before TUI starts), race conditions (all state flows through the bubbletea message loop, never shared mutably), lifecycle coordination (HTTP server must shut down when TUI exits), and v2 API misuse (most online examples target v1, which is incompatible). All four risks have straightforward mitigations documented in the research. The project transitions from zero external dependencies to three Charm packages -- this is the first external dependency addition and should be treated deliberately.

## Key Findings

### Recommended Stack

The existing Go 1.25 + stdlib stack remains unchanged. Three Charm ecosystem packages are added, all at v2 to ensure mutual compatibility.

**Core technologies:**
- **Bubble Tea v2** (`charm.land/bubbletea/v2@v2.0.1`): TUI framework with Elm architecture -- the standard for interactive Go TUIs, no viable alternative
- **Lip Gloss v2** (`charm.land/lipgloss/v2@v2.0.0`): Terminal styling with automatic color downsampling -- companion to Bubble Tea, handles color profiles gracefully
- **Bubbles v2** (`charm.land/bubbles/v2@v2.0.0`): Pre-built components (viewport for scrolling, help for keybind display, key for keybind definitions) -- avoids reimplementing scroll logic

**Critical constraint:** All three Charm packages MUST be v2. Mixing v1 and v2 causes compile errors and runtime I/O conflicts. The v2 module paths use `charm.land/` not `github.com/charmbracelet/`.

**Not needed:** bubbles/list (too opinionated), bubbles/table, bubbles/textinput, glamour, wish. A hand-rolled cursor+offset list on top of viewport gives full control over the compact+expand rendering pattern.

### Expected Features

**Must have (table stakes -- v1.1):**
- Live streaming display of requests as they arrive
- Compact one-line-per-request list (method, path, status, timestamp)
- Expand/collapse request detail with Enter (headers, body, client IP, timing)
- Color-coded HTTP methods (GET=green, POST=blue, PUT=orange, DELETE=red)
- Color-coded status codes (2xx=green, 4xx=yellow, 5xx=red)
- Keyboard navigation (j/k, arrows, page up/down)
- Clear visible logs (c key, display only, file preserved)
- Status bar with keybind hints and request counter
- JSON file logging continues unchanged

**Should have (add after v1.1 stabilizes):**
- Auto-scroll to newest with manual override (pause when user scrolls up)
- JSON body pretty-printing with indentation
- Relative timestamps ("2s ago") in compact view
- Responsive layout with graceful truncation on narrow terminals

**Defer (v2+):**
- Request filtering/search (use log file + jq for now)
- Mouse support
- Split pane view
- Color theme configuration

### Architecture Approach

Three-file structure (`main.go`, `handler.go`, `tui.go`) in a single package. No subdirectories -- this is a small tool. The HTTP server moves to a background goroutine; `tea.Program.Run()` takes the main goroutine. A buffered channel (cap 64) bridges request data from HTTP handlers into the TUI event loop via a blocking `tea.Cmd`. File logging is fully independent of the TUI.

**Major components:**
1. **main.go** — Orchestration: creates channel, starts HTTP server goroutine, runs TUI, coordinates shutdown
2. **handler.go** — `RequestEntry` struct + modified `handleRequest()` that logs to file AND sends to channel (non-blocking)
3. **tui.go** — Bubble Tea Model/Update/View, message types, Lip Gloss styles, all rendering logic

### Critical Pitfalls

1. **Stdout logging corrupts TUI** — Remove `os.Stdout` from slog MultiWriter before TUI starts. File-only logging while TUI is active. Must be done in Phase 1 scaffolding.
2. **Race conditions between HTTP handlers and TUI** — Never share mutable state. All request data flows through the channel into `Update()`. Always test with `go run -race`.
3. **Bubble Tea v2 API differences** — `View()` returns `tea.View` not `string`, keys use `tea.KeyPressMsg` not `tea.KeyMsg`, alt screen is a View field not a command. Follow only v2 docs.
4. **Uncoordinated HTTP/TUI lifecycle** — Use `context.Context` with cancellation. When TUI exits, shut down HTTP server via `srv.Shutdown()`. Design upfront, painful to retrofit.
5. **Panic leaves terminal in raw mode** — Add `recover()` wrapper in `main()` that resets terminal state. Essential during development.
6. **Unbounded body read (DoS)** — Wrap `r.Body` with `http.MaxBytesReader` (1 MB limit). Pre-existing issue that gets worse with TUI memory.

## Implications for Roadmap

Based on research, the project naturally divides into 4 phases following the architecture's dependency chain.

### Phase 1: Core TUI Scaffolding and HTTP Bridge

**Rationale:** Every other feature depends on the TUI being wired up and receiving requests. All 6 critical pitfalls must be addressed here. This is the foundation.
**Delivers:** A working TUI that displays plain-text request lines as they arrive. HTTP server runs in background. File logging works. Process exits cleanly on quit.
**Addresses:** Live streaming display, JSON file logging preserved, quit keybind
**Avoids:** Stdout/TUI conflict, race conditions, lifecycle coordination issues, v2 API misuse, panic recovery, unbounded body read
**Stack:** Bubble Tea v2 (core), channel bridge pattern, file-only slog
**Key tasks:**
- Add Charm v2 dependencies to go.mod
- Extract handler to `handler.go` with `RequestEntry` struct and channel parameter
- Add `MaxBytesReader` to body reads
- Create minimal `tui.go` with Model/Update/View receiving requests via channel
- Move HTTP server to goroutine, TUI on main goroutine
- Coordinate shutdown with context cancellation
- Add panic recovery wrapper

### Phase 2: Compact List Rendering with Styles

**Rationale:** Once the bridge works, the next priority is making the list scannable. Color-coding and formatting are pure rendering changes with no new state management.
**Delivers:** Color-coded compact request list with method badges, formatted paths, timestamps, visual separation between entries.
**Addresses:** Color-coded HTTP methods, color-coded status codes, visual separation, compact one-line list view
**Uses:** Lip Gloss v2 for all styling
**Avoids:** Using fmt.Sprintf instead of lipgloss (technical debt pattern)

### Phase 3: Interactive Features (Navigate, Expand, Clear)

**Rationale:** Requires the styled list from Phase 2 as the visual foundation. Adds the core interactivity that differentiates this from `tail -f`.
**Delivers:** Cursor-based navigation, expand/collapse detail view with headers/body/timing, clear display command, help footer.
**Addresses:** Keyboard navigation, expand/collapse detail, clear visible logs, help indicator, status bar with request counter
**Avoids:** Expand/collapse bugs when new requests arrive (cursor stability), displaying unbounded body content in detail view (truncate to 4KB display)

### Phase 4: Polish and Hardening

**Rationale:** Refinement pass after core features are stable. Addresses performance under sustained load and UX edge cases.
**Delivers:** Zero-state message ("Waiting for requests..."), terminal resize handling, scroll position indicator, bounded memory (capped request list), listening port in status bar.
**Addresses:** Auto-scroll with override, responsive layout, request counter
**Avoids:** Unbounded memory growth (cap at 1000 entries), View performance degradation (render only visible rows)

### Phase Ordering Rationale

- **Phase 1 before anything:** The channel bridge and TUI scaffolding are the substrate. Nothing visual can be built without them. All critical pitfalls map to this phase.
- **Phase 2 before Phase 3:** Styling must exist before interactive navigation, because expand/collapse rendering depends on the compact line format being defined.
- **Phase 3 is the feature core:** This is where the tool becomes genuinely useful beyond `tail -f | jq`. It is the largest phase by feature count.
- **Phase 4 is separate from Phase 3:** Hardening and polish should not be mixed with feature development. Bounded memory and resize handling are stability concerns, not features.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 1:** Needs `/gsd:research-phase` — the Bubble Tea v2 API is new (Feb 2026), and the channel-to-Cmd bridge pattern needs exact implementation details verified against v2. The v2 `tea.View` return type and lifecycle coordination are not yet widely documented in community examples.

Phases with standard patterns (skip research-phase):
- **Phase 2:** Standard Lip Gloss styling — well-documented, many examples available. Color maps and border styles are straightforward.
- **Phase 3:** Standard Elm architecture state management — expand/collapse via map, cursor+offset scrolling. Well-established bubbletea pattern.
- **Phase 4:** Standard hardening — ring buffer, viewport optimization, resize handling. Generic Go patterns, nothing TUI-specific.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Charm v2 is the only viable Go TUI framework. Versions verified against pkg.go.dev. Compatibility matrix confirmed. |
| Features | HIGH | Feature set derived from competitor analysis (logtui, lazyjournal, browser devtools) and PROJECT.md requirements. Clear MVP boundary. |
| Architecture | HIGH | Channel bridge pattern is documented in official bubbletea examples. Three-file structure matches project scale. Data flow is straightforward. |
| Pitfalls | HIGH | All pitfalls sourced from official docs, issue trackers, and community post-mortems. Each has a concrete prevention strategy. |

**Overall confidence:** HIGH

### Gaps to Address

- **Bubble Tea v2 is 1 week old:** The v2.0.1 release is from late February 2026. While the API is documented, community patterns and edge-case discoveries are still emerging. Phase 1 implementation should be treated as slightly exploratory.
- **No automated TUI testing strategy:** Research did not cover how to test bubbletea programs (golden file tests, `teatest` package). This should be investigated if test coverage is desired.
- **`--no-tui` fallback flag:** Research identified that removing stdout from slog breaks pipe-based workflows. A `--no-tui` flag reverting to v1.0 stdout behavior was suggested but not scoped. Consider for Phase 4.

## Sources

### Primary (HIGH confidence)
- [Bubble Tea v2 pkg.go.dev](https://pkg.go.dev/charm.land/bubbletea/v2) — v2 API reference, View type, Program.Send()
- [Lip Gloss v2 pkg.go.dev](https://pkg.go.dev/charm.land/lipgloss/v2) — v2 styling API, color downsampling
- [Bubbles v2 pkg.go.dev](https://pkg.go.dev/charm.land/bubbles/v2) — viewport, help, key components
- [Bubble Tea v2 Release Notes](https://github.com/charmbracelet/bubbletea/releases/tag/v2.0.0) — breaking changes from v1
- [Bubble Tea v2 Upgrade Guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md) — v1-to-v2 migration
- [Charm v2 Blog Post](https://charm.land/blog/v2/) — official announcement
- [Issue #25: Injecting messages from outside](https://github.com/charmbracelet/bubbletea/issues/25) — Program.Send() and channel patterns

### Secondary (MEDIUM confidence)
- [Tips for building Bubble Tea programs](https://leg100.github.io/en/posts/building-bubbletea-programs/) — community patterns for concurrency and architecture
- [logtui](https://github.com/jnatten/logtui) — closest competitor, feature comparison
- [lazyjournal](https://github.com/Lifailon/lazyjournal) — TUI log viewer patterns and keyboard conventions

### Tertiary (LOW confidence)
- Stack Overflow / community blog posts — mostly v1 content, used only to identify anti-patterns to avoid

---
*Research completed: 2026-03-06*
*Ready for roadmap: yes*
