package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestRequestChannel verifies handleRequest sends RequestData to the channel.
func TestRequestChannel(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	ch := make(chan RequestData, 1)
	handler := handleRequest(logger, ch)

	req := httptest.NewRequest(http.MethodPost, "/hello", strings.NewReader("body"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	select {
	case rd := <-ch:
		if rd.Method != "POST" {
			t.Errorf("expected Method POST, got %s", rd.Method)
		}
		if rd.URI != "/hello" {
			t.Errorf("expected URI /hello, got %s", rd.URI)
		}
		if rd.Status != 200 {
			t.Errorf("expected Status 200, got %d", rd.Status)
		}
		if rd.Timestamp.IsZero() {
			t.Error("expected non-zero Timestamp")
		}
		if rd.ResponseTime <= 0 {
			t.Error("expected positive ResponseTime")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for RequestData on channel")
	}

	if rec.Code != 200 {
		t.Errorf("expected HTTP 200, got %d", rec.Code)
	}
}

// TestRequestChannelNonBlocking verifies handler doesn't block when channel is full.
func TestRequestChannelNonBlocking(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	ch := make(chan RequestData, 1)

	// Fill the channel
	ch <- RequestData{}

	handler := handleRequest(logger, ch)

	done := make(chan struct{})
	go func() {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != 200 {
			t.Errorf("expected HTTP 200, got %d", rec.Code)
		}
		close(done)
	}()

	select {
	case <-done:
		// success -- handler returned without blocking
	case <-time.After(2 * time.Second):
		t.Fatal("handler blocked on full channel")
	}
}

// TestFileOnlyLogging verifies handler writes JSON to the provided logger.
func TestFileOnlyLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	ch := make(chan RequestData, 1)
	handler := handleRequest(logger, ch)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log JSON: %v\nraw: %s", err, buf.String())
	}

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

// TestFormatRequestLine verifies the format "HH:MM:SS METHOD /path STATUS TIMEms".
func TestFormatRequestLine(t *testing.T) {
	ts := time.Date(2026, 3, 6, 14, 30, 45, 0, time.UTC)
	rd := RequestData{
		Timestamp:    ts,
		Method:       "GET",
		URI:          "/hello",
		Status:       200,
		ResponseTime: 42 * time.Millisecond,
	}

	line := formatRequestLine(rd)
	expected := "14:30:45 GET /hello 200 42ms"
	if line != expected {
		t.Errorf("expected %q, got %q", expected, line)
	}
}

// TestFormatRequestLineURITruncation verifies URI > 40 chars is truncated.
func TestFormatRequestLineURITruncation(t *testing.T) {
	ts := time.Date(2026, 3, 6, 14, 0, 0, 0, time.UTC)
	longURI := "/this-is-a-very-long-uri-that-exceeds-forty-characters-easily"
	rd := RequestData{
		Timestamp:    ts,
		Method:       "POST",
		URI:          longURI,
		Status:       200,
		ResponseTime: 5 * time.Millisecond,
	}

	line := formatRequestLine(rd)
	// URI should be truncated to 37 chars + "..."
	truncated := longURI[:37] + "..."
	expected := "14:00:00 POST " + truncated + " 200 5ms"
	if line != expected {
		t.Errorf("expected %q, got %q", expected, line)
	}
}

// TestFormatRequestLineSubMillisecond verifies <1ms for sub-millisecond times.
func TestFormatRequestLineSubMillisecond(t *testing.T) {
	ts := time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC)
	rd := RequestData{
		Timestamp:    ts,
		Method:       "GET",
		URI:          "/fast",
		Status:       200,
		ResponseTime: 500 * time.Microsecond,
	}

	line := formatRequestLine(rd)
	expected := "10:00:00 GET /fast 200 <1ms"
	if line != expected {
		t.Errorf("expected %q, got %q", expected, line)
	}
}
