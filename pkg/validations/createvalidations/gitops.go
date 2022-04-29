package createvalidations

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func ValidateGitOps(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster, spec *cluster.Spec, cliConfig *config.CliConfig) error {
	if spec.FluxConfig != nil {
		err := validateAuthenticationGitProvider(spec, cliConfig)
		if err != nil {
			return err
		}
	}

	if spec.GitOpsConfig == nil || spec.Cluster.IsSelfManaged() {
		logger.V(5).Info("skipping ValidateGitOps")
		return nil
	}

	existingGitOps, err := k.SearchEksaGitOpsConfig(ctx, spec.Cluster.Spec.GitOpsRef.Name, cluster.KubeconfigFile, spec.Cluster.Namespace)
	if err != nil {
		return err
	}
	if len(existingGitOps) > 0 {
		return fmt.Errorf("gitOpsConfig %s already exists", spec.Cluster.Spec.GitOpsRef.Name)
	}

	err = validateWorkloadFields(ctx, k, cluster, spec)
	if err != nil {
		return fmt.Errorf("workload cluster gitOpsConfig is invalid: %v", err)
	}

	return err
}

func validateAuthenticationGitProvider(clusterSpec *cluster.Spec, cliConfig *config.CliConfig) error {
	fluxConfig := clusterSpec.FluxConfig
	if fluxConfig.Spec.Git == nil {
		return nil
	}
	if cliConfig.GitPassword == "" && cliConfig.GitPrivateKeyFile == "" {
		return fmt.Errorf("for git provider in the Flux config either set a password using %s env var or path to private key"+
			" file using %s", config.EksaGitPasswordTokenEnv, config.EksaGitPrivateKeyTokenEnv)
	}
	if cliConfig.GitPrivateKeyFile != "" {
		if !validations.FileExistsAndIsNotEmpty(cliConfig.GitPrivateKeyFile) {
			return fmt.Errorf("private key file does not exist at %s or is empty", cliConfig.GitPrivateKeyFile)
		}
	}
	return nil
}

func validateWorkloadFields(ctx context.Context, k validations.KubectlClient, cluster *types.Cluster, spec *cluster.Spec) error {
	mgmtCluster, err := k.GetEksaCluster(ctx, cluster, cluster.Name)
	if err != nil {
		return err
	}
	mgmtGitOps, err := k.GetEksaGitOpsConfig(ctx, mgmtCluster.Spec.GitOpsRef.Name, cluster.KubeconfigFile, mgmtCluster.Namespace)
	if err != nil {
		return err
	}

	if !mgmtGitOps.Spec.Equal(&spec.GitOpsConfig.Spec) {
		return fmt.Errorf("expected gitOpsConfig to be the same between management and its workload clusters")
	}

	return nil
}
