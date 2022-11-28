package clusterapi

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

// Workers represents the provider specific CAPI spec for an eks-a cluster's workers.
type Workers[M Object[M]] struct {
	Groups []WorkerGroup[M]
}

// WorkerObjects returns a list of API objects for concrete provider-specific collection of worker groups.
func (w *Workers[M]) WorkerObjects() []kubernetes.Object {
	objs := make([]kubernetes.Object, 0, len(w.Groups)*3)
	for _, g := range w.Groups {
		objs = append(objs, g.Objects()...)
	}

	return objs
}

// UpdateImmutableObjectNames checks if any immutable objects have changed by comparing the new definition
// with the current state of the cluster. If they had, it generates a new name for them by increasing a monotonic number
// at the end of the name.
func (w *Workers[M]) UpdateImmutableObjectNames(
	ctx context.Context,
	client kubernetes.Client,
	machineTemplateRetriever ObjectRetriever[M],
	machineTemplateComparator ObjectComparator[M],
) error {
	for _, g := range w.Groups {
		if err := g.UpdateImmutableObjectNames(ctx, client, machineTemplateRetriever, machineTemplateComparator); err != nil {
			return err
		}
	}

	return nil
}

// WorkerGroup represents the provider specific CAPI spec for an eks-a worker group.
type WorkerGroup[M Object[M]] struct {
	KubeadmConfigTemplate   *kubeadmv1.KubeadmConfigTemplate
	MachineDeployment       *clusterv1.MachineDeployment
	ProviderMachineTemplate M
}

// Objects returns a list of API objects for a provider-specific of the worker group.
func (g *WorkerGroup[M]) Objects() []kubernetes.Object {
	return []kubernetes.Object{
		g.KubeadmConfigTemplate,
		g.MachineDeployment,
		g.ProviderMachineTemplate,
	}
}

// UpdateImmutableObjectNames checks if any immutable objects have changed by comparing the new definition
// with the current state of the cluster. If they had, it generates a new name for them by increasing a monotonic number
// at the end of the name.
// This process is performed to the provider machine template and the kubeadmconfigtemplate.
// The kubeadmconfigtemplate is not immutable at the API level but we treat it as such for consistency.
func (g *WorkerGroup[M]) UpdateImmutableObjectNames(
	ctx context.Context,
	client kubernetes.Client,
	machineTemplateRetriever ObjectRetriever[M],
	machineTemplateComparator ObjectComparator[M],
) error {
	currentMachineDeployment := &clusterv1.MachineDeployment{}
	err := client.Get(ctx, g.MachineDeployment.Name, g.MachineDeployment.Namespace, currentMachineDeployment)
	if apierrors.IsNotFound(err) {
		// MachineDeployment doesn't exist, this is a new cluster so machine templates should use their default name
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "reading current machine deployment from API")
	}

	g.ProviderMachineTemplate.SetName(currentMachineDeployment.Spec.Template.Spec.InfrastructureRef.Name)
	if err = EnsureNewNameIfChanged(ctx, client, machineTemplateRetriever, machineTemplateComparator, g.ProviderMachineTemplate); err != nil {
		return err
	}
	g.MachineDeployment.Spec.Template.Spec.InfrastructureRef.Name = g.ProviderMachineTemplate.GetName()

	g.KubeadmConfigTemplate.SetName(currentMachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef.Name)
	if err = EnsureNewNameIfChanged(ctx, client, GetKubeadmConfigTemplate, KubeadmConfigTemplateEqual, g.KubeadmConfigTemplate); err != nil {
		return err
	}
	g.MachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef.Name = g.KubeadmConfigTemplate.Name

	return nil
}

// DeepCopy generates a new WorkerGroup copying the contexts of the receiver.
func (g *WorkerGroup[M]) DeepCopy() *WorkerGroup[M] {
	return &WorkerGroup[M]{
		MachineDeployment:       g.MachineDeployment.DeepCopy(),
		KubeadmConfigTemplate:   g.KubeadmConfigTemplate.DeepCopy(),
		ProviderMachineTemplate: g.ProviderMachineTemplate.DeepCopy(),
	}
}

// GetKubeadmConfigTemplate retrieves a KubeadmConfigTemplate using a client
// Implements ObjectRetriever.
func GetKubeadmConfigTemplate(ctx context.Context, client kubernetes.Client, name, namespace string) (*kubeadmv1.KubeadmConfigTemplate, error) {
	k := &kubeadmv1.KubeadmConfigTemplate{}
	if err := client.Get(ctx, name, namespace, k); err != nil {
		return nil, err
	}

	return k, nil
}

// KubeadmConfigTemplateEqual returns true only if the new version of a KubeadmConfigTemplate
// involves changes with respect to the old one when applied to the cluster.
// Implements ObjectComparator.
func KubeadmConfigTemplateEqual(new, old *kubeadmv1.KubeadmConfigTemplate) bool {
	// DeepDerivative treats empty map (length == 0) as unset field. We need to manually compare certain fields
	// such as taints, so that setting it to empty will trigger machine recreate
	return kubeadmConfigTemplateTaintsEqual(new, old) && kubeadmConfigTemplateExtraArgsEqual(new, old) &&
		equality.Semantic.DeepDerivative(new.Spec, old.Spec)
}

func kubeadmConfigTemplateTaintsEqual(new, old *kubeadmv1.KubeadmConfigTemplate) bool {
	return new.Spec.Template.Spec.JoinConfiguration == nil ||
		old.Spec.Template.Spec.JoinConfiguration == nil ||
		anywherev1.TaintsSliceEqual(
			new.Spec.Template.Spec.JoinConfiguration.NodeRegistration.Taints,
			old.Spec.Template.Spec.JoinConfiguration.NodeRegistration.Taints,
		)
}

func kubeadmConfigTemplateExtraArgsEqual(new, old *kubeadmv1.KubeadmConfigTemplate) bool {
	return new.Spec.Template.Spec.JoinConfiguration == nil ||
		old.Spec.Template.Spec.JoinConfiguration == nil ||
		anywherev1.MapEqual(
			new.Spec.Template.Spec.JoinConfiguration.NodeRegistration.KubeletExtraArgs,
			old.Spec.Template.Spec.JoinConfiguration.NodeRegistration.KubeletExtraArgs,
		)
}
