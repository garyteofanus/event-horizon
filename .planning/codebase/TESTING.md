# Testing Patterns

**Analysis Date:** 2026-03-06

## Test Framework

**Runner:**
- Go standard `testing` package (built-in)
- Config: None required; `go test` discovers `*_test.go` files automatically

**Assertion Library:**
- Use standard library comparisons (`if got != want`) or add `testify` if needed
- No assertion library currently in use

**Run Commands:**
```bash
go test ./...              # Run all tests
go test -v ./...           # Verbose output
go test -cover ./...       # Coverage summary
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out  # HTML coverage report
```

## Test File Organization

**Location:**
- Co-located: place `main_test.go` alongside `main.go` in the project root
- If packages are extracted, place `foo_test.go` alongside `foo.go` in the same package directory

**Naming:**
- `{source}_test.go` (e.g., `main_test.go`)

**Structure:**
```
blackhole-server/
├── main.go
└── main_test.go        # Does not exist yet -- needs to be created
```

## Current Test Coverage

**Status: No tests exist.** There are zero `*_test.go` files in the repository. This is the highest-priority quality gap.

## Recommended Test Structure

**Suite Organization:**
```go
func TestEchoHandler(t *testing.T) {
    t.Run("returns method and URI in response", func(t *testing.T) {
        // arrange
        req := httptest.NewRequest(http.MethodGet, "/test-path", nil)
        rec := httptest.NewRecorder()

        // act
        handler(rec, req)

        // assert
        if rec.Code != http.StatusOK {
            t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
        }
    })
}
```

**Patterns:**
- Use `t.Run` subtests for related cases
- Use `httptest.NewRequest` and `httptest.NewRecorder` for HTTP handler testing
- Follow arrange/act/assert structure

## Mocking

**Framework:** No mocking framework needed for current scope.

**Patterns:**
- Use `net/http/httptest` package for request/response simulation:
  ```go
  req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"key":"value"}`))
  req.Header.Set("Content-Type", "application/json")
  rec := httptest.NewRecorder()
  ```

**What to Mock:**
- External HTTP calls (if added in future) via `httptest.NewServer`
- Time functions if deterministic output is needed (inject `time.Now` as a dependency)

**What NOT to Mock:**
- The HTTP handler itself -- test it directly with `httptest`
- Standard library functions

## Fixtures and Factories

**Test Data:**
```go
// Define inline in tests for this project's scope
body := `{"example": "payload"}`
req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(body))
```

**Location:**
- For a single-file project, inline test data in `main_test.go`
- If the project grows, use `testdata/` directory for larger fixtures (Go convention)

## Coverage

**Requirements:** None enforced currently.

**Recommended minimum:** Aim for coverage of the handler's key behaviors:
- GET request echo
- POST request with body echo
- Header sorting and echo
- Content-Type response header

**View Coverage:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Types

**Unit Tests:**
- Test the HTTP handler function directly using `httptest`
- This is the primary test type for this project

**Integration Tests:**
- Use `httptest.NewServer` to spin up a real HTTP server in tests:
  ```go
  srv := httptest.NewServer(http.HandlerFunc(handler))
  defer srv.Close()
  resp, err := http.Get(srv.URL + "/test")
  ```

**E2E Tests:**
- Not applicable for this project's scope

## Prerequisite for Testability

The handler logic is currently an anonymous function inside `main()`. To enable testing, extract it into a named function:

```go
// In main.go
func echoHandler(w http.ResponseWriter, r *http.Request) {
    // ... current handler logic ...
}

func main() {
    // ...
    http.HandleFunc("/", echoHandler)
    // ...
}
```

This is the **required first step** before any tests can be written.

## Common Patterns

**Async Testing:**
- Not applicable; handler is synchronous

**Error Testing:**
```go
func TestEchoHandler_BodyReadError(t *testing.T) {
    req := httptest.NewRequest(http.MethodPost, "/", &errorReader{})
    rec := httptest.NewRecorder()
    echoHandler(rec, req)
    // verify graceful handling
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (int, error) {
    return 0, fmt.Errorf("simulated read error")
}
```

---

*Testing analysis: 2026-03-06*
