---
phase: quick-7
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - .gitignore
  - output/bin/.gitkeep
  - output/logs/.gitkeep
  - main.go
  - main_test.go
  - README.md
  - CLAUDE.md
autonomous: true
requirements: [OUTDIR-01]
---

<objective>
Create dedicated output directories for built binaries and log files, wire the default log path to the log output directory, and ignore generated artifacts in git.

Purpose: Keep build artifacts and runtime logs out of the repo root while preserving a usable checked-in folder structure.
Output: `output/bin/` and `output/logs/` exist, generated contents are gitignored, builds target `output/bin/`, and the default log file path is under `output/logs/`.
</objective>

<execution_context>
@/Users/teo/.codex/get-shit-done/workflows/execute-plan.md
@/Users/teo/.codex/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@CLAUDE.md
@main.go
@main_test.go
@README.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add output directory layout and wire defaults</name>
  <files>.gitignore, output/bin/.gitkeep, output/logs/.gitkeep, main.go, main_test.go, README.md, CLAUDE.md</files>
  <action>
1. Create tracked `output/bin/` and `output/logs/` directories with `.gitkeep` placeholders.

2. Add a new `.gitignore` that ignores generated contents inside both output directories while keeping the `.gitkeep` files tracked.

3. Update `main.go` so the default log file path becomes `output/logs/requests.log` and ensure the parent directory exists before opening the log file.

4. Update `main_test.go` to cover the new default log path and the log-file-open helper that creates parent directories.

5. Update `README.md` and `CLAUDE.md` so build instructions target `output/bin/event-horizon` and documentation reflects the new default log file path.
  </action>
  <verify>
    <automated>go test ./... && go build -o output/bin/event-horizon . && test -d output/bin && test -d output/logs && test -f output/bin/.gitkeep && test -f output/logs/.gitkeep && rm output/bin/event-horizon</automated>
  </verify>
  <done>`output/bin/` and `output/logs/` exist, generated contents are ignored, builds target `output/bin/`, and the default log path is `output/logs/requests.log`.</done>
</task>

</tasks>

<verification>
- `go test ./...` passes
- `go build -o output/bin/event-horizon .` succeeds and the binary can be removed without touching tracked placeholders
- `.gitignore` ignores generated contents under `output/bin/` and `output/logs/` while preserving `.gitkeep`
- `LOG_FILE` still overrides the default path when set
</verification>

<success_criteria>
- Repo has dedicated `output/bin/` and `output/logs/` directories
- Generated binary and log files are ignored by git
- Default runtime log file path is `output/logs/requests.log`
- Build and docs point users at `output/bin/event-horizon`
</success_criteria>

<output>
After completion, create `.planning/quick/7-create-an-output-folder-to-store-binary-/7-SUMMARY.md`
</output>
