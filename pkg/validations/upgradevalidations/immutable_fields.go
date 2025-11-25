package upgradevalidations

import (
	"context"
	"errors"
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

	if err := ValidateGitOpsImmutableFields(ctx, k, cluster, spec, prevSpec); err != nil {
		return err
	}

	if !nSpec.ControlPlaneConfiguration.Endpoint.Equal(oSpec.ControlPlaneConfiguration.Endpoint, nSpec.DatacenterRef.Kind) {
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

	// We don't want users to be able to toggle  off SkipUpgrade until we've understood the
	// implications so we are temporarily disallowing it.

	oCNI := prevSpec.Spec.ClusterNetwork.CNIConfig
	nCNI := spec.Cluster.Spec.ClusterNetwork.CNIConfig
	if oCNI != nil && oCNI.Cilium != nil && !oCNI.Cilium.IsManaged() && nCNI.Cilium.IsManaged() {
		return fmt.Errorf("spec.clusterNetwork.cniConfig.cilium.skipUpgrade cannot be toggled off")
	}

	if !nSpec.ProxyConfiguration.Equal(oSpec.ProxyConfiguration) {
		return fmt.Errorf("spec.proxyConfiguration is immutable")
	}

	oldETCD := oSpec.ExternalEtcdConfiguration
	newETCD := nSpec.ExternalEtcdConfiguration
	if oldETCD != nil && newETCD == nil || oldETCD == nil && newETCD != nil {
		return errors.New("adding or removing external etcd during upgrade is not supported")
	}

	oldAWSIamConfigRef := &v1alpha1.Ref{}

	for _, oIdentityProvider := range oSpec.IdentityProviderRefs {
		switch oIdentityProvider.Kind {
		case v1alpha1.AWSIamConfigKind:
			oIdentityProvider := oIdentityProvider // new variable scoped to the for loop Ref: https://github.com/golang/go/wiki/CommonMistakes#using-reference-to-loop-iterator-variable
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
	if oSpec.ManagementCluster.Name != nSpec.ManagementCluster.Name {
		return fmt.Errorf("management cluster name is immutable")
	}

	return provider.ValidateNewSpec(ctx, cluster, spec)
}

func ValidateGitOpsImmutableFields(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster, clusterSpec *cluster.Spec, oldCluster *v1alpha1.Cluster) error {
	if oldCluster.Spec.GitOpsRef == nil {
		return nil
	}

	if !clusterSpec.Cluster.Spec.GitOpsRef.Equal(oldCluster.Spec.GitOpsRef) {
		return errors.New("once cluster.spec.gitOpsRef is set, it is immutable")
	}

	switch clusterSpec.Cluster.Spec.GitOpsRef.Kind {
	case v1alpha1.GitOpsConfigKind:
		prevGitOps, err := k.GetEksaGitOpsConfig(ctx, clusterSpec.Cluster.Spec.GitOpsRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace)
		if err != nil {
			return err
		}

		if prevGitOps.Spec.Flux.Github.Owner != clusterSpec.GitOpsConfig.Spec.Flux.Github.Owner {
			return errors.New("gitOps spec.flux.github.owner is immutable")
		}
		if prevGitOps.Spec.Flux.Github.Repository != clusterSpec.GitOpsConfig.Spec.Flux.Github.Repository {
			return errors.New("gitOps spec.flux.github.repository is immutable")
		}
		if prevGitOps.Spec.Flux.Github.Personal != clusterSpec.GitOpsConfig.Spec.Flux.Github.Personal {
			return errors.New("gitOps spec.flux.github.personal is immutable")
		}
		if prevGitOps.Spec.Flux.Github.FluxSystemNamespace != clusterSpec.GitOpsConfig.Spec.Flux.Github.FluxSystemNamespace {
			return errors.New("gitOps spec.flux.github.fluxSystemNamespace is immutable")
		}
		if prevGitOps.Spec.Flux.Github.Branch != clusterSpec.GitOpsConfig.Spec.Flux.Github.Branch {
			return errors.New("gitOps spec.flux.github.branch is immutable")
		}
		if prevGitOps.Spec.Flux.Github.ClusterConfigPath != clusterSpec.GitOpsConfig.Spec.Flux.Github.ClusterConfigPath {
			return errors.New("gitOps spec.flux.github.clusterConfigPath is immutable")
		}

	case v1alpha1.FluxConfigKind:
		prevGitOps, err := k.GetEksaFluxConfig(ctx, clusterSpec.Cluster.Spec.GitOpsRef.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace)
		if err != nil {
			return err
		}

		if prevGitOps.Spec.Git != nil {
			if prevGitOps.Spec.Git.RepositoryUrl != clusterSpec.FluxConfig.Spec.Git.RepositoryUrl {
				return errors.New("fluxConfig spec.fluxConfig.spec.git.repositoryUrl is immutable")
			}
			if prevGitOps.Spec.Git.SshKeyAlgorithm != clusterSpec.FluxConfig.Spec.Git.SshKeyAlgorithm {
				return errors.New("fluxConfig spec.fluxConfig.spec.git.sshKeyAlgorithm is immutable")
			}
		}

		if prevGitOps.Spec.Github != nil {
			if prevGitOps.Spec.Github.Repository != clusterSpec.FluxConfig.Spec.Github.Repository {
				return errors.New("fluxConfig spec.github.repository is immutable")
			}

			if prevGitOps.Spec.Github.Owner != clusterSpec.FluxConfig.Spec.Github.Owner {
				return errors.New("fluxConfig spec.github.owner is immutable")
			}

			if prevGitOps.Spec.Github.Personal != clusterSpec.FluxConfig.Spec.Github.Personal {
				return errors.New("fluxConfig spec.github.personal is immutable")
			}
		}

		if prevGitOps.Spec.Branch != clusterSpec.FluxConfig.Spec.Branch {
			return errors.New("fluxConfig spec.branch is immutable")
		}

		if prevGitOps.Spec.ClusterConfigPath != clusterSpec.FluxConfig.Spec.ClusterConfigPath {
			return errors.New("fluxConfig spec.clusterConfigPath is immutable")
		}

		if prevGitOps.Spec.SystemNamespace != clusterSpec.FluxConfig.Spec.SystemNamespace {
			return errors.New("fluxConfig spec.systemNamespace is immutable")
		}
	}

	return nil
}
