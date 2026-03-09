# event-horizon

A minimal HTTP server that swallows every request and shows them in a real-time terminal UI.

## What it does

event-horizon captures every incoming HTTP request -- method, URI, headers, body -- logs it as structured JSON, and displays it live in a terminal interface. Every request gets an empty 200 OK response. Nothing escapes past the event horizon.

## Features

- Real-time TUI log viewer with color-coded HTTP methods and status codes
- Structured JSON logging to file
- Expanded request detail view (headers, body, client IP, response time)
- JSON body pretty-printing with syntax highlighting
- Clipboard support (copy body or full request) via OSC52
- Keyboard-driven: j/k navigate, enter/space expand, c/C copy, f format toggle, x clear, q quit
- Zero-config: just run and send requests

## Quick start

```bash
go install github.com/garyteofanus/event-horizon@latest
event-horizon
```

Or clone and build:

```bash
git clone https://github.com/garyteofanus/event-horizon.git
cd event-horizon
go build -o event-horizon .
./event-horizon
```

Then in another terminal:

```bash
curl http://localhost:8080/hello
```

## Configuration

| Variable   | Default        | Description                    |
|------------|----------------|--------------------------------|
| `PORT`     | `8080`         | HTTP listen port               |
| `LOG_FILE` | `requests.log` | Path to structured JSON log    |

## Keybindings

| Key             | Action                        |
|-----------------|-------------------------------|
| `j` / `k`      | Move selection down / up      |
| `up` / `down`   | Move selection down / up      |
| `enter` / `space` | Toggle expanded detail view |
| `c`             | Copy request body to clipboard |
| `C`             | Copy full request to clipboard |
| `f`             | Toggle JSON body formatting   |
| `x`             | Clear all requests            |
| `q`             | Quit                          |

## Public access

Expose event-horizon publicly with a Cloudflare quick tunnel (no account required):

```bash
event-horizon &
cloudflared tunnel --url http://localhost:8080
```

Use the printed `https://...trycloudflare.com` URL to access the server from anywhere.

## License

[MIT](LICENSE)
