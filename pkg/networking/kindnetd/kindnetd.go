package kindnetd

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/types"
)

// Client allows to interact with the Kubernetes API.
type Client interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
}

// Kindnetd allows to install and upgrade kindnetd in a an EKS-A cluster.
type Kindnetd struct {
	*Upgrader
	*Installer
}

// NewKindnetd constructs a new Kindnetd.
func NewKindnetd(client Client, reader manifests.FileReader) *Kindnetd {
	return &Kindnetd{
		Installer: NewInstaller(client, reader),
		Upgrader:  NewUpgrader(client, reader),
	}
}

// Install install kindnetd CNI in an eks-a docker cluster.
func (c *Kindnetd) Install(ctx context.Context, cluster *types.Cluster, spec *cluster.Spec, _ []string) error {
	return c.Installer.Install(ctx, cluster, spec)
}
