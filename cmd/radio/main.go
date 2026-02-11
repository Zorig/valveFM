package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"radio-tui/internal/config"
	"radio-tui/internal/player"
	"radio-tui/internal/radio"
	"radio-tui/internal/ui"
)

func main() {
	api, err := radio.NewClient("ValveFM/1.0 (terminal radio)")
	if err != nil {
		fmt.Fprintln(os.Stderr, "radio api error:", err)
		os.Exit(1)
	}

	playerInstance, playerErr := player.New()
	favorites, favErr := config.LoadFavorites()
	cfg := config.LoadConfig()

	model := ui.NewModel(api, playerInstance, favorites, playerErr, favErr, cfg.Theme)
	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
