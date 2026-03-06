# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

A minimal HTTP echo server in Go (no external dependencies). It logs every incoming request (method, URI, headers, body) to stdout and echoes the same information back to the client as plain text.

## Commands

- **Run:** `go run main.go` (listens on `:8080` by default, override with `PORT` env var)
- **Build:** `go build -o echo-server .`

## Architecture

Single-file server (`main.go`) using only the Go standard library. One catch-all handler on `/` handles all routes and methods.

## Public Access (cloudflared)

To expose the echo server publicly via a Cloudflare quick tunnel:

1. Start the server: `go run main.go &`
2. Start the tunnel: `cloudflared tunnel --url http://localhost:8080`
3. Use the printed `https://...trycloudflare.com` URL to access the server from anywhere
4. To stop: kill both processes (`fg` then Ctrl-C, or `pkill -f "go run main.go"` and `pkill cloudflared`)

No Cloudflare account required. The URL changes each time you restart the tunnel.
