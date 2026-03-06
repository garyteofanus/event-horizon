package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// handleRequest returns an http.HandlerFunc that logs every incoming request
// as structured JSON via the provided slog.Logger and responds with empty 200 OK.
func handleRequest(logger *slog.Logger) http.HandlerFunc {
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
		// Empty 200 OK -- do not write to w
	}
}

func main() {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	http.HandleFunc("/", handleRequest(logger))

	logger.LogAttrs(context.Background(), slog.LevelInfo, "server starting",
		slog.String("port", port),
	)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.LogAttrs(context.Background(), slog.LevelError, "server failed",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
}
