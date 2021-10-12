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
	anywhereTypes "github.com/aws/eks-anywhere/pkg/types"
)

const isMgmt = true

var excludeReconcilation = map[string]bool{"kubeadmcontrolplane": true, "etcdadmcluster": true}

type Reconciler interface {
	Reconcile(ctx context.Context, objectKey types.NamespacedName, dryRun bool) error
}

type clusterReconciler struct {
	Log logr.Logger
	ResourceFetcher
	ResourceUpdater
	vsphereTemplate VsphereTemplate
	dockerTemplate  DockerTemplate
}

func NewClusterReconciler(resourceFetcher ResourceFetcher, resourceUpdater ResourceUpdater, now anywhereTypes.NowFunc, log logr.Logger) *clusterReconciler {
	return &clusterReconciler{
		Log:             log,
		ResourceFetcher: resourceFetcher,
		ResourceUpdater: resourceUpdater,
		vsphereTemplate: VsphereTemplate{
			ResourceFetcher: resourceFetcher,
			now:             now,
		},
		dockerTemplate: DockerTemplate{
			ResourceFetcher: resourceFetcher,
			now:             now,
		},
	}
}

func (cor *clusterReconciler) Reconcile(ctx context.Context, objectKey types.NamespacedName, dryRun bool) error {
	cs, err := cor.FetchCluster(ctx, objectKey)
	if err != nil {
		return err
	}
	spec, err := cor.FetchAppliedSpec(ctx, cs)
	if err != nil {
		return err
	}
	switch cs.Spec.DatacenterRef.Kind {
	case anywherev1.VSphereDatacenterKind:
		vdc := &anywherev1.VSphereDatacenterConfig{}
		cpVmc := &anywherev1.VSphereMachineConfig{}
		etcdVmc := &anywherev1.VSphereMachineConfig{}
		workerVmc := &anywherev1.VSphereMachineConfig{}
		err := cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.DatacenterRef.Name}, vdc)
		if err != nil {
			return err
		}
		err = cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.ControlPlaneConfiguration.MachineGroupRef.Name}, cpVmc)
		if err != nil {
			return err
		}
		if len(cs.Spec.WorkerNodeGroupConfigurations) != 1 {
			return fmt.Errorf("expects WorkerNodeGroupConfigurations's length to be 1, but found %d", len(cs.Spec.WorkerNodeGroupConfigurations))
		}
		err = cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name}, workerVmc)
		if err != nil {
			return err
		}
		if cs.Spec.ExternalEtcdConfiguration != nil {
			err = cor.FetchObject(ctx, types.NamespacedName{Namespace: objectKey.Namespace, Name: cs.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name}, etcdVmc)
			if err != nil {
				return err
			}
		}
		resources, err := cor.vsphereTemplate.TemplateResources(ctx, cs, spec, *vdc, *cpVmc, *workerVmc, *etcdVmc)
		if err != nil {
			return err
		}
		return cor.applyTemplates(ctx, resources, isMgmt, dryRun)
	case anywherev1.DockerDatacenterKind:
		resources, err := cor.dockerTemplate.TemplateResources(ctx, cs, spec)
		if err != nil {
			return err
		}
		return cor.applyTemplates(ctx, resources, isMgmt, dryRun)
	default:
		return fmt.Errorf("unsupport Provider %s", cs.Spec.DatacenterRef.Kind)
	}
}

func (cor *clusterReconciler) applyTemplates(ctx context.Context, resources []*unstructured.Unstructured, isMgmt bool, dryRun bool) error {
	for _, resource := range resources {
		kind := resource.GetKind()
		name := resource.GetName()
		if isMgmt && skipReconciliation(resource.GetKind()) {
			cor.Log.Info("skipping object for management cluster", "kind", kind, "name", name, "dryRun", dryRun)
			continue
		}
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

func skipReconciliation(kind string) bool {
	if _, ok := excludeReconcilation[strings.ToLower(kind)]; ok {
		return true
	}
	return false
}
