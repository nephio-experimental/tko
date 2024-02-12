package tui

// A simple program that opens the alternate screen buffer then counts down
// from 5 and then exits.

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	ticks int
}

type tickMsg time.Time

func Start() {
	p := tea.NewProgram(Model{ticks: 5}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (self Model) Init() tea.Cmd {
	return tick()
}

func (self Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return self, tea.Quit
		}

	case tickMsg:
		self.ticks--
		if self.ticks <= 0 {
			return self, tea.Quit
		}
		return self, tick()
	}

	return self, nil
}

func (self Model) View() string {
	return fmt.Sprintf("\n\n     Hi. This program will exit in %d seconds...", self)
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
