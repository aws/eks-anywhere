package executables_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

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
	expectedParam := []string{"hardware", "get", "--tinkerbell-cert-url", tinkerbellCertUrl, "--tinkerbell-grpc-authority", tinkerbellGrpcAuthority, "--format", "json"}
	expectCommand(e, ctx, expectedParam...).to().Return(bytes.Buffer{}, nil)
	if _, err := tink.GetHardware(ctx); err != nil {
		t.Errorf("Tink.GetHardware() error = %v, want nil", err)
	}
}

func TestTinkGetWorkflow(t *testing.T) {
	tink, ctx, e := newTink(t)
	expectedParam := []string{"workflow", "get", "--tinkerbell-cert-url", tinkerbellCertUrl, "--tinkerbell-grpc-authority", tinkerbellGrpcAuthority, "--format", "json"}
	expectCommand(e, ctx, expectedParam...).to().Return(bytes.Buffer{}, nil)
	if _, err := tink.GetWorkflow(ctx); err != nil {
		t.Errorf("Tink.GetWorkflow() error = %v, want nil", err)
	}
}

func TestTinkDeleteWorkflowSingle(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	mockExecutable := mocks.NewMockExecutable(ctrl)

	tink := executables.NewTink(mockExecutable, tinkerbellCertUrl, tinkerbellGrpcAuthority)

	ctx := context.Background()
	workflowID := "abc"
	expectedParam := []string{
		"workflow", "delete",
		"--tinkerbell-cert-url", tinkerbellCertUrl,
		"--tinkerbell-grpc-authority", tinkerbellGrpcAuthority,
		workflowID,
	}

	expectCommand(mockExecutable, ctx, expectedParam...).to().Return(bytes.Buffer{}, nil)

	err := tink.DeleteWorkflow(ctx, workflowID)
	g.Expect(err).ToNot(gomega.HaveOccurred())
}

func TestTinkDeleteWorkflowMulti(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	mockExecutable := mocks.NewMockExecutable(ctrl)

	tink := executables.NewTink(mockExecutable, tinkerbellCertUrl, tinkerbellGrpcAuthority)

	ctx := context.Background()
	workflowIDs := []string{"abc", "def", "ghi"}
	expectedParam := []string{
		"workflow", "delete",
		"--tinkerbell-cert-url", tinkerbellCertUrl,
		"--tinkerbell-grpc-authority", tinkerbellGrpcAuthority,
	}
	expectedParam = append(expectedParam, workflowIDs...)

	expectCommand(mockExecutable, ctx, expectedParam...).to().Return(bytes.Buffer{}, nil)

	err := tink.DeleteWorkflow(ctx, workflowIDs...)
	g.Expect(err).ToNot(gomega.HaveOccurred())
}

func TestTinkDeleteWorkflowCmdError(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	mockExecutable := mocks.NewMockExecutable(ctrl)

	tink := executables.NewTink(mockExecutable, tinkerbellCertUrl, tinkerbellGrpcAuthority)

	ctx := context.Background()
	workflowIDs := []string{"abc", "def", "ghi"}
	expectedParam := []string{
		"workflow", "delete",
		"--tinkerbell-cert-url", tinkerbellCertUrl,
		"--tinkerbell-grpc-authority", tinkerbellGrpcAuthority,
	}
	expectedParam = append(expectedParam, workflowIDs...)

	expect := errors.New("hello world")
	expectCommand(mockExecutable, ctx, expectedParam...).to().Return(bytes.Buffer{}, expect)

	err := tink.DeleteWorkflow(ctx, workflowIDs...)
	g.Expect(err).To(gomega.HaveOccurred())
}
