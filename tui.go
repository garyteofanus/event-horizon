package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

var (
	colorGreen         = lipgloss.Color("10")
	colorBlue          = lipgloss.Color("12")
	colorRed           = lipgloss.Color("9")
	colorYellow        = lipgloss.Color("11")
	colorCyan          = lipgloss.Color("14")
	colorMuted         = lipgloss.Color("245")
	colorMutedBright   = lipgloss.Color("246")
	colorNeutral       = lipgloss.Color("252")
	colorBorder        = lipgloss.Color("240")
	headerStyle        = lipgloss.NewStyle().Bold(true)
	separatorStyle     = lipgloss.NewStyle().Faint(true)
	timestampStyle     = lipgloss.NewStyle().Foreground(colorMuted)
	pathStyle          = lipgloss.NewStyle()
	timeStyle          = lipgloss.NewStyle().Foreground(colorMutedBright)
	defaultMethodStyle = lipgloss.NewStyle().Bold(true).Foreground(colorNeutral)
	defaultStatusStyle = lipgloss.NewStyle().Bold(true).Foreground(colorNeutral)
	rowBaseStyle       = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(colorBorder).
				PaddingLeft(1)
	alternatingRowStyle = lipgloss.NewStyle().Faint(true)
)

func methodStyle(method string) lipgloss.Style {
	switch method {
	case "GET":
		return lipgloss.NewStyle().Bold(true).Foreground(colorGreen)
	case "POST":
		return lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	case "DELETE":
		return lipgloss.NewStyle().Bold(true).Foreground(colorRed)
	case "PUT":
		return lipgloss.NewStyle().Bold(true).Foreground(colorYellow)
	case "PATCH":
		return lipgloss.NewStyle().Bold(true).Foreground(colorCyan)
	default:
		return defaultMethodStyle
	}
}

func statusStyle(code int) lipgloss.Style {
	switch {
	case code >= 200 && code < 300:
		return lipgloss.NewStyle().Bold(true).Foreground(colorGreen)
	case code >= 300 && code < 400:
		return lipgloss.NewStyle().Bold(true).Foreground(colorCyan)
	case code >= 400 && code < 500:
		return lipgloss.NewStyle().Bold(true).Foreground(colorYellow)
	case code >= 500 && code < 600:
		return lipgloss.NewStyle().Bold(true).Foreground(colorRed)
	default:
		return defaultStatusStyle
	}
}

func formatPath(uri string) string {
	if len(uri) > 40 {
		return uri[:37] + "..."
	}
	return uri
}

func formatResponseTime(d time.Duration) string {
	if d < time.Millisecond {
		return "<1ms"
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

func renderRequestRow(r RequestData, rowIndex int, width int) string {
	content := strings.Join([]string{
		timestampStyle.Render(r.Timestamp.Format("15:04:05")),
		methodStyle(r.Method).Render(r.Method),
		pathStyle.Render(formatPath(r.URI)),
		statusStyle(r.Status).Render(fmt.Sprintf("%d", r.Status)),
		timeStyle.Render(formatResponseTime(r.ResponseTime)),
	}, " ")

	row := rowBaseStyle.Render(content)
	if rowIndex%2 == 1 {
		row = alternatingRowStyle.Render(row)
	}

	return row
}

// renderView builds the raw string for the TUI display.
// Extracted as a helper for testability (tea.View may not expose string content easily).
func renderView(m model) string {
	var s strings.Builder

	// Header bar
	s.WriteString(headerStyle.Render(fmt.Sprintf("blackhole :%s -> %s", m.port, m.logPath)))
	s.WriteString("\n")

	// Separator
	sepWidth := 40
	if m.width > sepWidth {
		sepWidth = m.width
	}
	s.WriteString(separatorStyle.Render(strings.Repeat("─", sepWidth)))
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
		for i, r := range reqs {
			s.WriteString(renderRequestRow(r, i, sepWidth))
			s.WriteString("\n")
		}
	}

	// Separator
	s.WriteString(separatorStyle.Render(strings.Repeat("─", sepWidth)))
	s.WriteString("\n")

	// Status line
	s.WriteString(fmt.Sprintf("%d requests . q to quit", len(m.requests)))

	return s.String()
}
