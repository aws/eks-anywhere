package resource

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	vspherev3 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1alpha3"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	kubeadmnv1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type ResourceFetcher interface {
	MachineDeployment(ctx context.Context, cs *anywherev1.Cluster) (*clusterv1.MachineDeployment, error)
	VSphereMachineTemplate(ctx context.Context, cs *anywherev1.Cluster) (*vspherev3.VSphereMachineTemplate, error)
	FetchObject(ctx context.Context, objectKey types.NamespacedName, obj client.Object) error
	FetchObjectByName(ctx context.Context, name string, namespace string, obj client.Object) error
	Fetch(ctx context.Context, name string, namespace string, kind string, apiVersion string) (*unstructured.Unstructured, error)
	FetchCluster(ctx context.Context, objectKey types.NamespacedName) (*anywherev1.Cluster, error)
	ExistingVSphereDatacenterConfig(ctx context.Context, cs *anywherev1.Cluster) (*anywherev1.VSphereDatacenterConfig, error)
	ExistingVSphereWorkerMachineConfig(ctx context.Context, cs *anywherev1.Cluster) (*anywherev1.VSphereMachineConfig, error)
	ControlPlane(ctx context.Context, cs *anywherev1.Cluster) (*kubeadmnv1alpha3.KubeadmControlPlane, error)
	FetchAppliedSpec(ctx context.Context, cs *anywherev1.Cluster) (*cluster.Spec, error)
}

type capiResourceFetcher struct {
	client client.Reader
	log    logr.Logger
}

func NewCAPIResourceFetcher(client client.Reader, Log logr.Logger) *capiResourceFetcher {
	return &capiResourceFetcher{
		client: client,
		log:    Log,
	}
}

func (r *capiResourceFetcher) FetchObjectByName(ctx context.Context, name string, namespace string, obj client.Object) error {
	err := r.FetchObject(ctx, types.NamespacedName{Namespace: namespace, Name: name}, obj)
	if err != nil {
		return err
	}
	return nil
}

func (r *capiResourceFetcher) FetchObject(ctx context.Context, objectKey types.NamespacedName, obj client.Object) error {
	err := r.client.Get(ctx, objectKey, obj)
	if err != nil {
		return err
	}
	return nil
}

func (r *capiResourceFetcher) fetchClusterKind(ctx context.Context, objectKey types.NamespacedName) (string, error) {
	supportedKinds := []string{anywherev1.ClusterKind, anywherev1.VSphereDatacenterKind, anywherev1.DockerDatacenterKind, anywherev1.VSphereMachineConfigKind}
	for _, kind := range supportedKinds {
		obj := &unstructured.Unstructured{}
		obj.SetKind(kind)
		obj.SetAPIVersion(anywherev1.GroupVersion.String())
		err := r.FetchObject(ctx, objectKey, obj)
		if err != nil && !apierrors.IsNotFound(err) {
			return "", err
		}
		if err == nil {
			return obj.GetKind(), nil
		}
	}
	return "", fmt.Errorf("no object found for %v", objectKey)
}

func (r *capiResourceFetcher) FetchCluster(ctx context.Context, objectKey types.NamespacedName) (*anywherev1.Cluster, error) {
	r.log.Info("looking up cluster", "objectKey", objectKey)
	kind, err := r.fetchClusterKind(ctx, objectKey)
	if err != nil {
		return nil, err
	}
	switch kind {
	case anywherev1.ClusterKind:
		cluster := &anywherev1.Cluster{}
		if err := r.FetchObject(ctx, objectKey, cluster); err != nil {
			return nil, err
		}
		return cluster, nil
	default:
		return r.fetchClusterForRef(ctx, objectKey, kind)
	}
}

func (r *capiResourceFetcher) clusterByName(ctx context.Context, namespace, name string) (*clusterv1.Cluster, error) {
	cluster := &clusterv1.Cluster{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}
	if err := r.FetchObject(ctx, key, cluster); err != nil {
		return nil, err
	}
	return cluster, nil
}

