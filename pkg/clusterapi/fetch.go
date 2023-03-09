package clusterapi

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// KubeClient is a kubernetes API client.
type KubeClient interface {
	Get(ctx context.Context, name, namespace string, obj kubernetes.Object) error
}

func MachineDeploymentInCluster(ctx context.Context, kubeclient KubeClient, clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) (*clusterv1.MachineDeployment, error) {
	md := &clusterv1.MachineDeployment{}
	err := kubeclient.Get(ctx, MachineDeploymentName(clusterSpec.Cluster, workerNodeGroupConfig), constants.EksaSystemNamespace, md)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return md, nil
}

func KubeadmConfigTemplateInCluster(ctx context.Context, kubeclient KubeClient, md *clusterv1.MachineDeployment) (*bootstrapv1.KubeadmConfigTemplate, error) {
	if md == nil {
		return nil, nil
	}

	kct := &bootstrapv1.KubeadmConfigTemplate{}
	err := kubeclient.Get(ctx, md.Spec.Template.Spec.Bootstrap.ConfigRef.Name, constants.EksaSystemNamespace, kct)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return kct, nil
}
