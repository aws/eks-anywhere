package upgradevalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func ValidateImmutableFields(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster, spec *cluster.Spec, provider providers.Provider) error {
	prevSpec, err := k.GetEksaCluster(ctx, cluster, spec.Cluster.Name)
	if err != nil {
		return err
	}

	if prevSpec.Name != spec.Cluster.Name {
		return fmt.Errorf("cluster name is immutable. previous name %s, new name %s", prevSpec.Name, spec.Cluster.Name)
	}

	if prevSpec.Namespace != spec.Cluster.Namespace {
		if !(prevSpec.Namespace == "default" && spec.Cluster.Namespace == "") {
			return fmt.Errorf("cluster namespace is immutable")
		}
	}

	oSpec := prevSpec.Spec
	nSpec := spec.Cluster.Spec

	if !nSpec.DatacenterRef.Equal(&oSpec.DatacenterRef) {
		return fmt.Errorf("spec.dataCenterRef.name is immutable")
	}

	if !nSpec.GitOpsRef.Equal(oSpec.GitOpsRef) {
		return fmt.Errorf("spec.gitOpsRef is immutable")
	}

	if nSpec.GitOpsRef != nil {
		if nSpec.GitOpsRef.Kind == v1alpha1.GitOpsConfigKind {
			prevGitOps, err := k.GetEksaGitOpsConfig(ctx, nSpec.GitOpsRef.Name, cluster.KubeconfigFile, spec.Cluster.Namespace)
			if err != nil {
				return err
			}

			if prevGitOps.Spec.Flux.Github.Owner != spec.GitOpsConfig.Spec.Flux.Github.Owner {
				return fmt.Errorf("gitOps spec.flux.github.owner is immutable")
			}
			if prevGitOps.Spec.Flux.Github.Repository != spec.GitOpsConfig.Spec.Flux.Github.Repository {
				return fmt.Errorf("gitOps spec.flux.github.repository is immutable")
			}
			if prevGitOps.Spec.Flux.Github.Personal != spec.GitOpsConfig.Spec.Flux.Github.Personal {
				return fmt.Errorf("gitOps spec.flux.github.personal is immutable")
			}
			if spec.GitOpsConfig.Spec.Flux.Github.FluxSystemNamespace != "" && prevGitOps.Spec.Flux.Github.FluxSystemNamespace != spec.GitOpsConfig.Spec.Flux.Github.FluxSystemNamespace {
				return fmt.Errorf("gitOps spec.flux.github.fluxSystemNamespace is immutable")
			}
			if spec.GitOpsConfig.Spec.Flux.Github.Branch != "" && prevGitOps.Spec.Flux.Github.Branch != spec.GitOpsConfig.Spec.Flux.Github.Branch {
				return fmt.Errorf("gitOps spec.flux.github.branch is immutable")
			}
			if spec.GitOpsConfig.Spec.Flux.Github.ClusterConfigPath != "" && prevGitOps.Spec.Flux.Github.ClusterConfigPath != spec.GitOpsConfig.Spec.Flux.Github.ClusterConfigPath {
				return fmt.Errorf("gitOps spec.flux.github.clusterConfigPath is immutable")
			}
		}

		if nSpec.GitOpsRef.Kind == v1alpha1.FluxConfigKind {
			prevGitOps, err := k.GetEksaFluxConfig(ctx, nSpec.GitOpsRef.Name, cluster.KubeconfigFile, spec.Cluster.Namespace)
			if err != nil {
				return err
			}

			if prevGitOps.Spec.Git != nil {
				if prevGitOps.Spec.Git.RepositoryUrl != spec.FluxConfig.Spec.Git.RepositoryUrl {
					return fmt.Errorf("fluxConfig spec.fluxConfig.spec.git.repositoryUrl is immutable")
				}
			}

			if prevGitOps.Spec.Github != nil {
				if prevGitOps.Spec.Github.Repository != spec.FluxConfig.Spec.Github.Repository {
					return fmt.Errorf("fluxConfig spec.github.repository is immutable")
				}

				if prevGitOps.Spec.Github.Owner != spec.FluxConfig.Spec.Github.Owner {
					return fmt.Errorf("fluxConfig spec.github.owner is immutable")
				}

				if prevGitOps.Spec.Github.Personal != spec.FluxConfig.Spec.Github.Personal {
					return fmt.Errorf("fluxConfig spec.github.personal is immutable")
				}
			}

			if prevGitOps.Spec.Branch != spec.FluxConfig.Spec.Branch {
				return fmt.Errorf("fluxConfig spec.branch is immutable")
			}

			if prevGitOps.Spec.ClusterConfigPath != spec.FluxConfig.Spec.ClusterConfigPath {
				return fmt.Errorf("fluxConfig spec.clusterConfigPath is immutable")
			}

			if prevGitOps.Spec.SystemNamespace != spec.FluxConfig.Spec.SystemNamespace {
				return fmt.Errorf("fluxConfig spec.systemNamespace is immutable")
			}
		}
	}

	if !nSpec.ControlPlaneConfiguration.Endpoint.Equal(oSpec.ControlPlaneConfiguration.Endpoint) {
		return fmt.Errorf("spec.controlPlaneConfiguration.endpoint is immutable")
	}

	/* compare all clusterNetwork fields individually, since we do allow updating updating fields for configuring plugins such as CiliumConfig through the cli*/
	if !nSpec.ClusterNetwork.Pods.Equal(&oSpec.ClusterNetwork.Pods) {
		return fmt.Errorf("spec.clusterNetwork.Pods is immutable")
	}
	if !nSpec.ClusterNetwork.Services.Equal(&oSpec.ClusterNetwork.Services) {
		return fmt.Errorf("spec.clusterNetwork.Services is immutable")
	}
	if !nSpec.ClusterNetwork.DNS.Equal(&oSpec.ClusterNetwork.DNS) {
		return fmt.Errorf("spec.clusterNetwork.DNS is immutable")
	}
	if !v1alpha1.CNIPluginSame(nSpec.ClusterNetwork, oSpec.ClusterNetwork) {
		return fmt.Errorf("spec.clusterNetwork.CNI/CNIConfig is immutable")
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

	oldAWSIamConfigRef := &v1alpha1.Ref{}

	for _, oIdentityProvider := range oSpec.IdentityProviderRefs {
		switch oIdentityProvider.Kind {
		case v1alpha1.AWSIamConfigKind:
			oldAWSIamConfigRef = &oIdentityProvider
		}
	}
	for _, nIdentityProvider := range nSpec.IdentityProviderRefs {
		switch nIdentityProvider.Kind {
		case v1alpha1.AWSIamConfigKind:
			newAWSIamConfigRef := &nIdentityProvider
			prevAwsIam, err := k.GetEksaAWSIamConfig(ctx, nIdentityProvider.Name, cluster.KubeconfigFile, spec.Cluster.Namespace)
			if err != nil {
				return err
			}
			if !prevAwsIam.Spec.Equal(&spec.AWSIamConfig.Spec) || !oldAWSIamConfigRef.Equal(newAWSIamConfigRef) {
				return fmt.Errorf("aws iam identity provider is immutable")
			}
		}
	}

	if spec.Cluster.IsSelfManaged() != prevSpec.IsSelfManaged() {
		return fmt.Errorf("management flag is immutable")
	}

	return provider.ValidateNewSpec(ctx, cluster, spec)
}
