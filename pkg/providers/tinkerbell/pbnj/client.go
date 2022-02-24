package pbnj

import (
	"context"
	"os"

	"github.com/tinkerbell/cluster-api-provider-tinkerbell/pbnj/client"
	"github.com/tinkerbell/pbnj/api/v1"
)

const (
	PbnjGrpcAuth = "PBNJ_GRPC_AUTHORITY"
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

func (p *Pbnj) ValidateBMCSecretCreds(ctx context.Context, bmcInfo BmcSecretConfig) error {
	_, err := p.pbnj.MachinePower(ctx, p.NewPowerRequest(bmcInfo))
	if err != nil {
		return err
	}

	return nil
}

func (p *Pbnj) NewPowerRequest(bmcInfo BmcSecretConfig) *v1.PowerRequest {
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
		PowerAction: v1.PowerAction_POWER_ACTION_STATUS,
	}
}
