package cilium

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/types"
)

// InstallerForSpec allows to configure Cilium for a particular EKS-A cluster
// It's a stateful version of installer, with a fixed Cilium config.
type InstallerForSpec struct {
	installer Installer
	config    Config
}

// NewInstallerForSpec constructs a new InstallerForSpec.
func NewInstallerForSpec(client KubernetesClient, templater InstallTemplater, config Config) *InstallerForSpec {
	return &InstallerForSpec{
		installer: *NewInstaller(client, templater),
		config:    config,
	}
}

// Install installs Cilium in an cluster.
func (i *InstallerForSpec) Install(ctx context.Context, cluster *types.Cluster) error {
	return i.installer.Install(ctx, cluster, i.config.Spec, i.config.AllowedNamespaces)
}

// InstallTemplater generates a Cilium manifest for installation.
type InstallTemplater interface {
	GenerateManifest(ctx context.Context, spec *cluster.Spec, opts ...ManifestOpt) ([]byte, error)
}

// Installer allows to configure Cilium in a cluster.
type Installer struct {
	templater InstallTemplater
	k8s       KubernetesClient
}

// NewInstaller constructs a new Installer.
func NewInstaller(client KubernetesClient, templater InstallTemplater) *Installer {
	return &Installer{
		templater: templater,
		k8s:       client,
	}
}

// Install configures Cilium in an EKS-A cluster.
func (i *Installer) Install(ctx context.Context, cluster *types.Cluster, spec *cluster.Spec, allowedNamespaces []string) error {
	manifest, err := i.templater.GenerateManifest(ctx,
		spec,
		WithPolicyAllowedNamespaces(allowedNamespaces),
	)
	if err != nil {
		return fmt.Errorf("generating Cilium manifest for install: %v", err)
	}

	if err = i.k8s.Apply(ctx, cluster, manifest); err != nil {
		return fmt.Errorf("applying Cilium manifest for install: %v", err)
	}

	return nil
}
