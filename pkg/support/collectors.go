package supportbundle

import (
	"github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
)

type collectorFactory struct{}

func NewCollectorFactory() *collectorFactory {
	return &collectorFactory{}
}

func (c *collectorFactory) DefaultCollectors() []*v1beta2.Collect {
	return []*v1beta2.Collect{
		{
			ClusterInfo: &v1beta2.ClusterInfo{},
		},
		{
			ClusterResources: &v1beta2.ClusterResources{},
		},
		{
			Secret: &v1beta2.Secret{
				Namespace:    "eksa-system",
				Name:         "eksa-license",
				IncludeValue: true,
				Key:          "license",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "capd-system",
				Name:      "logs/capd-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "capi-kubeadm-bootstrap-system",
				Name:      "logs/capi-kubeadm-bootstrap-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "capi-kubeadm-control-plane-system",
				Name:      "logs/capi-kubeadm-control-plane-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "capi-system",
				Name:      "logs/capi-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "capi-webhook-system",
				Name:      "logs/capi-webhook-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "cert-manager",
				Name:      "logs/cert-manager",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "eksa-system",
				Name:      "logs/eksa-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "default",
				Name:      "logs/default",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "etcdadm-bootstrap-provider-system",
				Name:      "logs/etcdadm-bootstrap-provider-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "etcdadm-controller-system",
				Name:      "logs/etcdadm-controller-system",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "kube-node-lease",
				Name:      "logs/kube-node-lease",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "kube-public",
				Name:      "logs/kube-public",
			},
		},
		{
			Logs: &v1beta2.Logs{
				Namespace: "kube-system",
				Name:      "logs/kube-system",
			},
		},
	}
}
