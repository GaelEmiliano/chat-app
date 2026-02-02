package framing

import (
	"bufio"
	"context"
	"fmt"
	"io"
)

// LineWriter writes newline-delimited frames to an io.Writer.
// Each frame is written as: <payload>\n
type LineWriter struct {
	writer *bufio.Writer
}

// NewLineWriter creates a LineWriter that buffers writes internally.
// The caller is responsible for concurrency control at a higher level;
// LineWriter itself is not safe for concurrent use.
func NewLineWriter(writer io.Writer) *LineWriter {
	return &LineWriter{
		writer: bufio.NewWriter(writer),
	}
}

// WriteFrame writes a single frame followed by a newline delimiter.
// It respects context cancellation before attempting the write.
func (lw *LineWriter) WriteFrame(ctx context.Context, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if _, err := lw.writer.Write(payload); err != nil {
		return fmt.Errorf("write payload: %w", err)
	}
	if err := lw.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("write delimiter: %w", err)
	}
	if err := lw.writer.Flush(); err != nil {
		return fmt.Errorf("flush writer: %w", err)
	}

	return nil
}
