package resource

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
	anywhereTypes "github.com/aws/eks-anywhere/pkg/types"
)

type Reconciler interface {
	Reconcile(ctx context.Context, objectKey types.NamespacedName, dryRun bool) error
}

type clusterReconciler struct {
	Log logr.Logger
	ResourceFetcher
	ResourceUpdater
	vsphereTemplate      VsphereTemplate
	cloudStackTemplate   CloudStackTemplate
	dockerTemplate       DockerTemplate
	awsIamConfigTemplate AWSIamConfigTemplate
}

func NewClusterReconciler(resourceFetcher ResourceFetcher, resourceUpdater ResourceUpdater, now anywhereTypes.NowFunc, log logr.Logger) *clusterReconciler {
	return &clusterReconciler{
		Log:             log,
		ResourceFetcher: resourceFetcher,
		ResourceUpdater: resourceUpdater,
		vsphereTemplate: VsphereTemplate{
			ResourceFetcher: resourceFetcher,
			ResourceUpdater: resourceUpdater,
			now:             now,
		},
		cloudStackTemplate: CloudStackTemplate{
			ResourceFetcher: resourceFetcher,
			ResourceUpdater: resourceUpdater,
			now:             now,
		},
		dockerTemplate: DockerTemplate{
			ResourceFetcher: resourceFetcher,
			now:             now,
		},
		awsIamConfigTemplate: AWSIamConfigTemplate{
			ResourceFetcher: resourceFetcher,
		},
	}
}

func (cor *clusterReconciler) Reconcile(ctx context.Context, objectKey types.NamespacedName, dryRun bool) error {
	var resources []*unstructured.Unstructured
	cs, err := cor.FetchCluster(ctx, objectKey)
	if err != nil {
		return err
	}
	spec, err := cor.FetchAppliedSpec(ctx, cs)
	if err != nil {
		return err
	}
	err = cor.fetchIdentityProviderRefs(ctx, spec, objectKey.Namespace)
	if err != nil {
		return err
	}

	switch cs.Spec.DatacenterRef.Kind {
	case anywherev1.VSphereDatacenterKind:
		vdc := &anywherev1.VSphereDatacenterConfig{}
		cpVmc := &anywherev1.VSphereMachineConfig{}
		etcdVmc := &anywherev1.VSphereMachineConfig{}
		workerVmc := &anywherev1.VSphereMachineConfig{}
		workerVmcs := make(map[string]anywherev1.VSphereMachineConfig, len(cs.Spec.WorkerNodeGroupConfigurations))
		err := cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.DatacenterRef.Name}, vdc)
		if err != nil {
			return err
		}
		err = cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.ControlPlaneConfiguration.MachineGroupRef.Name}, cpVmc)
		if err != nil {
			return err
		}
		for _, workerNodeGroupConfiguration := range cs.Spec.WorkerNodeGroupConfigurations {
			err = cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: workerNodeGroupConfiguration.MachineGroupRef.Name}, workerVmc)
			if err != nil {
				return err
			}
			workerVmcs[workerNodeGroupConfiguration.MachineGroupRef.Name] = *workerVmc
		}
		if cs.Spec.ExternalEtcdConfiguration != nil {
			err = cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name}, etcdVmc)
			if err != nil {
				return err
			}
		}
		r, err := cor.vsphereTemplate.TemplateResources(ctx, cs, spec, *vdc, *cpVmc, *etcdVmc, workerVmcs)
		if err != nil {
			return err
		}
		resources = append(resources, r...)
	case anywherev1.CloudStackDatacenterKind:
		if !features.IsActive(features.CloudStackProvider()) {
			return fmt.Errorf("cloudstack provider is not supported in eks-a controller")
		}
		csdc := &anywherev1.CloudStackDatacenterConfig{}
		cpCsmc := &anywherev1.CloudStackMachineConfig{}
		etcdCsmc := &anywherev1.CloudStackMachineConfig{}
		workerCsmc := &anywherev1.CloudStackMachineConfig{}
		workerCsmcs := make(map[string]anywherev1.CloudStackMachineConfig, len(cs.Spec.WorkerNodeGroupConfigurations))
		err := cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.DatacenterRef.Name}, csdc)
		if err != nil {
			return err
		}
		err = cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.ControlPlaneConfiguration.MachineGroupRef.Name}, cpCsmc)
		if err != nil {
			return err
		}
		if len(cs.Spec.WorkerNodeGroupConfigurations) != 1 {
			return fmt.Errorf("expects WorkerNodeGroupConfigurations's length to be 1, but found %d", len(cs.Spec.WorkerNodeGroupConfigurations))
		}
		err = cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name}, workerCsmc)
		if err != nil {
			return err
		}
		workerCsmcs[cs.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name] = *workerCsmc
		if cs.Spec.ExternalEtcdConfiguration != nil {
			err = cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name}, etcdCsmc)
			if err != nil {
				return err
			}
		}
		r, err := cor.cloudStackTemplate.TemplateResources(ctx, cs, spec, *csdc, *cpCsmc, *etcdCsmc, workerCsmcs)
		if err != nil {
			return err
		}
		resources = append(resources, r...)
	case anywherev1.DockerDatacenterKind:
		r, err := cor.dockerTemplate.TemplateResources(ctx, cs, spec)
		if err != nil {
			return err
		}
		resources = append(resources, r...)
	default:
		return fmt.Errorf("unsupport Provider %s", cs.Spec.DatacenterRef.Kind)
	}

	// Reconcile IdentityProviders
	for _, identityProvider := range cs.Spec.IdentityProviderRefs {
		switch identityProvider.Kind {
		case anywherev1.AWSIamConfigKind:
			r, err := cor.awsIamConfigTemplate.TemplateResources(ctx, spec)
			if err != nil {
				return err
			}
			resources = append(resources, r...)
		}
	}
	return cor.applyTemplates(ctx, resources, dryRun)
}

func (cor *clusterReconciler) applyTemplates(ctx context.Context, resources []*unstructured.Unstructured, dryRun bool) error {
	for _, resource := range resources {
		kind := resource.GetKind()
		name := resource.GetName()
		cor.Log.Info("applying object", "kind", kind, "name", name, "dryRun", dryRun)
		fetch, err := cor.Fetch(ctx, resource.GetName(), resource.GetNamespace(), resource.GetKind(), resource.GetAPIVersion())
		if err == nil {
			resource.SetResourceVersion(fetch.GetResourceVersion())
			if err := cor.ApplyUpdatedTemplate(ctx, resource, dryRun); err != nil {
				return err
			}
			continue
		}
		if statusError, isStatus := err.(*errors.StatusError); isStatus && statusError.Status().Reason == metav1.StatusReasonNotFound {
			if err := cor.ForceApplyTemplate(ctx, resource, dryRun); err != nil {
				return err
			}
			continue
		}
		return err
	}
	return nil
}

func (cor *clusterReconciler) fetchIdentityProviderRefs(ctx context.Context, cs *cluster.Spec, namespace string) error {
	for _, identityProvider := range cs.Cluster.Spec.IdentityProviderRefs {
		switch identityProvider.Kind {
		case anywherev1.AWSIamConfigKind:
			awsIamConfig, err := cor.AWSIamConfig(ctx, &identityProvider, namespace)
			if err != nil {
				return err
			}
			cs.AWSIamConfig = awsIamConfig
		case anywherev1.OIDCConfigKind:
			oidcConfig, err := cor.OIDCConfig(ctx, &identityProvider, namespace)
			if err != nil {
				return err
			}
			cs.OIDCConfig = oidcConfig
		}
	}
	return nil
}
