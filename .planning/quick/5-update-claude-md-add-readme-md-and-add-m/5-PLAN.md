---
phase: quick-5
plan: 01
type: execute
wave: 1
depends_on: []
files_modified: [CLAUDE.md, README.md, LICENSE]
autonomous: true
requirements: [DOCS-01, DOCS-02, DOCS-03]
must_haves:
  truths:
    - "CLAUDE.md accurately describes the current architecture (multi-file, TUI, bubbletea, lipgloss)"
    - "README.md presents the project clearly for GitHub visitors"
    - "MIT license file exists at repo root"
  artifacts:
    - path: "CLAUDE.md"
      provides: "Updated project guidance for Claude Code"
    - path: "README.md"
      provides: "Public-facing GitHub README"
    - path: "LICENSE"
      provides: "MIT license"
---

<objective>
Update CLAUDE.md to reflect the current state of the project, create a proper public-facing README.md for the GitHub repo, and add an MIT LICENSE file.

Purpose: The project has evolved significantly (TUI log viewer, structured logging, clipboard, JSON formatting) but CLAUDE.md still describes a single-file server. The repo also lacks a README and license.
Output: Three files updated/created at repo root.
</objective>

<execution_context>
@/Users/teo/.claude/get-shit-done/workflows/execute-plan.md
@/Users/teo/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@CLAUDE.md
@main.go
@handler.go
@tui.go
@go.mod
</context>

<tasks>

<task type="auto">
  <name>Task 1: Update CLAUDE.md and add LICENSE</name>
  <files>CLAUDE.md, LICENSE</files>
  <action>
Rewrite CLAUDE.md to accurately describe the current project:

**Overview section:** Event-horizon is an HTTP server + TUI log viewer in Go. It captures every incoming HTTP request (method, URI, headers, body), logs structured JSON to a file, and displays requests in a real-time terminal UI built with Bubble Tea. Responds with empty 200 OK to all requests (swallowing input like a black hole).

**Architecture section:** Multi-file structure:
- main.go: Orchestration -- starts HTTP server in background goroutine, runs Bubble Tea TUI on main goroutine, bridges them via buffered channel
- handler.go: HTTP handler + RequestData type -- logs structured JSON via slog, sends to TUI channel (non-blocking)
- tui.go: Bubble Tea model -- live request list with lipgloss styling, expanded detail view, clipboard support (OSC52), JSON pretty-printing with syntax highlighting

**Dependencies:** Bubble Tea v2 (charm.land/bubbletea/v2), Lipgloss v2 (charm.land/lipgloss/v2). No other direct dependencies.

**Commands section:** Keep existing run/build commands. Add:
- Test: `go test ./...`
- Test verbose: `go test -v ./...`

**Key features list:** Structured JSON logging to file, real-time TUI with color-coded methods/status, keyboard navigation (j/k, enter/space expand), copy request body (c) or full request (C) via OSC52 clipboard, JSON body formatting with syntax highlighting (f to toggle), clear requests (x)

**Environment variables:** PORT (default 8080), LOG_FILE (default requests.log)

**Keep** the Public Access (cloudflared) section as-is.

For LICENSE: Create standard MIT license file. Copyright (c) 2026 Gary Teofanus. Use the full MIT license text.
  </action>
  <verify>cat CLAUDE.md | grep -q "bubbletea" && cat CLAUDE.md | grep -q "handler.go" && cat CLAUDE.md | grep -q "tui.go" && test -f LICENSE && echo "PASS"</verify>
  <done>CLAUDE.md accurately describes multi-file architecture, TUI features, dependencies, and commands. LICENSE exists with MIT text.</done>
</task>

<task type="auto">
  <name>Task 2: Create README.md</name>
  <files>README.md</files>
  <action>
Create a clean, public-facing README.md for https://github.com/garyteofanus/event-horizon.

Structure:

1. **Title and tagline:** `# event-horizon` followed by a one-liner: "A minimal HTTP server that swallows every request and shows them in a real-time terminal UI."

2. **What it does (2-3 sentences):** event-horizon captures every incoming HTTP request -- method, URI, headers, body -- logs it as structured JSON, and displays it live in a terminal interface. Every request gets an empty 200 OK response. Nothing escapes past the event horizon.

3. **Features (bullet list):**
   - Real-time TUI log viewer with color-coded HTTP methods and status codes
   - Structured JSON logging to file
   - Expanded request detail view (headers, body, client IP, response time)
   - JSON body pretty-printing with syntax highlighting
   - Clipboard support (copy body or full request) via OSC52
   - Keyboard-driven: j/k navigate, enter/space expand, c/C copy, f format toggle, x clear, q quit
   - Zero-config: just run and send requests

4. **Quick start:**
   ```bash
   go install github.com/garyteofanus/event-horizon@latest
   event-horizon
   ```
   Or clone and build:
   ```bash
   git clone https://github.com/garyteofanus/event-horizon.git
   cd event-horizon
   go build -o event-horizon .
   ./event-horizon
   ```
   Then in another terminal: `curl http://localhost:8080/hello`

5. **Configuration:** Table with PORT and LOG_FILE env vars.

6. **Keybindings:** Clean table of all keybindings (j/k or arrows, enter/space, c, C, f, x, q).

7. **Public access section:** Brief mention of cloudflared quick tunnel with the 2-line command.

8. **License:** MIT

Keep the tone concise and technical. No badges, no screenshots placeholder, no contributing section. Just clean information.
  </action>
  <verify>test -f README.md && head -1 README.md | grep -q "event-horizon" && echo "PASS"</verify>
  <done>README.md exists with project description, features, quick start, configuration, keybindings, and license reference.</done>
</task>

</tasks>

<verification>
- CLAUDE.md mentions bubbletea, lipgloss, handler.go, tui.go, main.go
- README.md has quick start, features, keybindings sections
- LICENSE contains MIT text and correct copyright
- All three files exist at repo root
</verification>

<success_criteria>
- CLAUDE.md accurately reflects the current multi-file architecture with TUI
- README.md is a clean, informative public-facing README
- LICENSE file contains MIT license with correct attribution
</success_criteria>

<output>
After completion, create `.planning/quick/5-update-claude-md-add-readme-md-and-add-m/5-SUMMARY.md`
</output>
