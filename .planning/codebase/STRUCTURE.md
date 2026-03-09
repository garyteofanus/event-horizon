# Codebase Structure

**Analysis Date:** 2026-03-06

## Directory Layout

```
event-horizon/
├── .claude/                # Claude Code local settings
│   └── settings.local.json
├── .planning/              # GSD planning documents
│   └── codebase/           # Codebase analysis docs (this file)
├── CLAUDE.md               # Claude Code project instructions
├── go.mod                  # Go module definition
└── main.go                 # Entire application source
```

## Directory Purposes

**Root (`/`):**
- Purpose: Contains the entire application -- there are no subdirectories for source code
- Contains: Go module file, single source file, project documentation
- Key files: `main.go`, `go.mod`

**`.claude/`:**
- Purpose: Claude Code tool configuration
- Contains: Local settings JSON
- Key files: `settings.local.json`

**`.planning/codebase/`:**
- Purpose: GSD codebase analysis documents
- Contains: Architecture and structure markdown files
- Generated: Yes (by GSD mapping)
- Committed: Project-dependent

## Key File Locations

**Entry Points:**
- `main.go`: Sole entry point -- contains `func main()` and the HTTP handler

**Configuration:**
- `go.mod`: Go module definition (module name: `event-horizon`, Go 1.25.0)
- `CLAUDE.md`: Project instructions for Claude Code

**Core Logic:**
- `main.go:19-63`: The catch-all HTTP handler (anonymous function)
- `main.go:13-17`: Server bootstrap and PORT configuration

**Testing:**
- No test files exist in this project

## Naming Conventions

**Files:**
- Go source: lowercase `main.go` (standard Go convention)
- Documentation: UPPERCASE.md (`CLAUDE.md`)
- Module: `go.mod` (Go standard)

**Directories:**
- Dot-prefixed for tooling: `.claude/`, `.planning/`

## Where to Add New Code

**New Handler / Route:**
- If the server grows, extract handlers into a `handlers/` directory or add named handler functions in `main.go`
- Register new routes via `http.HandleFunc` in `main()` alongside the existing catch-all

**New Feature (e.g., middleware, structured logging):**
- For a single-file project this small, add directly to `main.go`
- If the file exceeds ~200 lines, split into packages: `handlers/`, `middleware/`, `server/`

**Tests:**
- Create `main_test.go` in the project root (co-located, standard Go convention)
- Use `httptest.NewServer` or `httptest.NewRecorder` for handler testing

**Utilities:**
- For shared helpers, create a `pkg/` or `internal/` directory following Go conventions

## Special Directories

**`.planning/`:**
- Purpose: GSD analysis and planning documents
- Generated: Yes
- Committed: Project-dependent

**`.claude/`:**
- Purpose: Claude Code local tool settings
- Generated: Yes
- Committed: No (typically gitignored)

---

*Structure analysis: 2026-03-06*
