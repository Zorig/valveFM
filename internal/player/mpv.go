package player

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type Player struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	backend string
	path    string
	lastURL string
}

func New() (*Player, error) {
	if path, backend := findBundledPlayer(); path != "" {
		return &Player{backend: backend, path: path}, nil
	}
	if path, backend := findDownloadedPlayer(); path != "" {
		return &Player{backend: backend, path: path}, nil
	}
	if path, err := exec.LookPath("mpv"); err == nil {
		return &Player{backend: "mpv", path: path}, nil
	}
	if path, err := exec.LookPath("ffplay"); err == nil {
		return &Player{backend: "ffplay", path: path}, nil
	}
	return nil, errors.New("mpv or ffplay not found (bundle one or add to PATH)")
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
		cmd = exec.Command(p.path, "--no-video", "--quiet", url)
	case "ffplay":
		cmd = exec.Command(p.path, "-nodisp", "-autoexit", "-loglevel", "quiet", url)
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

func findBundledPlayer() (string, string) {
	exe, err := os.Executable()
	if err != nil {
		return "", ""
	}
	dir := filepath.Dir(exe)

	candidates := []struct {
		backend string
		name    string
	}{
		{backend: "mpv", name: "mpv"},
		{backend: "mpv", name: "mpv.exe"},
		{backend: "ffplay", name: "ffplay"},
		{backend: "ffplay", name: "ffplay.exe"},
	}

	for _, candidate := range candidates {
		path := filepath.Join(dir, candidate.name)
		if isExecutable(path) {
			return path, candidate.backend
		}
	}

	return "", ""
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	if strings.HasSuffix(strings.ToLower(path), ".exe") {
		return true
	}
	return info.Mode()&0o111 != 0
}
