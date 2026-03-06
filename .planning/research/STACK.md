# Stack Research

**Domain:** TUI Log Viewer (additions to existing Go HTTP blackhole server)
**Researched:** 2026-03-06
**Confidence:** HIGH

## Existing Stack (DO NOT change)

These are validated from v1.0 and remain unchanged:

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.25.0 | Runtime and build |
| `log/slog` + `JSONHandler` | stdlib | Structured JSON logging |
| `io.MultiWriter` | stdlib | Dual output (stdout + file) |
| `net/http` | stdlib | HTTP server |

## Recommended Stack Additions

### Core Technologies

| Technology | Version | Module Path | Purpose | Why Recommended |
|------------|---------|-------------|---------|-----------------|
| Bubble Tea | v2.0.1 | `charm.land/bubbletea/v2` | TUI framework (Elm architecture) | The standard Go TUI framework. v2 released Feb 27, 2026 with new cell-based renderer, declarative View type, and progressive keyboard enhancements. Powers 25,000+ open source apps. No viable alternative in Go for interactive TUIs. |
| Lip Gloss | v2.0.0 | `charm.land/lipgloss/v2` | Terminal styling (colors, borders, padding) | Companion styling library to Bubble Tea. Provides CSS-like composable styles with automatic color downsampling. v2 released Feb 24, 2026 with deterministic styles and lockstep I/O with Bubble Tea v2. |
| Bubbles | v2.0.0 | `charm.land/bubbles/v2` | Pre-built TUI components (viewport, help, key) | Official component library for Bubble Tea. Provides viewport (scrolling), help (keybind display), and key (keybind definitions) out of the box. Avoids reimplementing scroll logic. |

### Specific Bubbles Sub-packages Needed

| Sub-package | Import Path | Purpose | When to Use |
|-------------|-------------|---------|-------------|
| `viewport` | `charm.land/bubbles/v2/viewport` | Vertical scrollable content area | Core component: the scrollable log list. Handles page up/down, mouse wheel, position tracking. |
| `key` | `charm.land/bubbles/v2/key` | Keybinding definitions | Define custom keybinds (expand/collapse, clear, quit) in a structured way that integrates with help component. |
| `help` | `charm.land/bubbles/v2/help` | Help bar showing available keybinds | Bottom status bar showing available keys. Auto-generates from key.Binding definitions. |

### Sub-packages NOT Needed

| Sub-package | Why Not |
|-------------|---------|
| `list` | Too opinionated for this use case (includes filtering, pagination, spinner). We need a custom log entry list, not a generic browsable list. Building on `viewport` gives more control. |
| `table` | Log entries are not tabular data. The compact one-line view is better served by styled strings in a viewport. |
| `textinput` / `textarea` | No text input needed in the TUI. |
| `spinner` | No loading states needed. |
| `filepicker` | Not applicable. |

## Key v2 API Changes (from v1)

These are critical for implementation -- do NOT follow v1 tutorials.

### View() returns `tea.View`, not `string`

```go
// v1 (WRONG for v2)
func (m Model) View() string { return "..." }

// v2 (CORRECT)
func (m Model) View() tea.View {
    return tea.NewView(m.renderContent())
}
```

### Key messages split into KeyPress and KeyRelease

```go
// v1 (WRONG for v2)
case tea.KeyMsg:
    if msg.String() == "q" { ... }

// v2 (CORRECT)
case tea.KeyPressMsg:
    if msg.Text == "q" { ... }
    // Use msg.Code for special keys, msg.Mod for modifiers
```

### Mouse messages split into specific types

```go
// v2 mouse types
case tea.MouseClickMsg:    // mouse button pressed
case tea.MouseWheelMsg:    // scroll wheel
case tea.MouseMotionMsg:   // mouse moved
case tea.MouseReleaseMsg:  // button released
```

### Terminal features are declarative (View fields, not Commands)

```go
// v1: tea.EnterAltScreen command
// v2: set in View
func (m Model) View() tea.View {
    v := tea.NewView(content)
    v.AltScreen = true    // declarative
    v.MouseMode = tea.MouseAllMotion
    return v
}
```

## Lip Gloss v2 Styling Patterns

### Color-coded HTTP methods

```go
var methodColors = map[string]lipgloss.Style{
    "GET":    lipgloss.NewStyle().Foreground(lipgloss.Color("#61AFEF")).Bold(true),
    "POST":   lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379")).Bold(true),
    "PUT":    lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B")).Bold(true),
    "DELETE": lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75")).Bold(true),
    "PATCH":  lipgloss.NewStyle().Foreground(lipgloss.Color("#C678DD")).Bold(true),
}
```

### Border and layout styles

```go
// Rounded borders for expanded detail view
detailStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("63")).
    Padding(0, 1)

// Status bar at bottom
statusStyle := lipgloss.NewStyle().
    Background(lipgloss.Color("235")).
    Foreground(lipgloss.Color("252")).
    Padding(0, 1).
    Width(termWidth)
```

### Automatic color downsampling

Lip Gloss v2 automatically detects terminal color profile and downsamples. Use `lipgloss.Print`/`lipgloss.Println` or render through Bubble Tea's View -- colors degrade gracefully from TrueColor to ANSI 256 to ANSI 16 without manual handling.

## Integration Architecture

### Stdout ownership change

The TUI takes over stdout. The current `io.MultiWriter(os.Stdout, logFile)` must change:

