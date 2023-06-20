package cilium

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"strings"
	"time"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed network_policy.yaml
var networkPolicyAllowAll string

const (
	maxRetries           = 10
	defaultBackOffPeriod = 5 * time.Second
)

type Helm interface {
	Template(ctx context.Context, ociURI, version, namespace string, values interface{}, kubeVersion string) ([]byte, error)
	RegistryLogin(ctx context.Context, registry, username, password string) error
}

type Templater struct {
	helm Helm
}

func NewTemplater(helm Helm) *Templater {
	return &Templater{
		helm: helm,
	}
}

func (t *Templater) GenerateUpgradePreflightManifest(ctx context.Context, spec *cluster.Spec) ([]byte, error) {
	v := templateValues(spec)
	v.set(true, "preflight", "enabled")
	v.set(spec.VersionsBundle.Cilium.Cilium.Image(), "preflight", "image", "repository")
	v.set(spec.VersionsBundle.Cilium.Cilium.Tag(), "preflight", "image", "tag")
	v.set(false, "agent")
	v.set(false, "operator", "enabled")

	uri, version := getChartUriAndVersion(spec)

	kubeVersion, err := getKubeVersionString(spec)
	if err != nil {
		return nil, err
	}

	manifest, err := t.helm.Template(ctx, uri, version, namespace, v, kubeVersion)
	if err != nil {
		return nil, fmt.Errorf("failed generating cilium upgrade preflight manifest: %v", err)
	}

	return manifest, nil
}

// ManifestOpt allows to modify options for a cilium manifest.
type ManifestOpt func(*ManifestConfig)

type ManifestConfig struct {
	values      values
	retrier     *retrier.Retrier
	kubeVersion string
	namespaces  []string
}

// WithKubeVersion allows to generate the Cilium manifest for a different kubernetes version
// than the one specified in the cluster spec. Useful for upgrades scenarios where Cilium is upgraded before
// the kubernetes components.
func WithKubeVersion(kubeVersion string) ManifestOpt {
	return func(c *ManifestConfig) {
		c.kubeVersion = kubeVersion
	}
}

// WithRetrier introduced for optimizing unit tests.
func WithRetrier(retrier *retrier.Retrier) ManifestOpt {
	return func(c *ManifestConfig) {
		c.retrier = retrier
	}
}

// WithUpgradeFromVersion allows to specify the compatibility Cilium version to use in the manifest.
// This is necessary for Cilium upgrades.
func WithUpgradeFromVersion(version semver.Version) ManifestOpt {
	return func(c *ManifestConfig) {
		c.values.set(fmt.Sprintf("%d.%d", version.Major, version.Minor), "upgradeCompatibility")
	}
}

// WithPolicyAllowedNamespaces allows to specify which namespaces traffic should be allowed when using
// and "Always" policy enforcement mode.
func WithPolicyAllowedNamespaces(namespaces []string) ManifestOpt {
	return func(c *ManifestConfig) {
		c.namespaces = namespaces
	}
}

