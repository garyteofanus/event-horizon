# Phase 6: Copy request body and format body in expanded view - Context

**Gathered:** 2026-03-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Add clipboard copy for request body/full request, and JSON pretty-printing with syntax highlighting in the expanded detail view. Enhances the existing expanded view from Phase 5 — no new navigation or layout changes.

</domain>

<decisions>
## Implementation Decisions

### Copy mechanism
- Two copy modes: `c` copies body only, `C` (Shift+C) copies full request (headers, body, client IP, response time)
- Copy works on the selected request even when collapsed — no need to expand first
- Flash message in footer after copy: "Copied!" for success, "No body to copy" when body is empty (GET requests etc.)
- Flash message disappears after ~1-2 seconds
- Remove `c` from clear binding — only `x` clears now

### Body formatting
- Auto-detect JSON bodies only (no XML, form data, etc.)
- Toggle between raw and formatted display with `f` key
- Default to formatted when JSON is detected
- Global toggle — one state affects all requests, not per-request
- Syntax highlighting with colors for formatted JSON: distinct colors for keys, strings, numbers/bools using existing lipgloss palette

### Format display
- Pretty-print with 2-space indentation
- Show full formatted JSON, no truncation — existing scroll/viewport handles overflow
- Body section label indicates content type: "Body (JSON)" when formatted JSON, "Body (raw)" for non-JSON or when format toggle is off
- Footer shows current format toggle state (e.g., "format: on" or "format: off")

### Keybinding updates
- `c` = copy body to clipboard
- `C` (Shift+C) = copy full request to clipboard
- `f` = toggle body format (raw/formatted)
- `x` = clear all requests (remove `c` from clear)
- Help footer updated: "j/k move | enter expand | c copy | C copy all | f format | x clear | q quit"

### Claude's Discretion
- Clipboard implementation approach (OS-level clipboard access in Go)
- Exact JSON syntax highlighting color choices (within existing palette)
- Flash message timing and animation
- How "format: on/off" integrates into the existing footer layout

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `renderExpandedDetails` in tui.go:247 — already renders body section, needs modification to support formatting
- `renderDetailSection` in tui.go:258 — generic label+value renderer with text wrapping, body label needs to become dynamic
- `blankFallback` in tui.go:293 — handles empty body display, can be extended for "No body to copy" logic
- `renderFooter` in tui.go:367 — needs flash message support and format toggle state display
- Existing lipgloss color palette (colorGreen, colorBlue, colorYellow, colorCyan, colorMuted, etc.) — reuse for JSON syntax highlighting

### Established Patterns
- lipgloss v2 styles defined as package-level vars — add new styles for JSON syntax highlighting
- `model` struct holds TUI state — add `formatBody bool` and flash message state
- `Update` handles key events via switch on `msg.String()` — add cases for "c", "C", "f"
- `renderView` builds display string via strings.Builder — consistent approach

### Integration Points
- `model.Update` in tui.go:41 — add key handlers for c, C, f; modify "c"/"x" clear binding to "x" only
- `renderExpandedDetails` — modify to accept format toggle state and render formatted/highlighted JSON
- `renderFooter` — add flash message display and format toggle indicator
- `RequestData.Body` in handler.go:22 — already stores body as string, no changes needed

</code_context>

<specifics>
## Specific Ideas

- Copy feedback should feel immediate — flash "Copied!" in the footer like vim's yank feedback
- The format toggle should feel like a view mode switch, not a transformation — the data doesn't change, just how it's displayed
- Help footer format: "j/k move | enter expand | c copy | C copy all | f format | x clear | q quit"

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 06-copy-request-body-and-format-body-in-expanded-view*
*Context gathered: 2026-03-09*
