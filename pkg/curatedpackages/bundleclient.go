package curatedpackages

import (
	"context"
	"fmt"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// BundleClient abstracts the retrieval of package bundle information.
type BundleClient interface {
	ActiveOrLatest(ctx context.Context) (*packagesv1.PackageBundle, error)
}

// BundleClientOptions alter the behavior of the bundle client when retrieving
// package bundle information.
type BundleClientOptions struct {
	// ClusterName identifies the cluster with which to communicate.
	ClusterName string
	// KubeConfig specifies a kubeconfig file to use.
	//
	// Defaults to the KUBECONFIG envrionment variable, and can fall back to
	// ~/.kube/config or similar.
	KubeConfig string
	// KubeVersion specifies a kubernetes version requirement for package bundles.
	//
	// Required if consulting a registry for package bundle information.
	KubeVersion string
	// Registry as an OCI Registry URL.
	//
	// Defaults to the EKS-A ECR. Used when consulting a registry for package
	// bundle information.
	Registry string
	// RegistryClient will be used if supplied.
	//
	// Specifying a custom registry client is optional, but could be needed
	// when specifying a custom registry mirror that requires authentication.
	RegistryClient bundle.RegistryClient
	// RESTConfigurator abstracts construction of a *rest.Config, used in Kube
	// client initialization.
	RESTConfigurator RESTConfigurator
}

// NewBundleClient constructs a BundleClient based on the value of source.
func NewBundleClient(source BundleSource, opts BundleClientOptions) (BundleClient, error) {
	switch source {
	case Registry:
		rc := opts.RegistryClient
		if rc == nil {
			rc = bundle.NewRegistryClient(artifacts.NewRegistryPuller())
		}
		kubeVersion := opts.KubeVersion
		if kubeVersion == "" {
			return nil, fmt.Errorf("kube version is required when source is registry")
		}
		registry := opts.Registry
		if registry == "" {
			registry = "public.ecr.aws/q0f6t3x4" // TODO pull this from the EKS-A bundle
		}
		return NewRegistryBundleClient(rc, registry, kubeVersion), nil
	case Cluster:
		kubeClient, err := NewKubeClientFromFilename(opts.KubeConfig, opts.RESTConfigurator)
		if err != nil {
			return nil, err
		}
		return NewClusterBundleClient(kubeClient, opts.ClusterName), nil
	default:
		return nil, fmt.Errorf("unhandled package bundle source %q", source)
	}
}

// clusterBundleClient implements BundleClient by querying an existing cluster.
type clusterBundleClient struct {
	kubeClient  client.Client
	clusterName string
}

var _ BundleClient = (*clusterBundleClient)(nil)

// NewClusterBundleClient retrieves package bundle info from an existing cluster.
func NewClusterBundleClient(kubeClient client.Client, clusterName string) *clusterBundleClient {
	return &clusterBundleClient{kubeClient: kubeClient, clusterName: clusterName}
}

// ActiveOrLatest implements BundleClient.
func (c *clusterBundleClient) ActiveOrLatest(ctx context.Context) (*packagesv1.PackageBundle, error) {
	pbc, err := c.ActivePackageBundleController(ctx)
	if err != nil {
		return nil, err
	}
	packageBundle := &packagesv1.PackageBundle{}
	packageBundleKey := client.ObjectKey{
		Namespace: packagesv1.PackageNamespace,
		Name:      pbc.Spec.ActiveBundle,
	}
	err = c.kubeClient.Get(ctx, packageBundleKey, packageBundle)
	if err != nil {
		return nil, fmt.Errorf("getting active package bundle %q: %w", packageBundleKey, err)
	}

	return packageBundle, nil
}

func (c *clusterBundleClient) ActivePackageBundleController(ctx context.Context) (*packagesv1.PackageBundleController, error) {
	pbc := &packagesv1.PackageBundleController{}
	pbcKey := client.ObjectKey{
		Namespace: constants.EksaPackagesName,
		Name:      c.clusterName,
	}
	err := c.kubeClient.Get(ctx, pbcKey, pbc)
	if err != nil {
		return nil, fmt.Errorf("getting package bundle controller: %w", err)
	}

	return pbc, nil
}

func (c *clusterBundleClient) UpgradeBundle(ctx context.Context, newVersion string) error {
	if newVersion == "" {
		return fmt.Errorf("no bundle version specified")
	}

	pbc, err := c.ActivePackageBundleController(ctx)
	if err != nil {
		return err
	}

	if strings.EqualFold(newVersion, pbc.Spec.ActiveBundle) {
		return fmt.Errorf("version %q is already active", newVersion)
	}
	pbc.Spec.ActiveBundle = newVersion

	err = c.kubeClient.Update(ctx, pbc)
	if err != nil {
		return fmt.Errorf("updating package bundle version: %w", err)
	}

	return nil
}

// registryBundleClient implements BundleClient by querying an OCI registry.
type registryBundleClient struct {
	regClient   bundle.RegistryClient
	kubeVersion string
	registry    string
}

var _ BundleClient = (*registryBundleClient)(nil)

// NewRegistryBundleClient retrieves package bundle info from an OCI registry.
func NewRegistryBundleClient(registryClient bundle.RegistryClient, registry, kubeVersion string) *registryBundleClient {
	return &registryBundleClient{
		regClient:   registryClient,
		kubeVersion: kubeVersion,
		registry:    registry,
	}
}

// ActiveOrLatest implements BundleClient.
func (c *registryBundleClient) ActiveOrLatest(ctx context.Context) (*packagesv1.PackageBundle, error) {
	return c.regClient.LatestBundle(ctx, c.registry, c.kubeVersion)
}
