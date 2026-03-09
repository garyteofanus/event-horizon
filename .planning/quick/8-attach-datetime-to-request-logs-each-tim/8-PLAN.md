---
phase: quick-8
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - .gitignore
  - bin/.gitkeep
  - logs/.gitkeep
  - main.go
  - main_test.go
  - README.md
  - CLAUDE.md
  - AGENTS.md
  - output/bin/.gitkeep
  - output/logs/.gitkeep
autonomous: true
requirements: [LOGPATH-01]
---

<objective>
Move build and log output directories to root-level `bin/` and `logs/`, and make the default log file name include the program start timestamp so previous runs remain preserved.

Purpose: Keep generated artifacts in predictable root folders and preserve a history of past log sessions instead of reusing one default log file.
Output: root-level `bin/` and `logs/` directories exist, generated contents are gitignored, the default log path is timestamped per start under `logs/`, and docs/tests reflect the new behavior.
</objective>

<execution_context>
@/Users/teo/.codex/get-shit-done/workflows/execute-plan.md
@/Users/teo/.codex/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@CLAUDE.md
@AGENTS.md
@main.go
@main_test.go
@README.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Move output layout to root and timestamp default log files</name>
  <files>.gitignore, bin/.gitkeep, logs/.gitkeep, output/bin/.gitkeep, output/logs/.gitkeep, main.go, main_test.go, README.md, CLAUDE.md, AGENTS.md</files>
  <action>
1. Replace the `output/bin` and `output/logs` convention with root-level `bin/` and `logs/` directories, keeping tracked `.gitkeep` placeholders and removing the old tracked output placeholders.

2. Update `.gitignore` so generated contents under `bin/` and `logs/` are ignored while the directory placeholders stay tracked.

3. Change the default log path behavior so each process start writes to a timestamped file under `logs/` (for example `logs/requests-20260309-170000.log`) while `LOG_FILE` still overrides the default path when set.

4. Keep parent-directory creation when opening the log file and add/update tests to cover the timestamped default path and custom override behavior.

5. Update `README.md`, `CLAUDE.md`, and `AGENTS.md` so build instructions point to `bin/event-horizon` and log documentation reflects the root `logs/` folder plus timestamped default filenames.
  </action>
  <verify>
    <automated>go test ./... && go build -o bin/event-horizon . && test -d bin && test -d logs && test -f bin/.gitkeep && test -f logs/.gitkeep && rm bin/event-horizon</automated>
  </verify>
  <done>Build artifacts use root `bin/`, default logs use root `logs/` with per-start timestamps, and generated contents are ignored by git.</done>
</task>

</tasks>

<verification>
- `go test ./...` passes
- `go build -o bin/event-horizon .` succeeds and the binary can be removed cleanly
- `.gitignore` ignores generated contents under `bin/` and `logs/`
- Default log filenames include the startup datetime while `LOG_FILE` still overrides the default path
</verification>

<success_criteria>
- Repo uses root-level `bin/` and `logs/` directories
- Generated binary and log files are ignored by git
- Default log path is timestamped per program start under `logs/`
- Docs and tests reflect the new root-folder layout and logging behavior
</success_criteria>

<output>
After completion, create `.planning/quick/8-attach-datetime-to-request-logs-each-tim/8-SUMMARY.md`
</output>
