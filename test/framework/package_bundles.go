package framework

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
)

// PackagesUpgradeFromBundleVar picks the package bundle the package upgrade tests install packages from before
// upgrading to the newest bundle:
//   - "prod": the current prod bundle for the cluster's Kubernetes version, the bundle customers upgrade from.
//   - otherwise, a bundle name in the PackageBundleController bundle repository or a full repository:tag reference.
//
// The tests skip when it's unset. Set it only where the cluster pulls from the staging or prod registries: switching
// bundles also replaces the packages controller, and dev registries don't have the controller images of most older
// bundles, so the controller wouldn't come back.
const PackagesUpgradeFromBundleVar = "T_PACKAGES_UPGRADE_FROM_BUNDLE"

// prodPackageBundlesRepository is where released package bundles are published.
const prodPackageBundlesRepository = "public.ecr.aws/eks-anywhere/eks-anywhere-packages-bundles"

var ecrPrivateRegistryRegex = regexp.MustCompile(`^\d{12}\.dkr\.ecr\.([a-z0-9-]+)\.amazonaws\.com$`)

// WaitForActivePackageBundle waits for the cluster's PackageBundleController to have an active
// bundle and returns it. Right after cluster creation this is the newest bundle the controller found.
func (e *ClusterE2ETest) WaitForActivePackageBundle() string {
	ctx := context.Background()
	var active string
	err := retrier.New(10*time.Minute, retrier.WithRetryPolicy(retrier.BackOffPolicy(10*time.Second))).Retry(func() error {
		pbc, err := e.KubectlClient.GetPackageBundleController(ctx, e.KubeconfigFilePath(), e.ClusterName)
		if err != nil {
			return err
		}
		if pbc.Spec.ActiveBundle == "" {
			return fmt.Errorf("PackageBundleController %s has no active bundle yet", e.ClusterName)
		}
		active = pbc.Spec.ActiveBundle
		return nil
	})
	if err != nil {
		e.T.Fatalf("Waiting for an active package bundle: %v", err)
	}
	return active
}

// CheckPackagesUpgradeFromBundle skips the test if T_PACKAGES_UPGRADE_FROM_BUNDLE isn't set. Call it before
// creating the cluster.
func CheckPackagesUpgradeFromBundle(t *testing.T) {
	if os.Getenv(PackagesUpgradeFromBundleVar) == "" {
		t.Skipf("%s isn't set. Set it to \"prod\" (or a bundle) on clusters that pull from the staging or prod registries.",
			PackagesUpgradeFromBundleVar)
	}
}

// PackageBundleToUpgradeFrom applies the bundle picked by T_PACKAGES_UPGRADE_FROM_BUNDLE, which the package
// upgrade tests install packages from before upgrading them to newest, and returns its name.
func (e *ClusterE2ETest) PackageBundleToUpgradeFrom(newest string) string {
	ctx := context.Background()
	from := os.Getenv(PackagesUpgradeFromBundleVar)
	if from == "" {
		e.T.Fatalf("%s is required", PackagesUpgradeFromBundleVar)
	}
	if from == "prod" {
		kubeVersion := strings.ReplaceAll(string(e.ClusterConfig.Cluster.Spec.KubernetesVersion), ".", "-")
		return e.applyUpgradeFromBundle(ctx, prodPackageBundlesRepository, fmt.Sprintf("v%s-latest", kubeVersion), newest)
	}
	if i := strings.LastIndex(from, ":"); i > 0 && strings.Contains(from, "/") {
		return e.applyUpgradeFromBundle(ctx, from[:i], from[i+1:], newest)
	}
	return e.applyUpgradeFromBundle(ctx, e.packageBundleRepository(ctx), from, newest)
}

