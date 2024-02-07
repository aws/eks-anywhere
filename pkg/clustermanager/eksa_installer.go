package clustermanager

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/exp/maps"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

// EKSAInstallerOpt updates an EKSAInstaller.
type EKSAInstallerOpt func(*EKSAInstaller)

// EKSAInstaller allows to install eks-a components in a cluster.
type EKSAInstaller struct {
	client                KubernetesClient
	reader                manifests.FileReader
	deploymentWaitTimeout time.Duration
}

// NewEKSAInstaller constructs a new EKSAInstaller.
func NewEKSAInstaller(client KubernetesClient, reader manifests.FileReader, opts ...EKSAInstallerOpt) *EKSAInstaller {
	i := &EKSAInstaller{
		client:                client,
		reader:                reader,
		deploymentWaitTimeout: DefaultDeploymentWait,
	}

	for _, o := range opts {
		o(i)
	}

	return i
}

// WithEKSAInstallerNoTimeouts disables the timeout when waiting for a deployment to be ready.
func WithEKSAInstallerNoTimeouts() EKSAInstallerOpt {
	return func(i *EKSAInstaller) {
		i.deploymentWaitTimeout = time.Duration(math.MaxInt64)
	}
}

// Install configures and applies eks-a components in a cluster accordingly to a spec.
func (i *EKSAInstaller) Install(ctx context.Context, log logr.Logger, cluster *types.Cluster, managementComponents *cluster.ManagementComponents, spec *cluster.Spec) error {
	if err := i.createEKSAComponents(ctx, log, cluster, managementComponents, spec); err != nil {
		return fmt.Errorf("applying EKSA components: %v", err)
	}

	if err := i.applyBundles(ctx, log, cluster, spec); err != nil {
		return fmt.Errorf("applying EKSA bundles: %v", err)
	}

	// We need to update this config map with the new upgrader images whenever we
	// apply a new Bundles object to the cluster in order to support in-place upgrades.
	cm, err := i.getUpgraderImagesFromBundle(ctx, cluster, spec)
	if err != nil {
		return fmt.Errorf("getting upgrader images from bundle: %v", err)
	}

	if err = i.client.Apply(ctx, cluster.KubeconfigFile, cm); err != nil {
		return fmt.Errorf("applying upgrader images config map: %v", err)
	}

	if err := i.applyReleases(ctx, log, cluster, spec); err != nil {
		return fmt.Errorf("applying EKSA releases: %v", err)
	}

	return nil
}

func (i *EKSAInstaller) getUpgraderImagesFromBundle(ctx context.Context, cluster *types.Cluster, cl *cluster.Spec) (*corev1.ConfigMap, error) {
	upgraderImages := make(map[string]string)
	for _, versionBundle := range cl.Bundles.Spec.VersionsBundles {
		eksD := versionBundle.EksD
		eksdVersion := fmt.Sprintf("%s-eks-%s-%s", eksD.KubeVersion, eksD.ReleaseChannel, strings.Split(eksD.Name, "-")[4])
		if _, ok := upgraderImages[eksdVersion]; !ok {
			upgraderImages[eksdVersion] = versionBundle.Upgrader.Upgrader.URI
		}
	}

	upgraderConfigMap, err := i.client.GetConfigMap(ctx, cluster.KubeconfigFile, constants.UpgraderConfigMapName, constants.EksaSystemNamespace)
	if err != nil {
		if executables.IsKubectlNotFoundError(err) {
			return newUpgraderConfigMap(upgraderImages), nil
		}
		return nil, err
	}

	for version, image := range upgraderImages {
		upgraderConfigMap.Data[version] = image
	}

	return upgraderConfigMap, nil
}

