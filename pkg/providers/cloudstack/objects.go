package cloudstack

import (
	"context"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/utils/unstructured"
	"github.com/go-logr/logr"
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"time"

	"k8s.io/apimachinery/pkg/api/equality"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta2"
)

func buildTemplateBuilder(ctx context.Context, clusterSpec *cluster.Spec, kubeClient kubernetes.Client) (*CloudStackTemplateBuilder, error) {
	machineConfigMap := map[string]*v1alpha1.CloudStackMachineConfig{}

	for _, ref := range clusterSpec.Cluster.MachineConfigRefs() {
		machineConfig := &v1alpha1.CloudStackMachineConfig{}
		if err := kubeClient.Get(ctx, ref.Name, clusterSpec.Cluster.Namespace, machineConfig); err != nil {
			return nil, err
		}
		machineConfigMap[ref.Name] = machineConfig
	}
	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.CloudStackMachineConfigSpec, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, wnConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = machineConfigMap[wnConfig.MachineGroupRef.Name].Spec
	}

	cp := machineConfigMap[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	var etcdSpec *v1alpha1.CloudStackMachineConfigSpec
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcd := machineConfigMap[clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
		etcdSpec = &etcd.Spec
	}

	return NewCloudStackTemplateBuilder(&clusterSpec.CloudStackDatacenter.Spec, &cp.Spec, etcdSpec, workerNodeGroupMachineSpecs, time.Now), nil
}

