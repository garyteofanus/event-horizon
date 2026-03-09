package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

func colorString(c any) string {
	return fmt.Sprint(c)
}

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

func TestModelUpdateNavigation(t *testing.T) {
	m := model{
		requests: []RequestData{
			{Method: "GET", URI: "/one"},
			{Method: "POST", URI: "/two"},
			{Method: "DELETE", URI: "/three"},
		},
		selectedIndex: 0,
		expandedIndex: -1,
	}

	result, _ := m.Update(tea.KeyPressMsg{Code: 'j'})
	updated := result.(model)
	if updated.selectedIndex != 1 {
		t.Fatalf("expected selection to move down to 1, got %d", updated.selectedIndex)
	}

	result, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	updated = result.(model)
	if updated.selectedIndex != 2 {
		t.Fatalf("expected selection to move down to 2 with arrow key, got %d", updated.selectedIndex)
	}

	result, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	updated = result.(model)
	if updated.selectedIndex != 1 {
		t.Fatalf("expected selection to move up to 1 with arrow key, got %d", updated.selectedIndex)
	}

	result, _ = updated.Update(tea.KeyPressMsg{Code: 'k'})
	updated = result.(model)
	if updated.selectedIndex != 0 {
		t.Fatalf("expected selection to move up to 0 with k, got %d", updated.selectedIndex)
	}
}

func TestModelUpdateNavigationBounds(t *testing.T) {
	m := model{
		requests:       []RequestData{{URI: "/one"}, {URI: "/two"}},
		selectedIndex:  0,
		expandedIndex:  -1,
		scrollOffset:   0,
	}

	result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	updated := result.(model)
	if updated.selectedIndex != 0 {
		t.Fatalf("expected selection to clamp at first row, got %d", updated.selectedIndex)
	}

	updated.selectedIndex = 1
	result, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	updated = result.(model)
	if updated.selectedIndex != 1 {
		t.Fatalf("expected selection to clamp at last row, got %d", updated.selectedIndex)
	}
}

func TestModelUpdateToggleExpand(t *testing.T) {
	m := model{
		requests:       []RequestData{{URI: "/one"}, {URI: "/two"}},
		selectedIndex:  1,
		expandedIndex:  -1,
	}

	result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated := result.(model)
	if updated.expandedIndex != 1 {
		t.Fatalf("expected selected row to expand, got %d", updated.expandedIndex)
	}
	if updated.selectedIndex != 1 {
		t.Fatalf("expected selection to stay on expanded row, got %d", updated.selectedIndex)
	}

	result, _ = updated.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	updated = result.(model)
	if updated.expandedIndex != -1 {
		t.Fatalf("expected expanded row to collapse, got %d", updated.expandedIndex)
	}
}

func TestModelUpdateClearRequests(t *testing.T) {
	m := model{
		requests:       []RequestData{{URI: "/one"}, {URI: "/two"}},
		selectedIndex:  1,
		expandedIndex:  1,
		scrollOffset:   3,
	}

	result, _ := m.Update(tea.KeyPressMsg{Code: 'c'})
	updated := result.(model)

	if len(updated.requests) != 0 {
		t.Fatalf("expected requests to be cleared, got %d", len(updated.requests))
	}
	if updated.selectedIndex != 0 {
		t.Fatalf("expected selectedIndex reset to 0, got %d", updated.selectedIndex)
	}
	if updated.expandedIndex != -1 {
		t.Fatalf("expected expandedIndex reset to -1, got %d", updated.expandedIndex)
	}
	if updated.scrollOffset != 0 {
		t.Fatalf("expected scrollOffset reset to 0, got %d", updated.scrollOffset)
	}
}

func TestModelUpdateAppendPreservesSelection(t *testing.T) {
	m := model{
		reqCh:          make(chan RequestData),
		requests:       []RequestData{{URI: "/one"}, {URI: "/two"}},
		selectedIndex:  0,
		expandedIndex:  1,
		scrollOffset:   0,
	}

	result, cmd := m.Update(requestMsg(RequestData{URI: "/three"}))
	updated := result.(model)

	if len(updated.requests) != 3 {
		t.Fatalf("expected appended request, got %d rows", len(updated.requests))
	}
	if updated.selectedIndex != 0 {
		t.Fatalf("expected selection to stay on existing row, got %d", updated.selectedIndex)
	}
	if updated.expandedIndex != 1 {
		t.Fatalf("expected expanded index to stay stable, got %d", updated.expandedIndex)
	}
	if cmd == nil {
		t.Fatal("expected waitForRequest command after append")
	}
}

func TestModelViewEmpty(t *testing.T) {
	m := model{port: "8080", logPath: "requests.log"}
	view := stripANSI(renderView(m))

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
	view := stripANSI(renderView(m))

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
	view := stripANSI(renderView(m))

	if !strings.Contains(view, "blackhole :9090 -> /tmp/test.log") {
		t.Errorf("view should contain header 'blackhole :9090 -> /tmp/test.log', got:\n%s", view)
	}
}

