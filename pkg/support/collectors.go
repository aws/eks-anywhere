package supportbundle

type collectorFactory struct{}

func NewCollectorFactory() *collectorFactory {
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