// Upgrade re-installs the eksa components in a cluster if the VersionBundle defined in the
// new spec has a different eks-a components version. Workload clusters are ignored.
func (i *EKSAInstaller) Upgrade(ctx context.Context, log logr.Logger, c *types.Cluster, currentManagementComponents, newManagementComponents *cluster.ManagementComponents, newSpec *cluster.Spec) (*types.ChangeDiff, error) {
	log.V(1).Info("Checking for EKS-A components upgrade")
	if !newSpec.Cluster.IsSelfManaged() {
		log.V(1).Info("Skipping EKS-A components upgrade, not a self-managed cluster")
		return nil, nil
	}
	changeDiff := EksaChangeDiff(currentManagementComponents, newManagementComponents)
	if changeDiff == nil {
		log.V(1).Info("Nothing to upgrade for controller and CRDs")
		return nil, nil
	}
	log.V(1).Info("Starting EKS-A components upgrade")
	oldVersion := currentManagementComponents.Eksa.Version
	newVersion := newManagementComponents.Eksa.Version
	if err := i.createEKSAComponents(ctx, log, c, newManagementComponents, newSpec); err != nil {
		return nil, fmt.Errorf("upgrading EKS-A components from version %v to version %v: %v", oldVersion, newVersion, err)
	}

	return changeDiff, nil
}

// createEKSAComponents creates eksa components and applies the objects to the cluster.
func (i *EKSAInstaller) createEKSAComponents(ctx context.Context, log logr.Logger, cluster *types.Cluster, managementComponents *cluster.ManagementComponents, spec *cluster.Spec) error {
	generator := EKSAComponentGenerator{log: log, reader: i.reader}
	components, err := generator.buildEKSAComponentsSpec(managementComponents, spec)
	if err != nil {
		return err
	}

	objs := make([]runtime.Object, 0, len(components.rest)+1)
	objs = append(objs, components.deployment)
	for _, o := range components.rest {
		objs = append(objs, o)
	}

	for _, o := range objs {
		if err = i.client.Apply(ctx, cluster.KubeconfigFile, o); err != nil {
			return fmt.Errorf("applying eksa components: %v", err)
		}
	}

	if err := i.client.WaitForDeployment(ctx, cluster, i.deploymentWaitTimeout.String(), "Available", constants.EksaControllerManagerDeployment, constants.EksaSystemNamespace); err != nil {
		return fmt.Errorf("waiting for eksa-controller-manager: %v", err)
	}

	return nil
}

// applyBundles applies the bundles to the cluster.
func (i *EKSAInstaller) applyBundles(ctx context.Context, log logr.Logger, cluster *types.Cluster, spec *cluster.Spec) error {
	bundleObj, err := yaml.Marshal(spec.Bundles)
	if err != nil {
		return fmt.Errorf("outputting bundle yaml: %v", err)
	}

	log.V(1).Info("Applying Bundles to cluster")
	if err := i.client.ApplyKubeSpecFromBytes(ctx, cluster, bundleObj); err != nil {
		return fmt.Errorf("applying bundle spec: %v", err)
	}

	return nil
}

// applyReleases applies the releases to the cluster.
func (i *EKSAInstaller) applyReleases(ctx context.Context, log logr.Logger, cluster *types.Cluster, spec *cluster.Spec) error {
	releaseObj, err := yaml.Marshal(spec.EKSARelease)
	if err != nil {
		return fmt.Errorf("outputting release yaml: %v", err)
	}

	log.V(1).Info("Applying EKSA Release to cluster")
	if err := i.client.ApplyKubeSpecFromBytes(ctx, cluster, releaseObj); err != nil {
		return fmt.Errorf("applying EKSA release spec: %v", err)
	}

	return nil
}

// EKSAComponentGenerator generates and configures eks-a components.
type EKSAComponentGenerator struct {
	log    logr.Logger
	reader manifests.FileReader
}

func (g *EKSAComponentGenerator) buildEKSAComponentsSpec(managamentComponents *cluster.ManagementComponents, spec *cluster.Spec) (*eksaComponents, error) {
	components, err := g.parseEKSAComponentsSpec(managamentComponents)
	if err != nil {
		return nil, err
	}

	g.configureEKSAComponents(components, spec)

	return components, nil
}

func (g *EKSAComponentGenerator) configureEKSAComponents(c *eksaComponents, spec *cluster.Spec) {
	// TODO(g-gaston): we should do this with a custom ControllerManagerConfig.
	// This requires wider changes in the controller manager setup and config manifest,
	// so leaving this for later.
	setManagerFlags(c.deployment, spec)
	setManagerEnvVars(c.deployment, spec)
}

