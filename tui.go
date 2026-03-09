package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// requestMsg is the tea.Msg type for channel messages carrying request data.
type requestMsg RequestData

// flashExpiredMsg signals that the flash message should be cleared.
type flashExpiredMsg struct{}

// waitForRequest returns a tea.Cmd that blocks on the channel until a request arrives.
func waitForRequest(ch <-chan RequestData) tea.Cmd {
	return func() tea.Msg {
		return requestMsg(<-ch)
	}
}

// model is the bubbletea TUI model for displaying live HTTP requests.
type model struct {
	reqCh          <-chan RequestData
	requests       []RequestData
	port           string
	logPath        string
	width          int
	height         int
	selectedIndex  int
	expandedIndex  int
	scrollOffset   int
	flashMessage   string
	formatBody     bool
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
		case "j", "down":
			if len(m.requests) == 0 {
				return m, nil
			}
			m.selectedIndex = minInt(m.selectedIndex+1, len(m.requests)-1)
			m.scrollOffset = clampScrollOffset(m.scrollOffset, len(m.requests))
			return m, nil
		case "k", "up":
			if len(m.requests) == 0 {
				return m, nil
			}
			m.selectedIndex = maxInt(m.selectedIndex-1, 0)
			m.scrollOffset = clampScrollOffset(m.scrollOffset, len(m.requests))
			return m, nil
		case "enter", "space":
			if len(m.requests) == 0 {
				return m, nil
			}
			if m.expandedIndex == m.selectedIndex {
				m.expandedIndex = -1
			} else {
				m.expandedIndex = m.selectedIndex
			}
			m.scrollOffset = clampScrollOffset(m.scrollOffset, len(m.requests))
			return m, nil
		case "x":
			m.requests = nil
			m.selectedIndex = 0
			m.expandedIndex = -1
			m.scrollOffset = 0
			return m, nil
		case "c":
			if len(m.requests) == 0 {
				return m, nil
			}
			r := m.requests[clampIndex(m.selectedIndex, len(m.requests))]
			tickCmd := tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return flashExpiredMsg{} })
			if strings.TrimSpace(r.Body) == "" {
				m.flashMessage = "No body to copy"
				return m, tickCmd
			}
			m.flashMessage = "Copied!"
			return m, tea.Batch(tea.SetClipboard(r.Body), tickCmd)
		case "f":
			m.formatBody = !m.formatBody
			return m, nil
		case "C":
			if len(m.requests) == 0 {
				return m, nil
			}
			r := m.requests[clampIndex(m.selectedIndex, len(m.requests))]
			fullText := formatFullRequest(r)
			m.flashMessage = "Copied!"
			tickCmd := tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return flashExpiredMsg{} })
			return m, tea.Batch(tea.SetClipboard(fullText), tickCmd)
		}
	case requestMsg:
		m.requests = append(m.requests, RequestData(msg))
		if len(m.requests) == 1 {
			m.selectedIndex = 0
		}
		m.selectedIndex = clampIndex(m.selectedIndex, len(m.requests))
		if m.expandedIndex >= len(m.requests) {
			m.expandedIndex = -1
		}
		m.scrollOffset = clampScrollOffset(m.scrollOffset, len(m.requests))
		return m, waitForRequest(m.reqCh)
	case flashExpiredMsg:
		m.flashMessage = ""
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.selectedIndex = clampIndex(m.selectedIndex, len(m.requests))
		if m.expandedIndex >= len(m.requests) {
			m.expandedIndex = -1
		}
		m.scrollOffset = clampScrollOffset(m.scrollOffset, len(m.requests))
		return m, nil
	}
	return m, nil
}

func clampIndex(index int, length int) int {
	if length == 0 {
		return 0
	}
	if index < 0 {
		return 0
	}
	if index >= length {
		return length - 1
	}
	return index
}

