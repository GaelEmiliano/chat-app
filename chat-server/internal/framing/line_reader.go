package framing

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

// ErrFrameTooLarge is returned when a single frame exceeds the configured limit.
var ErrFrameTooLarge = errors.New("frame exceeds maximum allowed size")

// LineReader reads newline-delimited frames from an io.Reader.
// A frame is defined as a sequence of bytes terminated by the '\n' character.
// The delimiter is not included in the returned frame.
type LineReader struct {
	scanner       *bufio.Scanner
	maxFrameBytes int
}

// NewLineReader creates a LineReader with a strict maximum frame size.
// The limit applies to the frame payload only (excluding the '\n' delimiter).
func NewLineReader(reader io.Reader, maxFrameBytes int) *LineReader {
	scanner := bufio.NewScanner(reader)

	// bufio.Scanner has a small default buffer; we must raise it explicitly.
	// We also cap it to maxFrameBytes to avoid unbounded memory usage.
	initialBuffer := make([]byte, 0, min(maxFrameBytes, 64*1024))
	scanner.Buffer(initialBuffer, maxFrameBytes)

	return &LineReader{
		scanner:       scanner,
		maxFrameBytes: maxFrameBytes,
	}
}

// ReadFrame blocks until a full frame is read, the connection is closed,
// or an error occurs.
//
// Possible errors:
//   - io.EOF: the underlying reader was closed cleanly
//   - ErrFrameTooLarge: a frame exceeded the configured maximum size
//   - any other error reported by the underlying reader
func (lr *LineReader) ReadFrame() ([]byte, error) {
	if lr.scanner.Scan() {
		frame := lr.scanner.Bytes()
		// Copy the bytes because Scanner reuses its buffer.
		copied := make([]byte, len(frame))
		copy(copied, frame)
		return copied, nil
	}

	if err := lr.scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return nil, fmt.Errorf("%w (max=%d bytes)", ErrFrameTooLarge, lr.maxFrameBytes)
		}
		return nil, err
	}

	return nil, io.EOF
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