// applyUpgradeFromBundle pulls the bundle at repository:tag, applies it to the cluster and returns its name.
func (e *ClusterE2ETest) applyUpgradeFromBundle(ctx context.Context, repository, tag, newest string) string {
	e.T.Logf("Pulling the package bundle to upgrade from, %s:%s", repository, tag)
	data, err := pullPackageBundle(ctx, repository, tag)
	if err != nil {
		e.T.Fatalf("Pulling package bundle %s:%s: %v", repository, tag, err)
	}
	bundle := &packagesv1.PackageBundle{}
	if err := yaml.Unmarshal(data, bundle); err != nil {
		e.T.Fatalf("Parsing package bundle %s:%s: %v", repository, tag, err)
	}
	if bundle.Name == newest {
		e.T.Logf("Package bundle %s is already the newest bundle, so packages won't change; only checking their state survives", newest)
		return newest
	}
	e.applyPackageBundle(ctx, bundle.Name, data)
	e.T.Logf("Upgrading packages from package bundle %s to %s", bundle.Name, newest)
	return bundle.Name
}

func (e *ClusterE2ETest) packageBundleRepository(ctx context.Context) string {
	pbc, err := e.KubectlClient.GetPackageBundleController(ctx, e.KubeconfigFilePath(), e.ClusterName)
	if err != nil {
		e.T.Fatalf("Getting PackageBundleController: %v", err)
	}
	return pbc.GetBundleURI()
}

// ActivatePackageBundle sets bundle as the cluster's active package bundle. If the bundle isn't in the
// cluster yet, it's pulled from the PackageBundleController bundle repository and applied first, which is
// the documented way to activate a bundle other than the newest one.
func (e *ClusterE2ETest) ActivatePackageBundle(bundle string) {
	ctx := context.Background()
	kubeconfig := e.KubeconfigFilePath()

	bundles, err := e.KubectlClient.GetPackageBundleList(ctx, kubeconfig)
	if err != nil {
		e.T.Fatalf("Getting package bundles: %v", err)
	}
	if findPackageBundle(bundles, bundle) == nil {
		pbc, err := e.KubectlClient.GetPackageBundleController(ctx, kubeconfig, e.ClusterName)
		if err != nil {
			e.T.Fatalf("Getting PackageBundleController: %v", err)
		}
		e.T.Logf("Pulling package bundle %s from %s", bundle, pbc.GetBundleURI())
		data, err := pullPackageBundle(ctx, pbc.GetBundleURI(), bundle)
		if err != nil {
			e.T.Fatalf("Pulling package bundle %s: %v", bundle, err)
		}
		e.applyPackageBundle(ctx, bundle, data)
	}

	// The PackageBundleController webhook rejects bundles it doesn't see yet, so retry until its cache has it.
	e.T.Logf("Activating package bundle %s", bundle)
	patch := fmt.Sprintf(`{"spec":{"activeBundle":%q}}`, bundle)
	err = retrier.New(2*time.Minute, retrier.WithRetryPolicy(retrier.BackOffPolicy(5*time.Second))).Retry(func() error {
		_, err := e.KubectlClient.ExecuteCommand(ctx, "patch", "packagebundlecontrollers.packages.eks.amazonaws.com", e.ClusterName,
			"--namespace", constants.EksaPackagesName, "--type", "merge", "-p", patch, "--kubeconfig", kubeconfig)
		return err
	})
	if err != nil {
		e.T.Fatalf("Activating package bundle %s: %v", bundle, err)
	}
}

// applyPackageBundle applies a bundle.yaml to the packages namespace.
func (e *ClusterE2ETest) applyPackageBundle(ctx context.Context, name string, data []byte) {
	if _, err := e.KubectlClient.ExecuteFromYaml(ctx, data, "apply", "-f", "-",
		"--namespace", constants.EksaPackagesName, "--kubeconfig", e.KubeconfigFilePath()); err != nil {
		e.T.Fatalf("Applying package bundle %s: %v", name, err)
	}
}

// WaitForPackagesControllerAtBundle waits for the packages controller, which manages itself as the
// eks-anywhere-packages package, to be installed from bundle and rolled out, so its webhooks serve
// requests before packages are created or changed.
func (e *ClusterE2ETest) WaitForPackagesControllerAtBundle(bundle string) {
	version := e.PackageBundleVersion(bundle, EksaPackagesInstallation)
	if err := e.waitForPackageInstalledAtVersion(EksaPackagesInstallation, version, 15*time.Minute); err != nil {
		e.T.Fatalf("Waiting for the packages controller to be installed at version %s from bundle %s: %v. If this cluster "+
			"pulls from the dev registries, which don't have the controller images of older bundles, don't set %s.",
			version.Name, bundle, err, PackagesUpgradeFromBundleVar)
	}
	if err := e.KubectlClient.WaitForResourceRolledout(context.Background(), e.Cluster(), "10m",
		EksaPackagesInstallation, constants.EksaPackagesName, "deployment"); err != nil {
		e.T.Fatalf("Waiting for the packages controller rollout: %v", err)
	}
}

