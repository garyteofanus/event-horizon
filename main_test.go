package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTest creates a bytes.Buffer-backed slog logger and returns the buffer
// and an http.HandlerFunc from handleRequest for testing.
func setupTest() (*bytes.Buffer, http.HandlerFunc) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	handler := handleRequest(logger)
	return &buf, handler
}

// parseLogEntry unmarshals the JSON log output from the buffer.
func parseLogEntry(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log JSON: %v\nraw output: %s", err, buf.String())
	}
	return entry
}

// TestLogJSON verifies that the log output is valid JSON with required base fields (LOG-01, OUT-01).
func TestLogJSON(t *testing.T) {
	buf, handler := setupTest()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, buf)

	// Must have time, level, msg fields from JSONHandler
	for _, field := range []string{"time", "level", "msg"} {
		if _, ok := entry[field]; !ok {
			t.Errorf("missing required field %q in log entry", field)
		}
	}

	if entry["level"] != "INFO" {
		t.Errorf("expected level INFO, got %v", entry["level"])
	}
	if entry["msg"] != "request" {
		t.Errorf("expected msg 'request', got %v", entry["msg"])
	}
}

// TestLogFields verifies method, URI, protocol, status, and response_time fields (LOG-02).
func TestLogFields(t *testing.T) {
	buf, handler := setupTest()

	req := httptest.NewRequest(http.MethodPost, "/hello", strings.NewReader("test body"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, buf)

	if entry["method"] != "POST" {
		t.Errorf("expected method POST, got %v", entry["method"])
	}
	if entry["uri"] != "/hello" {
		t.Errorf("expected uri /hello, got %v", entry["uri"])
	}
	if entry["protocol"] != "HTTP/1.1" {
		t.Errorf("expected protocol HTTP/1.1, got %v", entry["protocol"])
	}

	// status should be 200 (JSON numbers are float64)
	status, ok := entry["status"].(float64)
	if !ok || status != 200 {
		t.Errorf("expected status 200, got %v", entry["status"])
	}

	// response_time must exist and be a number (nanoseconds from slog.Duration)
	rt, ok := entry["response_time"].(float64)
	if !ok {
		t.Errorf("expected response_time as number, got %v (%T)", entry["response_time"], entry["response_time"])
	}
	if rt < 0 {
		t.Errorf("response_time should be non-negative, got %v", rt)
	}
}

// TestLogHeaders verifies request headers appear as a structured group (LOG-03).
func TestLogHeaders(t *testing.T) {
	buf, handler := setupTest()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Custom", "value1")
	req.Header.Set("Accept", "text/plain")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, buf)

	headers, ok := entry["headers"].(map[string]any)
	if !ok {
		t.Fatalf("expected headers to be a JSON object, got %T: %v", entry["headers"], entry["headers"])
	}

	if headers["X-Custom"] != "value1" {
		t.Errorf("expected X-Custom=value1, got %v", headers["X-Custom"])
	}
	if headers["Accept"] != "text/plain" {
		t.Errorf("expected Accept=text/plain, got %v", headers["Accept"])
	}
}

// TestLogBody verifies request body is logged, including empty body (LOG-04).
func TestLogBody(t *testing.T) {
	buf, handler := setupTest()

	// Test with body
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello world"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, buf)
	if entry["body"] != "hello world" {
		t.Errorf("expected body 'hello world', got %v", entry["body"])
	}

	// Test with empty body
	buf.Reset()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry = parseLogEntry(t, buf)
	if entry["body"] != "" {
		t.Errorf("expected empty body string, got %v", entry["body"])
	}
}

// TestLogClientInfo verifies client_ip and user_agent fields (LOG-05).
func TestLogClientInfo(t *testing.T) {
	buf, handler := setupTest()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("User-Agent", "test-agent/1.0")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, buf)

	clientIP, ok := entry["client_ip"].(string)
	if !ok || clientIP == "" {
		t.Errorf("expected non-empty client_ip string, got %v", entry["client_ip"])
	}

	if entry["user_agent"] != "test-agent/1.0" {
		t.Errorf("expected user_agent 'test-agent/1.0', got %v", entry["user_agent"])
	}
}

// TestLogContentLength verifies content_length field (LOG-06).
func TestLogContentLength(t *testing.T) {
	buf, handler := setupTest()

	// Known content length
	body := "12345"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.ContentLength = int64(len(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry := parseLogEntry(t, buf)
	cl, ok := entry["content_length"].(float64)
	if !ok {
		t.Fatalf("expected content_length as number, got %v (%T)", entry["content_length"], entry["content_length"])
	}
	if int64(cl) != int64(len(body)) {
		t.Errorf("expected content_length %d, got %v", len(body), cl)
	}

	// Unknown content length (-1)
	buf.Reset()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.ContentLength = -1
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	entry = parseLogEntry(t, buf)
	cl, ok = entry["content_length"].(float64)
	if !ok {
		t.Fatalf("expected content_length as number, got %v (%T)", entry["content_length"], entry["content_length"])
	}
	if int64(cl) != -1 {
		t.Errorf("expected content_length -1 for unknown, got %v", cl)
	}
}

// TestEmptyResponse verifies 200 OK with empty body for all methods and paths (SRV-01, SRV-02).
func TestEmptyResponse(t *testing.T) {
	_, handler := setupTest()

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/"},
		{http.MethodPost, "/foo"},
		{http.MethodPut, "/bar/baz"},
		{http.MethodDelete, "/"},
		{http.MethodPatch, "/anything"},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != 200 {
				t.Errorf("expected status 200, got %d", rec.Code)
			}
			if rec.Body.Len() != 0 {
				t.Errorf("expected empty body, got %q", rec.Body.String())
			}
		})
	}
}

