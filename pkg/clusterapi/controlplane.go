package clusterapi

import (
	"context"

	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

type ProviderControlPlane interface {
	Objects() []kubernetes.Object
}

type NoObjectsProviderControlPlane struct{}

func (NoObjectsProviderControlPlane) Objects() []kubernetes.Object {
	return nil
}

type Object interface {
	kubernetes.Object
	comparable
}

type ControlPlane[C, M Object, P ProviderControlPlane] struct {
	Cluster                 *clusterv1.Cluster
	KubeadmControlPlane     *controlplanev1.KubeadmControlPlane
	EtcdCluster             *etcdv1.EtcdadmCluster
	ProviderCluster         C
	ProviderMachineTemplate M
	EtcdMachineTemplate     M
	Provider                P
}

func (cp *ControlPlane[C, M, P]) Objects() []kubernetes.Object {
	objs := cp.Provider.Objects()
	objs = append(objs, cp.Cluster, cp.KubeadmControlPlane)
	if cp.EtcdCluster != nil {
		objs = append(objs, cp.EtcdCluster)
	}

	return objs
}

func (cp *ControlPlane[C, M, P]) UpdateImmutableObjectNames(
	ctx context.Context,
	client kubernetes.Client,
	machineTemplateComparator func(context.Context, kubernetes.Client, M) (bool, error),
) error {
	currentKCP := &controlplanev1.KubeadmControlPlane{}
	err := client.Get(ctx, cp.KubeadmControlPlane.Name, cp.KubeadmControlPlane.Namespace, currentKCP)
	if apierrors.IsNotFound(err) {
		// KubeadmControlPlane doesn't exist, this is a new cluster so machine templates should use their default name
		return nil
	}
	if err != nil {
		return err
	}

	cp.ProviderMachineTemplate.SetName(currentKCP.Spec.MachineTemplate.InfrastructureRef.Name)
	if err = ensureUniqueObjectName(ctx, client, machineTemplateComparator, cp.ProviderMachineTemplate); err != nil {
		return err
	}

	if cp.EtcdCluster == nil {
		return nil
	}

	currentEtcdCluster := &etcdv1.EtcdadmCluster{}
	err = client.Get(ctx, cp.EtcdCluster.Name, cp.EtcdCluster.Namespace, currentEtcdCluster)
	if apierrors.IsNotFound(err) {
		// EtcdadmCluster doesn't exist, this is a new cluster so machine templates should use their default name
		return nil
	}
	if err != nil {
		return err
	}

	cp.EtcdMachineTemplate.SetName(currentEtcdCluster.Spec.InfrastructureTemplate.Name)
	if err = ensureUniqueObjectName(ctx, client, machineTemplateComparator, cp.EtcdMachineTemplate); err != nil {
		return err
	}

	return nil
}

func ensureUniqueObjectName[M Object](ctx context.Context,
	client kubernetes.Client,
	machineTemplateComparator func(context.Context, kubernetes.Client, M) (bool, error),
	machineTemplate M,
) error {
	equal, err := machineTemplateComparator(ctx, client, machineTemplate)
	if err != nil {
		return err
	}

	if !equal {
		newName, err := IncrementName(machineTemplate.GetName())
		if err != nil {
			return err
		}

		machineTemplate.SetName(newName)
	}

	return nil
}
