package diagnostics

import (
	"fmt"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers"
)

type collectorFactory struct {
	DiagnosticCollectorImage string
}

func NewCollectorFactory(diagnosticCollectorImage string) *collectorFactory {
	return &collectorFactory{
		DiagnosticCollectorImage: diagnosticCollectorImage,
	}
}

func NewDefaultCollectorFactory() *collectorFactory {
	return &collectorFactory{}
}

func (c *collectorFactory) DefaultCollectors() []*Collect {
	collectors := []*Collect{
		{
			ClusterInfo: &clusterInfo{},
		},
		{
			ClusterResources: &clusterResources{},
		},
		{
			Secret: &secret{
				Namespace:    "eksa-system",
				SecretName:   "eksa-license",
				IncludeValue: true,
				Key:          "license",
			},
		},
	}
	collectors = append(collectors, c.defaultLogCollectors()...)
	return collectors
}

func (c *collectorFactory) EksaHostCollectors(machineConfigs []providers.MachineConfig) []*Collect {
	var collectors []*Collect
	collectorsMap := c.getCollectorsMap()

	// we don't want to duplicate the collectors if multiple machine configs have the same OS family
	osFamiliesSeen := map[v1alpha1.OSFamily]bool{}
	for _, config := range machineConfigs {
		if _, seen := osFamiliesSeen[config.OSFamily()]; !seen {
			collectors = append(collectors, collectorsMap[config.OSFamily()]...)
			osFamiliesSeen[config.OSFamily()] = true
		}
	}
	return collectors
}

func (c *collectorFactory) DataCenterConfigCollectors(datacenter v1alpha1.Ref) []*Collect {
	switch datacenter.Kind {
	case v1alpha1.VSphereDatacenterKind:
		return c.eksaVsphereCollectors()
	case v1alpha1.DockerDatacenterKind:
		return c.eksaDockerCollectors()
	case v1alpha1.CloudStackDatacenterKind:
		return c.eksaCloudstackCollectors()
	default:
		return nil
	}
}

func (c *collectorFactory) eksaVsphereCollectors() []*Collect {
	vsphereLogs := []*Collect{
		{
			Logs: &logs{
				Namespace: constants.CapvSystemNamespace,
				Name:      logpath(constants.CapvSystemNamespace),
			},
		},
	}
	return append(vsphereLogs, c.vsphereCrdCollectors()...)
}

func (c *collectorFactory) eksaCloudstackCollectors() []*Collect {
	cloudstackLogs := []*Collect{
		{
			Logs: &logs{
				Namespace: constants.CapcSystemNamespace,
				Name:      logpath(constants.CapcSystemNamespace),
			},
		},
	}
	return append(cloudstackLogs, c.cloudstackCrdCollectors()...)
}

func (c *collectorFactory) eksaDockerCollectors() []*Collect {
	return []*Collect{
		{
			Logs: &logs{
				Namespace: constants.CapdSystemNamespace,
				Name:      logpath(constants.CapdSystemNamespace),
			},
		},
	}
}

func (c *collectorFactory) ManagementClusterCollectors() []*Collect {
	var collectors []*Collect
	collectors = append(collectors, c.managementClusterCrdCollectors()...)
	collectors = append(collectors, c.managementClusterLogCollectors()...)
	return collectors
}

func (c *collectorFactory) PackagesCollectors() []*Collect {
	var collectors []*Collect
	collectors = append(collectors, c.packagesCrdCollectors()...)
	collectors = append(collectors, c.packagesLogCollectors()...)
	return collectors
}

func (c *collectorFactory) getCollectorsMap() map[v1alpha1.OSFamily][]*Collect {
	return map[v1alpha1.OSFamily][]*Collect{
		v1alpha1.Ubuntu:       c.ubuntuHostCollectors(),
		v1alpha1.Bottlerocket: c.bottleRocketHostCollectors(),
	}
}

func (c *collectorFactory) bottleRocketHostCollectors() []*Collect {
	return []*Collect{}
}

func (c *collectorFactory) ubuntuHostCollectors() []*Collect {
	return []*Collect{
		{
			CopyFromHost: &copyFromHost{
				Name:      hostlogPath("cloud-init"),
				Namespace: constants.EksaDiagnosticsNamespace,
				Image:     c.DiagnosticCollectorImage,
				HostPath:  "/var/log/cloud-init.log",
			},
		},
		{
			CopyFromHost: &copyFromHost{
				Name:      hostlogPath("cloud-init-output"),
				Namespace: constants.EksaDiagnosticsNamespace,
				Image:     c.DiagnosticCollectorImage,
				HostPath:  "/var/log/cloud-init-output.log",
			},
		},
		{
			CopyFromHost: &copyFromHost{
				Name:      hostlogPath("syslog"),
				Namespace: constants.EksaDiagnosticsNamespace,
				Image:     c.DiagnosticCollectorImage,
				HostPath:  "/var/log/syslog",
				Timeout:   time.Minute.String(),
			},
		},
	}
}

