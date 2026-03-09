# Phase 6: Copy Request Body and Format Body in Expanded View - Research

**Researched:** 2026-03-09
**Domain:** TUI clipboard operations, JSON formatting, syntax highlighting
**Confidence:** HIGH

## Summary

Phase 6 adds two features to the existing TUI: (1) clipboard copy for request body/full request, and (2) JSON pretty-printing with syntax highlighting in the expanded detail view. Both features build on the existing `tui.go` model, styles, and rendering patterns.

The critical discovery is that **Bubbletea v2 has native clipboard support** via `tea.SetClipboard(s string) Cmd`, which uses OSC52 -- no external clipboard dependency needed. This keeps the project dependency-light. For JSON formatting, Go's stdlib `encoding/json` provides `json.Indent` for pretty-printing existing JSON strings, and lipgloss styles (already in use) handle syntax highlighting via manual token rendering.

**Primary recommendation:** Use `tea.SetClipboard()` for clipboard, `json.Indent` for pretty-printing, and hand-rolled lipgloss-based token coloring for JSON syntax highlighting. Flash messages use `tea.Tick` for auto-dismiss timing.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Two copy modes: `c` copies body only, `C` (Shift+C) copies full request (headers, body, client IP, response time)
- Copy works on the selected request even when collapsed -- no need to expand first
- Flash message in footer after copy: "Copied!" for success, "No body to copy" when body is empty (GET requests etc.)
- Flash message disappears after ~1-2 seconds
- Remove `c` from clear binding -- only `x` clears now
- Auto-detect JSON bodies only (no XML, form data, etc.)
- Toggle between raw and formatted display with `f` key
- Default to formatted when JSON is detected
- Global toggle -- one state affects all requests, not per-request
- Syntax highlighting with colors for formatted JSON: distinct colors for keys, strings, numbers/bools using existing lipgloss palette
- Pretty-print with 2-space indentation
- Show full formatted JSON, no truncation -- existing scroll/viewport handles overflow
- Body section label indicates content type: "Body (JSON)" when formatted JSON, "Body (raw)" for non-JSON or when format toggle is off
- Footer shows current format toggle state (e.g., "format: on" or "format: off")
- Keybindings: `c` = copy body, `C` = copy full request, `f` = toggle format, `x` = clear (remove `c` from clear)
- Help footer: "j/k or arrows move | enter/space expand | c copy | C copy all | f format | x clear | q quit"

### Claude's Discretion
- Clipboard implementation approach (OS-level clipboard access in Go)
- Exact JSON syntax highlighting color choices (within existing palette)
- Flash message timing and animation
- How "format: on/off" integrates into the existing footer layout

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| charm.land/bubbletea/v2 | v2.0.1 | TUI framework + clipboard | Already in use; `tea.SetClipboard()` provides native OSC52 clipboard |
| charm.land/lipgloss/v2 | v2.0.0 | Terminal styling | Already in use; reuse for JSON syntax highlighting styles |
| encoding/json (stdlib) | go1.25 | JSON validation + pretty-print | `json.Valid()` for detection, `json.Indent` for formatting |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| time (stdlib) | go1.25 | Flash message timing | `tea.Tick` uses `time.Duration` for auto-dismiss |
| strings (stdlib) | go1.25 | String building | Already used for view rendering |
| fmt (stdlib) | go1.25 | Formatting copy output | Building full-request copy string |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| tea.SetClipboard (OSC52) | atotto/clipboard (exec pbcopy) | OSC52 works over SSH, no external dep; but not all terminals support it. Acceptable tradeoff for this dev tool. |
| Hand-rolled JSON coloring | chroma/pygments library | Adds heavy dependency for one feature; lipgloss styles are sufficient for JSON tokens |
| json.Indent (reformat string) | json.MarshalIndent (unmarshal+remarshal) | json.Indent works directly on []byte without round-trip; simpler and preserves original types |

## Architecture Patterns

### Model State Additions
```go
// Add to model struct in tui.go
type model struct {
    // ... existing fields ...
    formatBody   bool          // global toggle: true = formatted JSON, false = raw
    flashMessage string        // current flash text ("Copied!", "No body to copy", etc.)
    flashTimer   int           // countdown ticks remaining (0 = no flash)
}
```

### Pattern 1: Clipboard via tea.SetClipboard (OSC52)
**What:** Bubbletea v2 native clipboard command using OSC52 escape sequences
**When to use:** Any clipboard copy operation in the TUI
**Example:**
```go
// Source: https://pkg.go.dev/charm.land/bubbletea/v2 (SetClipboard)
case "c":
    if len(m.requests) == 0 {
        return m, nil
    }
    r := m.requests[m.selectedIndex]
    if strings.TrimSpace(r.Body) == "" {
        m.flashMessage = "No body to copy"
        return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
            return flashExpiredMsg{}
        })
    }
    m.flashMessage = "Copied!"
    return m, tea.Batch(
        tea.SetClipboard(r.Body),
        tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
            return flashExpiredMsg{}
        }),
    )
```

