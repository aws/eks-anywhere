package cmd

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/aws/eks-anywhere/pkg/addonmanager/addonclients"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	fluxclient "github.com/aws/eks-anywhere/pkg/clients/flux"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/networking"
	"github.com/aws/eks-anywhere/pkg/providers/factory"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type upgradeClusterOptions struct {
	clusterOptions
	wConfig    string
	forceClean bool
}

func (uc *upgradeClusterOptions) kubeConfig(clusterName string) string {
	if uc.wConfig == "" {
		return filepath.Join(clusterName, fmt.Sprintf(kubeconfigPattern, clusterName))
	}
	return uc.wConfig
}

var uc = &upgradeClusterOptions{}

var upgradeClusterCmd = &cobra.Command{
	Use:          "cluster",
	Short:        "Upgrade workload cluster",
	Long:         "This command is used to upgrade workload clusters",
	PreRunE:      preRunUpgradeCluster,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := uc.upgradeCluster(cmd.Context()); err != nil {
			return fmt.Errorf("failed to upgrade cluster: %v", err)
		}
		return nil
	},
}

func preRunUpgradeCluster(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func init() {
	upgradeCmd.AddCommand(upgradeClusterCmd)
	upgradeClusterCmd.Flags().StringVarP(&uc.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	upgradeClusterCmd.Flags().StringVarP(&uc.wConfig, "w-config", "w", "", "Kubeconfig file to use when upgrading a workload cluster")
	upgradeClusterCmd.Flags().BoolVar(&uc.forceClean, "force-cleanup", false, "Force deletion of previously created bootstrap cluster")
	upgradeClusterCmd.Flags().StringVar(&uc.bundlesOverride, "bundles-override", "", "Override default Bundles manifest (not recommended)")
	err := upgradeClusterCmd.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func (uc *upgradeClusterOptions) upgradeCluster(ctx context.Context) error {
	if _, err := uc.commonValidations(ctx); err != nil {
		return fmt.Errorf("common validations failed due to: %v", err)
	}
	clusterSpec, err := newClusterSpec(uc.clusterOptions)
	if err != nil {
		return err
	}

	writer, err := filewriter.NewWriter(clusterSpec.Name)
	if err != nil {
		return fmt.Errorf("unable to write: %v", err)
	}

	eksaToolsImage := clusterSpec.VersionsBundle.Eksa.CliTools
	image := eksaToolsImage.VersionedImage()
	executableBuilder, err := executables.NewExecutableBuilder(ctx, image)
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	clusterawsadm := executableBuilder.BuildClusterAwsAdmExecutable()
	kind := executableBuilder.BuildKindExecutable(writer)
	clusterctl := executableBuilder.BuildClusterCtlExecutable(writer)
	kubectl := executableBuilder.BuildKubectlExecutable()
	govc := executableBuilder.BuildGovcExecutable(writer)
	docker := executables.BuildDockerExecutable()
	flux := executableBuilder.BuildFluxExecutable()

	providerFactory := &factory.ProviderFactory{
		AwsClient:            clusterawsadm,
		DockerClient:         docker,
		DockerKubectlClient:  kubectl,
		VSphereGovcClient:    govc,
		VSphereKubectlClient: kubectl,
		Writer:               writer,
	}
	provider, err := providerFactory.BuildProvider(uc.fileName, clusterSpec.Cluster)
	if err != nil {
		return err
	}

	bootstrapper := bootstrapper.New(&bootstrapperClient{kind, kubectl})

	clusterManager := clustermanager.New(
		&clusterManagerClient{
			clusterctl,
			kubectl,
		},
		networking.NewCilium(),
		writer,
	)

	gitOpts, err := addonclients.NewGitOptions(ctx, clusterSpec.Cluster, clusterSpec.GitOpsConfig, writer)
	if err != nil {
		return fmt.Errorf("failed to set up git options: %v", err)
	}

	addonClient := addonclients.NewFluxAddonClient(
		&fluxclient.FluxKubectl{
			Flux:    flux,
			Kubectl: kubectl,
		},
		gitOpts,
	)

	upgradeCluster := workflows.NewUpgrade(
		bootstrapper,
		provider,
		clusterManager,
		addonClient,
		writer,
	)

	workloadCluster := &types.Cluster{
		Name:           clusterSpec.Name,
		KubeconfigFile: uc.kubeConfig(clusterSpec.Name),
	}

	validationOpts := &upgradevalidations.UpgradeValidationOpts{
		Kubectl:         kubectl,
		Spec:            clusterSpec,
		WorkloadCluster: workloadCluster,
		Provider:        provider,
	}
	upgradeValidations := upgradevalidations.New(validationOpts)

	err = upgradeCluster.Run(ctx, clusterSpec, workloadCluster, upgradeValidations, uc.forceClean)
	if err == nil {
		writer.CleanUpTemp()
	}
	return err
}

func (uc *upgradeClusterOptions) commonValidations(ctx context.Context) (cluster *v1alpha1.Cluster, err error) {
	clusterConfig, err := commonValidation(ctx, uc.fileName)
	if err != nil {
		return nil, err
	}
	if !validations.KubeConfigExists(clusterConfig.Name, clusterConfig.Name, uc.wConfig, kubeconfigPattern) {
		return nil, fmt.Errorf("KubeConfig doesn't exists for cluster %s", clusterConfig.Name)
	}
	return clusterConfig, nil
}
