package player

import (
	"errors"
	"io"
	"os/exec"
	"sync"
)

type Player struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	backend string
	lastURL string
}

func New() (*Player, error) {
	if _, err := exec.LookPath("mpv"); err == nil {
		return &Player{backend: "mpv"}, nil
	}
	if _, err := exec.LookPath("ffplay"); err == nil {
		return &Player{backend: "ffplay"}, nil
	}
	return nil, errors.New("mpv or ffplay not found in PATH")
}

func (p *Player) Play(url string) error {
	if url == "" {
		return errors.New("stream url is required")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	_ = p.stopLocked()
	p.lastURL = url

	var cmd *exec.Cmd
	switch p.backend {
	case "mpv":
		cmd = exec.Command("mpv", "--no-video", "--quiet", url)
	case "ffplay":
		cmd = exec.Command("ffplay", "-nodisp", "-autoexit", "-loglevel", "quiet", url)
	default:
		return errors.New("no audio backend available")
	}

	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return err
	}

	p.cmd = cmd
	go func(local *exec.Cmd) {
		_ = local.Wait()
		p.mu.Lock()
		if p.cmd == local {
			p.cmd = nil
		}
		p.mu.Unlock()
	}(cmd)

	return nil
}

func (p *Player) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stopLocked()
}

func (p *Player) stopLocked() error {
	if p.cmd == nil {
		return nil
	}
	if p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
	p.cmd = nil
	return nil
}

func (p *Player) IsPlaying() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cmd != nil
}

func (p *Player) LastURL() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastURL
}
