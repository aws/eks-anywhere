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

type Workers[M Object] struct {
	Groups []WorkerGroup[M]
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

type WorkerGroup[M Object] struct {
	KubeadmConfigTemplate   *kubeadmv1.KubeadmConfigTemplate
	MachineDeployment       *clusterv1.MachineDeployment
	ProviderMachineTemplate M
}

// UpdateImmutableObjectNames checks if any immutable objects have changed by comparing the new definition
// with the current state of the cluster. If they had, it generates a new name for them by increasing a monotonic number
// at the end of the name.
// This process is performed to the provider machine template and the kubeadmconfigtemplate.
// The kubeadmconfigtemplate is not immutable at the API level but we treat is a such for consistency
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

	g.KubeadmConfigTemplate.SetName(currentMachineDeployment.Spec.Template.Spec.Bootstrap.ConfigRef.Name)
	if err = EnsureNewNameIfChanged(ctx, client, GetKubeadmConfigTemplate, KubeadmConfigTemplateEqual, g.KubeadmConfigTemplate); err != nil {
		return err
	}

	return nil
}

// GetKubeadmConfigTemplate retrieves a KubeadmConfigTemplate using a client
// Implements ObjectRetriever
func GetKubeadmConfigTemplate(ctx context.Context, client kubernetes.Client, name, namespace string) (*kubeadmv1.KubeadmConfigTemplate, error) {
	k := &kubeadmv1.KubeadmConfigTemplate{}
	if err := client.Get(ctx, name, namespace, k); err != nil {
		return nil, err
	}

	return k, nil
}

// KubeadmConfigTemplateEqual returns true only if the new version of a KubeadmConfigTemplate
// involves changes with respect6 to the old one when applied to the cluster
// Implements ObjectComparator
func KubeadmConfigTemplateEqual(new, old *kubeadmv1.KubeadmConfigTemplate) bool {
	// DeepDerivative treats empty map (length == 0) as unset field. We need to manually compare certain fields
	// such as taints, so that setting it to empty will trigger machine recreate
	return kubeadmConfigTemplateTaintsEqual(new, old) &&
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