func ControlPlaneObjects(ctx context.Context, clusterSpec *cluster.Spec, log logr.Logger, kubeClient kubernetes.Client) ([]kubernetes.Object, error) {
	machineConfigMap := map[string]*v1alpha1.CloudStackMachineConfig{}

	log.Info("[DEBUG] Getting MachineConfigRefs for cluster")
	for _, ref := range clusterSpec.Cluster.MachineConfigRefs() {
		machineConfig := &v1alpha1.CloudStackMachineConfig{}
		if err := kubeClient.Get(ctx, ref.Name, clusterSpec.Cluster.Namespace, machineConfig); err != nil {
			return nil, err
		}
		machineConfigMap[ref.Name] = machineConfig
	}
	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.CloudStackMachineConfigSpec, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, wnConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = machineConfigMap[wnConfig.MachineGroupRef.Name].Spec
	}

	log.Info("[DEBUG] Building template builder for CP objects")
	templateBuilder, err := buildTemplateBuilder(ctx, clusterSpec, kubeClient)
	if err != nil {
		return nil, err
	}
	clusterName := clusterSpec.Cluster.ObjectMeta.Name


	cpOpt := func(values map[string]interface{}) {
		log.Info("[DEBUG] Populating cpopts for CP objects")
		values["controlPlaneTemplateName"] = common.CPMachineTemplateName(clusterName, time.Now)
		controlPlaneUser := machineConfigMap[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec.Users[0]
		values["cloudstackControlPlaneSshAuthorizedKey"] = controlPlaneUser.SshAuthorizedKeys[0]

		if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
			etcdUser := machineConfigMap[clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec.Users[0]
			values["cloudstackEtcdSshAuthorizedKey"] = etcdUser.SshAuthorizedKeys[0]
		}

		values["etcdTemplateName"] = common.EtcdMachineTemplateName(clusterName, time.Now)
	}
	log.Info("[DEBUG] Generating CAPI Spec Control Plane")
	controlPlaneSpec, err := templateBuilder.GenerateCAPISpecControlPlane(clusterSpec, cpOpt)
	//log.Info("[DEBUG]", "cp yaml", string(controlPlaneSpec))
	log.Info("[DEBUG] Converting yaml to client objects")
	objs, err := unstructured.YamlToUnstructured(controlPlaneSpec)
	//log.Info("[DEBUG]", "cp objs", objs)
	if err != nil {
		return nil, err
	}
	etcdadmCluster := &etcdv1.EtcdadmCluster{}
	kubeadmCP := &controlplanev1.KubeadmControlPlane{}
	cloudstackCluster := &cloudstackv1.CloudStackCluster{}
	capiCluster := &clusterv1.Cluster{}
	log.Info("[DEBUG] Parsing cluster to identify generated objects")
	for _, obj := range objs {
		log.Info(fmt.Sprintf("[DEBUG] Parsing object of kind %s", obj.GetObjectKind().GroupVersionKind().Kind))
		switch obj.GetObjectKind().GroupVersionKind().Kind {
		case "Cluster":
			log.Info("[DEBUG] Found Cluster object")
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), capiCluster); err != nil {
				return nil, err
			}
		case "CloudStackCluster":
			log.Info("[DEBUG] Found CloudStackCluster object")
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), cloudstackCluster); err != nil {
				return nil, err
			}
		case "KubeadmControlPlane":
			log.Info("[DEBUG] Found KubeadmControlPlane object")
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), kubeadmCP); err != nil {
				return nil, err
			}
		case "EtcdadmCluster":
			log.Info("[DEBUG] Found EtcdadmCluster object")
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), etcdadmCluster); err != nil {
				return nil, err
			}
		}
	}
	log.Info("[DEBUG] Generating new CP CSMT")
	newCpMachineTemplate := CloudStackMachineTemplate(clusterapi.ControlPlaneMachineTemplateName(clusterSpec), clusterSpec.CloudStackMachineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])

	oldCpMachineTemplate, err := oldControlPlaneMachineTemplate(ctx, kubeClient, clusterSpec)
	if err != nil {
		return nil, err
	}

	log.Info("[DEBUG] Setting proper name on CSMT")
	newCpMachineTemplate.SetName(NewMachineTemplateName(newCpMachineTemplate, oldCpMachineTemplate))
	log.Info(fmt.Sprintf("[DEBUG] Set proper name on CSMT: %s\n", newCpMachineTemplate.Name))
	objects := []kubernetes.Object{capiCluster, cloudstackCluster, kubeadmCP, newCpMachineTemplate}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		newEtcdMachineTemplate := CloudStackMachineTemplate(clusterapi.EtcdMachineTemplateName(clusterSpec), clusterSpec.CloudStackMachineConfigs[clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name])
		oldEtcdMachineTemplate, err := oldEtcdMachineTemplate(ctx, kubeClient, clusterSpec.Cluster.Spec.ExternalEtcdConfiguration, clusterSpec.Cluster.ClusterName)
		if err != nil {
			return nil, err
		}
		newEtcdMachineTemplate.SetName(NewMachineTemplateName(newCpMachineTemplate, oldEtcdMachineTemplate))
		objects = append(objects, newEtcdMachineTemplate, etcdadmCluster)
	}

	//log.Info(fmt.Sprintf("[DEBUG] Returning generated CP objects: %s\n", objects))
	return objects, nil
}

func WorkersObjects(ctx context.Context, clusterSpec *cluster.Spec, kubeClient kubernetes.Client) ([]kubernetes.Object, error) {
	workerMachineTemplates, kubeadmConfigTemplates, err := WorkersMachineAndConfigTemplate(ctx, kubeClient, clusterSpec)
	if err != nil {
		return nil, err
	}

	machineDeployments := MachineDeployments(clusterSpec, kubeadmConfigTemplates, workerMachineTemplates)

	return concatWorkersObjects(machineDeployments, kubeadmConfigTemplates, workerMachineTemplates), nil
}