```go
// v1.0: slog writes to stdout + file
writer := io.MultiWriter(os.Stdout, logFile)

// v1.1: slog writes to file ONLY; TUI owns stdout
logger := slog.New(slog.NewJSONHandler(logFile, nil))
```

### Request flow into TUI

The HTTP handler sends log entries to the TUI via `Program.Send()`:

```go
// In the HTTP handler:
program.Send(RequestMsg{
    Method: r.Method,
    Path:   r.RequestURI,
    // ... other fields
})

// In the TUI Update():
case RequestMsg:
    m.entries = append(m.entries, msg)
    // re-render viewport content
```

`Program.Send()` is thread-safe -- safe to call from HTTP handler goroutines.

### Startup order

1. Open log file, create slog logger (file-only)
2. Create Bubble Tea program with TUI model
3. Start HTTP server in background goroutine
4. Run `program.Run()` (blocks on main goroutine, owns terminal)

## Installation

```bash
# Add Charm ecosystem dependencies
go get charm.land/bubbletea/v2@latest
go get charm.land/lipgloss/v2@latest
go get charm.land/bubbles/v2@latest

# Tidy module
go mod tidy
```

This transitions the project from zero external dependencies to three (plus transitive deps). This is the first external dependency addition.

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| `charm.land/bubbletea/v2` | `github.com/rivo/tview` | Never for this project. tview uses immediate-mode rendering; bubbletea's Elm architecture is better for reactive UIs that update from external events (HTTP requests). tview also has no v2 with modern terminal features. |
| `charm.land/bubbletea/v2` | `github.com/gdamore/tcell` | Never for this project. tcell is low-level (cell-by-cell rendering). Bubbletea sits on top of it and provides the application architecture. Using tcell directly means reimplementing everything bubbletea gives you. |
| `charm.land/bubbles/v2/viewport` | Custom scroll implementation | Never. Viewport handles terminal resize, mouse wheel, page navigation, boundary checking. Reimplementing this is error-prone and pointless. |
| `charm.land/lipgloss/v2` | Raw ANSI escape codes | Never. Lipgloss handles color profile detection, downsampling, and provides a clean API. Raw ANSI codes break on terminals with limited color support. |
| `charm.land/bubbles/v2/viewport` | `charm.land/bubbles/v2/list` | Not for this project. The list component includes filtering, pagination controls, and a status bar that conflict with our custom UI. Viewport is the right primitive -- it just scrolls content, and we render the content ourselves. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `github.com/charmbracelet/bubbletea` (v1) | v1 module path; deprecated API. v2 has different View return type, message types, and rendering pipeline. | `charm.land/bubbletea/v2` |
| `github.com/charmbracelet/lipgloss` (v1) | v1 fights with bubbletea v2 over I/O. v2 versions are designed to work together. | `charm.land/lipgloss/v2` |
| `github.com/charmbracelet/bubbles` (v1) | Must match bubbletea v2. v1 bubbles are incompatible with v2 tea.Model interface. | `charm.land/bubbles/v2` |
| `charm.land/glamour/v2` | Markdown rendering. Not needed -- log entries are structured data, not markdown. |  Direct lipgloss styling |
| `charm.land/wish/v2` | SSH server for TUI apps. Not needed -- this is a local terminal app. | Nothing |
| `charm.land/bubbles/v2/list` | Over-featured for log display. Includes fuzzy filtering, pagination, spinner -- all things explicitly out of scope. | `charm.land/bubbles/v2/viewport` with custom rendering |

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| `charm.land/bubbletea/v2@v2.0.1` | `charm.land/lipgloss/v2@v2.0.0` | Designed together; shared I/O model in v2. Must use v2 of both. |
| `charm.land/bubbletea/v2@v2.0.1` | `charm.land/bubbles/v2@v2.0.0` | Bubbles v2 implements bubbletea v2's Model interface. |
| `charm.land/bubbletea/v2@v2.0.1` | Go 1.25.0 | Verified: bubbletea v2 requires Go 1.22+ minimum. Go 1.25 is well above. |
| `charm.land/lipgloss/v2@v2.0.0` | Go 1.25.0 | Same Go minimum as bubbletea. |

**CRITICAL: All three Charm packages must be v2.** Mixing v1 and v2 will cause compile errors (incompatible interfaces) and runtime I/O conflicts.

## Sources

- [Bubble Tea v2.0.0 Release](https://github.com/charmbracelet/bubbletea/releases/tag/v2.0.0) -- Release notes, Feb 27, 2026
- [charm.land/bubbletea/v2 - Go Packages](https://pkg.go.dev/charm.land/bubbletea/v2) -- Official API docs, v2.0.1
- [charm.land/lipgloss/v2 - Go Packages](https://pkg.go.dev/charm.land/lipgloss/v2) -- Official API docs, v2.0.0
- [charm.land/bubbles/v2 - Go Packages](https://pkg.go.dev/charm.land/bubbles/v2) -- Official API docs, v2.0.0
- [Bubble Tea v2: What's New](https://github.com/charmbracelet/bubbletea/discussions/1374) -- v2 migration overview
- [Lip Gloss v2: What's New](https://github.com/charmbracelet/lipgloss/discussions/506) -- v2 changes
- [Charm v2 Blog Post](https://charm.land/blog/v2/) -- Official announcement

---
*Stack research for: TUI Log Viewer (Charm ecosystem additions)*
*Researched: 2026-03-06*
