package executables

import (
	"context"
	"fmt"
)

const (
	tinkPath                   = "tink"
	TinkerbellCertUrlKey       = "TINKERBELL_CERT_URL"
	TinkerbellGrpcAuthorityKey = "TINKERBELL_GRPC_AUTHORITY"
)

type Tink struct {
	Executable
	tinkerbellCertUrl       string
	tinkerbellGrpcAuthority string
	envMap                  map[string]string
}

func NewTink(executable Executable, tinkerbellCertUrl, tinkerbellGrpcAuthority string) *Tink {
	return &Tink{
		Executable:              executable,
		tinkerbellCertUrl:       tinkerbellCertUrl,
		tinkerbellGrpcAuthority: tinkerbellGrpcAuthority,
		envMap: map[string]string{
			TinkerbellCertUrlKey:       tinkerbellCertUrl,
			TinkerbellGrpcAuthorityKey: tinkerbellGrpcAuthority,
		},
	}
}

func (t *Tink) PushHardware(ctx context.Context, hardware []byte) error {
	params := []string{"hardware", "push"}
	if _, err := t.Command(ctx, params...).WithStdIn(hardware).WithEnvVars(t.envMap).Run(); err != nil {
		return fmt.Errorf("error pushing hardware: %v", err)
	}
	return nil
}

func (t *Tink) ListHardware(ctx context.Context) error {
	params := []string{"hardware", "list"}
	if _, err := t.Command(ctx, params...).WithEnvVars(t.envMap).Run(); err != nil {
		return fmt.Errorf("error getting hardware list: %v", err)
	}
	return nil
}
