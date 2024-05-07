package irpc

import (
	"fmt"
	"io"
)

// mostly for testing
// returns two io.ReadWriters that are interconnected. write on one result in read on the other one
func NewDoubleEndedPipe() (endA, endB *PipeEnd) {
	ra, wa := io.Pipe()
	rb, wb := io.Pipe()

	// we cross the readers/writers
	return &PipeEnd{PipeReader: ra, PipeWriter: wb}, &PipeEnd{PipeReader: rb, PipeWriter: wa}
}

// one end of pipe
// implements io.ReadWriter
type PipeEnd struct {
	*io.PipeReader
	*io.PipeWriter
}

func (p *PipeEnd) Close() error {
	if err := p.PipeReader.Close(); err != nil {
		return fmt.Errorf("failed to close reader: %w", err)
	}
	if err := p.PipeWriter.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}
