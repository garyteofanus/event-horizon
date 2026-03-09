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

func TestRequestChannelIncludesDetails(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	ch := make(chan RequestData, 1)
	handler := handleRequest(logger, ch)

	req := httptest.NewRequest(http.MethodPost, "/details?debug=1", strings.NewReader("hello world"))
	req.Header.Set("X-Debug", "yes")
	req.Header.Add("X-Trace", "abc")
	req.Header.Add("X-Trace", "def")
	req.RemoteAddr = "203.0.113.5:4321"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	select {
	case rd := <-ch:
		if rd.Body != "hello world" {
			t.Fatalf("expected body to be preserved, got %q", rd.Body)
		}
		if rd.ClientIP != "203.0.113.5" {
			t.Fatalf("expected client IP 203.0.113.5, got %q", rd.ClientIP)
		}
		if got := rd.Headers.Values("X-Trace"); len(got) != 2 || got[0] != "abc" || got[1] != "def" {
			t.Fatalf("expected X-Trace header values to be preserved, got %#v", got)
		}
		if rd.Headers.Get("X-Debug") != "yes" {
			t.Fatalf("expected X-Debug header to be preserved, got %q", rd.Headers.Get("X-Debug"))
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for RequestData on channel")
	}
}

func TestRequestChannelDetailsSurviveNonBlockingSend(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	ch := make(chan RequestData, 2)
	handler := handleRequest(logger, ch)

	firstReq := httptest.NewRequest(http.MethodPost, "/first", strings.NewReader("first-body"))
	firstReq.Header.Set("X-Request", "one")
	firstReq.RemoteAddr = "198.51.100.8:1111"
	handler.ServeHTTP(httptest.NewRecorder(), firstReq)

	secondReq := httptest.NewRequest(http.MethodPut, "/second", strings.NewReader("second-body"))
	secondReq.Header.Set("X-Request", "two")
	secondReq.RemoteAddr = "198.51.100.9:2222"
	handler.ServeHTTP(httptest.NewRecorder(), secondReq)

	first := <-ch
	second := <-ch

	if first.Body != "first-body" || second.Body != "second-body" {
		t.Fatalf("expected request bodies to survive channel send, got %q and %q", first.Body, second.Body)
	}
	if first.Headers.Get("X-Request") != "one" || second.Headers.Get("X-Request") != "two" {
		t.Fatalf("expected headers to survive channel send, got %q and %q", first.Headers.Get("X-Request"), second.Headers.Get("X-Request"))
	}
	if first.ClientIP != "198.51.100.8" || second.ClientIP != "198.51.100.9" {
		t.Fatalf("expected client IPs to survive channel send, got %q and %q", first.ClientIP, second.ClientIP)
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

