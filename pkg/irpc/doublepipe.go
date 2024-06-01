package irpc

import (
	"errors"
	"io"
)

// mostly for testing
// returns two io.ReadWriteClosers that are interconnected. data written on one end results in data read on the other one
func NewDoubleEndedPipe() (endA, endB *PipeEnd) {
	ra, wa := io.Pipe()
	rb, wb := io.Pipe()

	// cross the readers/writers
	return &PipeEnd{PipeReader: ra, PipeWriter: wb}, &PipeEnd{PipeReader: rb, PipeWriter: wa}
}

// one end of pipe
// implements io.ReadWriteCloser
type PipeEnd struct {
	*io.PipeReader
	*io.PipeWriter
}

// closes the underlying pipes
// any further read/write results in error
func (p *PipeEnd) Close() error {
	return errors.Join(
		p.PipeReader.Close(),
		p.PipeWriter.Close(),
	)
}