// WaitForPackageBundleActive waits for the PackageBundleController to use bundle. When bundle isn't the newest,
// the controller reports "upgrade available", which also counts as active.
func (e *ClusterE2ETest) WaitForPackageBundleActive(bundle string) {
	ctx := context.Background()
	err := retrier.New(5*time.Minute, retrier.WithRetryPolicy(retrier.BackOffPolicy(10*time.Second))).Retry(func() error {
		pbc, err := e.KubectlClient.GetPackageBundleController(ctx, e.KubeconfigFilePath(), e.ClusterName)
		if err != nil {
			return err
		}
		if pbc.Spec.ActiveBundle != bundle {
			return fmt.Errorf("PackageBundleController %s has active bundle %s", e.ClusterName, pbc.Spec.ActiveBundle)
		}
		switch pbc.Status.State {
		case packagesv1.BundleControllerStateActive, packagesv1.BundleControllerStateUpgradeAvailable:
			return nil
		}
		return fmt.Errorf("PackageBundleController %s is %q: %s", e.ClusterName, pbc.Status.State, pbc.Status.Detail)
	})
	if err != nil {
		e.T.Fatalf("Waiting for package bundle %s to be active: %v", bundle, err)
	}
}

// PackageBundleVersion returns the version of packageName that bundle installs by default.
func (e *ClusterE2ETest) PackageBundleVersion(bundle, packageName string) packagesv1.SourceVersion {
	bundles, err := e.KubectlClient.GetPackageBundleList(context.Background(), e.KubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("Getting package bundles: %v", err)
	}
	b := findPackageBundle(bundles, bundle)
	if b == nil {
		e.T.Fatalf("Package bundle %s not found in the cluster", bundle)
	}
	pkg, err := b.FindPackage(packageName)
	if err != nil {
		e.T.Fatalf("Finding package %s: %v", packageName, err)
	}
	version, err := b.FindVersion(pkg, packagesv1.Latest)
	if err != nil {
		e.T.Fatalf("Finding the version of package %s: %v", packageName, err)
	}
	return version
}

// WaitForPackageInstalledAtVersion waits for the Package packageName to be installed from version's chart. It
// matches the chart digest, not the version name: bundles can ship a new chart under the same version name, and
// the same chart under different names (dev and prod bundles name the same controller chart differently).
func (e *ClusterE2ETest) WaitForPackageInstalledAtVersion(packageName string, version packagesv1.SourceVersion, timeout time.Duration) {
	if err := e.waitForPackageInstalledAtVersion(packageName, version, timeout); err != nil {
		e.T.Fatalf("Waiting for package %s to be installed at version %s: %v", packageName, version.Name, err)
	}
}

func (e *ClusterE2ETest) waitForPackageInstalledAtVersion(packageName string, version packagesv1.SourceVersion, timeout time.Duration) error {
	ctx := context.Background()
	namespace := fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName)
	e.T.Logf("Waiting for package %s to be installed at version %s (%s)", packageName, version.Name, version.Digest)
	return retrier.New(timeout, retrier.WithRetryPolicy(retrier.BackOffPolicy(15*time.Second))).Retry(func() error {
		pkg := &packagesv1.Package{}
		if err := e.KubectlClient.GetObject(ctx, "packages.packages.eks.amazonaws.com", packageName, namespace, e.KubeconfigFilePath(), pkg); err != nil {
			return err
		}
		if pkg.Status.State != packagesv1.StateInstalled || pkg.Status.Source.Digest != version.Digest {
			return fmt.Errorf("package %s is %q at version %q (%s), detail: %q",
				packageName, pkg.Status.State, pkg.Status.CurrentVersion, pkg.Status.Source.Digest, pkg.Status.Detail)
		}
		return nil
	})
}

