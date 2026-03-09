package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
)

const defaultLogPath = "output/logs/requests.log"

func resolveLogPath() string {
	if lf := os.Getenv("LOG_FILE"); lf != "" {
		return lf
	}
	return defaultLogPath
}

func openLogFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func main() {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	logPath := resolveLogPath()

	logFile, err := openLogFile(logPath)
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

	// Start HTTP server in background goroutine
	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Run TUI on main goroutine
	p := tea.NewProgram(model{
		reqCh:      reqCh,
		port:       port,
		logPath:    logPath,
		formatBody: true,
	})
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}
