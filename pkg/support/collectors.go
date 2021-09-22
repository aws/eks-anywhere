package supportbundle

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

type collectorFactory struct {
	DiagnosticCollectorImage string
}

func NewCollectorFactory(diagnosticCollectorImage string) *collectorFactory {
	return &collectorFactory{
		DiagnosticCollectorImage: diagnosticCollectorImage,
	}
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
				Namespace: "capd-system",
				Name:      "logs/capd-system",
			},
		},
		{
			Logs: &logs{
				Namespace: "capi-kubeadm-bootstrap-system",
				Name:      "logs/capi-kubeadm-bootstrap-system",
			},
		},
		{
			Logs: &logs{
				Namespace: "capi-kubeadm-control-plane-system",
				Name:      "logs/capi-kubeadm-control-plane-system",
			},
		},
		{
			Logs: &logs{
				Namespace: "capi-system",
				Name:      "logs/capi-system",
			},
		},
		{
			Logs: &logs{
				Namespace: "capi-webhook-system",
				Name:      "logs/capi-webhook-system",
			},
		},
		{
			Logs: &logs{
				Namespace: "cert-manager",
				Name:      "logs/cert-manager",
			},
		},
		{
			Logs: &logs{
				Namespace: "eksa-system",
				Name:      "logs/eksa-system",
			},
		},
		{
			Logs: &logs{
				Namespace: "default",
				Name:      "logs/default",
			},
		},
		{
			Logs: &logs{
				Namespace: "etcdadm-bootstrap-provider-system",
				Name:      "logs/etcdadm-bootstrap-provider-system",
			},
		},
		{
			Logs: &logs{
				Namespace: "etcdadm-controller-system",
				Name:      "logs/etcdadm-controller-system",
			},
		},
		{
			Logs: &logs{
				Namespace: "kube-node-lease",
				Name:      "logs/kube-node-lease",
			},
		},
		{
			Logs: &logs{
				Namespace: "kube-public",
				Name:      "logs/kube-public",
			},
		},
		{
			Logs: &logs{
				Namespace: "kube-system",
				Name:      "logs/kube-system",
			},
		},
	}
}

func (c *collectorFactory) EksaHostCollectors(osFamilies map[v1alpha1.OSFamily]bool) []*Collect {
	for family := range osFamilies {
		switch family {
		case v1alpha1.Ubuntu:
			return c.ubuntuHostCollectors()
		}
	}
	return nil
}

func (c *collectorFactory) ubuntuHostCollectors() []*Collect {
	return []*Collect{
		{
			CopyFromHost: &copyFromHost{
				Name:      "Test Collector API Server",
				Namespace: constants.EksaSystemNamespace,
				Image:     c.DiagnosticCollectorImage,
				HostPath:  "/var/log/kube-apiserver.log",
			},
		},
	}
}