func setManagerFlags(d *appsv1.Deployment, spec *cluster.Spec) {
	gates := []string{}
	for _, g := range managerEnabledGates(spec) {
		gates = append(gates, fmt.Sprintf("%s=true", g))
	}

	args := d.Spec.Template.Spec.Containers[0].Args
	if len(gates) > 0 {
		args = append(args, fmt.Sprintf("--feature-gates=%s", strings.Join(gates, ",")))
	}

	d.Spec.Template.Spec.Containers[0].Args = args
}

func setManagerEnvVars(d *appsv1.Deployment, spec *cluster.Spec) {
	envVars := d.Spec.Template.Spec.Containers[0].Env
	proxy := spec.Cluster.ProxyConfiguration()
	if proxy != nil {
		proxyEnvVarNames := maps.Keys(proxy)
		sort.Strings(proxyEnvVarNames)
		for _, name := range proxyEnvVarNames {
			envVars = append(envVars, v1.EnvVar{Name: name, Value: proxy[name]})
		}
	}

	// TODO: remove this feature flag if we decide to support in-place upgrades for vSphere provider.
	if features.IsActive(features.VSphereInPlaceUpgradeEnabled()) {
		envVars = append(envVars, v1.EnvVar{Name: features.VSphereInPlaceEnvVar, Value: "true"})
	}

	d.Spec.Template.Spec.Containers[0].Env = envVars
}

func managerEnabledGates(spec *cluster.Spec) []string {
	return nil
}

func fullLifeCycleControllerForProvider(cluster *anywherev1.Cluster) bool {
	// TODO(g-gaston): inject a dependency where this check can be delegated
	// We can use some kind of configurator registering that allow to decouple this generator
	// from the logic that drives the gates.
	return cluster.Spec.DatacenterRef.Kind == anywherev1.VSphereDatacenterKind ||
		cluster.Spec.DatacenterRef.Kind == anywherev1.DockerDatacenterKind ||
		cluster.Spec.DatacenterRef.Kind == anywherev1.SnowDatacenterKind ||
		cluster.Spec.DatacenterRef.Kind == anywherev1.NutanixDatacenterKind ||
		cluster.Spec.DatacenterRef.Kind == anywherev1.TinkerbellDatacenterKind ||
		cluster.Spec.DatacenterRef.Kind == anywherev1.CloudStackDatacenterKind
}

func (g *EKSAComponentGenerator) parseEKSAComponentsSpec(managementComponents *cluster.ManagementComponents) (*eksaComponents, error) {
	componentsManifest, err := bundles.ReadManifest(g.reader, managementComponents.Eksa.Components)
	if err != nil {
		return nil, fmt.Errorf("loading manifest for eksa components: %v", err)
	}

	p := yamlutil.NewParser(g.log)
	err = p.RegisterMappings(
		yamlutil.NewMapping(
			"Deployment", func() yamlutil.APIObject {
				return &appsv1.Deployment{}
			},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("registering yaml mappings for eksa components: %v", err)
	}
	p.RegisterMappingForAnyKind(func() yamlutil.APIObject {
		return &unstructured.Unstructured{}
	})

	components := &eksaComponents{}

	if err = p.Parse(componentsManifest.Content, components); err != nil {
		return nil, fmt.Errorf("parsing eksa components: %v", err)
	}

	return components, nil
}

type eksaComponents struct {
	deployment *appsv1.Deployment
	rest       []*unstructured.Unstructured
}

func (c *eksaComponents) BuildFromParsed(lookup yamlutil.ObjectLookup) error {
	for _, obj := range lookup {
		if obj.GetObjectKind().GroupVersionKind().Kind == "Deployment" {
			c.deployment = obj.(*appsv1.Deployment)
		} else {
			c.rest = append(c.rest, obj.(*unstructured.Unstructured))
		}
	}
	return nil
}

// EksaChangeDiff computes the version diff in eksa components between two specs.
func EksaChangeDiff(currentManagementComponents, newManagementComponents *cluster.ManagementComponents) *types.ChangeDiff {
	if currentManagementComponents.Eksa.Version != newManagementComponents.Eksa.Version {
		return &types.ChangeDiff{
			ComponentReports: []types.ComponentChangeDiff{
				{
					ComponentName: "EKS-A Management",
					NewVersion:    newManagementComponents.Eksa.Version,
					OldVersion:    currentManagementComponents.Eksa.Version,
				},
			},
		}
	}
	return nil
}
