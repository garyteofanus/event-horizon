# Phase 5: Interactive Features and Polish - Research

**Researched:** 2026-03-06
**Domain:** Bubble Tea v2 interaction state, expandable request rows, resize-safe rendering
**Confidence:** HIGH

## Summary

Phase 5 should stay inside the existing Bubble Tea model and rendering seams rather than introducing new framework layers. The current code already has the right foundation: `model` owns terminal width/height, `renderView` is a deterministic render helper, and `RequestData` already covers the compact row. The real work is extending in-memory request data for detail view, adding explicit selection/expansion state, and making scrolling operate on rendered row height instead of raw request count once expanded content exists.

The most important planning decision is state semantics. If the phase is planned loosely, the implementation will drift on three questions: what exactly is selected, whether more than one row may be expanded, and what "clear visible entries" means when the viewport only shows a subset of `m.requests`. The lowest-risk plan is:

- keep one selected row via `selectedIndex`
- allow one expanded row at a time via `expandedIndex` (or `-1` when none)
- treat clear as clearing the TUI's in-memory request list, never the log file
- compute viewport from row heights so resize and expansion share one render path

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| INTR-01 | Navigate entries with `j/k` or arrow keys | Add cursor state and key handling in `model.Update` |
| INTR-02 | Expand/collapse a selected entry to show details | Extend `RequestData` with detail fields already available in the handler |
| INTR-03 | Clear visible log entries with a keybind | Clear only TUI memory; leave file logging and future requests untouched |
| INTR-04 | Footer shows available keybindings | Replace current status line with status + help footer composition |
| ROBU-01 | Resize re-renders correctly | Reuse `tea.WindowSizeMsg` plus deterministic row-height-aware rendering |

## Current Codebase Fit

### Existing Seams

- `handler.go`
  - `RequestData` is the bridge payload from HTTP handler to TUI
  - the handler already captures body, headers, client IP, and response timing for file logging
  - detail fields are available at capture time but are not yet sent to the TUI
- `tui.go`
  - `model` already stores `requests`, `width`, and `height`
  - `Update` already handles `tea.KeyPressMsg`, `requestMsg`, and `tea.WindowSizeMsg`
  - `renderView` already centralizes layout and is the correct place to keep resize-safe output logic
  - `renderRequestRow` is the right seam to split compact row rendering from expanded detail rendering
- `tui_test.go`
  - tests already assert render output through `renderView`
  - ANSI-safe assertions already exist, so interaction tests can stay string-based instead of spinning a full program

### Gaps To Close

The current bridge type is too small for INTR-02. `RequestData` will need, at minimum:

- `Headers http.Header` or a render-ready header representation
- `Body string`
- `ClientIP string`
- possibly `Protocol string` if planners want detail view to mirror the file log more closely

The current clipping logic assumes each request consumes one screen line. Once rows expand, clipping must be based on rendered row height, not `len(reqs)`.

## Recommended State Model

Add these fields to `model`:

- `selectedIndex int`
- `expandedIndex int`
- `scrollOffset int`

Recommended semantics:

- `selectedIndex` points to the active request and is clamped after append, clear, and resize
- `expandedIndex == selectedIndex` means the selected row is expanded
- `expandedIndex = -1` means all rows are collapsed
- `scrollOffset` tracks the first request index in the viewport

Why this shape:

- one selected row is required for keyboard navigation
- one expanded row keeps viewport math simple and predictable
- append-only request arrival means index-based state is stable enough for this phase
- clearing can atomically reset all three fields without hidden stale state

If multi-expand is desired later, it should be a future phase. It adds viewport complexity without being required by the roadmap.

## Architecture Patterns

### Pattern 1: Detail Data Stays In The Existing Bridge

Do not read the log file back into the TUI and do not invent a second transport path. The handler already has the needed detail fields in memory when the request arrives. Extend `RequestData` once and keep the TUI consuming the same channel.

### Pattern 2: Separate Row Summary From Row Detail

Keep two render helpers in `tui.go`:

- `renderRequestRow(...)` for the compact one-line summary
- `renderExpandedDetails(...)` for headers/body/client IP/response time

Then compose them in a higher-level helper such as `renderRequestBlock(...)`. This keeps `renderView` orchestration-focused and makes row-height calculation testable.

### Pattern 3: Viewport Logic Should Follow Rendered Block Height

Once a row can expand to multiple lines, visible content should be calculated from block height. A practical approach is:

1. build per-request rendered blocks
2. measure each block with `strings.Count(block, "\n") + 1`
3. select the slice that fits into available content height
4. render only that slice

That same logic should run after request append, key navigation, expand/collapse, clear, and resize.

### Pattern 4: Footer Is A Stable Layout Region

The current footer is only a status line. Phase 5 should reserve the bottom region for:

- request count / selection summary
- keybinding help

Keep it fixed-height so resize logic stays deterministic.

## Interaction Contract

Recommended keybind set:

- `j` / `down`: move selection down
- `k` / `up`: move selection up
- `enter` or `space`: expand/collapse selected request
- `x` or `c`: clear in-memory requests
- `q` / `ctrl+c`: quit

