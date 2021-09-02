package supportbundle

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
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

func osSystemLogCollectors(osFamily v1alpha1.OSFamily) []*v1beta2.Collect {
	switch osFamily {
	case v1alpha1.Ubuntu:
		return ubuntuHostCollectors()
	case v1alpha1.Bottlerocket:
		return bottlerocketHostCollectors()
	default:
		return nil
	}
}

func ubuntuHostCollectors() []*v1beta2.Collect {
	return []*v1beta2.Collect{
		{
			CopyFromHost: &v1beta2.CopyFromHost{
				CollectorMeta: v1beta2.CollectorMeta{},
				Name:          "authLogs",
				Namespace:     constants.EksaSystemNamespace,
				Image:         "busybox:latest",
				HostPath:      "/var/log/auth.log",
			},
		}, {
			CopyFromHost: &v1beta2.CopyFromHost{
				CollectorMeta: v1beta2.CollectorMeta{},
				Name:          "daemonLogs",
				Namespace:     constants.EksaSystemNamespace,
				Image:         "busybox:latest",
				HostPath:      "/var/log/daemon.log",
			},
		}, {
			CopyFromHost: &v1beta2.CopyFromHost{
				CollectorMeta: v1beta2.CollectorMeta{},
				Name:          "debugLogs",
				Namespace:     constants.EksaSystemNamespace,
				Image:         "busybox:latest",
				HostPath:      "/var/log/debug",
			},
		}, {
			CopyFromHost: &v1beta2.CopyFromHost{
				CollectorMeta: v1beta2.CollectorMeta{},
				Name:          "systemLogs",
				Namespace:     constants.EksaSystemNamespace,
				Image:         "busybox:latest",
				HostPath:      "/var/log/syslog",
			},
		}, {
			CopyFromHost: &v1beta2.CopyFromHost{
				CollectorMeta: v1beta2.CollectorMeta{},
				Name:          "cloudInit",
				Namespace:     constants.EksaSystemNamespace,
				Image:         "busybox:latest",
				HostPath:      "/var/log/cloud-init.log",
			},
		}, {
			CopyFromHost: &v1beta2.CopyFromHost{
				CollectorMeta: v1beta2.CollectorMeta{},
				Name:          "cloudInitOutput",
				Namespace:     constants.EksaSystemNamespace,
				Image:         "busybox:latest",
				HostPath:      "/var/log/cloud-init-output.log",
			},
		}, {
			CopyFromHost: &v1beta2.CopyFromHost{
				CollectorMeta: v1beta2.CollectorMeta{},
				Name:          "kernelMesgBuffer",
				Namespace:     constants.EksaSystemNamespace,
				Image:         "busybox:latest",
				HostPath:      "/var/log/dmesg",
			},
		}, {
			CopyFromHost: &v1beta2.CopyFromHost{
				CollectorMeta:   v1beta2.CollectorMeta{},
				Name:            "cniLogCollector",
				Namespace:       constants.EksaSystemNamespace,
				Image:           "busybox:latest",
				HostPath:        "/etc/cni",
			},
		},
	}
}

func bottlerocketHostCollectors() []*v1beta2.Collect {
	return []*v1beta2.Collect{}
}
