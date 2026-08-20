package player

import (
	"bytes"
	"io"
	"math/rand"
	"testing"
)

// TestDecodeMP3_NoPanic ensures decodeMP3 never panics on malformed input:
// go-mp3 historically panics on certain broken MPEG2 frames, and we convert
// that into a normal error so the composite backend can fall back.
func TestDecodeMP3_NoPanic(t *testing.T) {
	inputs := [][]byte{
		nil,
		[]byte("not an mp3 at all"),
		{0xff, 0xfb, 0x90, 0x64, 0x00, 0x00}, // truncated MPEG2 frame header
		{0xff, 0xf3, 0x94, 0x64, 0x00, 0x00}, // MPEG2 layer III, broken scale factors
	}

	seeded := rand.New(rand.NewSource(42))
	for i := 0; i < 100; i++ {
		buf := make([]byte, seeded.Intn(512))
		_, _ = seeded.Read(buf)
		inputs = append(inputs, buf)
	}

	for i, in := range inputs {
		_, _, err := decodeMP3(io.NopCloser(bytes.NewReader(in)))
		// Regardless of whether go-mp3 returns an error or panics internally,
		// we must not panic here — that is the regression this guards against.
		_ = err
		_ = i
	}
}
