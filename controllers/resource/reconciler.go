package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers/common"
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
	nutanixTemplate      NutanixTemplate
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
			log:             log,
		},
		dockerTemplate: DockerTemplate{
			ResourceFetcher: resourceFetcher,
			now:             now,
		},
		awsIamConfigTemplate: AWSIamConfigTemplate{
			ResourceFetcher: resourceFetcher,
		},
		nutanixTemplate: NutanixTemplate{
			ResourceFetcher: resourceFetcher,
			now:             now,
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
		// max len = len(workers) + CP + etcd
		spec.VSphereMachineConfigs = make(map[string]*anywherev1.VSphereMachineConfig, len(cs.Spec.WorkerNodeGroupConfigurations)+2)
		cpVmc := &anywherev1.VSphereMachineConfig{}
		etcdVmc := &anywherev1.VSphereMachineConfig{}
		workerVmc := &anywherev1.VSphereMachineConfig{}
		workerVmcs := make(map[string]anywherev1.VSphereMachineConfig, len(cs.Spec.WorkerNodeGroupConfigurations))
		err := cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.DatacenterRef.Name}, vdc)
		if err != nil {
			return err
		}
		spec.VSphereDatacenter = vdc
		err = cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.ControlPlaneConfiguration.MachineGroupRef.Name}, cpVmc)
		if err != nil {
			return err
		}
		spec.VSphereMachineConfigs[cs.Spec.ControlPlaneConfiguration.MachineGroupRef.Name] = cpVmc
		for _, workerNodeGroupConfiguration := range cs.Spec.WorkerNodeGroupConfigurations {
			err = cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: workerNodeGroupConfiguration.MachineGroupRef.Name}, workerVmc)
			if err != nil {
				return err
			}
			workerVmcs[workerNodeGroupConfiguration.MachineGroupRef.Name] = *workerVmc
			spec.VSphereMachineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name] = workerVmc
		}
		if cs.Spec.ExternalEtcdConfiguration != nil {
			err = cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name}, etcdVmc)
			if err != nil {
				return err
			}
			spec.VSphereMachineConfigs[cs.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name] = etcdVmc
		}
		r, err := cor.vsphereTemplate.TemplateResources(ctx, cs, spec, *vdc, *cpVmc, *etcdVmc, workerVmcs)
		if err != nil {
			return err
		}
		resources = append(resources, r...)
	case anywherev1.CloudStackDatacenterKind:
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
		for _, workerNodeGroupConfiguration := range cs.Spec.WorkerNodeGroupConfigurations {
			err = cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: workerNodeGroupConfiguration.MachineGroupRef.Name}, workerCsmc)
			if err != nil {
				return err
			}
			workerCsmcs[workerNodeGroupConfiguration.MachineGroupRef.Name] = *workerCsmc
		}
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
	case anywherev1.NutanixDatacenterKind:
		dcConf := &anywherev1.NutanixDatacenterConfig{}
		if err := cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.DatacenterRef.Name}, dcConf); err != nil {
			return err
		}

		controlPlaneMachineConf := &anywherev1.NutanixMachineConfig{}
		if err := cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.ControlPlaneConfiguration.MachineGroupRef.Name}, controlPlaneMachineConf); err != nil {
			return err
		}

		workerMachineConfs := make(map[string]anywherev1.NutanixMachineConfig, len(cs.Spec.WorkerNodeGroupConfigurations))
		for _, workerNodeGroupConfiguration := range cs.Spec.WorkerNodeGroupConfigurations {
			workerMachineConf := &anywherev1.NutanixMachineConfig{}
			if err := cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: workerNodeGroupConfiguration.MachineGroupRef.Name}, workerMachineConf); err != nil {
				return err
			}
			workerMachineConfs[workerNodeGroupConfiguration.MachineGroupRef.Name] = *workerMachineConf
		}

		etcdMachineConf := &anywherev1.NutanixMachineConfig{}
		if cs.Spec.ExternalEtcdConfiguration != nil {
			if err := cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name}, etcdMachineConf); err != nil {
				return err
			}
		}
		r, err := cor.nutanixTemplate.TemplateResources(ctx, cs, spec, *dcConf, *controlPlaneMachineConf, *etcdMachineConf, workerMachineConfs)
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
			// Block controller from updating the self-managed cluster IAM Configmap when reconciling for workload cluster spec.
			if cs.IsSelfManaged() {
				r, err := cor.awsIamConfigTemplate.TemplateResources(ctx, spec)
				if err != nil {
					return err
				}
				resources = append(resources, r...)
			}
		}
	}
	return cor.applyTemplates(ctx, cs, resources, dryRun)
}

func (cor *clusterReconciler) applyTemplates(ctx context.Context, cs *anywherev1.Cluster, resources []*unstructured.Unstructured, dryRun bool) error {
	for _, resource := range resources {
		kind := resource.GetKind()
		name := resource.GetName()
		if cs.IsSelfManaged() && (strings.HasPrefix(name, common.CPMachineTemplateBase(cs.Name)) || strings.HasPrefix(name, common.EtcdMachineTemplateBase(cs.Name))) {
			continue
		}
		cor.Log.Info("applying object", "kind", kind, "name", name, "dryRun", dryRun)
		fetch, err := cor.Fetch(ctx, resource.GetName(), resource.GetNamespace(), resource.GetKind(), resource.GetAPIVersion())
		if err == nil {
			resource.SetResourceVersion(fetch.GetResourceVersion())

			// We want to preserve annotations. It's possible that some CAPI objects have extra annotations
			// that were added manually to handle difficult upgrade/recovery scenarios and we don't want the
			// controller to remove them.
			// Ideally we would use something like server side apply instead of an plain replace, which would
			// extend this to labels, etc. But given this is a legacy controller and will be removed shortly,
			// this is the path of least resistance.
			annotations := fetch.GetAnnotations()
			if annotations == nil {
				annotations = make(map[string]string)
			}

			for k, v := range resource.GetAnnotations() {
				annotations[k] = v
			}

			resource.SetAnnotations(annotations)

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
