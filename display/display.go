package display

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type model struct {
    text string
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m model) View() string {
    return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(m.text)
}

func display() {
    p := tea.NewProgram(model{text: "Hello, TUI!"})
    if _, err := p.Run(); err != nil {
        panic(err)
    }
}
