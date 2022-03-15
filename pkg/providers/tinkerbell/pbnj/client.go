package pbnj

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tinkerbell/cluster-api-provider-tinkerbell/pbnj/client"
	v1 "github.com/tinkerbell/pbnj/api/v1"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	PbnjGrpcAuth = "PBNJ_GRPC_AUTHORITY"
)

type PowerState string

const (
	PowerStateOn      PowerState = "on"
	PowerStateOff     PowerState = "off"
	PowerStateUnknown PowerState = "unknown"
)

const (
	machineAlreadyPoweredOffErrorMsg = "server is already off"
)

type Pbnj struct {
	pbnj *client.PbnjClient
}

type BmcSecretConfig struct {
	Host     string
	Username string
	Password string
	Vendor   string
}

// NewPBNJClient client establishes a connection with PBnJ which requires "PBNJ_GRPC_AUTHORITY" as an env variables and returns a PBnJClient instance
func NewPBNJClient(pbnjGrpcAuth string) (*Pbnj, error) {
	os.Setenv(PbnjGrpcAuth, pbnjGrpcAuth)
	defer os.Unsetenv(PbnjGrpcAuth)

	conn, _ := client.SetupConnection()

	mClient := v1.NewMachineClient(conn)
	tClient := v1.NewTaskClient(conn)
	pbnjObj := &Pbnj{client.NewPbnjClient(mClient, tClient)}

	return pbnjObj, nil
}

func (p *Pbnj) GetPowerState(ctx context.Context, bmcInfo BmcSecretConfig) (PowerState, error) {
	status, err := p.pbnj.MachinePower(ctx, NewPowerRequest(bmcInfo, v1.PowerAction_POWER_ACTION_STATUS))
	if err != nil {
		return PowerStateUnknown, err
	}

	result := strings.ToLower(status.Result)

	if strings.Contains(result, string(PowerStateOn)) {
		return PowerStateOn, nil
	} else if strings.Contains(result, string(PowerStateOff)) {
		return PowerStateOff, nil
	}

	return PowerStateUnknown, nil
}

func (p *Pbnj) PowerOff(ctx context.Context, bmcInfo BmcSecretConfig) error {
	logger.V(4).Info("Attempting to power off machine", "machine", bmcInfo.Host)
	_, err := p.pbnj.MachinePower(ctx, NewPowerRequest(bmcInfo, v1.PowerAction_POWER_ACTION_OFF))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), machineAlreadyPoweredOffErrorMsg) {
			logger.Info("WARNING: Machine is already powered off", "machine", bmcInfo.Host, "msg", err.Error())
			return nil
		}

		return fmt.Errorf("failed to power off machine: %s %v", bmcInfo.Host, err)
	}

	logger.V(4).Info("Successfully powered off machine", "machine", bmcInfo.Host)
	return nil
}

func (p *Pbnj) PowerOn(ctx context.Context, bmcInfo BmcSecretConfig) error {
	_, err := p.pbnj.MachinePower(ctx, NewPowerRequest(bmcInfo, v1.PowerAction_POWER_ACTION_ON))
	if err != nil {
		return fmt.Errorf("failed to power on machine: %s %v", bmcInfo.Host, err)
	}

	return nil
}

func NewPowerRequest(bmcInfo BmcSecretConfig, powerAction v1.PowerAction) *v1.PowerRequest {
	return &v1.PowerRequest{
		Authn: &v1.Authn{
			Authn: &v1.Authn_DirectAuthn{
				DirectAuthn: &v1.DirectAuthn{
					Host: &v1.Host{
						Host: bmcInfo.Host,
					},
					Username: bmcInfo.Username,
					Password: bmcInfo.Password,
				},
			},
		},
		Vendor: &v1.Vendor{
			Name: bmcInfo.Vendor,
		},
		PowerAction: powerAction,
	}
}
