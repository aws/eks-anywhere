package clusterapi

import (
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

type ProviderControlPlane interface {
	Objects() []kubernetes.Object
}

type ControlPlane[P ProviderControlPlane] struct {
	Cluster             *clusterv1.Cluster
	KubeadmControlPlane *controlplanev1.KubeadmControlPlane
	EtcdCluster         *etcdv1.EtcdadmCluster
	Provider            P
}

func (cp *ControlPlane[P]) Objects() []kubernetes.Object {
	objs := cp.Provider.Objects()
	objs = append(objs, cp.Cluster, cp.KubeadmControlPlane)
	if cp.EtcdCluster != nil {
		objs = append(objs, cp.EtcdCluster)
	}

	return objs
}

func ProcessCluster[P ProviderControlPlane](cp *ControlPlane[P], lookup yamlutil.ObjectLookup) {
	for _, obj := range lookup {
		if obj.GetObjectKind().GroupVersionKind().Kind == "Cluster" {
			cp.Cluster = obj.(*clusterv1.Cluster)
			return
		}
	}
}

func ProcessKubeadmControlPlane[P ProviderControlPlane](cp *ControlPlane[P], lookup yamlutil.ObjectLookup) {
	if cp.Cluster == nil {
		return
	}

	kcp := lookup.GetFromRef(*cp.Cluster.Spec.ControlPlaneRef)
	if kcp == nil {
		return
	}

	cp.KubeadmControlPlane = kcp.(*controlplanev1.KubeadmControlPlane)
}

func ProcessEtcdCluster[P ProviderControlPlane](cp *ControlPlane[P], lookup yamlutil.ObjectLookup) {
	if cp.Cluster == nil || cp.Cluster.Spec.ManagedExternalEtcdRef == nil {
		return
	}

	etcdCluster := lookup.GetFromRef(*cp.Cluster.Spec.ManagedExternalEtcdRef)
	if etcdCluster == nil {
		return
	}

	cp.EtcdCluster = etcdCluster.(*etcdv1.EtcdadmCluster)
}

func RegisterControlPlaneMappings[T any](parser *yamlutil.Parser[T]) error {
	err := parser.RegisterMappings(map[string]yamlutil.APIObjectGenerator{
		"Cluster": func() yamlutil.APIObject {
			return &clusterv1.Cluster{}
		},
		"KubeadmControlPlane": func() yamlutil.APIObject {
			return &controlplanev1.KubeadmControlPlane{}
		},
		"EtdcCluster": func() yamlutil.APIObject {
			return &etcdv1.EtcdadmCluster{}
		},
	})
	if err != nil {
		return err
	}

	return nil
}
