package components

import (
	"strings"
	"time"

	"github.com/localpulse/localpulse/monitor"
	"github.com/localpulse/localpulse/ui"
)

type LogEntry struct {
	Timestamp time.Time
	Message   string
	IsError   bool
	IsSuccess bool
}

type LogPanel struct {
	Entries  []LogEntry
	MaxLines int
	Width    int
	Height   int
	styles   *ui.Styles
}

func NewLogPanel(maxLines int, styles *ui.Styles) *LogPanel {
	if maxLines <= 0 {
		maxLines = 100
	}
	return &LogPanel{
		Entries:  make([]LogEntry, 0),
		MaxLines: maxLines,
		styles:   styles,
	}
}

func (p *LogPanel) AddEntry(message string, isError, isSuccess bool) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Message:   message,
		IsError:   isError,
		IsSuccess: isSuccess,
	}

	p.Entries = append(p.Entries, entry)

	if len(p.Entries) > p.MaxLines {
		p.Entries = p.Entries[len(p.Entries)-p.MaxLines:]
	}
}

func (p *LogPanel) AddRequest(result monitor.RequestResult, endpoint string) {
	var message string
	var isError, isSuccess bool

	if result.IsError {
		message = endpoint + " ERROR: " + result.ErrorMessage
		isError = true
	} else {
		status := "OK"
		if result.StatusCode >= 400 {
			status = "ERR"
			isError = true
		} else {
			isSuccess = true
		}
		message = endpoint + " " + status + " " + ui.FormatLatency(result.Latency.Nanoseconds())
	}

	p.AddEntry(message, isError, isSuccess)
}

func (p *LogPanel) SetSize(width, height int) {
	p.Width = width
	p.Height = height
}

func (p *LogPanel) View() string {
	if len(p.Entries) == 0 {
		empty := p.styles.Theme.ColorMuted("Waiting for requests...")
		return p.styles.Panel.Width(p.Width).Height(p.Height).Render(empty)
	}

	start := 0
	if len(p.Entries) > p.Height-2 {
		start = len(p.Entries) - (p.Height - 2)
	}

	var lines []string
	for i := start; i < len(p.Entries); i++ {
		entry := p.Entries[i]
		timeStr := entry.Timestamp.Format("15:04:05")

		var line string
		if entry.IsError {
			line = p.styles.LogError.Render(timeStr+" ") + entry.Message
		} else if entry.IsSuccess {
			line = p.styles.LogSuccess.Render(timeStr+" ") + entry.Message
		} else {
			line = p.styles.LogLine.Render(timeStr+" ") + entry.Message
		}

		if p.Width > 0 && len(line) > p.Width-4 {
			line = line[:p.Width-7] + "..."
		}

		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	return p.styles.Panel.Width(p.Width).Height(p.Height).Render(content)
}

func (p *LogPanel) Clear() {
	p.Entries = p.Entries[:0]
}
