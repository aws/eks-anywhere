package validations

import (
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Opts struct {
	Kubectl           KubectlClient
	Spec              *cluster.Spec
	WorkloadCluster   *types.Cluster
	ManagementCluster *types.Cluster
	Provider          providers.Provider
	TlsValidator      TlsValidator
	CliConfig         *config.CliConfig
	Version           VersionClient
}

func (o *Opts) SetDefaults() {
	if o.TlsValidator == nil {
		o.TlsValidator = crypto.NewTlsValidator()
	}
}
