---
phase: quick
plan: 2
type: execute
wave: 1
depends_on: []
files_modified:
  - go.mod
  - CLAUDE.md
  - .planning/codebase/STACK.md
  - .planning/codebase/STRUCTURE.md
  - .planning/codebase/TESTING.md
  - .planning/codebase/INTEGRATIONS.md
  - .planning/codebase/CONCERNS.md
autonomous: true
must_haves:
  truths:
    - "go.mod declares module blackhole-server"
    - "go build -o blackhole-server . compiles successfully"
    - "CLAUDE.md references blackhole-server everywhere echo-server appeared"
    - "Parent directory is named blackhole-server"
  artifacts:
    - path: "go.mod"
      provides: "Module declaration"
      contains: "module blackhole-server"
    - path: "CLAUDE.md"
      provides: "Updated project docs"
      contains: "blackhole-server"
  key_links: []
---

<objective>
Rename the Go module from echo-server to blackhole-server, update all references across documentation and planning codebase files, and rename the parent directory.

Purpose: Rebrand the project from echo-server to blackhole-server.
Output: All files updated, module renamed, directory renamed.
</objective>

<execution_context>
@/Users/teo/.claude/get-shit-done/workflows/execute-plan.md
@/Users/teo/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@go.mod
@CLAUDE.md
@.planning/codebase/STACK.md
@.planning/codebase/STRUCTURE.md
@.planning/codebase/TESTING.md
@.planning/codebase/INTEGRATIONS.md
@.planning/codebase/CONCERNS.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Rename Go module and update all file references</name>
  <files>go.mod, CLAUDE.md, .planning/codebase/STACK.md, .planning/codebase/STRUCTURE.md, .planning/codebase/TESTING.md, .planning/codebase/INTEGRATIONS.md, .planning/codebase/CONCERNS.md</files>
  <action>
1. Update `go.mod` line 1: change `module echo-server` to `module blackhole-server`.

2. Update `CLAUDE.md`:
   - Line 7: change "A minimal HTTP echo server" to "A minimal HTTP blackhole server" and update the description: instead of "echoes the same information back to the client as plain text" say "responds with an empty 200 OK (swallowing all input like a blackhole)".
   - Line 12: change `go build -o echo-server .` to `go build -o blackhole-server .`
   - Line 20: change "expose the echo server publicly" to "expose the blackhole server publicly"
   - Line 21: update comment about starting "the echo server" to "the server"

3. Update `.planning/codebase/STACK.md`: replace all occurrences of `echo-server` with `blackhole-server` (module definition reference, build command).

4. Update `.planning/codebase/STRUCTURE.md`: replace `echo-server/` directory name reference with `blackhole-server/` and module name reference.

5. Update `.planning/codebase/TESTING.md`: replace `echo-server/` directory reference with `blackhole-server/`.

6. Update `.planning/codebase/INTEGRATIONS.md`: replace "echo server" with "blackhole server" in the description.

7. Update `.planning/codebase/CONCERNS.md`: replace "echo server" with "blackhole server" where it describes the server design.

Do NOT update files under `.planning/phases/`, `.planning/research/`, or `.planning/quick/1-*/` -- those are historical records and should remain accurate to when they were written.
  </action>
  <verify>
    <automated>cd /Users/teo/Developer/echo-server && grep -c "module blackhole-server" go.mod && grep -c "blackhole-server" CLAUDE.md && go vet ./...</automated>
  </verify>
  <done>go.mod declares module blackhole-server. CLAUDE.md references blackhole-server in all relevant places. go vet passes. Codebase docs updated.</done>
</task>

<task type="auto">
  <name>Task 2: Rename parent directory from echo-server to blackhole-server</name>
  <files></files>
  <action>
Rename the parent directory:
```
mv /Users/teo/Developer/echo-server /Users/teo/Developer/blackhole-server
```

Then verify the build still works from the new location:
```
cd /Users/teo/Developer/blackhole-server && go build -o blackhole-server .
```

Clean up the built binary after verification:
```
rm /Users/teo/Developer/blackhole-server/blackhole-server
```
  </action>
  <verify>
    <automated>cd /Users/teo/Developer/blackhole-server && go build -o blackhole-server . && rm blackhole-server && echo "BUILD OK"</automated>
  </verify>
  <done>Directory is /Users/teo/Developer/blackhole-server. Project builds successfully from new location.</done>
</task>

</tasks>

<verification>
- `grep "module blackhole-server" go.mod` returns a match
- `go build -o blackhole-server .` compiles without errors from /Users/teo/Developer/blackhole-server
- `grep -r "echo-server" go.mod CLAUDE.md .planning/codebase/` returns no matches
- Working directory is /Users/teo/Developer/blackhole-server
</verification>

<success_criteria>
- Go module is blackhole-server
- Parent directory is blackhole-server
- All active documentation references blackhole-server
- Project compiles and runs correctly
</success_criteria>

<output>
After completion, create `.planning/quick/2-rename-folder-and-go-module-to-blackhole/2-SUMMARY.md`
</output>
