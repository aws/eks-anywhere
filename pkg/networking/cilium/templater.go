package cilium

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/semver"
)

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
	v.set(false, "agent")
	v.set(false, "operator", "enabled")

	manifest, err := c.helm.Template(ctx, "placeholder", spec.VersionsBundle.Cilium.Version, namespace, v) // TODO: use real oci uri
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

	manifest, err := c.helm.Template(ctx, "placeholder", newSpec.VersionsBundle.Cilium.Version, namespace, v) // TODO: use real oci uri
	if err != nil {
		return nil, fmt.Errorf("failed generating cilium upgrade manifest: %v", err)
	}

	return manifest, nil
}

func (c *Templater) GenerateManifest(ctx context.Context, spec *cluster.Spec) ([]byte, error) {
	v := templateValues(spec)

	manifest, err := c.helm.Template(ctx, "placeholder", spec.VersionsBundle.Cilium.Version, namespace, v) // TODO: use real oci uri
	if err != nil {
		return nil, fmt.Errorf("failed generating cilium manifest: %v", err)
	}

	return manifest, nil
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
	return values{
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
				"repository": spec.VersionsBundle.Cilium.Operator.Image(),
				"tag":        spec.VersionsBundle.Cilium.Operator.Tag(),
			},
			"prometheus": values{
				"enabled": true,
			},
		},
	}
}
