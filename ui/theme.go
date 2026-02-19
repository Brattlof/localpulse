package ui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Name       string
	IsDark     bool
	Background lipgloss.Color
	Foreground lipgloss.Color

	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color

	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color

	Muted  lipgloss.Color
	Border lipgloss.Color
	Card   lipgloss.Color
}

var (
	DarkTheme = Theme{
		Name:       "dark",
		IsDark:     true,
		Background: lipgloss.Color("#1a1b26"),
		Foreground: lipgloss.Color("#c0caf5"),
		Primary:    lipgloss.Color("#7aa2f7"),
		Secondary:  lipgloss.Color("#bb9af7"),
		Accent:     lipgloss.Color("#7dcfff"),
		Success:    lipgloss.Color("#9ece6a"),
		Warning:    lipgloss.Color("#e0af68"),
		Error:      lipgloss.Color("#f7768e"),
		Muted:      lipgloss.Color("#565f89"),
		Border:     lipgloss.Color("#3b4261"),
		Card:       lipgloss.Color("#24283b"),
	}

	LightTheme = Theme{
		Name:       "light",
		IsDark:     false,
		Background: lipgloss.Color("#ffffff"),
		Foreground: lipgloss.Color("#343b58"),
		Primary:    lipgloss.Color("#34548a"),
		Secondary:  lipgloss.Color("#6183bb"),
		Accent:     lipgloss.Color("#0f4b6e"),
		Success:    lipgloss.Color("#3d6756"),
		Warning:    lipgloss.Color("#8f6200"),
		Error:      lipgloss.Color("#8c4351"),
		Muted:      lipgloss.Color("#9699a3"),
		Border:     lipgloss.Color("#d4d4d4"),
		Card:       lipgloss.Color("#f2f2f2"),
	}
)

func DetectTheme() Theme {
	colorterm := os.Getenv("COLORTERM")
	term := os.Getenv("TERM")
	background := os.Getenv("COLORSCHEME")

	if background == "dark" || background == "Dark" {
		return DarkTheme
	}
	if background == "light" || background == "Light" {
		return LightTheme
	}

	if colorterm == "truecolor" || colorterm == "24bit" {
		if term == "xterm-256color" || term == "screen-256color" {
			return DarkTheme
		}
	}

	if term == "alacritty" || term == "kitty" || term == "wezterm" {
		return DarkTheme
	}

	return DarkTheme
}

func (t Theme) ColorSuccess(s string) string {
	return lipgloss.NewStyle().Foreground(t.Success).Render(s)
}

func (t Theme) ColorWarning(s string) string {
	return lipgloss.NewStyle().Foreground(t.Warning).Render(s)
}

func (t Theme) ColorError(s string) string {
	return lipgloss.NewStyle().Foreground(t.Error).Render(s)
}

func (t Theme) ColorPrimary(s string) string {
	return lipgloss.NewStyle().Foreground(t.Primary).Render(s)
}

func (t Theme) ColorMuted(s string) string {
	return lipgloss.NewStyle().Foreground(t.Muted).Render(s)
}

func (t Theme) ColorAccent(s string) string {
	return lipgloss.NewStyle().Foreground(t.Accent).Render(s)
}
