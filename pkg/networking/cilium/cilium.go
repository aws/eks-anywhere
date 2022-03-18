package cilium

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/templater"
)

const namespace = constants.KubeSystemNamespace

//go:embed network_policy.yaml
var networkPolicyAllowAll string

type Cilium struct {
	*Upgrader
}

func NewCilium(client Client, helm Helm) *Cilium {
	return &Cilium{
		Upgrader: NewUpgrader(client, helm),
	}
}

func (c *Cilium) GenerateManifest(ctx context.Context, clusterSpec *cluster.Spec, provider providers.Provider) ([]byte, error) {
	ciliumManifest, err := c.templater.GenerateManifest(ctx, clusterSpec)
	if err != nil {
		return nil, err
	}
	if clusterSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode != "always" {
		return ciliumManifest, nil
	}

	// if always mode, create NetworkPolicy objects for kube-system, EKS-A and CAPI components as required
	values := map[string]interface{}{
		"managementCluster":  clusterSpec.Cluster.IsSelfManaged(),
		"providerNamespaces": provider.GetDeployments(),
		"gitopsEnabled":      clusterSpec.Cluster.Spec.GitOpsRef != nil,
	}

	/* k8s 1.21 and up labels each namespace with key `kubernetes.io/metadata.name:` and value is the namespace's name.
	This can be used to create a networkPolicy that allows traffic only between pods within kube-system ns, which is ideal for workload clusters. (not needed
	for mgmt clusters).
	So we will create networkPolicy using this default label as namespaceSelector for all versions 1.21 and higher
	For 1.20 we will create a networkPolicy that allows allow traffic to/from kube-system pods, and document this. Users can still modify it and add new policies
	as needed*/
	k8sVersion, err := semver.New(clusterSpec.VersionsBundle.KubeDistro.Kubernetes.Tag)
	if err != nil {
		return nil, fmt.Errorf("error parsing kubernetes version %v: %v", clusterSpec.Cluster.Spec.KubernetesVersion, err)
	}
	if k8sVersion.Major == 1 && k8sVersion.Minor >= 21 {
		values["kubeSystemNSHasLabel"] = true
	}

	content, err := templater.Execute(networkPolicyAllowAll, values)
	if err != nil {
		return nil, err
	}
	return templater.AppendYamlResources(ciliumManifest, content), nil
}
