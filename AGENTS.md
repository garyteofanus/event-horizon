# AGENTS.md

This file provides guidance to AI agents working with code in this repository.

## Overview

A Go HTTP event-horizon server with a Bubble Tea TUI. It logs every incoming request (method, URI, headers, body) to structured JSON files under `logs/`, displays requests live in the terminal, and responds with an empty 200 OK.

## Commands

- **Run:** `go run .` (listens on `:8080` by default, override with `PORT` env var)
- **Build:** `go build -o bin/event-horizon .`

## Architecture

Multi-file Go server: `main.go` orchestrates the HTTP server and TUI, `handler.go` captures/logs requests, and `tui.go` renders the terminal UI. One catch-all handler on `/` handles all routes and methods.

## Public Access (cloudflared)

To expose event-horizon publicly via a Cloudflare quick tunnel:

1. Start the server: `go run . &`
2. Start the tunnel: `cloudflared tunnel --url http://localhost:8080`
3. Use the printed `https://...trycloudflare.com` URL to access the server from anywhere
4. To stop: kill both processes (`fg` then Ctrl-C, or `pkill -f "go run ."` and `pkill cloudflared`)

No Cloudflare account required. The URL changes each time you restart the tunnel.
