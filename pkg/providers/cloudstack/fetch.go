package cloudstack

import (
	"context"
	"fmt"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta2"
)

func oldControlPlaneMachineTemplate(ctx context.Context, kubeClient kubernetes.Client, oldSpec *cluster.Spec) (*cloudstackv1.CloudStackMachineTemplate, error) {
	cp := &controlplanev1.KubeadmControlPlane{}
	err := kubeClient.Get(ctx, clusterapi.KubeadmControlPlaneName(oldSpec), constants.EksaSystemNamespace, cp)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	mt := &cloudstackv1.CloudStackMachineTemplate{}
	err = kubeClient.Get(ctx, cp.Spec.MachineTemplate.InfrastructureRef.Name, constants.EksaSystemNamespace, mt)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mt, nil
}

func oldWorkerMachineTemplate(ctx context.Context, kubeclient kubernetes.Client, md *clusterv1.MachineDeployment) (*cloudstackv1.CloudStackMachineTemplate, error) {
	if md == nil {
		return nil, nil
	}

	mt := &cloudstackv1.CloudStackMachineTemplate{}
	err := kubeclient.Get(ctx, md.Spec.Template.Spec.InfrastructureRef.Name, constants.EksaSystemNamespace, mt)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mt, nil
}

func oldEtcdMachineTemplate(ctx context.Context, kubeclient kubernetes.Client, etcdConfig *anywherev1.ExternalEtcdConfiguration, clusterName string) (*cloudstackv1.CloudStackMachineTemplate, error) {
	if etcdConfig == nil {
		return nil, nil
	}
	etcdCluster := &etcdv1.EtcdadmCluster{}
	err := kubeclient.Get(ctx, fmt.Sprintf("%s-etcd", clusterName), constants.EksaSystemNamespace, etcdCluster)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	mt := &cloudstackv1.CloudStackMachineTemplate{}
	err = kubeclient.Get(ctx, etcdConfig.MachineGroupRef.Name, constants.EksaSystemNamespace, mt)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mt, nil
}