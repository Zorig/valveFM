package player

import (
	"bytes"
	"io"
	"strconv"
	"strings"
)

// icyReader wraps an HTTP radio stream body, strips ICY metadata blocks inline,
// and surfaces track titles via a callback. It implements io.ReadCloser so it
// can be passed directly to mp3.Decode.
type icyReader struct {
	rc        io.ReadCloser
	metaInt   int
	remaining int
	onTitle   func(string)
}

func newICYReader(rc io.ReadCloser, metaInt int, onTitle func(string)) *icyReader {
	return &icyReader{
		rc:        rc,
		metaInt:   metaInt,
		remaining: metaInt,
		onTitle:   onTitle,
	}
}

// Read returns only audio bytes; metadata blocks are stripped transparently.
func (r *icyReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	// Never read past the next metadata boundary in one call.
	toRead := len(p)
	if toRead > r.remaining {
		toRead = r.remaining
	}

	n, err := r.rc.Read(p[:toRead])
	r.remaining -= n

	if r.remaining == 0 {
		if metaErr := r.readMeta(); metaErr != nil && err == nil {
			err = metaErr
		}
		r.remaining = r.metaInt
	}

	return n, err
}

func (r *icyReader) Close() error {
	return r.rc.Close()
}

// readMeta reads one ICY metadata block from the stream.
// Format: 1 length byte (value * 16 = block size), then the block.
func (r *icyReader) readMeta() error {
	var lenBuf [1]byte
	if _, err := io.ReadFull(r.rc, lenBuf[:]); err != nil {
		return err
	}
	size := int(lenBuf[0]) * 16
	if size == 0 {
		return nil
	}
	meta := make([]byte, size)
	if _, err := io.ReadFull(r.rc, meta); err != nil {
		return err
	}
	if r.onTitle != nil {
		if title := parseICYTitle(meta); title != "" {
			r.onTitle(title)
		}
	}
	return nil
}

// parseICYTitle extracts the StreamTitle value from a raw ICY metadata block.
// Block format: StreamTitle='Artist - Song';StreamUrl='...';
func parseICYTitle(data []byte) string {
	s := string(bytes.TrimRight(data, "\x00"))
	const prefix = "StreamTitle='"
	idx := strings.Index(s, prefix)
	if idx < 0 {
		return ""
	}
	s = s[idx+len(prefix):]
	end := strings.Index(s, "';")
	if end < 0 {
		end = strings.LastIndex(s, "'")
	}
	if end <= 0 {
		return ""
	}
	return strings.TrimSpace(s[:end])
}

// parseICYMetaInt parses the icy-metaint header value into an int.
func parseICYMetaInt(value string) int {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n <= 0 {
		return 0
	}
	return n
}
