package snow

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

func ControlPlaneObjects(ctx context.Context, clusterSpec *cluster.Spec, kubeClient kubernetes.Client) ([]kubernetes.Object, error) {
	capasCredentialsSecret, err := capasCredentialsSecret(clusterSpec)
	if err != nil {
		return nil, err
	}

	snowCluster := SnowCluster(clusterSpec, capasCredentialsSecret)

	new := SnowMachineTemplate(clusterapi.ControlPlaneMachineTemplateName(clusterSpec.Cluster), clusterSpec.SnowMachineConfigs[clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])

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

	return []kubernetes.Object{capiCluster, snowCluster, kubeadmControlPlane, new, capasCredentialsSecret}, nil
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
	workerMachineTemplates map[string]*snowv1.AWSSnowMachineTemplate,
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

// NewMachineTemplateName compares the existing awssnowmachinetemplate object in the cluster with the
// awssnowmachinetemplate constructed from cluster spec, to figure out whether a new awssnowmachinetemplate
// needs to be created. Return the awssnowmachinetemplate name.
func NewMachineTemplateName(new, old *snowv1.AWSSnowMachineTemplate) string {
	if old == nil {
		return new.GetName()
	}

	if MachineTemplateDeepDerivative(new, old) {
		return old.GetName()
	}

	return clusterapi.IncrementNameWithFallbackDefault(old.GetName(), new.GetName())
}

// MachineTemplateDeepDerivative compares two awssnowmachinetemplates to determine if their spec fields are equal.
// DeepDerivative is used so that unset fields in new object are not compared. Although DeepDerivative treats
// new subset slice equal to the original slice. i.e. DeepDerivative([]int{1}, []int{1, 2}) returns true.
// Custom logic is added to justify this usecase since removing a device from the devices list shall trigger machine
// rollout and recreate or the snow cluster goes into a state where the machines on the removed device can’t be deleted.
func MachineTemplateDeepDerivative(new, old *snowv1.AWSSnowMachineTemplate) bool {
	if len(new.Spec.Template.Spec.Devices) != len(old.Spec.Template.Spec.Devices) {
		return false
	}
	return equality.Semantic.DeepDerivative(new.Spec, old.Spec)
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

// credentialsSecret generates the credentials secret(s) used for provisioning a snow cluster.
// - eks-a credentials secret: user managed secret referred from snowdatacenterconfig identityRef
// - snow credentials secret: eks-a creates, updates and deletes in eksa-system namespace. this secret is fully managed by eks-a. User shall treat it as a "read-only" object.
func capasCredentialsSecret(clusterSpec *cluster.Spec) (*v1.Secret, error) {
	if clusterSpec.SnowCredentialsSecret == nil {
		return nil, errors.New("snowCredentialsSecret in clusterSpec shall not be nil")
	}

	// we reconcile the snow credentials secret to be in sync with the eks-a credentials secret user manages.
	// notice for cli upgrade, we handle the eks-a credentials secret update in a separate step - under provider.UpdateSecrets
	// which runs before the actual cluster upgrade.
	// for controller secret, the user is responsible for making sure the eks-a credentials secret is created and up to date.
	credsB64, ok := clusterSpec.SnowCredentialsSecret.Data["credentials"]
	if !ok {
		return nil, fmt.Errorf("unable to retrieve credentials from secret [%s]", clusterSpec.SnowCredentialsSecret.GetName())
	}
	certsB64, ok := clusterSpec.SnowCredentialsSecret.Data["ca-bundle"]
	if !ok {
		return nil, fmt.Errorf("unable to retrieve ca-bundle from secret [%s]", clusterSpec.SnowCredentialsSecret.GetName())
	}

	return CAPASCredentialsSecret(clusterSpec, credsB64, certsB64), nil
}