func TestRenderRequestRowContainsCompactFields(t *testing.T) {
	row := stripANSI(renderRequestRow(RequestData{
		Timestamp:    time.Date(2026, 3, 6, 14, 32, 5, 0, time.UTC),
		Method:       "GET",
		URI:          "/api/users?page=1",
		Status:       200,
		ResponseTime: 2 * time.Millisecond,
	}, 0, 80))

	if !strings.Contains(row, "14:32:05 GET /api/users?page=1 200 2ms") {
		t.Fatalf("expected compact one-line row, got %q", row)
	}
	if strings.Count(row, "\n") != 0 {
		t.Fatalf("expected request row to stay on one line, got %q", row)
	}
	if !strings.HasPrefix(row, "│ ") {
		t.Fatalf("expected row border prefix, got %q", row)
	}
}

func TestMethodStyleMappings(t *testing.T) {
	tests := []struct {
		method   string
		expected string
	}{
		{method: "GET", expected: colorString(colorGreen)},
		{method: "POST", expected: colorString(colorBlue)},
		{method: "DELETE", expected: colorString(colorRed)},
		{method: "PUT", expected: colorString(colorYellow)},
		{method: "PATCH", expected: colorString(colorCyan)},
		{method: "OPTIONS", expected: colorString(colorNeutral)},
	}

	for _, tc := range tests {
		style := methodStyle(tc.method)
		if !style.GetBold() {
			t.Fatalf("expected bold style for method %s", tc.method)
		}
		if got := colorString(style.GetForeground()); got != tc.expected {
			t.Fatalf("expected method %s color %s, got %s", tc.method, tc.expected, got)
		}
	}
}

func TestStatusStyleMappings(t *testing.T) {
	tests := []struct {
		status   int
		expected string
	}{
		{status: 200, expected: colorString(colorGreen)},
		{status: 302, expected: colorString(colorCyan)},
		{status: 404, expected: colorString(colorYellow)},
		{status: 500, expected: colorString(colorRed)},
		{status: 102, expected: colorString(colorNeutral)},
	}

	for _, tc := range tests {
		style := statusStyle(tc.status)
		if !style.GetBold() {
			t.Fatalf("expected bold style for status %d", tc.status)
		}
		if got := colorString(style.GetForeground()); got != tc.expected {
			t.Fatalf("expected status %d color %s, got %s", tc.status, tc.expected, got)
		}
	}
}

func TestMethodStylesRenderANSI(t *testing.T) {
	rendered := methodStyle("GET").Render("GET")
	if !strings.Contains(rendered, "\x1b[") {
		t.Fatalf("expected ANSI-styled method render, got %q", rendered)
	}
}

func TestStatusStylesRenderANSI(t *testing.T) {
	rendered := statusStyle(200).Render("200")
	if !strings.Contains(rendered, "\x1b[") {
		t.Fatalf("expected ANSI-styled status render, got %q", rendered)
	}
}

func TestRenderViewSeparatesAdjacentRows(t *testing.T) {
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
				Status:       404,
				ResponseTime: 5 * time.Millisecond,
			},
		},
	}

	view := stripANSI(renderView(m))
	if strings.Count(view, "│ ") < 2 {
		t.Fatalf("expected each request row to include a border prefix, got:\n%s", view)
	}
}

func TestRenderViewPreservesBottomAppendOrder(t *testing.T) {
	m := model{
		port:    "8080",
		logPath: "requests.log",
		requests: []RequestData{
			{
				Timestamp:    time.Date(2026, 1, 1, 14, 32, 5, 0, time.UTC),
				Method:       "GET",
				URI:          "/first",
				Status:       200,
				ResponseTime: 2 * time.Millisecond,
			},
			{
				Timestamp:    time.Date(2026, 1, 1, 14, 32, 10, 0, time.UTC),
				Method:       "POST",
				URI:          "/second",
				Status:       200,
				ResponseTime: 5 * time.Millisecond,
			},
		},
	}

	view := stripANSI(renderView(m))
	first := strings.Index(view, "/first")
	second := strings.Index(view, "/second")
	if first == -1 || second == -1 || first >= second {
		t.Fatalf("expected older request before newer request, got:\n%s", view)
	}
}

func TestRenderViewAutoScrollStillShowsMostRecentRows(t *testing.T) {
	requests := make([]RequestData, 0, 5)
	for i := 0; i < 5; i++ {
		requests = append(requests, RequestData{
			Timestamp:    time.Date(2026, 1, 1, 14, 32, 5+i, 0, time.UTC),
			Method:       "GET",
			URI:          fmt.Sprintf("/req-%d", i),
			Status:       200,
			ResponseTime: 2 * time.Millisecond,
		})
	}

	m := model{
		port:     "8080",
		logPath:  "requests.log",
		requests: requests,
		height:   6,
	}

	view := stripANSI(renderView(m))
	if strings.Contains(view, "/req-0") || strings.Contains(view, "/req-1") {
		t.Fatalf("expected oldest requests to be clipped, got:\n%s", view)
	}
	if !strings.Contains(view, "/req-3") || !strings.Contains(view, "/req-4") {
		t.Fatalf("expected newest requests to remain visible, got:\n%s", view)
	}
}
