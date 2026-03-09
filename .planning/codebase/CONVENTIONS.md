# Coding Conventions

**Analysis Date:** 2026-03-06

## Naming Patterns

**Files:**
- Single `main.go` at project root (single-file server)
- Use lowercase, no separators for Go source files

**Functions:**
- Standard Go conventions: exported functions use PascalCase, unexported use camelCase
- Currently only `main()` exists; no exported functions

**Variables:**
- Short, idiomatic Go names: `p`, `k`, `v`, `w`, `r`
- Descriptive names for domain concepts: `port`, `timestamp`, `body`, `keys`

**Types:**
- No custom types defined; relies entirely on standard library types (`http.ResponseWriter`, `*http.Request`)

## Code Style

**Formatting:**
- `gofmt` (standard Go formatter) -- no custom configuration
- Use tabs for indentation (Go default)

**Linting:**
- No linter configuration detected (no `.golangci.yml`, no `staticcheck.conf`)
- Rely on `go vet` as minimum static analysis

## Import Organization

**Order:**
1. Standard library only (single block, alphabetically sorted)

**Path Aliases:**
- None; no external dependencies exist

**Pattern:** Group all stdlib imports in a single parenthesized block, sorted alphabetically. If external dependencies are added, use the standard Go convention of a blank line separating stdlib from third-party imports.

```go
import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)
```

## Error Handling

**Patterns:**
- Inline error checks with immediate handling (print and continue):
  ```go
  body, err := io.ReadAll(r.Body)
  if err != nil {
      fmt.Printf("Error reading body: %v\n", err)
  } else if len(body) > 0 {
      // use body
  }
  ```
- Fatal errors for startup failures: `log.Fatal(http.ListenAndServe(...))`
- No error wrapping (`fmt.Errorf` with `%w`) currently used
- No HTTP error responses sent to clients on failure (body read errors are logged but the response continues)

**Prescriptive guidance:**
- Use `log.Fatal` only for unrecoverable startup errors
- For request-scoped errors, log with `log.Printf` and return an appropriate HTTP status code via `http.Error(w, msg, code)`
- Wrap errors with `fmt.Errorf("context: %w", err)` when propagating

## Logging

**Framework:** Standard library (`fmt` and `log`)

**Patterns:**
- `fmt.Printf` / `fmt.Println` for request logging to stdout (structured with visual delimiters `=====`)
- `log.Printf` for server lifecycle messages (includes timestamp automatically)
- `log.Fatal` for fatal startup errors

**Prescriptive guidance:**
- Use `log.Printf` (not `fmt.Printf`) for all server-side logging to get automatic timestamps
- Use `fmt.Fprintf(w, ...)` only for writing HTTP responses

## Comments

**When to Comment:**
- Brief inline comments before logical sections within a function: `// Print headers sorted`, `// Print body`, `// Echo back`
- No function-level doc comments (only `main` exists)

**Prescriptive guidance:**
- Add GoDoc comments to all exported functions and types
- Use inline comments sparingly to explain "why", not "what"

## Function Design

**Size:** The single handler is ~45 lines, which is at the upper bound for a single function. If logic grows, extract into named functions.

**Parameters:** Use standard `http.HandlerFunc` signature `(http.ResponseWriter, *http.Request)`.

**Return Values:** Follow Go convention of `(result, error)` tuples for any extracted functions.

## Module Design

**Exports:** Single `main` package with no exported symbols. If the project grows:
- Extract handler logic into a separate package (e.g., `handler/`)
- Export handler constructors for testability

**Barrel Files:** Not applicable to Go.

## Environment Configuration

**Pattern:** Read from environment with a hardcoded default:
```go
port := "8080"
if p := os.Getenv("PORT"); p != "" {
    port = p
}
```

**Prescriptive guidance:** Continue this pattern for any new configuration. Do not use external config libraries for this project's scope.

---

*Convention analysis: 2026-03-06*
