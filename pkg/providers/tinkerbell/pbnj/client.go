package pbnj

import (
	"context"
	"os"

	pbnjClient "github.com/tinkerbell/cluster-api-provider-tinkerbell/pbnj/client"
	pbnjv1 "github.com/tinkerbell/pbnj/api/v1"

	"github.com/aws/eks-anywhere/pkg/hardware"
)

const (
	PbnjGrpcAuth = "PBNJ_GRPC_AUTHORITY"
)

type Pbnj struct {
	*pbnjClient.PbnjClient
}

func NewPBNJClient(pbnjGrpcAuth string) (*Pbnj, error) {
	os.Setenv(PbnjGrpcAuth, pbnjGrpcAuth)

	conn, _ := pbnjClient.SetupConnection()

	mClient := pbnjv1.NewMachineClient(conn)
	tClient := pbnjv1.NewTaskClient(conn)
	pbnjObj := &Pbnj{pbnjClient.NewPbnjClient(mClient, tClient)}

	os.Unsetenv(PbnjGrpcAuth)

	return pbnjObj, nil
}

func (p *Pbnj) ValidateBMCSecretCreds(ctx context.Context, bmcInfo hardware.BmcSecretConfig) error {
	powerRequest := &pbnjv1.PowerRequest{
		Authn: &pbnjv1.Authn{
			Authn: &pbnjv1.Authn_DirectAuthn{
				DirectAuthn: &pbnjv1.DirectAuthn{
					Host: &pbnjv1.Host{
						Host: bmcInfo.Host,
					},
					Username: bmcInfo.Username,
					Password: bmcInfo.Password,
				},
			},
		},
		Vendor: &pbnjv1.Vendor{
			Name: bmcInfo.Vendor,
		},
		PowerAction: pbnjv1.PowerAction_POWER_ACTION_STATUS,
	}

	_, err := p.MachinePower(ctx, powerRequest)
	if err != nil {
		return err
	}

	return nil
}