func (c *collectorFactory) defaultLogCollectors() []*Collect {
	return []*Collect{
		{
			Logs: &logs{
				Namespace: constants.EksaSystemNamespace,
				Name:      logpath(constants.EksaSystemNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.DefaultNamespace,
				Name:      logpath(constants.DefaultNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.KubeNodeLeaseNamespace,
				Name:      logpath(constants.KubeNodeLeaseNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.KubePublicNamespace,
				Name:      logpath(constants.KubePublicNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.KubeSystemNamespace,
				Name:      logpath(constants.KubeSystemNamespace),
			},
		},
	}
}

func (c *collectorFactory) packagesLogCollectors() []*Collect {
	return []*Collect{
		{
			Logs: &logs{
				Namespace: constants.EksaPackagesName,
				Name:      logpath(constants.EksaPackagesName),
			},
		},
	}
}

func (c *collectorFactory) managementClusterLogCollectors() []*Collect {
	return []*Collect{
		{
			Logs: &logs{
				Namespace: constants.CapiKubeadmBootstrapSystemNamespace,
				Name:      logpath(constants.CapiKubeadmBootstrapSystemNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.CapiKubeadmControlPlaneSystemNamespace,
				Name:      logpath(constants.CapiKubeadmControlPlaneSystemNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.CapiSystemNamespace,
				Name:      logpath(constants.CapiSystemNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.CapiWebhookSystemNamespace,
				Name:      logpath(constants.CapiWebhookSystemNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.CertManagerNamespace,
				Name:      logpath(constants.CertManagerNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.EtcdAdmBootstrapProviderSystemNamespace,
				Name:      logpath(constants.EtcdAdmBootstrapProviderSystemNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.EtcdAdmControllerSystemNamespace,
				Name:      logpath(constants.EtcdAdmControllerSystemNamespace),
			},
		},
	}
}

func (c *collectorFactory) managementClusterCrdCollectors() []*Collect {
	mgmtCrds := []string{
		"clusters.anywhere.eks.amazonaws.com",
		"bundles.anywhere.eks.amazonaws.com",
		"clusters.cluster.x-k8s.io",
		"machinedeployments.cluster.x-k8s.io",
		"machines.cluster.x-k8s.io",
		"machinehealthchecks.cluster.x-k8s.io",
		"kubeadmcontrolplane.controlplane.cluster.x-k8s.io",
	}
	return c.generateCrdCollectors(mgmtCrds)
}

func (c *collectorFactory) vsphereCrdCollectors() []*Collect {
	capvCrds := []string{
		"vsphereclusteridentities.infrastructure.cluster.x-k8s.io",
		"vsphereclusters.infrastructure.cluster.x-k8s.io",
		"vspheredatacenterconfigs.anywhere.eks.amazonaws.com",
		"vspheremachineconfigs.anywhere.eks.amazonaws.com",
		"vspheremachines.infrastructure.cluster.x-k8s.io",
		"vspheremachinetemplates.infrastructure.cluster.x-k8s.io",
		"vspherevms.infrastructure.cluster.x-k8s.io",
	}
	return c.generateCrdCollectors(capvCrds)
}

func (c *collectorFactory) cloudstackCrdCollectors() []*Collect {
	crds := []string{
		"cloudstackaffinitygroups.infrastructure.cluster.x-k8s.io",
		"cloudstackclusters.infrastructure.cluster.x-k8s.io",
		"cloudstackdatacenterconfigs.anywhere.eks.amazonaws.com",
		"cloudstackisolatednetworks.infrastructure.cluster.x-k8s.io",
		"cloudstackmachineconfigs.anywhere.eks.amazonaws.com",
		"cloudstackmachines.infrastructure.cluster.x-k8s.io",
		"cloudstackmachinestatecheckers.infrastructure.cluster.x-k8s.io",
		"cloudstackmachinetemplates.infrastructure.cluster.x-k8s.io",
		"cloudstackzones.infrastructure.cluster.x-k8s.io",
	}
	return c.generateCrdCollectors(crds)
}

func (c *collectorFactory) packagesCrdCollectors() []*Collect {
	packageCrds := []string{
		"packagebundlecontrollers.packages.eks.amazonaws.com",
		"packagebundles.packages.eks.amazonaws.com",
		"packagecontrollers.packages.eks.amazonaws.com",
		"packages.packages.eks.amazonaws.com",
	}
	return c.generateCrdCollectors(packageCrds)
}

func (c *collectorFactory) generateCrdCollectors(crds []string) []*Collect {
	var crdCollectors []*Collect
	for _, d := range crds {
		crdCollectors = append(crdCollectors, c.crdCollector(d))
	}
	return crdCollectors
}

func (c *collectorFactory) crdCollector(crdType string) *Collect {
	command := []string{"kubectl"}
	args := []string{"get", crdType, "-o", "json", "--all-namespaces"}
	collectorPath := crdPath(crdType)
	return &Collect{
		Run: &run{
			collectorMeta: collectorMeta{
				CollectorName: crdType,
			},
			Name:      collectorPath,
			Namespace: constants.EksaDiagnosticsNamespace,
			Image:     c.DiagnosticCollectorImage,
			Command:   command,
			Args:      args,
		},
	}
}

func logpath(namespace string) string {
	return fmt.Sprintf("logs/%s", namespace)
}

func hostlogPath(logType string) string {
	return fmt.Sprintf("hostLogs/%s", logType)
}

func crdPath(crdType string) string {
	return fmt.Sprintf("crds/%s", crdType)
}
