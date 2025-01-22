package display

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/notnil/chess"
)

type model struct {
    game *chess.Game
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
    return lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Render(m.game.Position().Board().Draw())
}

func Display(game *chess.Game) {
    p := tea.NewProgram(model{game: game})
    if _, err := p.Run(); err != nil {
        panic(err)
    }
}
