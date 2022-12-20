package snow

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

// ControlPlaneObjects generates the control plane objects for snow provider from clusterSpec.
func ControlPlaneObjects(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec, kubeClient kubernetes.Client) ([]kubernetes.Object, error) {
	cp, err := ControlPlaneSpec(ctx, log, kubeClient, clusterSpec)
	if err != nil {
		return nil, err
	}

	return cp.Objects(), nil
}

type (
	// Workers represents the Snow specific CAPI spec for worker nodes.
	Workers     = clusterapi.Workers[*snowv1.AWSSnowMachineTemplate]
	workerGroup = clusterapi.WorkerGroup[*snowv1.AWSSnowMachineTemplate]
)

// WorkersSpec generates a Snow specific CAPI spec for an eks-a cluster worker nodes.
// It talks to the cluster with a client to detect changes in immutable objects and generates new
// names for them.
func WorkersSpec(ctx context.Context, log logr.Logger, spec *cluster.Spec, kubeClient kubernetes.Client) (*Workers, error) {
	workerMachineTemplates, kubeadmConfigTemplates, err := WorkersMachineAndConfigTemplate(ctx, log, kubeClient, spec)
	if err != nil {
		return nil, err
	}

	machineDeployments := MachineDeployments(spec, kubeadmConfigTemplates, workerMachineTemplates)
	w := &Workers{
		Groups: make([]workerGroup, 0, len(spec.Cluster.Spec.WorkerNodeGroupConfigurations)),
	}
	for _, wc := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		w.Groups = append(w.Groups, workerGroup{
			MachineDeployment:       machineDeployments[wc.Name],
			KubeadmConfigTemplate:   kubeadmConfigTemplates[wc.Name],
			ProviderMachineTemplate: workerMachineTemplates[wc.Name],
		})
	}

	return w, nil
}

// WorkersObjects generates all the objects that compose a Snow specific CAPI spec for the worker nodes of an eks-a cluster.
func WorkersObjects(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec, kubeClient kubernetes.Client) ([]kubernetes.Object, error) {
	w, err := WorkersSpec(ctx, log, clusterSpec, kubeClient)
	if err != nil {
		return nil, err
	}

	return w.WorkerObjects(), nil
}

// WorkersMachineAndConfigTemplate generates the snowMachineTemplates and kubeadmConfigTemplates from clusterSpec.
func WorkersMachineAndConfigTemplate(ctx context.Context, log logr.Logger, kubeClient kubernetes.Client, clusterSpec *cluster.Spec) (map[string]*snowv1.AWSSnowMachineTemplate, map[string]*bootstrapv1.KubeadmConfigTemplate, error) {
	machines := make(map[string]*snowv1.AWSSnowMachineTemplate, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	configs := make(map[string]*bootstrapv1.KubeadmConfigTemplate, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))

	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		md, err := clusterapi.MachineDeploymentInCluster(ctx, kubeClient, clusterSpec, workerNodeGroupConfig)
		if err != nil {
			return nil, nil, err
		}

		// build worker machineTemplate with new clusterSpec
		newMachineTemplate := MachineTemplate(clusterapi.WorkerMachineTemplateName(clusterSpec, workerNodeGroupConfig), clusterSpec.SnowMachineConfigs[workerNodeGroupConfig.MachineGroupRef.Name], nil)

		// build worker kubeadmConfigTemplate with new clusterSpec
		newConfigTemplate, err := KubeadmConfigTemplate(log, clusterSpec, workerNodeGroupConfig)
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
