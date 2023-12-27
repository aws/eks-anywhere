package validations

import (
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/version"
)

type Opts struct {
	Kubectl            KubectlClient
	Spec               *cluster.Spec
	WorkloadCluster    *types.Cluster
	ManagementCluster  *types.Cluster
	Provider           providers.Provider
	TLSValidator       TlsValidator
	CliConfig          *config.CliConfig
	SkippedValidations map[string]bool
	CliVersion         string
	KubeClient         kubernetes.Client
}

func (o *Opts) SetDefaults() {
	if o.TLSValidator == nil {
		o.TLSValidator = crypto.NewTlsValidator()
	}
	if o.CliVersion == "" {
		o.CliVersion = version.Get().GitVersion
	}
}
