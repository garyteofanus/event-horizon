---
phase: quick-3
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - go.mod
  - CLAUDE.md
  - AGENTS.md
  - tui.go
  - tui_test.go
  - .planning/PROJECT.md
  - .planning/ROADMAP.md
  - .planning/REQUIREMENTS.md
  - .planning/STATE.md
  - .planning/codebase/STACK.md
  - .planning/codebase/STRUCTURE.md
  - .planning/codebase/TESTING.md
  - .planning/codebase/INTEGRATIONS.md
  - .planning/codebase/CONCERNS.md
  - .planning/research/SUMMARY.md
  - .planning/research/ARCHITECTURE.md
  - .planning/research/PITFALLS.md
  - .planning/research/STACK.md
  - .planning/research/FEATURES.md
autonomous: true
requirements: []

must_haves:
  truths:
    - "go.mod declares module event-horizon"
    - "go build and go test pass with new module name"
    - "TUI header shows 'event-horizon' instead of 'blackhole'"
    - "All documentation references event-horizon instead of blackhole-server"
  artifacts:
    - path: "go.mod"
      contains: "module event-horizon"
    - path: "CLAUDE.md"
      contains: "event-horizon"
    - path: "tui.go"
      contains: "event-horizon"
  key_links:
    - from: "go.mod"
      to: "all .go files"
      via: "module name"
      pattern: "module event-horizon"
---

<objective>
Rename the project from "blackhole-server" to "event-horizon" across the Go module, source code, documentation, and planning files.

Purpose: Rebrand to a clever celestial name -- the event horizon is the boundary of a black hole beyond which nothing escapes, fitting the server's "swallow everything" design.
Output: All internal references updated, project builds and tests pass.
</objective>

<execution_context>
@/Users/teo/.claude/get-shit-done/workflows/execute-plan.md
@/Users/teo/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@CLAUDE.md
@go.mod

Folder rename is OUT OF SCOPE. Only rename internal references.

The project was previously renamed from echo-server to blackhole-server (quick task 2).
This follows the same pattern: module name, binary name, docs, TUI header text.
</context>

<tasks>

<task type="auto">
  <name>Task 1: Rename Go module and source code references</name>
  <files>go.mod, tui.go, tui_test.go, CLAUDE.md, AGENTS.md</files>
  <action>
1. Update `go.mod` line 1: change `module blackhole-server` to `module event-horizon`.

2. Update `tui.go` line 419: change the header string from `"blackhole :%s -> %s"` to `"event-horizon :%s -> %s"`.

3. Update `tui_test.go`:
   - Line 257: change expected string from `"blackhole :9090 -> /tmp/test.log"` to `"event-horizon :9090 -> /tmp/test.log"`.
   - Line 258: update the error message similarly.
   - Line 516: change expected string from `"blackhole :8080 -> requests.log"` to `"event-horizon :8080 -> requests.log"`.

4. Update `CLAUDE.md`:
   - Line 7 (Overview): change "A minimal HTTP blackhole server" to "A minimal HTTP event-horizon server" and change "swallowing all input like a blackhole" to "swallowing all input like a black hole -- nothing escapes past the event horizon".
   - Line 12 (Build): change `go build -o blackhole-server .` to `go build -o event-horizon .`.
   - Line 20: change "expose the blackhole server publicly" to "expose event-horizon publicly".

5. Update `AGENTS.md` with the same changes as CLAUDE.md (it mirrors CLAUDE.md content).

6. Run `go build -o /dev/null .` and `go test ./...` to verify everything compiles and passes.
  </action>
  <verify>
    <automated>cd /Users/teo/Developer/blackhole-server && grep "module event-horizon" go.mod && go build -o /dev/null . && go test ./... 2>&1 | tail -5</automated>
  </verify>
  <done>go.mod declares module event-horizon. TUI header shows "event-horizon". CLAUDE.md and AGENTS.md reference event-horizon. All tests pass.</done>
</task>

<task type="auto">
  <name>Task 2: Update all planning and codebase documentation</name>
  <files>.planning/PROJECT.md, .planning/ROADMAP.md, .planning/REQUIREMENTS.md, .planning/STATE.md, .planning/codebase/STACK.md, .planning/codebase/STRUCTURE.md, .planning/codebase/TESTING.md, .planning/codebase/INTEGRATIONS.md, .planning/codebase/CONCERNS.md, .planning/research/SUMMARY.md, .planning/research/ARCHITECTURE.md, .planning/research/PITFALLS.md, .planning/research/STACK.md, .planning/research/FEATURES.md</files>
  <action>
Perform a systematic find-and-replace across all .planning/ files:

1. Replace `blackhole-server` with `event-horizon` (the module/binary name form).
2. Replace `Blackhole Server` with `Event Horizon` (the title-case display form).
3. Replace `blackhole server` with `event-horizon server` (the lowercase prose form).
4. Replace `blackhole :` with `event-horizon :` (the TUI header format used in plan/research docs).

Special care:
- In `.planning/STATE.md` quick tasks table: Do NOT rename the description of quick task 2 ("Rename folder and Go module to blackhole-server") since that is historical record. Only update current-state references.
- In `.planning/codebase/CONCERNS.md`: "blackhole server design" becomes "event-horizon server design".
- Leave quick task 2's own directory name and plan/summary untouched (historical).
- The `.planning/phases/` plan files from phases 03, 04, 05 are historical records of completed work -- do NOT modify them.

After all replacements, verify no stale "blackhole-server" references remain in actively-used docs (PROJECT.md, ROADMAP.md, REQUIREMENTS.md, codebase/*.md).
  </action>
  <verify>
    <automated>cd /Users/teo/Developer/blackhole-server && grep -rl "blackhole-server" .planning/PROJECT.md .planning/ROADMAP.md .planning/REQUIREMENTS.md .planning/codebase/ 2>/dev/null; echo "EXIT:$?"</automated>
  </verify>
  <done>All active planning docs reference "event-horizon" instead of "blackhole-server". Historical phase plans and quick task 2 records are preserved unchanged.</done>
</task>

</tasks>

<verification>
- `grep "module event-horizon" go.mod` matches
- `go build -o /dev/null .` succeeds
- `go test ./...` all pass
- `grep -r "blackhole-server" CLAUDE.md AGENTS.md go.mod tui.go tui_test.go` returns nothing
- Active .planning/ docs reference event-horizon, not blackhole-server
</verification>

<success_criteria>
- Go module is event-horizon
- Binary build target is event-horizon
- TUI header displays "event-horizon :PORT -> LOGPATH"
- All tests pass
- Documentation consistently uses event-horizon
</success_criteria>

<output>
After completion, create `.planning/quick/3-rename-project-to-a-clever-mythological-/3-SUMMARY.md`
</output>
