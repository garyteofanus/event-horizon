package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// requestMsg is the tea.Msg type for channel messages carrying request data.
type requestMsg RequestData

// waitForRequest returns a tea.Cmd that blocks on the channel until a request arrives.
func waitForRequest(ch <-chan RequestData) tea.Cmd {
	return func() tea.Msg {
		return requestMsg(<-ch)
	}
}

// model is the bubbletea TUI model for displaying live HTTP requests.
type model struct {
	reqCh    <-chan RequestData
	requests []RequestData
	port     string
	logPath  string
	width    int
	height   int
}

func (m model) Init() tea.Cmd {
	return waitForRequest(m.reqCh)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case requestMsg:
		m.requests = append(m.requests, RequestData(msg))
		return m, waitForRequest(m.reqCh)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

func (m model) View() tea.View {
	return tea.NewView(renderView(m))
}

// renderView builds the raw string for the TUI display.
// Extracted as a helper for testability (tea.View may not expose string content easily).
func renderView(m model) string {
	var s strings.Builder

	// Header bar
	s.WriteString(fmt.Sprintf("blackhole :%s -> %s\n", m.port, m.logPath))

	// Separator
	sepWidth := 40
	if m.width > sepWidth {
		sepWidth = m.width
	}
	s.WriteString(strings.Repeat("─", sepWidth))
	s.WriteString("\n")

	// Content area
	if len(m.requests) == 0 {
		s.WriteString("\n          Waiting for requests...\n\n")
	} else {
		// Auto-scroll: only show last N lines that fit
		reqs := m.requests
		if m.height > 0 {
			availableLines := m.height - 4 // header + 2 separators + status
			if availableLines > 0 && len(reqs) > availableLines {
				reqs = reqs[len(reqs)-availableLines:]
			}
		}
		for _, r := range reqs {
			s.WriteString(formatRequestLine(r))
			s.WriteString("\n")
		}
	}

	// Separator
	s.WriteString(strings.Repeat("─", sepWidth))
	s.WriteString("\n")

	// Status line
	s.WriteString(fmt.Sprintf("%d requests . q to quit", len(m.requests)))

	return s.String()
}
