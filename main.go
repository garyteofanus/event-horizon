package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	logPath := "requests.log"
	if lf := os.Getenv("LOG_FILE"); lf != "" {
		logPath = lf
	}

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).LogAttrs(
			context.Background(), slog.LevelError, "failed to open log file",
			slog.String("path", logPath),
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer logFile.Close()

	logger := slog.New(slog.NewJSONHandler(logFile, nil))

	reqCh := make(chan RequestData, 256)
	http.HandleFunc("/", handleRequest(logger, reqCh))

	logger.LogAttrs(context.Background(), slog.LevelInfo, "server starting",
		slog.String("port", port),
		slog.String("log_file", logPath),
	)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.LogAttrs(context.Background(), slog.LevelError, "server failed",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
}