func clampScrollOffset(offset int, length int) int {
	if length == 0 || offset < 0 {
		return 0
	}
	if offset >= length {
		return length - 1
	}
	return offset
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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
	colorSelected      = lipgloss.Color("39")
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
	selectedRowStyle   = lipgloss.NewStyle().
				Bold(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(colorSelected).
				PaddingLeft(1)
	detailLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(colorMutedBright)
	detailValueStyle = lipgloss.NewStyle().Foreground(colorNeutral)
	footerStyle      = lipgloss.NewStyle().Foreground(colorMutedBright)
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

func renderRequestRow(r RequestData, rowIndex int, width int, selected bool) string {
	content := strings.Join([]string{
		timestampStyle.Render(r.Timestamp.Format("15:04:05")),
		methodStyle(r.Method).Render(r.Method),
		pathStyle.Render(formatPath(r.URI)),
		statusStyle(r.Status).Render(fmt.Sprintf("%d", r.Status)),
		timeStyle.Render(formatResponseTime(r.ResponseTime)),
	}, " ")

	style := rowBaseStyle
	if selected {
		style = selectedRowStyle
	}

	row := style.Render(content)
	if rowIndex%2 == 1 && !selected {
		row = alternatingRowStyle.Render(row)
	}

	return row
}

func renderExpandedDetails(r RequestData, width int, formatBody bool) string {
	innerWidth := maxInt(width-4, 20)

	// Determine body content and label
	label := bodyLabel(r.Body, formatBody)
	var bodySection string
	if formatBody && isJSON(r.Body) {
		bodySection = renderHighlightedBodySection(label, prettyJSON(r.Body), innerWidth)
	} else {
		bodySection = renderDetailSection(label, blankFallback(r.Body), innerWidth)
	}

	sections := []string{
		renderDetailSection("Headers", formatHeaders(r.Headers), innerWidth),
		bodySection,
		renderDetailSection("Client IP", blankFallback(r.ClientIP), innerWidth),
		renderDetailSection("Response Time", formatResponseTime(r.ResponseTime), innerWidth),
	}
	return strings.Join(sections, "\n")
}

func renderHighlightedBodySection(label, value string, width int) string {
	lines := wrapText(value, width)
	if len(lines) == 0 {
		lines = []string{"-"}
	}

	var b strings.Builder
	b.WriteString(detailLabelStyle.Render(label))
	b.WriteString("\n")
	for _, line := range lines {
		b.WriteString("  ")
		b.WriteString(highlightJSONLine(line))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderDetailSection(label, value string, width int) string {
	lines := wrapText(value, width)
	if len(lines) == 0 {
		lines = []string{"-"}
	}

	var b strings.Builder
	b.WriteString(detailLabelStyle.Render(label))
	b.WriteString("\n")
	for _, line := range lines {
		b.WriteString("  ")
		b.WriteString(detailValueStyle.Render(line))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func formatHeaders(headers http.Header) string {
	if len(headers) == 0 {
		return "-"
	}

	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %s", key, strings.Join(headers.Values(key), ", ")))
	}
	return strings.Join(lines, "\n")
}

func formatFullRequest(r RequestData) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s %s\n", r.Method, r.URI))
	b.WriteString(fmt.Sprintf("Client IP: %s\n", blankFallback(r.ClientIP)))
	b.WriteString(fmt.Sprintf("Response Time: %s\n", formatResponseTime(r.ResponseTime)))
	b.WriteString(fmt.Sprintf("\nHeaders:\n%s\n", formatHeaders(r.Headers)))
	if strings.TrimSpace(r.Body) != "" {
		b.WriteString(fmt.Sprintf("\nBody:\n%s", r.Body))
	}
	return b.String()
}

func blankFallback(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func isJSON(s string) bool {
	return len(strings.TrimSpace(s)) > 0 && json.Valid([]byte(s))
}

func prettyJSON(s string) string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(s), "", "  "); err != nil {
		return s
	}
	return buf.String()
}

func bodyLabel(body string, formatOn bool) string {
	if formatOn && isJSON(body) {
		return "Body (JSON)"
	}
	return "Body (raw)"
}

var (
	jsonKeyStyle    = lipgloss.NewStyle().Foreground(colorCyan)
	jsonStringStyle = lipgloss.NewStyle().Foreground(colorGreen)
	jsonNumberStyle = lipgloss.NewStyle().Foreground(colorYellow)
	jsonBoolStyle   = lipgloss.NewStyle().Foreground(colorBlue)
	jsonNullStyle   = lipgloss.NewStyle().Foreground(colorMuted)
	jsonBraceStyle  = lipgloss.NewStyle().Foreground(colorNeutral)
)

// jsonTokenRe matches JSON tokens in pretty-printed output.
var jsonTokenRe = regexp.MustCompile(`("(?:[^"\\]|\\.)*")\s*:|("(?:[^"\\]|\\.)*")|(-?[0-9]+\.?[0-9]*(?:[eE][+-]?[0-9]+)?)|(\btrue\b|\bfalse\b)|(\bnull\b)|([{}\[\],:])|(\s+)`)

func highlightJSON(pretty string) string {
	lines := strings.Split(pretty, "\n")
	for i, line := range lines {
		lines[i] = highlightJSONLine(line)
	}
	return strings.Join(lines, "\n")
}

func highlightJSONLine(line string) string {
	if strings.TrimSpace(line) == "" {
		return line
	}

	var result strings.Builder

	// Preserve leading whitespace
	trimmed := strings.TrimLeft(line, " \t")
	leading := line[:len(line)-len(trimmed)]
	result.WriteString(leading)

	// Tokenize and highlight
	remaining := trimmed
	for len(remaining) > 0 {
		loc := jsonTokenRe.FindStringSubmatchIndex(remaining)
		if loc == nil || loc[0] != 0 {
			// No match at start -- write one char and advance
			if loc == nil {
				result.WriteString(remaining)
				break
			}
			result.WriteString(remaining[:loc[0]])
			remaining = remaining[loc[0]:]
			continue
		}

		fullMatch := remaining[loc[0]:loc[1]]

		switch {
		case loc[2] >= 0: // Group 1: key (string followed by colon)
			key := remaining[loc[2]:loc[3]]
			// fullMatch includes "key": -- render key styled, then the rest (whitespace + colon)
			after := fullMatch[len(key):]
			result.WriteString(jsonKeyStyle.Render(key))
			result.WriteString(jsonBraceStyle.Render(after))
		case loc[4] >= 0: // Group 2: string value
			str := remaining[loc[4]:loc[5]]
			result.WriteString(jsonStringStyle.Render(str))
		case loc[6] >= 0: // Group 3: number
			num := remaining[loc[6]:loc[7]]
			result.WriteString(jsonNumberStyle.Render(num))
		case loc[8] >= 0: // Group 4: bool
			b := remaining[loc[8]:loc[9]]
			result.WriteString(jsonBoolStyle.Render(b))
		case loc[10] >= 0: // Group 5: null
			n := remaining[loc[10]:loc[11]]
			result.WriteString(jsonNullStyle.Render(n))
		case loc[12] >= 0: // Group 6: structural chars
			ch := remaining[loc[12]:loc[13]]
			result.WriteString(jsonBraceStyle.Render(ch))
		case loc[14] >= 0: // Group 7: whitespace
			ws := remaining[loc[14]:loc[15]]
			result.WriteString(ws)
		default:
			result.WriteString(fullMatch)
		}

		remaining = remaining[loc[1]:]
	}

	return result.String()
}

func renderRequestBlock(r RequestData, index int, m model, width int) string {
	selected := len(m.requests) > 0 && index == m.selectedIndex
	parts := []string{renderRequestRow(r, index, width, selected)}
	if selected && m.expandedIndex == m.selectedIndex {
		parts = append(parts, renderExpandedDetails(r, width, m.formatBody))
	}
	return strings.Join(parts, "\n")
}

func visibleRequestBlocks(m model, contentHeight int) []string {
	blocks, _ := visibleRequestBlocksWithOffset(m, contentHeight)
	return blocks
}

func visibleRequestBlocksWithOffset(m model, contentHeight int) ([]string, int) {
	if len(m.requests) == 0 {
		return nil, 0
	}

	width := contentWidth(m.width)
	if contentHeight <= 0 {
		contentHeight = len(m.requests)
	}

	blockStrings := make([]string, len(m.requests))
	blockHeights := make([]int, len(m.requests))
	for i, req := range m.requests {
		blockStrings[i] = renderRequestBlock(req, i, m, width)
		blockHeights[i] = blockHeight(blockStrings[i])
	}

	selected := clampIndex(m.selectedIndex, len(m.requests))
	offset := clampScrollOffset(m.scrollOffset, len(m.requests))
	if selected < offset {
		offset = selected
	}

	for {
		total := 0
		for i := offset; i <= selected; i++ {
			total += blockHeights[i]
		}
		if total <= contentHeight || offset >= selected {
			break
		}
		offset++
	}

	visible := make([]string, 0, len(m.requests)-offset)
	usedHeight := 0
	for i := offset; i < len(blockStrings); i++ {
		nextHeight := blockHeights[i]
		if len(visible) > 0 && usedHeight+nextHeight > contentHeight {
			break
		}
		visible = append(visible, blockStrings[i])
		usedHeight += nextHeight
	}

	if len(visible) == 0 {
		visible = append(visible, blockStrings[selected])
		offset = selected
	}

	return visible, offset
}

func renderFooter(m model) string {
	selected := 0
	if len(m.requests) > 0 {
		selected = clampIndex(m.selectedIndex, len(m.requests)) + 1
	}

	formatState := "format: on"
	if !m.formatBody {
		formatState = "format: off"
	}
	status := fmt.Sprintf("%d requests | selected %d/%d | %s", len(m.requests), selected, len(m.requests), formatState)
	if m.flashMessage != "" {
		status += " | " + m.flashMessage
	}
	help := "j/k or arrows move | enter/space expand | c copy | C copy all | f format | x clear | q quit"
	return footerStyle.Render(status) + "\n" + footerStyle.Render(help)
}

func blockHeight(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func wrapText(value string, width int) []string {
	if width <= 0 {
		return []string{value}
	}

	var lines []string
	for _, rawLine := range strings.Split(value, "\n") {
		line := strings.TrimRight(rawLine, " ")
		if line == "" {
			lines = append(lines, "")
			continue
		}
		for len(line) > width {
			lines = append(lines, line[:width])
			line = line[width:]
		}
		lines = append(lines, line)
	}
	return lines
}

func contentWidth(width int) int {
	if width > 0 {
		return maxInt(width, 40)
	}
	return 40
}

// renderView builds the raw string for the TUI display.
// Extracted as a helper for testability (tea.View may not expose string content easily).
func renderView(m model) string {
	var s strings.Builder

	// Header bar
	s.WriteString(headerStyle.Render(fmt.Sprintf("event-horizon :%s -> %s", m.port, m.logPath)))
	s.WriteString("\n")

	// Separator
	sepWidth := contentWidth(m.width)
	s.WriteString(separatorStyle.Render(strings.Repeat("─", sepWidth)))
	s.WriteString("\n")

	// Content area
	if len(m.requests) == 0 {
		s.WriteString("\n          Waiting for requests...\n")
	} else {
		contentHeight := m.height - 5 // header + separators + 2-line footer
		for _, block := range visibleRequestBlocks(m, contentHeight) {
			s.WriteString(block)
			s.WriteString("\n")
		}
	}

	// Separator
	s.WriteString(separatorStyle.Render(strings.Repeat("─", sepWidth)))
	s.WriteString("\n")

	// Two-line footer: context and key help.
	s.WriteString(renderFooter(m))

	return s.String()
}
