package player

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
)

// GoPlayer plays MP3 HTTP streams using the high-level beep library.
// It handles resampling automatically, fixing pitch issues with different sample rates.
type GoPlayer struct {
	mu          sync.Mutex
	streamer    beep.StreamSeekCloser
	ctrl        *beep.Ctrl
	resp        *http.Response
	lastURL     string
	playing     bool
	initialized bool
}

// NewGoPlayer creates a GoPlayer instance.
func NewGoPlayer() *GoPlayer {
	return &GoPlayer{}
}

// initSpeaker initializes the audio device once.
// We standardize on 44100Hz for all playback.
func (g *GoPlayer) initSpeaker() error {
	if g.initialized {
		return nil
	}
	// 44100Hz sample rate, buffer size of ~100ms
	sr := beep.SampleRate(44100)
	if err := speaker.Init(sr, sr.N(time.Second/10)); err != nil {
		return fmt.Errorf("speaker init: %w", err)
	}
	g.initialized = true
	return nil
}

// Play opens an HTTP stream and starts playback.
func (g *GoPlayer) Play(url string) error {
	if url == "" {
		return errors.New("stream url is required")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Stop previous
	g.stopLocked()
	g.lastURL = url

	// Initialize speaker if needed (lazy)
	if err := g.initSpeaker(); err != nil {
		return err
	}

	// Request stream
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	req.Header.Set("User-Agent", "ValveFM/1.0")
	req.Header.Set("Icy-MetaData", "0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("stream open: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("stream HTTP %d", resp.StatusCode)
	}

	// Decode MP3 via beep (wraps go-mp3)
	streamer, format, err := mp3.Decode(resp.Body)
	if err != nil {
		resp.Body.Close()
		return fmt.Errorf("mp3 decode: %w", err)
	}

	// Resample to our standard 44100Hz rate
	resampled := beep.Resample(4, format.SampleRate, beep.SampleRate(44100), streamer)

	// Wrap in a Ctrl to allow pausing/stopping nicely
	ctrl := &beep.Ctrl{Streamer: resampled, Paused: false}

	// Play!
	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		// Callback when stream ends
		if g.ctrl == ctrl {
			g.playing = false
			g.cleanupLocked()
		}
	})))

	g.streamer = streamer
	g.ctrl = ctrl
	g.resp = resp
	g.playing = true

	return nil
}

// Stop halts playback immediately.
func (g *GoPlayer) Stop() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.stopLocked()
	return nil
}

func (g *GoPlayer) stopLocked() {
	// Stop existing playback by pausing the controller (which removes it from mixer eventually)
	// and closing the streamer/response.
	if g.ctrl != nil {
		g.ctrl.Paused = true
		g.ctrl = nil
	}
	g.cleanupLocked()
}

func (g *GoPlayer) cleanupLocked() {
	if g.streamer != nil {
		g.streamer.Close()
		g.streamer = nil
	}
	// Response body is closed by streamer.Close() usually, but be safe
	if g.resp != nil {
		g.resp.Body.Close()
		g.resp = nil
	}
	g.playing = false
}

func (g *GoPlayer) IsPlaying() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.playing
}

func (g *GoPlayer) LastURL() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.lastURL
}

func probeGoAudio() *GoPlayer {
	return NewGoPlayer()
}

var _ Backend = (*GoPlayer)(nil)
