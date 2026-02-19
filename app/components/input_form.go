package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/localpulse/localpulse/ui"
)

type InputFormState int

const (
	InputStateIdle InputFormState = iota
	InputStateActive
	InputStateSubmitted
	InputStateCancelled
)

type InputForm struct {
	textInput textinput.Model
	state     InputFormState
	width     int
	styles    *ui.Styles
}

func NewInputForm(styles *ui.Styles) *InputForm {
	ti := textinput.New()
	ti.Placeholder = "http://localhost:8080/api"
	ti.CharLimit = 256
	ti.Width = 40

	return &InputForm{
		textInput: ti,
		state:     InputStateIdle,
		styles:    styles,
	}
}

func (f *InputForm) Focus() {
	f.state = InputStateActive
	f.textInput.Focus()
}

func (f *InputForm) Blur() {
	f.state = InputStateIdle
	f.textInput.Blur()
}

func (f *InputForm) IsActive() bool {
	return f.state == InputStateActive
}

func (f *InputForm) SetWidth(width int) {
	f.width = width
	f.textInput.Width = width - 10
	if f.textInput.Width < 20 {
		f.textInput.Width = 20
	}
}

func (f *InputForm) Value() string {
	return f.textInput.Value()
}

func (f *InputForm) Reset() {
	f.textInput.SetValue("")
	f.state = InputStateIdle
}

func (f *InputForm) Init() tea.Cmd {
	return textinput.Blink
}

func (f *InputForm) Update(msg tea.Msg) (*InputForm, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if f.state == InputStateActive && f.textInput.Value() != "" {
				f.state = InputStateSubmitted
				return f, nil
			}
		case "esc":
			f.state = InputStateCancelled
			return f, nil
		}
	}

	f.textInput, cmd = f.textInput.Update(msg)
	return f, cmd
}

func (f *InputForm) View() string {
	if f.state == InputStateIdle {
		return ""
	}

	title := f.styles.CardTitle.Render("Add Endpoint")
	input := f.textInput.View()

	hint := f.styles.Theme.ColorMuted("Enter URL • Esc to cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		input,
		"",
		hint,
	)

	return f.styles.Panel.Width(f.width).Render(content)
}

func (f *InputForm) IsSubmitted() bool {
	return f.state == InputStateSubmitted
}

func (f *InputForm) IsCancelled() bool {
	return f.state == InputStateCancelled
}

func (f *InputForm) Acknowledge() {
	if f.state == InputStateSubmitted || f.state == InputStateCancelled {
		f.Reset()
	}
}