### Pattern 2: Flash Message with tea.Tick
**What:** Temporary status message that auto-clears after a duration
**When to use:** Copy feedback, error feedback
**Example:**
```go
// Source: https://pkg.go.dev/charm.land/bubbletea/v2 (Tick)
type flashExpiredMsg struct{}

// In Update:
case flashExpiredMsg:
    m.flashMessage = ""
    return m, nil
```

### Pattern 3: JSON Detection and Formatting
**What:** Validate JSON with `json.Valid()`, format with `json.Indent()`
**When to use:** Body display in expanded view when format toggle is on
**Example:**
```go
// Source: https://pkg.go.dev/encoding/json (Valid, Indent)
func isJSON(s string) bool {
    return json.Valid([]byte(s))
}

func prettyJSON(s string) string {
    var buf bytes.Buffer
    if err := json.Indent(&buf, []byte(s), "", "  "); err != nil {
        return s
    }
    return buf.String()
}
```

### Pattern 4: JSON Syntax Highlighting with Lipgloss
**What:** Color JSON tokens (keys, strings, numbers, bools, null) using existing palette
**When to use:** Rendering formatted JSON in expanded body section
**Recommended color mapping:**
```go
// Reuse existing palette colors
var (
    jsonKeyStyle    = lipgloss.NewStyle().Foreground(colorCyan)    // keys
    jsonStringStyle = lipgloss.NewStyle().Foreground(colorGreen)   // string values
    jsonNumberStyle = lipgloss.NewStyle().Foreground(colorYellow)  // numbers
    jsonBoolStyle   = lipgloss.NewStyle().Foreground(colorBlue)    // true/false
    jsonNullStyle   = lipgloss.NewStyle().Foreground(colorMuted)   // null
    jsonBraceStyle  = lipgloss.NewStyle().Foreground(colorNeutral) // {} [] , :
)
```
**Approach:** Walk the pretty-printed JSON string line by line, apply styles to tokens. A simple state machine or regex-based tokenizer is sufficient for well-formed JSON (which it is, since it came from `json.Indent`).

### Pattern 5: Full Request Copy Format
**What:** Format the complete request for clipboard
**When to use:** `C` (Shift+C) key handler
**Example:**
```go
func formatFullRequest(r RequestData) string {
    var b strings.Builder
    b.WriteString(fmt.Sprintf("%s %s\n", r.Method, r.URI))
    b.WriteString(fmt.Sprintf("Client IP: %s\n", r.ClientIP))
    b.WriteString(fmt.Sprintf("Response Time: %s\n", formatResponseTime(r.ResponseTime)))
    b.WriteString("\nHeaders:\n")
    b.WriteString(formatHeaders(r.Headers))
    if strings.TrimSpace(r.Body) != "" {
        b.WriteString("\n\nBody:\n")
        b.WriteString(r.Body)
    }
    return b.String()
}
```

### Anti-Patterns to Avoid
- **Unmarshaling JSON just to re-marshal:** Use `json.Indent` on the raw string bytes directly -- no need for `json.Unmarshal` + `json.MarshalIndent` roundtrip
- **Per-request format state:** The toggle is global, not per-request. Do not add a `formatted` field to `RequestData`
- **Blocking clipboard operations:** `tea.SetClipboard` returns a `Cmd` (async) -- never call clipboard in a blocking way from `Update`
- **Complex JSON parser for highlighting:** The JSON is already well-formed from `json.Indent`. A line-by-line regex tokenizer is sufficient; do not build a full recursive descent parser

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Clipboard access | exec.Command("pbcopy") wrapper | tea.SetClipboard() | Native to bubbletea v2, works over SSH via OSC52, zero extra dependencies |
| JSON validation | Manual brace/bracket counting | json.Valid() | Handles all edge cases (escaped chars, unicode, nested structures) |
| JSON pretty-printing | Manual indentation logic | json.Indent() | Handles all JSON types correctly, 2-space indent built in |
| Timed messages | goroutine + time.After | tea.Tick() | Integrates properly with bubbletea's Cmd/Msg architecture |

**Key insight:** Bubbletea v2 and Go stdlib already provide every primitive needed. The only custom code is the JSON syntax highlighter (coloring tokens), which is intentionally simple since the input is guaranteed well-formed.

## Common Pitfalls

