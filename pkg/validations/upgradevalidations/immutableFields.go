package upgradevalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
)

func ValidateImmutableFields(ctx context.Context, k ValidationsKubectlClient, cluster *types.Cluster, spec *cluster.Spec, provider providers.Provider) error {
	prevSpec, err := k.GetEksaCluster(ctx, cluster)
	if err != nil {
		return err
	}

	if prevSpec.Name != spec.Name {
		return fmt.Errorf("cluster name is immutable")
	}

	if prevSpec.Namespace != spec.Namespace {
		return fmt.Errorf("cluster namespace is immutable")
	}

	oSpec := prevSpec.Spec
	nSpec := spec.Spec

	if !nSpec.DatacenterRef.Equal(&oSpec.DatacenterRef) {
		return fmt.Errorf("spec.dataCenterRef.name is immutable")
	}

	if !nSpec.GitOpsRef.Equal(oSpec.GitOpsRef) {
		return fmt.Errorf("spec.gitOpsRef is immutable")
	}

	if nSpec.GitOpsRef != nil {
		prevGitOps, err := k.GetEksaGitOpsConfig(ctx, nSpec.GitOpsRef.Name, cluster.KubeconfigFile, spec.Namespace)
		if err != nil {
			return err
		}
		if !prevGitOps.Spec.Equal(&spec.GitOpsConfig.Spec) {
			return fmt.Errorf("gitOps is immutable")
		}
	}

	if !nSpec.ControlPlaneConfiguration.Endpoint.Equal(oSpec.ControlPlaneConfiguration.Endpoint) {
		return fmt.Errorf("spec.controlPlaneConfiguration.endpoint is immutable")
	}

	if !nSpec.ClusterNetwork.Equal(&oSpec.ClusterNetwork) {
		return fmt.Errorf("spec.clusterNetwork is immutable")
	}

	if !nSpec.ProxyConfiguration.Equal(oSpec.ProxyConfiguration) {
		return fmt.Errorf("spec.proxyConfiguration is immutable")
	}

	oldETCD := oSpec.ExternalEtcdConfiguration
	newETCD := nSpec.ExternalEtcdConfiguration
	if oldETCD != nil && newETCD != nil {
		if oldETCD.Count != newETCD.Count {
			return fmt.Errorf("spec.externalEtcdConfiguration is immutable")
		}
	} else if oldETCD != newETCD {
		return fmt.Errorf("spec.externalEtcdConfiguration is immutable")
	}

	if !v1alpha1.RefSliceEqual(nSpec.IdentityProviderRefs, oSpec.IdentityProviderRefs) {
		return fmt.Errorf("spec.identityProviderRefs is immutable")
	}
	if len(nSpec.IdentityProviderRefs) > 0 {
		// if it got here, it means we already validated for IdentityProviderRefs to ensure it only has one and is
		// of type OIDCConfig
		prevOIDC, err := k.GetEksaOIDCConfig(ctx, nSpec.IdentityProviderRefs[0].Name, cluster.KubeconfigFile, spec.Namespace)
		if err != nil {
			return err
		}
		if !prevOIDC.Spec.Equal(&spec.OIDCConfig.Spec) {
			return fmt.Errorf("oidc identity provider is immutable")
		}
	}

	return provider.ValidateNewSpec(ctx, cluster, spec)
}
