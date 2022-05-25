package snow

import (
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

func ControlPlaneObjects(ctx context.Context, clusterSpec *cluster.Spec, kubeClient kubernetes.Client) ([]runtime.Object, error) {
	snowCluster := SnowCluster(clusterSpec)
	new := SnowMachineTemplate(clusterSpec.SnowMachineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])

	old, err := oldControlPlaneMachineTemplate(ctx, kubeClient, clusterSpec)
	if err != nil {
		return nil, err
	}
	if err := UpdateMachineTemplateName(new, old); err != nil {
		return nil, err
	}

	kubeadmControlPlane, err := KubeadmControlPlane(clusterSpec, new)
	if err != nil {
		return nil, err
	}
	capiCluster := CAPICluster(clusterSpec, snowCluster, kubeadmControlPlane)

	return []runtime.Object{capiCluster, snowCluster, kubeadmControlPlane, new}, nil
}

func WorkersObjects(ctx context.Context, clusterSpec *cluster.Spec, kubeClient kubernetes.Client) ([]runtime.Object, error) {
	kubeadmConfigTemplates, err := KubeadmConfigTemplates(clusterSpec)
	if err != nil {
		return nil, err
	}

	workerMachineTemplates, err := WorkerMachineTemplates(ctx, kubeClient, clusterSpec)
	if err != nil {
		return nil, err
	}

	machineDeployments := MachineDeployments(clusterSpec, kubeadmConfigTemplates, workerMachineTemplates)

	return concatWorkersObjects(machineDeployments, kubeadmConfigTemplates, workerMachineTemplates), nil
}

func concatWorkersObjects(machineDeployments map[string]*clusterv1.MachineDeployment,
	kubeadmConfigTemplates map[string]*v1beta1.KubeadmConfigTemplate,
	workerMachineTemplates map[string]*snowv1.AWSSnowMachineTemplate,
) []runtime.Object {
	workersObjs := make([]runtime.Object, 0, len(machineDeployments)+len(kubeadmConfigTemplates)+len(workerMachineTemplates))
	for _, item := range machineDeployments {
		workersObjs = append(workersObjs, item)
	}
	for _, item := range kubeadmConfigTemplates {
		workersObjs = append(workersObjs, item)
	}
	for _, item := range workerMachineTemplates {
		workersObjs = append(workersObjs, item)
	}
	return workersObjs
}

func NewMachineTemplateName(new, old *snowv1.AWSSnowMachineTemplate) (string, error) {
	if old == nil {
		return new.GetName(), nil
	}

	if equality.Semantic.DeepDerivative(new.Spec, old.Spec) {
		return old.GetName(), nil
	}

	return clusterapi.IncrementName(old.GetName())
}

func UpdateMachineTemplateName(new, old *snowv1.AWSSnowMachineTemplate) error {
	name, err := NewMachineTemplateName(new, old)
	if err != nil {
		return err
	}
	new.SetName(name)
	return nil
}

func WorkerMachineTemplates(ctx context.Context, kubeClient kubernetes.Client, clusterSpec *cluster.Spec) (map[string]*snowv1.AWSSnowMachineTemplate, error) {
	m := map[string]*snowv1.AWSSnowMachineTemplate{}

	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		new := SnowMachineTemplate(clusterSpec.SnowMachineConfigs[workerNodeGroupConfig.MachineGroupRef.Name])

		old, err := oldWorkerMachineTemplate(ctx, kubeClient, clusterSpec, workerNodeGroupConfig)
		if err != nil {
			return nil, err
		}
		if err := UpdateMachineTemplateName(new, old); err != nil {
			return nil, err
		}

		m[workerNodeGroupConfig.MachineGroupRef.Name] = new
	}
	return m, nil
}
