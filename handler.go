package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// RequestData holds the essential fields from an HTTP request for the TUI display.
type RequestData struct {
	Timestamp    time.Time
	Method       string
	URI          string
	Status       int
	ResponseTime time.Duration
}

// handleRequest returns an http.HandlerFunc that logs every incoming request
// as structured JSON via the provided slog.Logger, sends a RequestData summary
// to reqCh (non-blocking), and responds with empty 200 OK.
func handleRequest(logger *slog.Logger, reqCh chan<- RequestData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Read body
		body, _ := io.ReadAll(r.Body)

		// Build header attrs
		headerAttrs := make([]slog.Attr, 0, len(r.Header))
		for name, values := range r.Header {
			if len(values) == 1 {
				headerAttrs = append(headerAttrs, slog.String(name, values[0]))
			} else {
				headerAttrs = append(headerAttrs, slog.Any(name, values))
			}
		}

		elapsed := time.Since(start)

		logger.LogAttrs(context.Background(), slog.LevelInfo, "request",
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("protocol", r.Proto),
			slog.Int("status", 200),
			slog.Duration("response_time", elapsed),
			slog.String("client_ip", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
			slog.Int64("content_length", r.ContentLength),
			slog.String("body", string(body)),
			slog.GroupAttrs("headers", headerAttrs...),
		)

		// Non-blocking send to TUI channel
		select {
		case reqCh <- RequestData{
			Timestamp:    start,
			Method:       r.Method,
			URI:          r.RequestURI,
			Status:       200,
			ResponseTime: elapsed,
		}:
		default:
		}
		// Empty 200 OK -- do not write to w
	}
}

// formatRequestLine formats a RequestData into a human-readable line:
// "HH:MM:SS METHOD /path STATUS TIMEms"
func formatRequestLine(d RequestData) string {
	ts := d.Timestamp.Format("15:04:05")

	uri := d.URI
	if len(uri) > 40 {
		uri = uri[:37] + "..."
	}

	var timing string
	if d.ResponseTime < time.Millisecond {
		timing = "<1ms"
	} else {
		timing = fmt.Sprintf("%dms", d.ResponseTime.Milliseconds())
	}

	return fmt.Sprintf("%s %s %s %d %s", ts, d.Method, uri, d.Status, timing)
}
