# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Event-horizon is an HTTP server + TUI log viewer in Go. It captures every incoming HTTP request (method, URI, headers, body), logs structured JSON to a file, and displays requests in a real-time terminal UI built with Bubble Tea. Responds with empty 200 OK to all requests (swallowing input like a black hole -- nothing escapes past the event horizon).

## Architecture

Multi-file structure:

- **main.go:** Orchestration -- starts HTTP server in background goroutine, runs Bubble Tea TUI on main goroutine, bridges them via buffered channel
- **handler.go:** HTTP handler + RequestData type -- logs structured JSON via slog, sends to TUI channel (non-blocking)
- **tui.go:** Bubble Tea model -- live request list with lipgloss styling, expanded detail view, clipboard support (OSC52), JSON pretty-printing with syntax highlighting

## Dependencies

- Bubble Tea v2 (`charm.land/bubbletea/v2`) -- terminal UI framework
- Lipgloss v2 (`charm.land/lipgloss/v2`) -- terminal styling

No other direct dependencies.

## Commands

- **Run:** `go run .` (listens on `:8080` by default, override with `PORT` env var)
- **Build:** `go build -o output/bin/event-horizon .`
- **Test:** `go test ./...`
- **Test verbose:** `go test -v ./...`

## Key Features

- Structured JSON logging to file
- Real-time TUI with color-coded methods and status codes
- Keyboard navigation (j/k or arrows, enter/space to expand)
- Copy request body (`c`) or full request (`C`) via OSC52 clipboard
- JSON body formatting with syntax highlighting (`f` to toggle)
- Clear requests (`x`), quit (`q`)

## Environment Variables

- `PORT` -- HTTP listen port (default `8080`)
- `LOG_FILE` -- Path to structured JSON log file (default `output/logs/requests.log`)

## Public Access (cloudflared)

To expose event-horizon publicly via a Cloudflare quick tunnel:

1. Start the server: `go run . &`
2. Start the tunnel: `cloudflared tunnel --url http://localhost:8080`
3. Use the printed `https://...trycloudflare.com` URL to access the server from anywhere
4. To stop: kill both processes (`fg` then Ctrl-C, or `pkill -f "go run"` and `pkill cloudflared`)

No Cloudflare account required. The URL changes each time you restart the tunnel.