func (r *capiResourceFetcher) fetchClusterForRef(ctx context.Context, refId types.NamespacedName, kind string) (*anywherev1.Cluster, error) {
	clusters := &anywherev1.ClusterList{}
	o := &client.ListOptions{Raw: &metav1.ListOptions{TypeMeta: metav1.TypeMeta{Kind: anywherev1.ClusterKind, APIVersion: anywherev1.GroupVersion.String()}}, Namespace: refId.Namespace}
	err := r.client.List(ctx, clusters, o)
	if err != nil {
		return nil, err
	}
	for _, c := range clusters.Items {
		if kind == anywherev1.VSphereDatacenterKind || kind == anywherev1.DockerDatacenterKind {
			if c.Spec.DatacenterRef.Name == refId.Name {
				if _, err := r.clusterByName(ctx, constants.EksaSystemNamespace, c.Name); err == nil { // further validates a capi cluster exists
					return &c, nil
				}
			}
		}
		if kind == anywherev1.VSphereMachineConfigKind {
			for _, machineRef := range c.Spec.WorkerNodeGroupConfigurations {
				if machineRef.MachineGroupRef != nil && machineRef.MachineGroupRef.Name == refId.Name {
					if _, err := r.clusterByName(ctx, constants.EksaSystemNamespace, c.Name); err == nil { // further validates a capi cluster exists
						return &c, nil
					}
				}
			}
			if c.Spec.ControlPlaneConfiguration.MachineGroupRef != nil && c.Spec.ControlPlaneConfiguration.MachineGroupRef.Name == refId.Name {
				if _, err := r.clusterByName(ctx, constants.EksaSystemNamespace, c.Name); err == nil { // further validates a capi cluster exists
					return &c, nil
				}
			}
			if c.Spec.ExternalEtcdConfiguration != nil && c.Spec.ExternalEtcdConfiguration.MachineGroupRef != nil && c.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name == refId.Name {
				if _, err := r.clusterByName(ctx, constants.EksaSystemNamespace, c.Name); err == nil { // further validates a capi cluster exists
					return &c, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("eksa cluster not found for datacenterRef %v", refId)
}

func (r *capiResourceFetcher) machineDeployments(ctx context.Context, c *anywherev1.Cluster) ([]*clusterv1.MachineDeployment, error) {
	machineDeployments := &clusterv1.MachineDeploymentList{}
	req, err := labels.NewRequirement(clusterv1.ClusterLabelName, selection.Equals, []string{c.Name})
	if err != nil {
		return nil, err
	}
	o := &client.ListOptions{LabelSelector: labels.NewSelector().Add(*req), Namespace: constants.EksaSystemNamespace}
	err = r.client.List(ctx, machineDeployments, o)
	if err != nil {
		return nil, err
	}
	deployments := make([]*clusterv1.MachineDeployment, 0, len(machineDeployments.Items))
	for _, md := range machineDeployments.Items {
		deployments = append(deployments, &md)
	}
	return deployments, nil
}

func (r *capiResourceFetcher) MachineDeployment(ctx context.Context, cs *anywherev1.Cluster) (*clusterv1.MachineDeployment, error) {
	deployments, err := r.machineDeployments(ctx, cs)
	if err != nil {
		return nil, err
	}

	if len(deployments) < 1 {
		return nil, fmt.Errorf("no machine deployments found for cluster %s", cs.Name)
	}

	return deployments[0], nil
}

func (r *capiResourceFetcher) Fetch(ctx context.Context, name string, namespace string, kind string, apiVersion string) (*unstructured.Unstructured, error) {
	us := &unstructured.Unstructured{}
	us.SetKind(kind)
	us.SetAPIVersion(apiVersion)
	key := client.ObjectKey{Name: name, Namespace: namespace}
	if err := r.client.Get(ctx, key, us); err != nil {
		return nil, err
	}
	return us, nil
}

func (r *capiResourceFetcher) VSphereMachineTemplate(ctx context.Context, cs *anywherev1.Cluster) (*vspherev3.VSphereMachineTemplate, error) {
	md, err := r.MachineDeployment(ctx, cs)
	if err != nil {
		return nil, err
	}
	vsphereMachineTemplate := &vspherev3.VSphereMachineTemplate{}
	err = r.FetchObjectByName(ctx, md.Spec.Template.Spec.InfrastructureRef.Name, constants.EksaSystemNamespace, vsphereMachineTemplate)
	if err != nil {
		return nil, err
	}
	return vsphereMachineTemplate, nil
}

func (r *capiResourceFetcher) clusterBundle(ctx context.Context, cs *anywherev1.Cluster) (*releasev1alpha1.Bundles, error) {
	clusterBundle := &releasev1alpha1.Bundles{}
	err := r.FetchObjectByName(ctx, cs.Name, cs.Namespace, clusterBundle)
	if err != nil {
		return nil, err
	}
	return clusterBundle, nil
}

func (r *capiResourceFetcher) ControlPlane(ctx context.Context, cs *anywherev1.Cluster) (*kubeadmnv1alpha3.KubeadmControlPlane, error) {
	// Fetch capi cluster
	capiCluster := &clusterv1.Cluster{}
	err := r.FetchObjectByName(ctx, cs.Name, constants.EksaSystemNamespace, capiCluster)
	if err != nil {
		return nil, err
	}
	cpRef := capiCluster.Spec.ControlPlaneRef
	cp := &kubeadmnv1alpha3.KubeadmControlPlane{}
	err = r.FetchObjectByName(ctx, cpRef.Name, cpRef.Namespace, cp)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

func (r *capiResourceFetcher) FetchAppliedSpec(ctx context.Context, cs *anywherev1.Cluster) (*cluster.Spec, error) {
	bundle, err := r.clusterBundle(ctx, cs)
	if err != nil {
		return nil, err
	}
	spec, err := cluster.BuildSpecFromBundles(cs, bundle)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

func (r *capiResourceFetcher) ExistingVSphereDatacenterConfig(ctx context.Context, cs *anywherev1.Cluster) (*anywherev1.VSphereDatacenterConfig, error) {
	vsMachineTemplate, err := r.VSphereMachineTemplate(ctx, cs)
	if err != nil {
		return nil, err
	}
	return MapMachineTemplateToVSphereDatacenterConfigSpec(vsMachineTemplate)
}

func (r *capiResourceFetcher) ExistingVSphereWorkerMachineConfig(ctx context.Context, cs *anywherev1.Cluster) (*anywherev1.VSphereMachineConfig, error) {
	vsMachineTemplate, err := r.VSphereMachineTemplate(ctx, cs)
	if err != nil {
		return nil, err
	}
	return MapMachineTemplateToVSphereWorkerMachineConfigSpec(vsMachineTemplate)
}

func MapMachineTemplateToVSphereDatacenterConfigSpec(vsMachineTemplate *vspherev3.VSphereMachineTemplate) (*anywherev1.VSphereDatacenterConfig, error) {
	vsSpec := &anywherev1.VSphereDatacenterConfig{}
	vsSpec.Spec.Thumbprint = vsMachineTemplate.Spec.Template.Spec.Thumbprint
	vsSpec.Spec.Server = vsMachineTemplate.Spec.Template.Spec.Server
	vsSpec.Spec.Datacenter = vsMachineTemplate.Spec.Template.Spec.Datacenter

	if len(vsMachineTemplate.Spec.Template.Spec.Network.Devices) == 0 {
		return nil, fmt.Errorf("networkName under devices not found on object %s", vsMachineTemplate.Kind)
	}
	vsSpec.Spec.Network = vsMachineTemplate.Spec.Template.Spec.Network.Devices[0].NetworkName

	return vsSpec, nil
}

func MapMachineTemplateToVSphereWorkerMachineConfigSpec(vsMachineTemplate *vspherev3.VSphereMachineTemplate) (*anywherev1.VSphereMachineConfig, error) {
	vsSpec := &anywherev1.VSphereMachineConfig{}
	vsSpec.Spec.MemoryMiB = int(vsMachineTemplate.Spec.Template.Spec.MemoryMiB)
	vsSpec.Spec.DiskGiB = int(vsMachineTemplate.Spec.Template.Spec.DiskGiB)
	vsSpec.Spec.NumCPUs = int(vsMachineTemplate.Spec.Template.Spec.NumCPUs)
	vsSpec.Spec.Template = vsMachineTemplate.Spec.Template.Spec.Template
	vsSpec.Spec.ResourcePool = vsMachineTemplate.Spec.Template.Spec.ResourcePool
	vsSpec.Spec.Datastore = vsMachineTemplate.Spec.Template.Spec.Datastore
	vsSpec.Spec.Folder = vsMachineTemplate.Spec.Template.Spec.Folder
	vsSpec.Spec.StoragePolicyName = vsMachineTemplate.Spec.Template.Spec.StoragePolicyName

	return vsSpec, nil
}
