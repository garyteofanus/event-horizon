package main

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestModelUpdateRequestMsg(t *testing.T) {
	ch := make(chan RequestData, 1)
	m := model{reqCh: ch, port: "8080", logPath: "requests.log"}

	msg := requestMsg(RequestData{
		Timestamp:    time.Now(),
		Method:       "GET",
		URI:          "/hello",
		Status:       200,
		ResponseTime: 5 * time.Millisecond,
	})

	result, cmd := m.Update(msg)
	updated := result.(model)

	if len(updated.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(updated.requests))
	}
	if updated.requests[0].Method != "GET" {
		t.Errorf("expected method GET, got %s", updated.requests[0].Method)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd (re-registered waitForRequest)")
	}
}

func TestModelUpdateQuit(t *testing.T) {
	m := model{port: "8080", logPath: "requests.log"}

	// Test "q" key
	msg := tea.KeyPressMsg{Code: 'q'}
	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Error("expected non-nil cmd for quit")
	}
}

func TestModelViewEmpty(t *testing.T) {
	m := model{port: "8080", logPath: "requests.log"}
	view := renderView(m)

	if !strings.Contains(view, "Waiting for requests...") {
		t.Errorf("empty view should contain 'Waiting for requests...', got:\n%s", view)
	}
}

func TestModelViewWithRequests(t *testing.T) {
	m := model{
		port:    "8080",
		logPath: "requests.log",
		requests: []RequestData{
			{
				Timestamp:    time.Date(2026, 1, 1, 14, 32, 5, 0, time.UTC),
				Method:       "GET",
				URI:          "/api/users",
				Status:       200,
				ResponseTime: 2 * time.Millisecond,
			},
			{
				Timestamp:    time.Date(2026, 1, 1, 14, 32, 10, 0, time.UTC),
				Method:       "POST",
				URI:          "/api/data",
				Status:       200,
				ResponseTime: 5 * time.Millisecond,
			},
		},
	}
	view := renderView(m)

	if !strings.Contains(view, "14:32:05 GET /api/users 200 2ms") {
		t.Errorf("view should contain first request line, got:\n%s", view)
	}
	if !strings.Contains(view, "14:32:10 POST /api/data 200 5ms") {
		t.Errorf("view should contain second request line, got:\n%s", view)
	}
	if !strings.Contains(view, "2 requests") {
		t.Errorf("view should contain '2 requests' in status, got:\n%s", view)
	}
}

func TestModelViewHeader(t *testing.T) {
	m := model{port: "9090", logPath: "/tmp/test.log"}
	view := renderView(m)

	if !strings.Contains(view, "blackhole :9090 -> /tmp/test.log") {
		t.Errorf("view should contain header 'blackhole :9090 -> /tmp/test.log', got:\n%s", view)
	}
}
