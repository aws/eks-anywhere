package executables

import "context"

type localExecutableBuilder struct{}

func newLocalExecutableBuilder() localExecutableBuilder {
	return localExecutableBuilder{}
}

func (b localExecutableBuilder) Build(binaryPath string) Executable {
	return NewExecutable(binaryPath)
}

func (b localExecutableBuilder) Init(_ context.Context) (Closer, error) {
	return NoOpClose, nil
}

func NoOpClose(ctx context.Context) error {
	return nil
}