Recommended behavior details:

- moving selection does not auto-expand
- expanding a row keeps the cursor on that row
- new requests still append at bottom; selection does not jump to newest if the user has moved off the tail
- if the selected row is cleared, selection resets to the last remaining row or zero
- if no requests exist, navigation and expand keys are no-ops and the empty state remains visible

## Clear Behavior Semantics

This is the one area the plan should make explicit before execution.

The requirement text says "visible log entries only", but the current TUI has no filtering, pagination, or hidden backing store beyond `m.requests`. If clear is implemented as "delete only rows currently on screen", old clipped entries would suddenly reappear, which is surprising and creates index churn.

**Recommendation:** define clear as "clear all requests currently retained in the TUI memory buffer; do not touch the file log and do not affect future incoming requests." This matches user intent, keeps behavior simple, and avoids a viewport-only deletion edge case.

## Resize Handling

ROBU-01 is mostly architectural discipline, not a new dependency.

What already works:

- `tea.WindowSizeMsg` is already handled
- rendering is already pure through `renderView`

What must change:

- available content height must subtract header, separators, and footer/help lines
- expanded rows must wrap or truncate safely within `m.width`
- viewport recalculation must happen after every `tea.WindowSizeMsg`

Recommended approach:

- use `lipgloss.Width`/`lipgloss.Size` aware helpers when styling wrapped detail blocks
- perform width-sensitive truncation before ANSI styling where possible
- keep detail rendering textual and narrow-friendly instead of trying to make a multi-column inspector

## Standard Stack

| Library | Version | Purpose | Why |
|---------|---------|---------|-----|
| `charm.land/bubbletea/v2` | existing | input handling, resize events, model updates | already adopted in Phase 3 |
| `charm.land/lipgloss/v2` | existing | styling for selected and expanded states | already adopted in Phase 4 |
| Go stdlib `net/http` | existing | source of request detail data | handler already owns all required inputs |

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| terminal resize plumbing | custom SIGWINCH handling | Bubble Tea `tea.WindowSizeMsg` | already integrated and portable |
| ANSI width math by raw string length | manual escape-sequence counting | Lip Gloss width-aware rendering + pre-style truncation | avoids broken wrapping and alignment |
| secondary request store for detail view | re-parsing log file or shadow copies | extend `RequestData` on the channel bridge | simpler and consistent with current architecture |
| complex cursor widget | custom framework layer | a few fields on `model` plus conditional row styling | scope is too small to justify abstraction |

## Common Pitfalls

### Pitfall 1: Expanding `RequestData` With Unstable Render-Only Strings

Store structured detail data where reasonable. If headers are flattened too early into one giant string, formatting changes become harder to test and wrap on resize.

### Pitfall 2: Keeping The Old "one request == one line" Scroll Logic

That logic will fail as soon as an expanded row consumes several lines. The planner should budget explicit viewport work, not treat resize as a tiny patch.

### Pitfall 3: Coupling Selection To New Request Arrival

Auto-jumping the cursor to every new request makes manual inspection frustrating. Cursor state and append-at-bottom behavior should be independent.

### Pitfall 4: Letting Clear Semantics Depend On Current Terminal Height

If clear deletes only the currently rendered slice, the same keypress has different meaning after resize. Tie clear to TUI memory, not viewport height.

### Pitfall 5: Forgetting Empty-State and Boundary Conditions

Navigation, expand, and clear must all behave safely with zero requests, one request, and after a clear followed by new incoming traffic.

## Validation Architecture

This phase is still strongly unit-testable because `renderView` remains the primary seam and Bubble Tea input messages can be sent directly to `model.Update`.

Recommended automated checks:

- `Update` tests for:
  - `j/k` and arrow-key navigation
  - selection clamping at top and bottom
  - expand/collapse toggling
  - clear resetting requests and interaction state
  - resize updating dimensions without panics
- render tests for:
  - selected row has visible selected treatment
  - expanded row shows headers, body, client IP, and response time
  - footer contains the documented keybindings
  - empty state still renders correctly with footer present
  - long bodies/headers do not corrupt layout under narrow widths
  - resized views keep separators/footer intact
- integration-level model tests for:
  - appending new requests while a non-tail selection is active
  - expanded row remains valid after new requests append
  - clear followed by new request produces a clean rebuilt view

Recommended manual verification:

- run `go run .`
- send multiple requests, navigate with `j/k` and arrows, and expand/collapse details
- clear the list, confirm the UI empties but `requests.log` still contains prior entries
- resize the terminal aggressively while rows are expanded and while new requests are arriving

## Implementation Recommendation

Plan this phase as one focused execution plan unless the planner wants to split validation/manual verification out:

1. extend `RequestData` and handler bridge payload with detail fields
2. add selection, expansion, clear, and viewport state to `model`
3. refactor rendering into summary block + detail block + footer help
4. replace line-count clipping with block-height-aware viewport logic
5. expand `tui_test.go` to cover input, detail rendering, clear semantics, and resize

The codebase is already shaped correctly for this work. The planner should spend most of its precision on interaction semantics and viewport math, not on dependency or file-structure decisions.

