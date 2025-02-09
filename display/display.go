package display

import (
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/notnil/chess"

	"fmt"
	"strings"
)

type GameMetadata struct {
	White string
	Black string
}

type analysisTab struct {
	white string
	black string
	eval  string
	board string
	evals []string
}

func (t analysisTab) printAnalysisTab() string {
	return fmt.Sprintf("%-30s", "White: "+t.white) + fmt.Sprintf("%30s", "Black: "+t.black) + t.eval + t.board
}

type model struct {
	game        *chess.Game
	tabs        []string
	tabContent  []string
	analysisTab analysisTab
	activeTab   int
	move        int
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
		case "l", "n", "tab":
			m.activeTab = min(m.activeTab+1, len(m.tabs)-1)
			return m, nil
		case "h", "p", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		case "left":
			switch m.activeTab {
			case 0:
				if m.move > 0 {
					m.move--
					m.analysisTab.board = m.game.Positions()[m.move].Board().Draw()
					m.analysisTab.eval = m.analysisTab.evals[m.move]
					m.tabContent[0] = m.analysisTab.printAnalysisTab()
				}
			}
			return m, nil
		case "right":
			switch m.activeTab {
			case 0:
				if m.move < len(m.game.Positions())-1 {
					m.move++
					if m.move == len(m.game.Positions())-1 {
						m.analysisTab.eval = m.game.Outcome().String()
					} else {
						m.analysisTab.eval = m.analysisTab.evals[m.move]
					}
					m.analysisTab.board = m.game.Positions()[m.move].Board().Draw()
					m.tabContent[0] = m.analysisTab.printAnalysisTab()
				}
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.tabs)-1, i == m.activeTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(m.tabContent[m.activeTab]))
	return docStyle.Render(doc.String())
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

func Display(game *chess.Game, md GameMetadata) {
	tabs := []string{"Analysis", "Game Selector"}

	evals := GetGameEvaluations(game)
	at := analysisTab{
		white: md.White,
		black: md.Black,
		eval:  evals[len(game.Positions())-1],
		board: game.Position().Board().Draw(),
		evals: evals,
	}
	tabContent := []string{
		at.printAnalysisTab(),
		"Game Selector Tab",
	}
	move := len(game.Positions()) - 1
	p := tea.NewProgram(model{game: game, tabs: tabs, tabContent: tabContent, analysisTab: at, move: move})
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GetGameEvaluations(game *chess.Game) []string {
	// Initialize stockfish
	stockfish := exec.Command("stockfish")
	stdin, err := stockfish.StdinPipe()
	if err != nil {
		panic(fmt.Errorf("error loading stockfish stdin: %w", err))
	}
	stdout, err := stockfish.StdoutPipe()
	if err != nil {
		panic(fmt.Errorf("error loading stockfish stdout: %w", err))
	}
	if err := stockfish.Start(); err != nil {
		panic(fmt.Errorf("error starting stockfish: %w", err))
	}
	defer stockfish.Wait()
	defer stockfish.Process.Kill()

	// Helper functions
	sendCommand := func(cmd string) {
		if _, err := stdin.Write([]byte(cmd + "\n")); err != nil {
			panic(fmt.Errorf("failed to send command to Stockfish: %w", err))
		}
	}

	readStockfishOutput := func() string {
		buf := make([]byte, 3072)
		n, err := stdout.Read(buf)
		if err != nil {
			panic(fmt.Errorf("failed to read Stockfish output: %w", err))
		}
		return string(buf[:n])
	}

	sendCommand("setoption name Threads value 4")

	// Get evaluations for each position
	positions := game.Positions()
	evals := make([]string, len(positions))

	for i, pos := range positions {
		sendCommand("position fen " + pos.String())
		sendCommand("eval")

		// Wait for and parse evaluation
		time.Sleep(100 * time.Millisecond)
		output := readStockfishOutput()

		// Parse the final evaluation line
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Final evaluation") {
				// Extract the evaluation value
				if parts := strings.Split(line, " "); len(parts) > 2 {
					evals[i] = parts[2]
				}
				break
			}
		}
	}

	fmt.Println(len(evals), len(positions))
	return evals
}
