package diagnostics

import (
	"fmt"
	"time"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers"
)

const (
	capdSystem                       = "capd-system"
	capiKubeadmBootstrapSystem       = "capi-kubeadm-bootstrap-system"
	capiKubeadmControlPlaneSystem    = "capi-kubeadm-control-plane-system"
	capiSystem                       = "capi-system"
	capiWebhookSystem                = "capi-webhook-system"
	certManager                      = "cert-manager"
	cefaultNamespace                 = "default"
	etcdAdminBootstrapProviderSystem = "etcdadm-bootstrap-provider-system"
	etcdAdminControllerSystem        = "etcdadm-controller-system"
	kubeNodeLease                    = "kube-node-lease"
	kubePublic                       = "kube-public"
	kubeSystem                       = "kube-system"
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
	return []*Collect{
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
		{
			Logs: &logs{
				Namespace: capdSystem,
				Name:      logpath(capdSystem),
			},
		},
		{
			Logs: &logs{
				Namespace: capiKubeadmBootstrapSystem,
				Name:      logpath(capiKubeadmBootstrapSystem),
			},
		},
		{
			Logs: &logs{
				Namespace: capiKubeadmControlPlaneSystem,
				Name:      logpath(capiKubeadmControlPlaneSystem),
			},
		},
		{
			Logs: &logs{
				Namespace: capiSystem,
				Name:      logpath(capiSystem),
			},
		},
		{
			Logs: &logs{
				Namespace: capiWebhookSystem,
				Name:      logpath(capiWebhookSystem),
			},
		},
		{
			Logs: &logs{
				Namespace: certManager,
				Name:      logpath(certManager),
			},
		},
		{
			Logs: &logs{
				Namespace: constants.EksaSystemNamespace,
				Name:      logpath(constants.EksaSystemNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: cefaultNamespace,
				Name:      logpath(cefaultNamespace),
			},
		},
		{
			Logs: &logs{
				Namespace: etcdAdminBootstrapProviderSystem,
				Name:      logpath(etcdAdminBootstrapProviderSystem),
			},
		},
		{
			Logs: &logs{
				Namespace: etcdAdminControllerSystem,
				Name:      logpath(etcdAdminControllerSystem),
			},
		},
		{
			Logs: &logs{
				Namespace: kubeNodeLease,
				Name:      logpath(kubeNodeLease),
			},
		},
		{
			Logs: &logs{
				Namespace: kubePublic,
				Name:      logpath(kubePublic),
			},
		},
		{
			Logs: &logs{
				Namespace: kubeSystem,
				Name:      logpath(kubeSystem),
			},
		},
	}
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

func (c *collectorFactory) getCollectorsMap() map[v1alpha1.OSFamily][]*Collect {
	return map[v1alpha1.OSFamily][]*Collect{
		v1alpha1.Ubuntu:       c.ubuntuHostCollectors(),
		v1alpha1.Bottlerocket: c.modelRocketHostCollectors(),
	}
}

func (c *collectorFactory) modelRocketHostCollectors() []*Collect {
	return []*Collect{}
}

func (c *collectorFactory) ubuntuHostCollectors() []*Collect {
	return []*Collect{
		{
			CopyFromHost: &copyFromHost{
				Name:      "CloudInitLog",
				Namespace: constants.EksaSystemNamespace,
				Image:     c.DiagnosticCollectorImage,
				HostPath:  "/var/log/cloud-init.log",
			},
		},
		{
			CopyFromHost: &copyFromHost{
				Name:      "CloudInitOutputLog",
				Namespace: constants.EksaSystemNamespace,
				Image:     c.DiagnosticCollectorImage,
				HostPath:  "/var/log/cloud-init-output.log",
			},
		},
		{
			CopyFromHost: &copyFromHost{
				Name:      "Syslog",
				Namespace: constants.EksaSystemNamespace,
				Image:     c.DiagnosticCollectorImage,
				HostPath:  "/var/log/syslog",
				Timeout:   time.Minute.String(),
			},
		},
	}
}

func logpath(namespace string) string {
	return fmt.Sprintf("logs/%s", namespace)
}
