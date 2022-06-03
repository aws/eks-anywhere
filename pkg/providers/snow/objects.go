package snow

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
	name, err := NewMachineTemplateName(new, old)
	if err != nil {
		return nil, err
	}
	new.SetName(name)

	kubeadmControlPlane, err := KubeadmControlPlane(clusterSpec, new)
	if err != nil {
		return nil, err
	}
	capiCluster := CAPICluster(clusterSpec, snowCluster, kubeadmControlPlane)

	return []runtime.Object{capiCluster, snowCluster, kubeadmControlPlane, new}, nil
}

func WorkersObjects(ctx context.Context, clusterSpec *cluster.Spec, kubeClient kubernetes.Client) ([]runtime.Object, error) {
	oldWorkerNodeGroupsMap, err := oldWorkerNodeGroupsMap(ctx, kubeClient, clusterSpec)
	if err != nil {
		return nil, fmt.Errorf("building old worker node groups mapping: %v", err)
	}

	kubeadmConfigTemplates, err := KubeadmConfigTemplates(ctx, kubeClient, clusterSpec, oldWorkerNodeGroupsMap)
	if err != nil {
		return nil, err
	}

	workerMachineTemplates, err := WorkerMachineTemplates(ctx, kubeClient, clusterSpec, oldWorkerNodeGroupsMap)
	if err != nil {
		return nil, err
	}

	machineDeployments := MachineDeployments(clusterSpec, kubeadmConfigTemplates, workerMachineTemplates)

	return concatWorkersObjects(machineDeployments, kubeadmConfigTemplates, workerMachineTemplates), nil
}

func concatWorkersObjects(machineDeployments map[string]*clusterv1.MachineDeployment,
	kubeadmConfigTemplates map[string]*bootstrapv1.KubeadmConfigTemplate,
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

func WorkerMachineTemplates(ctx context.Context, kubeClient kubernetes.Client, clusterSpec *cluster.Spec, oldWorkerNodeGroupsMap map[string]v1alpha1.WorkerNodeGroupConfiguration) (map[string]*snowv1.AWSSnowMachineTemplate, error) {
	m := map[string]*snowv1.AWSSnowMachineTemplate{}

	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		new := SnowMachineTemplate(clusterSpec.SnowMachineConfigs[workerNodeGroupConfig.MachineGroupRef.Name])

		if oldWorkerNodeGroupsMap == nil {
			m[workerNodeGroupConfig.MachineGroupRef.Name] = new
			break
		}

		md, err := oldMachineDeployment(ctx, kubeClient, clusterSpec, workerNodeGroupConfig)
		if err != nil {
			return nil, err
		}

		old, err := oldWorkerMachineTemplate(ctx, kubeClient, clusterSpec, md)
		if err != nil {
			return nil, err
		}

		name, err := WorkerNodeGroupNewMachineTemplateName(new, old, workerNodeGroupConfig, oldWorkerNodeGroupsMap[workerNodeGroupConfig.Name])
		if err != nil {
			return nil, err
		}

		new.SetName(name)

		m[workerNodeGroupConfig.MachineGroupRef.Name] = new
	}

	return m, nil
}

func WorkerNodeGroupNewMachineTemplateName(newMt, oldMt *snowv1.AWSSnowMachineTemplate, newWng, oldWng v1alpha1.WorkerNodeGroupConfiguration) (string, error) {
	name, err := NewMachineTemplateName(newMt, oldMt)
	if err != nil {
		return "", err
	}

	if recreateMachineNeeded(newWng, oldWng) {
		name, err = clusterapi.IncrementName(oldMt.GetName())
		if err != nil {
			return "", err
		}
	}
	return name, nil
}

func KubeadmConfigTemplates(ctx context.Context, kubeClient kubernetes.Client, clusterSpec *cluster.Spec, oldWorkerNodeGroupsMap map[string]v1alpha1.WorkerNodeGroupConfiguration) (map[string]*bootstrapv1.KubeadmConfigTemplate, error) {
	m := make(map[string]*bootstrapv1.KubeadmConfigTemplate, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))

	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		new, err := kubeadmConfigTemplate(clusterSpec, workerNodeGroupConfig)
		if err != nil {
			return nil, err
		}

		if oldWorkerNodeGroupsMap == nil {
			m[workerNodeGroupConfig.Name] = new
			break
		}

		md, err := oldMachineDeployment(ctx, kubeClient, clusterSpec, workerNodeGroupConfig)
		if err != nil {
			return nil, err
		}

		old, err := oldKubeadmConfigTemplate(ctx, kubeClient, clusterSpec, md)
		if err != nil {
			return nil, err
		}

		if recreateMachineNeeded(workerNodeGroupConfig, oldWorkerNodeGroupsMap[workerNodeGroupConfig.Name]) {
			name, err := clusterapi.IncrementName(old.GetName())
			if err != nil {
				return nil, err
			}
			new.SetName(name)
		}

		m[workerNodeGroupConfig.Name] = new
	}

	return m, nil
}

func recreateMachineNeeded(new, old v1alpha1.WorkerNodeGroupConfiguration) bool {
	return !v1alpha1.TaintsSliceEqual(new.Taints, old.Taints) || !v1alpha1.LabelsMapEqual(new.Labels, old.Labels)
}
