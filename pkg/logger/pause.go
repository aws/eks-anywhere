package logger

import (
	"bytes"
	"io"
	"sync"
)

// PausableWriter facilitates pausing output from the underlying writer and resuming it later.
// It can be used with Init to control console output from the logger.
type PausableWriter struct {
	out io.Writer
	mtx sync.Mutex
}

func (w *PausableWriter) Write(p []byte) (int, error) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return w.out.Write(p)
}

// Pause will pause output to the underlying io.Writer and return a function that resumes output.
// When resumed, all data received during the pause will be written to the underlying io.Writer.
func (w *PausableWriter) Pause() func() error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	original := w.out

	var buffer bytes.Buffer
	w.out = &buffer

	return func() error {
		w.mtx.Lock()
		defer w.mtx.Unlock()

		// Always restore the original writer so we don't continually add to the buffer. This will
		// ensure that even if we fail to copy buffered content to the original writer, we won't
		// continue to buffer.
		defer func() { w.out = original }()

		// Copy the contents of the buffer to the destination before we reset the writer.
		_, err := io.Copy(original, &buffer)
		return err
	}
}

// NewPausableWriter creates a new PausableWriter that will write to w.
func NewPausableWriter(w io.Writer) *PausableWriter {
	return &PausableWriter{out: w}
}