### Pitfall 1: OSC52 Terminal Compatibility
**What goes wrong:** `tea.SetClipboard()` uses OSC52 which some terminals don't support (older terminals, some tmux configs)
**Why it happens:** OSC52 is a terminal escape sequence, not an OS-level clipboard API
**How to avoid:** This is acceptable for a developer tool. The copy still "works" silently -- the flash message shows "Copied!" regardless. No fallback needed.
**Warning signs:** User reports copy doesn't work in specific terminals

### Pitfall 2: Flash Message Race Conditions
**What goes wrong:** Multiple rapid copy presses queue multiple `flashExpiredMsg` ticks, causing premature clear
**Why it happens:** Each copy starts a new tick timer, but old timers still fire
**How to avoid:** Use a generation counter or simply always set flash text on copy and clear on any `flashExpiredMsg`. The worst case is the flash clears slightly early on rapid presses, which is fine UX.
**Warning signs:** Flash message disappears too quickly on second copy

### Pitfall 3: Clear Keybinding Change Breaks Tests
**What goes wrong:** Existing test `TestModelUpdateClearRequests` uses `'c'` key to clear
**Why it happens:** Phase 6 remaps `c` from clear to copy body
**How to avoid:** Update the test to use `'x'` instead of `'c'`. The `"x", "c"` case in Update becomes just `"x"`.
**Warning signs:** Test failure on `TestModelUpdateClearRequests`

### Pitfall 4: JSON Highlighting Breaks wrapText
**What goes wrong:** ANSI escape codes in highlighted JSON count toward `wrapText` width calculation
**Why it happens:** `wrapText` counts raw string length including escape codes
**How to avoid:** Apply syntax highlighting AFTER wrapping, or use lipgloss's width-aware rendering. Since the expanded detail section already uses `renderDetailSection` which calls `wrapText`, the highlighting should be applied at render time on wrapped lines, not before wrapping.
**Warning signs:** Lines wrap too early, content looks truncated

### Pitfall 5: Shift+C Key Detection
**What goes wrong:** `msg.String()` for Shift+C may return "C" or "shift+c" depending on bubbletea version
**Why it happens:** Bubbletea v2 key representation differs from v1
**How to avoid:** Test with actual `tea.KeyPressMsg{Code: 'C', Shift: true}` and check what `msg.String()` returns. In bubbletea v2, uppercase letters come through as their character with shift modifier.
**Warning signs:** Shift+C doesn't trigger copy-all, or triggers wrong handler

## Code Examples

### JSON Detection and Pretty-Print
```go
// Source: https://pkg.go.dev/encoding/json
import (
    "bytes"
    "encoding/json"
)

func isJSON(s string) bool {
    return json.Valid([]byte(s))
}

func prettyJSON(s string) string {
    var buf bytes.Buffer
    if err := json.Indent(&buf, []byte(s), "", "  "); err != nil {
        return s // fallback to raw on error
    }
    return buf.String()
}
```

### Body Label Logic
```go
func bodyLabel(body string, formatOn bool) string {
    if formatOn && isJSON(body) {
        return "Body (JSON)"
    }
    return "Body (raw)"
}
```

### Footer with Flash and Format State
```go
func renderFooter(m model) string {
    selected := 0
    if len(m.requests) > 0 {
        selected = clampIndex(m.selectedIndex, len(m.requests)) + 1
    }

    // Status line with format toggle
    formatState := "off"
    if m.formatBody {
        formatState = "on"
    }
    status := fmt.Sprintf("%d requests | selected %d/%d | format: %s",
        len(m.requests), selected, len(m.requests), formatState)

    // Flash message replaces or appends to status
    if m.flashMessage != "" {
        status += " | " + m.flashMessage
    }

    help := "j/k or arrows move | enter/space expand | c copy | C copy all | f format | x clear | q quit"
    return footerStyle.Render(status) + "\n" + footerStyle.Render(help)
}
```