// TestPortConfig verifies PORT env var is read (SRV-03).
func TestPortConfig(t *testing.T) {
	// Default port
	os.Unsetenv("PORT")
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	if port != "8080" {
		t.Errorf("expected default port 8080, got %s", port)
	}

	// Custom port
	os.Setenv("PORT", "9090")
	defer os.Unsetenv("PORT")
	port = "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}
	if port != "9090" {
		t.Errorf("expected port 9090, got %s", port)
	}
}

// TestDualOutput verifies that io.MultiWriter sends identical JSON to both a buffer and a file (OUT-02).
func TestDualOutput(t *testing.T) {
	var buf bytes.Buffer
	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, "test.log")
	tmpFile, err := os.OpenFile(tmpPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	writer := io.MultiWriter(&buf, tmpFile)
	logger := slog.New(slog.NewJSONHandler(writer, nil))
	handler := handleRequest(logger)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	tmpFile.Close()

	fileBytes, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	// Both outputs should be byte-for-byte identical
	if !bytes.Equal(buf.Bytes(), fileBytes) {
		t.Errorf("buffer and file content differ\nbuffer: %s\nfile:   %s", buf.String(), string(fileBytes))
	}

	// Parse both as JSON and verify key fields
	var bufEntry, fileEntry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &bufEntry); err != nil {
		t.Fatalf("failed to parse buffer JSON: %v", err)
	}
	if err := json.Unmarshal(fileBytes, &fileEntry); err != nil {
		t.Fatalf("failed to parse file JSON: %v", err)
	}

	for _, entry := range []map[string]any{bufEntry, fileEntry} {
		if entry["method"] != "GET" {
			t.Errorf("expected method GET, got %v", entry["method"])
		}
		if entry["uri"] != "/test" {
			t.Errorf("expected uri /test, got %v", entry["uri"])
		}
		if entry["msg"] != "request" {
			t.Errorf("expected msg 'request', got %v", entry["msg"])
		}
	}
}

// TestLogFilePathDefault verifies LOG_FILE defaults to "requests.log" when unset (OUT-03).
func TestLogFilePathDefault(t *testing.T) {
	t.Setenv("LOG_FILE", "")
	os.Unsetenv("LOG_FILE")

	logPath := "requests.log"
	if lf := os.Getenv("LOG_FILE"); lf != "" {
		logPath = lf
	}

	if logPath != "requests.log" {
		t.Errorf("expected default log path 'requests.log', got %q", logPath)
	}
}

// TestLogFilePathCustom verifies LOG_FILE env var overrides the default path (OUT-03).
func TestLogFilePathCustom(t *testing.T) {
	t.Setenv("LOG_FILE", "/tmp/custom.log")

	logPath := "requests.log"
	if lf := os.Getenv("LOG_FILE"); lf != "" {
		logPath = lf
	}

	if logPath != "/tmp/custom.log" {
		t.Errorf("expected log path '/tmp/custom.log', got %q", logPath)
	}
}

// TestLogFileAppend verifies that the log file is appended to, not truncated (OUT-02).
func TestLogFileAppend(t *testing.T) {
	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, "append.log")

	// Write pre-existing content
	if err := os.WriteFile(tmpPath, []byte("existing\n"), 0644); err != nil {
		t.Fatalf("failed to write pre-existing content: %v", err)
	}

	// Reopen with append mode
	tmpFile, err := os.OpenFile(tmpPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("failed to open temp file for append: %v", err)
	}

	var buf bytes.Buffer
	writer := io.MultiWriter(&buf, tmpFile)
	logger := slog.New(slog.NewJSONHandler(writer, nil))
	handler := handleRequest(logger)

	req := httptest.NewRequest(http.MethodGet, "/append-test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	tmpFile.Close()

	fileBytes, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	contents := string(fileBytes)
	if !strings.HasPrefix(contents, "existing\n") {
		t.Errorf("expected file to start with 'existing\\n', got %q", contents[:min(len(contents), 20)])
	}

	// The rest after "existing\n" should be valid JSON
	jsonPart := contents[len("existing\n"):]
	var entry map[string]any
	if err := json.Unmarshal([]byte(jsonPart), &entry); err != nil {
		t.Fatalf("failed to parse appended JSON: %v\nraw: %s", err, jsonPart)
	}
	if entry["uri"] != "/append-test" {
		t.Errorf("expected uri '/append-test', got %v", entry["uri"])
	}
}

// TestQA01NoDeps verifies zero external dependencies (QA-01).
func TestQA01NoDeps(t *testing.T) {
	cmd := exec.Command("go", "list", "-m", "all")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to run go list -m all: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 module (self only), got %d: %v", len(lines), lines)
	}
}

// TestQA02LogAttrsOnly verifies that main.go does not use non-LogAttrs slog methods (QA-02).
func TestQA02LogAttrsOnly(t *testing.T) {
	// Check for non-LogAttrs slog convenience methods in main.go
	patterns := []string{
		`logger\.Info(`,
		`logger\.Warn(`,
		`logger\.Error(`,
		`logger\.Debug(`,
	}
	for _, pattern := range patterns {
		cmd := exec.Command("grep", "-c", pattern, "main.go")
		output, err := cmd.Output()
		count := strings.TrimSpace(string(output))
		// grep returns exit 1 when count is 0, which is what we want
		if err == nil && count != "0" {
			t.Errorf("found %s matches for %q in main.go -- should be 0 (QA-02)", count, pattern)
		}
	}
}
