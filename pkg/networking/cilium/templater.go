package cilium

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed network_policy.yaml
var networkPolicyAllowAll string

type Helm interface {
	Template(ctx context.Context, ociURI, version, namespace string, values interface{}) ([]byte, error)
}

type Templater struct {
	helm Helm
}

func NewTemplater(helm Helm) *Templater {
	return &Templater{
		helm: helm,
	}
}

func (c *Templater) GenerateUpgradePreflightManifest(ctx context.Context, spec *cluster.Spec) ([]byte, error) {
	v := templateValues(spec)
	v.set(true, "preflight", "enabled")
	v.set(spec.VersionsBundle.Cilium.Cilium.Image(), "preflight", "image", "repository")
	v.set(spec.VersionsBundle.Cilium.Cilium.Tag(), "preflight", "image", "tag")
	v.set(false, "agent")
	v.set(false, "operator", "enabled")

	uri, version := getChartUriAndVersion(spec)

	manifest, err := c.helm.Template(ctx, uri, version, namespace, v)
	if err != nil {
		return nil, fmt.Errorf("failed generating cilium upgrade preflight manifest: %v", err)
	}

	return manifest, nil
}

func (c *Templater) GenerateUpgradeManifest(ctx context.Context, currentSpec, newSpec *cluster.Spec) ([]byte, error) {
	currentVersion, err := semver.New(currentSpec.VersionsBundle.Cilium.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid version for Cilium in current spec: %v", err)
	}

	v := templateValues(newSpec)
	v.set(fmt.Sprintf("%d.%d", currentVersion.Major, currentVersion.Minor), "upgradeCompatibility")

	uri, version := getChartUriAndVersion(newSpec)

	manifest, err := c.helm.Template(ctx, uri, version, namespace, v)
	if err != nil {
		return nil, fmt.Errorf("failed generating cilium upgrade manifest: %v", err)
	}

	return manifest, nil
}

func (c *Templater) GenerateManifest(ctx context.Context, spec *cluster.Spec) ([]byte, error) {
	v := templateValues(spec)

	uri, version := getChartUriAndVersion(spec)

	manifest, err := c.helm.Template(ctx, uri, version, namespace, v)
	if err != nil {
		return nil, fmt.Errorf("failed generating cilium manifest: %v", err)
	}

	return manifest, nil
}

func (c *Templater) GenerateNetworkPolicyManifest(spec *cluster.Spec, namespaces []string) ([]byte, error) {
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

	/* k8s versions 1.21 and higher label each namespace with key `kubernetes.io/metadata.name:` and value is the namespace's name.
	This can be used to create a networkPolicy that allows traffic only between pods within kube-system ns, which is ideal for workload clusters. (not needed
	for mgmt clusters).
	So we will create networkPolicy using this default label as namespaceSelector for all versions 1.21 and higher
	For 1.20 we will create a networkPolicy that allows allow traffic to/from kube-system pods, and document this. Users can still modify it and add new policies
	as needed*/
	k8sVersion, err := semver.New(spec.VersionsBundle.KubeDistro.Kubernetes.Tag)
	if err != nil {
		return nil, fmt.Errorf("parsing kubernetes version %v: %v", spec.Cluster.Spec.KubernetesVersion, err)
	}
	if k8sVersion.Major == 1 && k8sVersion.Minor >= 21 {
		values["kubeSystemNSHasLabel"] = true
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

	if spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode != "" {
		val["policyEnforcementMode"] = spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode
	}
	return val
}

func getChartUriAndVersion(spec *cluster.Spec) (uri, version string) {
	chart := spec.VersionsBundle.Cilium.HelmChart
	uri = fmt.Sprintf("oci://%s", chart.Image())
	version = chart.Tag()
	return uri, version
}
