package snow

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
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

func oldWorkerMachineTemplate(ctx context.Context, kubeclient kubernetes.Client, clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) (*snowv1.AWSSnowMachineTemplate, error) {
	md := &clusterv1.MachineDeployment{}
	err := kubeclient.Get(ctx, clusterapi.MachineDeploymentName(clusterSpec, workerNodeGroupConfig), constants.EksaSystemNamespace, md)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	mt := &snowv1.AWSSnowMachineTemplate{}
	err = kubeclient.Get(ctx, md.Spec.Template.Spec.InfrastructureRef.Name, constants.EksaSystemNamespace, mt)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return mt, nil
}
