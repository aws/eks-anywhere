package hardware

// TeeWriter implements MachineWriter. It writes Machine instances to multiple writers similar to the tee
// unix tool writes to multiple output streams.
type TeeWriter []MachineWriter

// Attach a new MachineWriter instance to t.
func (t *TeeWriter) Attach(w MachineWriter) {
	*t = append(*t, w)
}

// Write m to all MachineWriters attached to t with t.Attach(...). If a MachineWriter returns an error
// Write immediately returns the error without attempting to write to any other writers.
func (t *TeeWriter) Write(m Machine) error {
	for _, writer := range *t {
		if err := writer.Write(m); err != nil {
			return err
		}
	}

	return nil
}

// NewTeeWriterWith creates a new *TeeWriter instance with all writers attached.
func NewTeeWriterWith(writers ...MachineWriter) *TeeWriter {
	var tee TeeWriter
	for _, w := range writers {
		tee.Attach(w)
	}
	return &tee
}
