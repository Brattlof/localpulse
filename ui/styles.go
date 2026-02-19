package ui

import (
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Theme Theme

	Title    lipgloss.Style
	Subtitle lipgloss.Style

	Card      lipgloss.Style
	CardTitle lipgloss.Style
	CardValue lipgloss.Style

	Endpoint         lipgloss.Style
	EndpointSelected lipgloss.Style
	EndpointStatus   lipgloss.Style

	LogLine    lipgloss.Style
	LogError   lipgloss.Style
	LogSuccess lipgloss.Style

	Help     lipgloss.Style
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style

	Border lipgloss.Style
	Panel  lipgloss.Style
}

func NewStyles(theme Theme) *Styles {
	return &Styles{
		Theme: theme,

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			Padding(0, 1),

		Subtitle: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Padding(0, 1),

		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Background(theme.Card).
			Padding(0, 1).
			Margin(0, 0),

		CardTitle: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Bold(true),

		CardValue: lipgloss.NewStyle().
			Foreground(theme.Foreground).
			Bold(true),

		Endpoint: lipgloss.NewStyle().
			Padding(0, 1),

		EndpointSelected: lipgloss.NewStyle().
			Background(theme.Primary).
			Foreground(theme.Background).
			Bold(true).
			Padding(0, 1),

		EndpointStatus: lipgloss.NewStyle().
			Padding(0, 1),

		LogLine: lipgloss.NewStyle().
			Foreground(theme.Muted),

		LogError: lipgloss.NewStyle().
			Foreground(theme.Error),

		LogSuccess: lipgloss.NewStyle().
			Foreground(theme.Success),

		Help: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Padding(0, 1),

		HelpKey: lipgloss.NewStyle().
			Foreground(theme.Accent).
			Bold(true),

		HelpDesc: lipgloss.NewStyle().
			Foreground(theme.Muted),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border),

		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Border).
			Padding(0, 1),
	}
}

func (s *Styles) TitleBar(title string, width int) string {
	return s.Title.
		Width(width).
		Align(lipgloss.Center).
		Render(title)
}

func (s *Styles) CardView(title, value string, width int) string {
	titleLine := s.CardTitle.Render(title)
	valueLine := s.CardValue.Render(value)
	content := lipgloss.JoinVertical(lipgloss.Left, titleLine, valueLine)
	return s.Card.Width(width).Render(content)
}

func (s *Styles) EndpointLine(name, status string, latency string, selected bool) string {
	var style lipgloss.Style
	if selected {
		style = s.EndpointSelected
	} else {
		style = s.Endpoint
	}

	line := lipgloss.JoinHorizontal(
		lipgloss.Top,
		s.EndpointStatus.Render(status),
		style.Render(name),
		s.Theme.ColorMuted(latency),
	)

	return style.Render(line)
}

func (s *Styles) HelpBar(keys []HelpKey, width int) string {
	var items []string
	for _, k := range keys {
		items = append(items, lipgloss.JoinHorizontal(
			lipgloss.Top,
			s.HelpKey.Render("["+k.Key+"]"),
			s.HelpDesc.Render(" "+k.Desc),
		))
	}

	separator := s.HelpDesc.Render(" • ")
	content := lipgloss.JoinHorizontal(lipgloss.Top, items...)
	if separator != "" {
		content = ""
		for i, item := range items {
			if i > 0 {
				content += separator
			}
			content += item
		}
	}

	return s.Help.Width(width).Render(content)
}

type HelpKey struct {
	Key  string
	Desc string
}

func DefaultHelpKeys() []HelpKey {
	return []HelpKey{
		{Key: "r", Desc: "refresh"},
		{Key: "↑↓", Desc: "navigate"},
		{Key: "+/-", Desc: "load"},
		{Key: "a", Desc: "add"},
		{Key: "d", Desc: "delete"},
		{Key: "q", Desc: "quit"},
	}
}
