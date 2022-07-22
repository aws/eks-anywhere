package snow

import (
	"context"

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
	new := SnowMachineTemplate(clusterapi.ControlPlaneMachineTemplateName(clusterSpec), clusterSpec.SnowMachineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])

	old, err := oldControlPlaneMachineTemplate(ctx, kubeClient, clusterSpec)
	if err != nil {
		return nil, err
	}

	new.SetName(NewMachineTemplateName(new, old))

	kubeadmControlPlane, err := KubeadmControlPlane(clusterSpec, new)
	if err != nil {
		return nil, err
	}
	capiCluster := CAPICluster(clusterSpec, snowCluster, kubeadmControlPlane)

	return []runtime.Object{capiCluster, snowCluster, kubeadmControlPlane, new}, nil
}

func WorkersObjects(ctx context.Context, clusterSpec *cluster.Spec, kubeClient kubernetes.Client) ([]runtime.Object, error) {
	workerMachineTemplates, kubeadmConfigTemplates, err := WorkersMachineAndConfigTemplate(ctx, kubeClient, clusterSpec)
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

func WorkersMachineAndConfigTemplate(ctx context.Context, kubeClient kubernetes.Client, clusterSpec *cluster.Spec) (map[string]*snowv1.AWSSnowMachineTemplate, map[string]*bootstrapv1.KubeadmConfigTemplate, error) {
	machines := make(map[string]*snowv1.AWSSnowMachineTemplate, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	configs := make(map[string]*bootstrapv1.KubeadmConfigTemplate, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))

	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		md, err := clusterapi.MachineDeploymentInCluster(ctx, kubeClient, clusterSpec, workerNodeGroupConfig)
		if err != nil {
			return nil, nil, err
		}

		// build worker machineTemplate with new clusterSpec
		newMachineTemplate := SnowMachineTemplate(clusterapi.WorkerMachineTemplateName(clusterSpec, workerNodeGroupConfig), clusterSpec.SnowMachineConfigs[workerNodeGroupConfig.MachineGroupRef.Name])

		// build worker kubeadmConfigTemplate with new clusterSpec
		newConfigTemplate, err := KubeadmConfigTemplate(clusterSpec, workerNodeGroupConfig)
		if err != nil {
			return nil, nil, err
		}

		// fetch the existing machineTemplate from cluster
		oldMachineTemplate, err := oldWorkerMachineTemplate(ctx, kubeClient, clusterSpec, md)
		if err != nil {
			return nil, nil, err
		}

		// fetch the existing kubeadmConfigTemplate from cluster
		oldConfigTemplate, err := clusterapi.KubeadmConfigTemplateInCluster(ctx, kubeClient, md)
		if err != nil {
			return nil, nil, err
		}

		// compare the old and new kubeadmConfigTemplate to determine whether to recreate new kubeadmConfigTemplate object
		configName := NewKubeadmConfigTemplateName(newConfigTemplate, oldConfigTemplate)

		// compare the old and new machineTemplate as well as kubeadmConfigTemplate to determine whether to recreate
		// new machineTemplate object
		machineName := NewWorkerMachineTemplateName(newMachineTemplate, oldMachineTemplate, newConfigTemplate, oldConfigTemplate)

		newConfigTemplate.SetName(configName)
		newMachineTemplate.SetName(machineName)

		configs[workerNodeGroupConfig.Name] = newConfigTemplate
		machines[workerNodeGroupConfig.Name] = newMachineTemplate
	}

	return machines, configs, nil
}

func NewMachineTemplateName(new, old *snowv1.AWSSnowMachineTemplate) string {
	if old == nil {
		return new.GetName()
	}

	if equality.Semantic.DeepDerivative(new.Spec, old.Spec) {
		return old.GetName()
	}

	return clusterapi.IncrementNameWithFallbackDefault(old.GetName(), new.GetName())
}

func NewWorkerMachineTemplateName(newMt, oldMt *snowv1.AWSSnowMachineTemplate, newKct, oldKct *bootstrapv1.KubeadmConfigTemplate) string {
	name := NewMachineTemplateName(newMt, oldMt)

	if oldKct == nil {
		return name
	}

	if recreateKubeadmConfigTemplateNeeded(newKct, oldKct) {
		name = clusterapi.IncrementNameWithFallbackDefault(oldMt.GetName(), newMt.GetName())
	}

	return name
}

func NewKubeadmConfigTemplateName(new, old *bootstrapv1.KubeadmConfigTemplate) string {
	if old == nil {
		return new.GetName()
	}

	if recreateKubeadmConfigTemplateNeeded(new, old) {
		return clusterapi.IncrementNameWithFallbackDefault(old.GetName(), new.GetName())
	}

	return old.GetName()
}

func recreateKubeadmConfigTemplateNeeded(new, old *bootstrapv1.KubeadmConfigTemplate) bool {
	// TODO: DeepDerivative treats empty map (length == 0) as unset field. We need to manually compare certain fields
	// such as taints, so that setting it to empty will trigger machine recreate
	if !v1alpha1.TaintsSliceEqual(new.Spec.Template.Spec.JoinConfiguration.NodeRegistration.Taints, old.Spec.Template.Spec.JoinConfiguration.NodeRegistration.Taints) {
		return true
	}
	return !equality.Semantic.DeepDerivative(new.Spec, old.Spec)
}