func concatWorkersObjects(machineDeployments map[string]*clusterv1.MachineDeployment,
	kubeadmConfigTemplates map[string]*bootstrapv1.KubeadmConfigTemplate,
	workerMachineTemplates map[string]*cloudstackv1.CloudStackMachineTemplate,
) []kubernetes.Object {
	workersObjs := make([]kubernetes.Object, 0, len(machineDeployments)+len(kubeadmConfigTemplates)+len(workerMachineTemplates))
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

func WorkersMachineAndConfigTemplate(ctx context.Context, kubeClient kubernetes.Client, clusterSpec *cluster.Spec) (map[string]*cloudstackv1.CloudStackMachineTemplate, map[string]*bootstrapv1.KubeadmConfigTemplate, error) {
	machines := make(map[string]*cloudstackv1.CloudStackMachineTemplate, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	configs := make(map[string]*bootstrapv1.KubeadmConfigTemplate, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))

	machineConfigMap := map[string]*v1alpha1.CloudStackMachineConfig{}

	for _, ref := range clusterSpec.Cluster.MachineConfigRefs() {
		machineConfig := &v1alpha1.CloudStackMachineConfig{}
		if err := kubeClient.Get(ctx, ref.Name, clusterSpec.Cluster.Namespace, machineConfig); err != nil {
			return nil, nil, err
		}
		machineConfigMap[ref.Name] = machineConfig
	}
	workerNodeGroupMachineSpecs := make(map[string]v1alpha1.CloudStackMachineConfigSpec, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, wnConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = machineConfigMap[wnConfig.MachineGroupRef.Name].Spec
	}

	kubeadmconfigTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	workloadTemplateNames := make(map[string]string, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))

	templateBuilder, err := buildTemplateBuilder(ctx, clusterSpec, kubeClient)

	for _, wnConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		kubeadmconfigTemplateNames[wnConfig.Name] = common.KubeadmConfigTemplateName(clusterSpec.Cluster.Name, wnConfig.MachineGroupRef.Name, time.Now)
		workloadTemplateNames[wnConfig.Name] = common.WorkerMachineTemplateName(clusterSpec.Cluster.Name, wnConfig.Name, time.Now)
		templateBuilder.WorkerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name] = workerNodeGroupMachineSpecs[wnConfig.MachineGroupRef.Name]
	}
	if err != nil {
		return nil, nil, err
	}
	workersSpec, err := templateBuilder.GenerateCAPISpecWorkers(clusterSpec, workloadTemplateNames, kubeadmconfigTemplateNames)
	objs, err := unstructured.YamlToUnstructured(workersSpec)
	if err != nil {
		return nil, nil, err
	}
	newConfigTemplate := &bootstrapv1.KubeadmConfigTemplate{}
	for _, obj := range objs {
		switch obj.GetObjectKind().GroupVersionKind().Kind {
		case "KubeadmConfigTemplate":
			if err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), newConfigTemplate); err != nil {
				return nil, nil, err
			}
		}
	}

	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		md, err := clusterapi.MachineDeploymentInCluster(ctx, kubeClient, clusterSpec, workerNodeGroupConfig)
		if err != nil {
			return nil, nil, err
		}

		// build worker machineTemplate with new clusterSpec
		newMachineTemplate := CloudStackMachineTemplate(clusterapi.WorkerMachineTemplateName(clusterSpec, workerNodeGroupConfig), clusterSpec.CloudStackMachineConfigs[workerNodeGroupConfig.MachineGroupRef.Name])

		// fetch the existing machineTemplate from cluster
		oldMachineTemplate, err := oldWorkerMachineTemplate(ctx, kubeClient, md)
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

func NewMachineTemplateName(new, old *cloudstackv1.CloudStackMachineTemplate) string {
	if old == nil {
		return new.GetName()
	}

	if equality.Semantic.DeepDerivative(new.Spec, old.Spec) {
		return old.GetName()
	}

	return clusterapi.IncrementNameWithFallbackDefault(old.GetName(), new.GetName())
}

func NewWorkerMachineTemplateName(newMt, oldMt *cloudstackv1.CloudStackMachineTemplate, newKct, oldKct *bootstrapv1.KubeadmConfigTemplate) string {
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
