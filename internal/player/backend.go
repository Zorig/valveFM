package player

import (
	"errors"
	"fmt"
	"sync"
)

// Backend is the common interface for all audio player backends.
type Backend interface {
	Play(url string) error
	Stop() error
	IsPlaying() bool
	LastURL() string
}

// CompositeBackend wraps multiple backends and selects the best one dynamically.
type CompositeBackend struct {
	mu      sync.Mutex
	gp      *GoPlayer
	ext     *Player
	active  Backend
	lastURL string
}

func (c *CompositeBackend) Play(url string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastURL = url

	// Stop any currently playing backend
	if c.active != nil {
		c.active.Stop()
	}

	var errGo error
	// 1. Try pure Go backend
	if c.gp != nil {
		if err := c.gp.Play(url); err == nil {
			c.active = c.gp
			return nil
		} else {
			errGo = err
		}
	}

	// 2. Fallback to external player (if available)
	// Useful for AAC/OGG streams that go-mp3 cannot decode
	if c.ext != nil {
		// If we're falling back from a Go error, we might want to ensure
		// the previous backend is fully stopped/cleaned up, which Stop() does.
		if err := c.ext.Play(url); err == nil {
			c.active = c.ext
			return nil
		} else {
			if errGo != nil {
				return fmt.Errorf("go-audio: %v, external: %v", errGo, err)
			}
			return err
		}
	}

	if errGo != nil {
		return fmt.Errorf("format not supported in pure Go; please install mpv or ffplay to listen (error: %v)", errGo)
	}
	return errors.New("no audio backend available; please install mpv or ffplay")
}

func (c *CompositeBackend) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active != nil {
		return c.active.Stop()
	}
	return nil
}

func (c *CompositeBackend) IsPlaying() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.active != nil && c.active.IsPlaying()
}

func (c *CompositeBackend) LastURL() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lastURL
}

// New returns a smart player that tries pure Go audio first,
// but falls back to system mpv/ffplay for unsupported formats (like AAC).
func New() (Backend, error) {
	gp := probeGoAudio()
	ext, _ := newExternal() // optional fallback

	if gp == nil && ext == nil {
		return nil, errors.New("no player backend available")
	}

	return &CompositeBackend{
		gp:  gp,
		ext: ext,
	}, nil
}