### Simple JSON Syntax Highlighter
```go
// Colorize a pretty-printed JSON string line by line.
// Input MUST be output of json.Indent (well-formed, indented).
func highlightJSON(pretty string) string {
    var result strings.Builder
    for i, line := range strings.Split(pretty, "\n") {
        if i > 0 {
            result.WriteString("\n")
        }
        result.WriteString(highlightJSONLine(line))
    }
    return result.String()
}

// highlightJSONLine applies lipgloss styles to a single JSON line.
// Handles: "key": value patterns, standalone values, braces/brackets.
func highlightJSONLine(line string) string {
    // Implementation: regex or simple state machine to identify:
    // - quoted keys (before colon) -> jsonKeyStyle
    // - quoted string values -> jsonStringStyle
    // - numeric values -> jsonNumberStyle
    // - true/false -> jsonBoolStyle
    // - null -> jsonNullStyle
    // - structural chars {}[],: -> jsonBraceStyle
    // - indentation whitespace -> pass through
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| exec.Command("pbcopy") | tea.SetClipboard() via OSC52 | Bubbletea v2 (2025) | No external dependency, works over SSH |
| Custom tick goroutines | tea.Tick() | Bubbletea v1+ | Proper integration with Elm architecture |
| json.MarshalIndent (roundtrip) | json.Indent (direct) | Always available | Simpler, no type information lost |

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | none (stdlib) |
| Quick run command | `go test ./... -count=1 -run TestPhase6` |
| Full suite command | `go test ./... -count=1` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| COPY-01 | `c` copies body of selected request to clipboard | unit | `go test -run TestCopyBody -count=1` | Wave 0 |
| COPY-02 | `C` copies full request to clipboard | unit | `go test -run TestCopyFull -count=1` | Wave 0 |
| COPY-03 | Flash "Copied!" after successful copy | unit | `go test -run TestFlashCopied -count=1` | Wave 0 |
| COPY-04 | Flash "No body to copy" for empty body | unit | `go test -run TestFlashNoBody -count=1` | Wave 0 |
| COPY-05 | Flash auto-dismisses after tick | unit | `go test -run TestFlashExpires -count=1` | Wave 0 |
| FMT-01 | JSON bodies detected and formatted by default | unit | `go test -run TestJSONFormat -count=1` | Wave 0 |
| FMT-02 | `f` toggles format on/off globally | unit | `go test -run TestFormatToggle -count=1` | Wave 0 |
| FMT-03 | Body label shows "(JSON)" or "(raw)" | unit | `go test -run TestBodyLabel -count=1` | Wave 0 |
| FMT-04 | JSON syntax highlighting applies distinct colors | unit | `go test -run TestJSONHighlight -count=1` | Wave 0 |
| KEY-01 | `x` clears, `c` no longer clears | unit | `go test -run TestClearOnlyX -count=1` | Wave 0 |
| KEY-02 | Footer shows updated keybinding help | unit | `go test -run TestFooterHelp -count=1` | Exists (needs update) |

### Sampling Rate
- **Per task commit:** `go test ./... -count=1`
- **Per wave merge:** `go test ./... -count=1 -race`
- **Phase gate:** Full suite green before verify

### Wave 0 Gaps
- [ ] Update `TestModelUpdateClearRequests` to use `'x'` instead of `'c'`
- [ ] Update `TestRenderViewShowsHelpFooter` for new keybinding text
- [ ] New test functions for copy, flash, format, and highlighting behaviors

## Open Questions

1. **Shift+C key representation in bubbletea v2**
   - What we know: Bubbletea v2 uses `tea.KeyPressMsg` with a `Code` field and modifier flags
   - What's unclear: Exact `msg.String()` output for Shift+C -- could be `"C"`, `"shift+c"`, or `"S-c"`
   - Recommendation: Test empirically during implementation. Check if `msg.String() == "C"` works, or if shift modifier needs explicit check. The `Code` rune for uppercase 'C' should be 67 with shift flag.

2. **ANSI width in wrapText**
   - What we know: `wrapText` counts raw string length
   - What's unclear: Whether highlighted JSON lines will wrap correctly
   - Recommendation: Apply highlighting after `wrapText`, or modify the detail section to skip wrapping for highlighted content (let the terminal handle it). Best approach: render raw text through wrapText, then apply highlighting to each wrapped line.

## Sources

### Primary (HIGH confidence)
- [charm.land/bubbletea/v2](https://pkg.go.dev/charm.land/bubbletea/v2) - SetClipboard, Tick, KeyPressMsg API
- [encoding/json](https://pkg.go.dev/encoding/json) - json.Valid, json.Indent
- Existing codebase: tui.go, handler.go, tui_test.go (direct code inspection)

### Secondary (MEDIUM confidence)
- [atotto/clipboard](https://github.com/atotto/clipboard) - Reviewed as alternative, confirmed tea.SetClipboard is better fit
- [Bubbletea v2 discussion](https://github.com/charmbracelet/bubbletea/discussions/1374) - v2 clipboard feature confirmation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - bubbletea v2 SetClipboard verified in official docs, json stdlib well-known
- Architecture: HIGH - patterns follow established model/update/view in existing tui.go
- Pitfalls: MEDIUM - OSC52 compatibility untested, Shift+C key representation needs empirical check

**Research date:** 2026-03-09
**Valid until:** 2026-04-09 (stable stack, no fast-moving dependencies)
