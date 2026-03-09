# Technology Stack

**Analysis Date:** 2026-03-06

## Languages

**Primary:**
- Go 1.25.0 - Entire application (`main.go`)

**Secondary:**
- None

## Runtime

**Environment:**
- Go 1.25.0

**Package Manager:**
- Go Modules
- Lockfile: No `go.sum` present (no external dependencies to lock)

## Frameworks

**Core:**
- Go standard library `net/http` - HTTP server (`main.go`)

**Testing:**
- Not configured (no test files detected)

**Build/Dev:**
- `go build` - Compiles to binary
- `go run` - Development execution

## Key Dependencies

**Critical:**
- None. Zero external dependencies. The application uses only Go standard library packages:
  - `fmt` - Formatted I/O
  - `io` - Body reading (`io.ReadAll`)
  - `log` - Server startup logging
  - `net/http` - HTTP server and handler
  - `os` - Environment variable access
  - `sort` - Header key sorting
  - `time` - Timestamp formatting

**Infrastructure:**
- None

## Configuration

**Environment:**
- `PORT` env var - Overrides default listen port (default: `8080`)
- No `.env` file present; configuration is purely via environment variables

**Build:**
- `go.mod` - Module definition (`event-horizon`)
- No build configuration files (Makefile, Dockerfile, etc.)

**Build commands:**
```bash
go run main.go          # Run directly
go build -o event-horizon . # Compile binary
```

## Platform Requirements

**Development:**
- Go 1.25.0+
- No other tooling required

**Production:**
- Compiled Go binary (statically linked, no runtime dependencies)
- Listens on TCP port (default `8080`, configurable via `PORT`)

---

*Stack analysis: 2026-03-06*
