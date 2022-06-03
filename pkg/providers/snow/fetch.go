package snow

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

func oldControlPlaneMachineTemplate(ctx context.Context, kubeClient kubernetes.Client, oldSpec *cluster.Spec) (*snowv1.AWSSnowMachineTemplate, error) {
	cp := &controlplanev1.KubeadmControlPlane{}
	err := kubeClient.Get(ctx, clusterapi.KubeadmControlPlaneName(oldSpec), constants.EksaSystemNamespace, cp)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	mt := &snowv1.AWSSnowMachineTemplate{}
	err = kubeClient.Get(ctx, cp.Spec.MachineTemplate.InfrastructureRef.Name, constants.EksaSystemNamespace, mt)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mt, nil
}

func oldMachineDeployment(ctx context.Context, kubeclient kubernetes.Client, clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) (*clusterv1.MachineDeployment, error) {
	md := &clusterv1.MachineDeployment{}
	err := kubeclient.Get(ctx, clusterapi.MachineDeploymentName(clusterSpec, workerNodeGroupConfig), constants.EksaSystemNamespace, md)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	return md, err
}

func oldWorkerMachineTemplate(ctx context.Context, kubeclient kubernetes.Client, clusterSpec *cluster.Spec, md *clusterv1.MachineDeployment) (*snowv1.AWSSnowMachineTemplate, error) {
	if md == nil {
		return nil, nil
	}

	mt := &snowv1.AWSSnowMachineTemplate{}
	err := kubeclient.Get(ctx, md.Spec.Template.Spec.InfrastructureRef.Name, constants.EksaSystemNamespace, mt)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mt, nil
}

func oldKubeadmConfigTemplate(ctx context.Context, kubeclient kubernetes.Client, clusterSpec *cluster.Spec, md *clusterv1.MachineDeployment) (*bootstrapv1.KubeadmConfigTemplate, error) {
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

func oldEksaCluster(ctx context.Context, kubeclient kubernetes.Client, clusterSpec *cluster.Spec) (*v1alpha1.Cluster, error) {
	cluster := &v1alpha1.Cluster{}
	err := kubeclient.Get(ctx, clusterSpec.Cluster.GetName(), constants.DefaultNamespace, cluster)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func oldWorkerNodeGroupsMap(ctx context.Context, kubeClient kubernetes.Client, clusterSpec *cluster.Spec) (map[string]v1alpha1.WorkerNodeGroupConfiguration, error) {
	eksaCluster, err := oldEksaCluster(ctx, kubeClient, clusterSpec)
	if err != nil || eksaCluster == nil {
		return nil, err
	}

	return cluster.BuildMapForWorkerNodeGroupsByName(eksaCluster.Spec.WorkerNodeGroupConfigurations), nil
}
