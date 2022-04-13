package hardware

// multiWriter implements MachineWriter. It writes Machine instances to multiple writers similar to the tee
// unix tool writes to multiple output streams.
type multiWriter []MachineWriter

// Write m to all MachineWriters attached to t with t.Attach(...). If a MachineWriter returns an error
// Write immediately returns the error without attempting to write to any other writers.
func (t multiWriter) Write(m Machine) error {
	for _, writer := range t {
		if err := writer.Write(m); err != nil {
			return err
		}
	}

	return nil
}

// MultiMachineWriter combines writers into a single MachineWriter instance. Passing no writers effectively creates a
// noop MachineWriter.
func MultiMachineWriter(writers ...MachineWriter) MachineWriter {
	var tee multiWriter
	tee = append(tee, writers...)
	return tee
}
