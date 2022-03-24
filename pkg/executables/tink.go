package executables

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tinkerbell/tink/protos/hardware"
	"github.com/tinkerbell/tink/protos/workflow"
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
		return fmt.Errorf("pushing hardware: %v", err)
	}
	return nil
}

func (t *Tink) GetHardware(ctx context.Context) ([]*hardware.Hardware, error) {
	params := []string{"hardware", "get", "--tinkerbell-cert-url", t.tinkerbellCertUrl, "--tinkerbell-grpc-authority", t.tinkerbellGrpcAuthority, "--format", "json"}
	data, err := t.Command(ctx, params...).Run()
	if err != nil {
		return nil, fmt.Errorf("getting hardware list: %v", err)
	}
	var hardwareList []*hardware.Hardware
	hardwareString := data.String()

	if len(hardwareString) > 0 {
		hardwareListData := map[string][]*hardware.Hardware{}

		if err = json.Unmarshal([]byte(hardwareString), &hardwareListData); err != nil {
			return nil, fmt.Errorf("unmarshling hardware json: %v", err)
		}
		if len(hardwareListData["data"]) > 0 {
			hardwareList = append(hardwareList, hardwareListData["data"]...)
		}
	}

	return hardwareList, nil
}

func (t *Tink) GetWorkflow(ctx context.Context) ([]*workflow.Workflow, error) {
	params := []string{"workflow", "get", "--tinkerbell-cert-url", t.tinkerbellCertUrl, "--tinkerbell-grpc-authority", t.tinkerbellGrpcAuthority, "--format", "json"}
	data, err := t.Command(ctx, params...).Run()
	if err != nil {
		return nil, fmt.Errorf("getting workflow list: %v", err)
	}
	var workflowList []*workflow.Workflow
	workflowString := data.String()

	if len(workflowString) > 0 {
		workflowListData := map[string][]*workflow.Workflow{}

		if err = json.Unmarshal([]byte(workflowString), &workflowListData); err != nil {
			return nil, fmt.Errorf("unmarshling workflow json: %v", err)
		}
		if len(workflowListData["data"]) > 0 {
			workflowList = append(workflowList, workflowListData["data"]...)
		}
	}

	return workflowList, nil
}