func findPackageBundle(bundles []packagesv1.PackageBundle, name string) *packagesv1.PackageBundle {
	for i := range bundles {
		if bundles[i].Name == name {
			return &bundles[i]
		}
	}
	return nil
}

// registryRepository returns a client for an OCI repository. Private ECR registries are accessed with the
// curated packages test credentials (EKSA_AWS_*), other registries anonymously.
func registryRepository(repository string) (*remote.Repository, error) {
	repo, err := remote.NewRepository(repository)
	if err != nil {
		return nil, err
	}
	client := &auth.Client{Client: retry.DefaultClient, Cache: auth.NewCache()}
	host := repo.Reference.Registry
	if m := ecrPrivateRegistryRegex.FindStringSubmatch(host); m != nil {
		username, password, err := ecrBasicAuth(m[1], os.Getenv(eksaPackagesAccessKey), os.Getenv(eksaPackagesSecretKey), os.Getenv(eksaPackagesSessionTokenKey))
		if err != nil {
			return nil, err
		}
		client.Credential = auth.StaticCredential(host, auth.Credential{Username: username, Password: password})
	}
	repo.Client = client
	return repo, nil
}

// pullPackageBundle returns the bundle.yaml of a package bundle artifact, like `oras pull` does.
func pullPackageBundle(ctx context.Context, repository, tag string) ([]byte, error) {
	repo, err := registryRepository(repository)
	if err != nil {
		return nil, err
	}
	desc, rc, err := repo.FetchReference(ctx, tag)
	if err != nil {
		return nil, fmt.Errorf("fetching manifest: %v", err)
	}
	defer rc.Close()
	manifestData, err := content.ReadAll(rc, desc)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %v", err)
	}
	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("parsing manifest: %v", err)
	}
	if len(manifest.Layers) == 0 {
		return nil, fmt.Errorf("bundle artifact %s:%s has no layers", repository, tag)
	}
	return content.FetchAll(ctx, repo, manifest.Layers[0])
}

//go:embed testdata/packages_bundle_upgrade.yaml
var packagesBundleUpgradeTemplate string

// BundleUpgradeTestPackages names the packages, and the state in them, that the package bundle upgrade test
// installs from an older bundle with ApplyBundleUpgradeTestPackages.
type BundleUpgradeTestPackages struct {
	PrometheusName      string
	PrometheusNamespace string
	PrometheusScrapeJob string
	MetalLBName         string
	MetalLBNamespace    string
	MetalLBPool         string
	MetalLBAddresses    string
	HelloName           string
	HelloNamespace      string
	HelloSubtitle       string
	EmissaryName        string
	EmissaryNamespace   string
	AdotName            string
	AdotNamespace       string
}

// ApplyBundleUpgradeTestPackages applies the Prometheus, MetalLB, hello-eks-anywhere, Emissary and ADOT
// packages used by the package bundle upgrade test.
func (e *ClusterE2ETest) ApplyBundleUpgradeTestPackages(p BundleUpgradeTestPackages) {
	manifest, err := templater.Execute(packagesBundleUpgradeTemplate, map[string]interface{}{
		"prometheusName":      p.PrometheusName,
		"prometheusNamespace": p.PrometheusNamespace,
		"prometheusScrapeJob": p.PrometheusScrapeJob,
		"metallbName":         p.MetalLBName,
		"metallbNamespace":    p.MetalLBNamespace,
		"metallbPool":         p.MetalLBPool,
		"metallbAddresses":    p.MetalLBAddresses,
		"helloName":           p.HelloName,
		"helloNamespace":      p.HelloNamespace,
		"helloSubtitle":       p.HelloSubtitle,
		"emissaryName":        p.EmissaryName,
		"emissaryNamespace":   p.EmissaryNamespace,
		"adotName":            p.AdotName,
		"adotNamespace":       p.AdotNamespace,
	})
	if err != nil {
		e.T.Fatalf("Rendering bundle upgrade test packages: %v", err)
	}
	e.ApplyPackageFile("bundle upgrade test packages", "", manifest)
}
