package ui

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	SparklineMaxHeight = 8
)

var SparklineBars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func Sparkline(data []float64, maxHeight int, theme Theme) string {
	if len(data) == 0 {
		return ""
	}

	if maxHeight <= 0 || maxHeight > SparklineMaxHeight {
		maxHeight = SparklineMaxHeight
	}

	var min, max float64
	for i, v := range data {
		if i == 0 || v < min {
			min = v
		}
		if i == 0 || v > max {
			max = v
		}
	}

	if max == min {
		max = min + 1
	}

	var result string
	for _, v := range data {
		normalized := (v - min) / (max - min)
		idx := int(normalized * float64(len(SparklineBars)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(SparklineBars) {
			idx = len(SparklineBars) - 1
		}
		result += string(SparklineBars[idx])
	}

	return lipgloss.NewStyle().Foreground(theme.Accent).Render(result)
}

func ProgressBar(percent float64, width int, theme Theme) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}

	filled := int(float64(width) * percent)
	if filled > width {
		filled = width
	}

	empty := width - filled

	filledStyle := lipgloss.NewStyle().Foreground(theme.Primary)
	emptyStyle := lipgloss.NewStyle().Foreground(theme.Muted)

	bar := filledStyle.Render(repeat("█", filled)) + emptyStyle.Render(repeat("░", empty))

	return bar
}

func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func StatusBar(healthy, slow, down int, theme Theme) string {
	total := healthy + slow + down
	if total == 0 {
		return theme.ColorMuted("No endpoints")
	}

	var parts []string

	if healthy > 0 {
		parts = append(parts, theme.ColorSuccess("●"+itoa(healthy)))
	}
	if slow > 0 {
		parts = append(parts, theme.ColorWarning("●"+itoa(slow)))
	}
	if down > 0 {
		parts = append(parts, theme.ColorError("●"+itoa(down)))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var result string
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

func FormatLatency(ns int64) string {
	us := ns / 1000
	ms := ns / 1000000

	if ms > 1000 {
		s := ms / 1000
		return itoa(int(s)) + "s"
	}
	if ms > 0 {
		return itoa(int(ms)) + "ms"
	}
	if us > 0 {
		return itoa(int(us)) + "µs"
	}
	return itoa(int(ns)) + "ns"
}

func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return itoa(int(bytes)) + "B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return itoa(int(bytes/div)) + string("KMGTPE"[exp]) + "B"
}

func FormatRPS(rps float64) string {
	if rps >= 1000 {
		return itoa(int(rps/1000)) + "k/s"
	}
	return itoa(int(rps)) + "/s"
}

func FormatPercent(p float64) string {
	return itoa(int(p)) + "%"
}

func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
