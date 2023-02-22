package kindnetd

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/types"
)

// InstallerForSpec allows to configure kindnetd for a particular EKS-A cluster
// It's a stateful version of installer, with a fixed Cilium config.
type InstallerForSpec struct {
	installer *Installer
	spec      *cluster.Spec
}

// NewInstallerForSpec constructs a new InstallerForSpec.
func NewInstallerForSpec(client Client, reader manifests.FileReader, spec *cluster.Spec) *InstallerForSpec {
	return &InstallerForSpec{
		installer: NewInstaller(client, reader),
		spec:      spec,
	}
}

// Install installs kindnetd in an cluster.
func (i *InstallerForSpec) Install(ctx context.Context, cluster *types.Cluster) error {
	return i.installer.Install(ctx, cluster, i.spec)
}

// Installer allows to configure kindnetd in a cluster.
type Installer struct {
	k8s    Client
	reader manifests.FileReader
}

// NewInstaller constructs a new Installer.
func NewInstaller(client Client, reader manifests.FileReader) *Installer {
	return &Installer{
		k8s:    client,
		reader: reader,
	}
}

// Install configures kindnetd in an EKS-A cluster.
func (i *Installer) Install(ctx context.Context, cluster *types.Cluster, spec *cluster.Spec) error {
	manifest, err := generateManifest(i.reader, spec)
	if err != nil {
		return fmt.Errorf("generating kindnetd manifest for install: %v", err)
	}

	if err = i.k8s.ApplyKubeSpecFromBytes(ctx, cluster, manifest); err != nil {
		return fmt.Errorf("applying kindnetd manifest for install: %v", err)
	}

	return nil
}
