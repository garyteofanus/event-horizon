# AGENTS.md

This file provides guidance to AI agents working with code in this repository.

## Overview

A minimal HTTP event-horizon server in Go (no external dependencies). It logs every incoming request (method, URI, headers, body) to stdout and responds with an empty 200 OK (swallowing all input like a black hole -- nothing escapes past the event horizon).

## Commands

- **Run:** `go run main.go` (listens on `:8080` by default, override with `PORT` env var)
- **Build:** `go build -o event-horizon .`

## Architecture

Single-file server (`main.go`) using only the Go standard library. One catch-all handler on `/` handles all routes and methods.

## Public Access (cloudflared)

To expose event-horizon publicly via a Cloudflare quick tunnel:

1. Start the server: `go run main.go &`
2. Start the tunnel: `cloudflared tunnel --url http://localhost:8080`
3. Use the printed `https://...trycloudflare.com` URL to access the server from anywhere
4. To stop: kill both processes (`fg` then Ctrl-C, or `pkill -f "go run main.go"` and `pkill cloudflared`)

No Cloudflare account required. The URL changes each time you restart the tunnel.
