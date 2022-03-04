package executables_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/executables/mocks"
)

const (
	tinkerbellCertUrl       = "http://1.2.3.4:42114/cert"
	tinkerbellGrpcAuthority = "1.2.3.4:42113"
	hardwareJson            = `
{
	"test": "123"
}	
`
)

var envMap = map[string]string{
	executables.TinkerbellCertUrlKey:       tinkerbellCertUrl,
	executables.TinkerbellGrpcAuthorityKey: tinkerbellGrpcAuthority,
}

func newTink(t *testing.T) (*executables.Tink, context.Context, *mocks.MockExecutable) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	executable := mocks.NewMockExecutable(ctrl)

	return executables.NewTink(executable, tinkerbellCertUrl, tinkerbellGrpcAuthority), ctx, executable
}

func TestTinkPushHardwareSuccess(t *testing.T) {
	tink, ctx, e := newTink(t)
	hardwareBytes := []byte(hardwareJson)
	expectedParam := []string{"hardware", "push"}
	expectCommand(e, ctx, expectedParam...).withStdIn(hardwareBytes).withEnvVars(envMap).to().Return(bytes.Buffer{}, nil)
	if err := tink.PushHardware(ctx, hardwareBytes); err != nil {
		t.Errorf("Tink.PushHardware() error = %v, want nil", err)
	}
}

func TestTinkGetHardware(t *testing.T) {
	tink, ctx, e := newTink(t)
	expectedParam := []string{"hardware", "get", "--format", "json"}
	expectCommand(e, ctx, expectedParam...).withEnvVars(envMap).to().Return(bytes.Buffer{}, nil)
	if _, err := tink.GetHardware(ctx); err != nil {
		t.Errorf("Tink.GetHardware() error = %v, want nil", err)
	}
}

func TestTinkWorkflow(t *testing.T) {
	tink, ctx, e := newTink(t)
	expectedParam := []string{"workflow", "get", "--format", "json"}
	expectCommand(e, ctx, expectedParam...).withEnvVars(envMap).to().Return(bytes.Buffer{}, nil)
	if _, err := tink.GetWorkflow(ctx); err != nil {
		t.Errorf("Tink.GetWorkflow() error = %v, want nil", err)
	}
}
