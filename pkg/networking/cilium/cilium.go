package cilium

import (
	"github.com/aws/eks-anywhere/pkg/constants"
)

const namespace = constants.KubeSystemNamespace

// InstallUpgradeTemplater is the composition of InstallTemplater and UpgradeTemplater.
type InstallUpgradeTemplater interface {
	InstallTemplater
	UpgradeTemplater
}

// Cilium allows to install and upgrade the Cilium CNI in clusters.
type Cilium struct {
	*Upgrader
	*Installer
}

// NewCilium constructs a new Cilium.
func NewCilium(client KubernetesClient, templater InstallUpgradeTemplater) *Cilium {
	return &Cilium{
		Installer: NewInstaller(client, templater),
		Upgrader:  NewUpgrader(client, templater),
	}
}
