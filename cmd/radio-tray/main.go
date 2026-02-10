package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/getlantern/systray"

	"radio-tui/internal/config"
	"radio-tui/internal/ipc"
	"radio-tui/internal/player"
	"radio-tui/internal/radio"
	"radio-tui/internal/ui"
)

const (
	cmdPlayPause = "PLAY_PAUSE"
	cmdNext      = "NEXT"
	cmdPrev      = "PREV"
	cmdQuit      = "QUIT"
	cmdStatus    = "STATUS"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("Valve FM")
	systray.SetTooltip("Valve FM")

	mPlayPause := systray.AddMenuItem("Play/Pause", "Toggle playback")
	mNext := systray.AddMenuItem("Next", "Next station")
	mPrev := systray.AddMenuItem("Previous", "Previous station")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit Valve FM")

	go func() {
		if err := runTUI(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		systray.Quit()
	}()

	go func() {
		for range mPlayPause.ClickedCh {
			_, _ = sendCommand(cmdPlayPause)
		}
	}()
	go func() {
		for range mNext.ClickedCh {
			_, _ = sendCommand(cmdNext)
		}
	}()
	go func() {
		for range mPrev.ClickedCh {
			_, _ = sendCommand(cmdPrev)
		}
	}()
	go func() {
		for range mQuit.ClickedCh {
			_, _ = sendCommand(cmdQuit)
			systray.Quit()
		}
	}()

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			status, err := sendCommand(cmdStatus)
			if err != nil {
				systray.SetTooltip("Valve FM (disconnected)")
				mPlayPause.Disable()
				mNext.Disable()
				mPrev.Disable()
				mQuit.Disable()
				continue
			}
			systray.SetTooltip("Valve FM " + status)
			mPlayPause.Enable()
			mNext.Enable()
			mPrev.Enable()
			mQuit.Enable()
		}
	}()
}

func onExit() {
	_, _ = sendCommand(cmdQuit)
}

func runTUI() error {
	api, err := radio.NewClient("ValveFM/1.0 (terminal radio)")
	if err != nil {
		return err
	}

	playerInstance, playerErr := player.New()
	favorites, favErr := config.LoadFavorites()

	model := ui.NewModel(api, playerInstance, favorites, playerErr, favErr)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func sendCommand(command string) (string, error) {
	ep, err := ipc.ResolveEndpoint()
	if err != nil {
		return "", err
	}

	conn, err := net.DialTimeout(ep.Network, ep.Address, 500*time.Millisecond)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_, err = fmt.Fprintln(conn, command)
	if err != nil {
		return "", err
	}

	_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "ERR ") {
		return "", errors.New(strings.TrimPrefix(line, "ERR "))
	}
	if line == "OK" {
		return "", nil
	}
	return line, nil
}