func (t *Templater) GenerateManifest(ctx context.Context, spec *cluster.Spec, opts ...ManifestOpt) ([]byte, error) {
	kubeVersion, err := getKubeVersionString(spec)
	if err != nil {
		return nil, err
	}

	c := &ManifestConfig{
		values:      templateValues(spec),
		kubeVersion: kubeVersion,
		retrier:     retrier.NewWithMaxRetries(maxRetries, defaultBackOffPeriod),
	}
	for _, o := range opts {
		o(c)
	}

	uri, version := getChartUriAndVersion(spec)
	var manifest []byte

	if spec.Cluster.Spec.RegistryMirrorConfiguration != nil {
		if spec.Cluster.Spec.RegistryMirrorConfiguration.Authenticate {
			username, password, err := config.ReadCredentials()
			if err != nil {
				return nil, err
			}
			endpoint := net.JoinHostPort(spec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint, spec.Cluster.Spec.RegistryMirrorConfiguration.Port)
			if err := t.helm.RegistryLogin(ctx, endpoint, username, password); err != nil {
				return nil, err
			}
		}
	}

	err = c.retrier.Retry(func() error {
		manifest, err = t.helm.Template(ctx, uri, version, namespace, c.values, c.kubeVersion)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed generating cilium manifest: %v", err)
	}

	if spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode == anywherev1.CiliumPolicyModeAlways {
		networkPolicyManifest, err := t.GenerateNetworkPolicyManifest(spec, c.namespaces)
		if err != nil {
			return nil, err
		}
		manifest = templater.AppendYamlResources(manifest, networkPolicyManifest)
	}

	return manifest, nil
}

func (t *Templater) GenerateNetworkPolicyManifest(spec *cluster.Spec, namespaces []string) ([]byte, error) {
	values := map[string]interface{}{
		"managementCluster":  spec.Cluster.IsSelfManaged(),
		"providerNamespaces": namespaces,
	}

	if spec.Cluster.Spec.GitOpsRef != nil {
		values["gitopsEnabled"] = true
		if spec.GitOpsConfig != nil {
			values["fluxNamespace"] = spec.GitOpsConfig.Spec.Flux.Github.FluxSystemNamespace
		}
	}

	return templater.Execute(networkPolicyAllowAll, values)
}

type values map[string]interface{}

func (c values) set(value interface{}, path ...string) {
	element := c
	for _, p := range path[:len(path)-1] {
		e, ok := element[p]
		if !ok {
			e = values{}
			element[p] = e
		}
		element = e.(values)
	}
	element[path[len(path)-1]] = value
}

func templateValues(spec *cluster.Spec) values {
	val := values{
		"cni": values{
			"chainingMode": "portmap",
		},
		"ipam": values{
			"mode": "kubernetes",
		},
		"identityAllocationMode": "crd",
		"prometheus": values{
			"enabled": true,
		},
		"rollOutCiliumPods": true,
		"tunnel":            "geneve",
		"image": values{
			"repository": spec.VersionsBundle.Cilium.Cilium.Image(),
			"tag":        spec.VersionsBundle.Cilium.Cilium.Tag(),
		},
		"operator": values{
			"image": values{
				// The chart expects an "incomplete" repository
				// and will add the necessary suffix ("-generic" in our case)
				"repository": strings.TrimSuffix(spec.VersionsBundle.Cilium.Operator.Image(), "-generic"),
				"tag":        spec.VersionsBundle.Cilium.Operator.Tag(),
			},
			"prometheus": values{
				"enabled": true,
			},
		},
	}

	if len(spec.Cluster.Spec.WorkerNodeGroupConfigurations) == 0 && spec.Cluster.Spec.ControlPlaneConfiguration.Count == 1 {
		val["operator"].(values)["replicas"] = 1
	}

	if spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode != "" {
		val["policyEnforcementMode"] = spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode
	}

	if spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.EgressMasqueradeInterfaces != "" {
		val["egressMasqueradeInterfaces"] = spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.EgressMasqueradeInterfaces
	}

	return val
}

func getChartUriAndVersion(spec *cluster.Spec) (uri, version string) {
	chart := spec.VersionsBundle.Cilium.HelmChart
	uri = fmt.Sprintf("oci://%s", chart.Image())
	version = chart.Tag()
	return uri, version
}

func getKubeVersion(spec *cluster.Spec) (*semver.Version, error) {
	k8sVersion, err := semver.New(spec.VersionsBundle.KubeDistro.Kubernetes.Tag)
	if err != nil {
		return nil, fmt.Errorf("parsing kubernetes version %v: %v", spec.Cluster.Spec.KubernetesVersion, err)
	}
	return k8sVersion, nil
}

func getKubeVersionString(spec *cluster.Spec) (string, error) {
	k8sVersion, err := getKubeVersion(spec)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.%d", k8sVersion.Major, k8sVersion.Minor), nil
}
