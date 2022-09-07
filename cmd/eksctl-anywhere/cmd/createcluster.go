package cmd

import (
	"fmt"
	"log"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type createClusterOptions struct {
	clusterOptions
	timeoutOptions
	forceClean            bool
	skipIpCheck           bool
	hardwareCSVPath       string
	tinkerbellBootstrapIP string
	installPackages       string
}

var cc = &createClusterOptions{}

var createClusterCmd = &cobra.Command{
	Use:          "cluster -f <cluster-config-file> [flags]",
	Short:        "Create workload cluster",
	Long:         "This command is used to create workload clusters",
	PreRunE:      bindFlagsToViper,
	SilenceUsage: true,
	RunE:         cc.createCluster,
}

func init() {
	createCmd.AddCommand(createClusterCmd)
	applyClusterOptionFlags(createClusterCmd.Flags(), &cc.clusterOptions)
	applyTimeoutFlags(createClusterCmd.Flags(), &cc.timeoutOptions)
	applyTinkerbellHardwareFlag(createClusterCmd.Flags(), &cc.hardwareCSVPath)
	createClusterCmd.Flags().StringVar(&cc.tinkerbellBootstrapIP, "tinkerbell-bootstrap-ip", "", "Override the local tinkerbell IP in the bootstrap cluster")
	createClusterCmd.Flags().BoolVar(&cc.forceClean, "force-cleanup", false, "Force deletion of previously created bootstrap cluster")
	createClusterCmd.Flags().BoolVar(&cc.skipIpCheck, "skip-ip-check", false, "Skip check for whether cluster control plane ip is in use")
	createClusterCmd.Flags().StringVar(&cc.installPackages, "install-packages", "", "Location of curated packages configuration files to install to the cluster")

	if err := createClusterCmd.MarkFlagRequired("filename"); err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func (cc *createClusterOptions) createCluster(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	clusterConfigFileExist := validations.FileExists(cc.fileName)
	if !clusterConfigFileExist {
		return fmt.Errorf("the cluster config file %s does not exist", cc.fileName)
	}

	clusterConfig, err := v1alpha1.GetAndValidateClusterConfig(cc.fileName)
	if err != nil {
		return fmt.Errorf("the cluster config file provided is invalid: %v", err)
	}

	if clusterConfig.Spec.DatacenterRef.Kind == v1alpha1.TinkerbellDatacenterKind {
		if err := checkTinkerbellFlags(cmd.Flags(), cc.hardwareCSVPath); err != nil {
			return err
		}
	}

	docker := executables.BuildDockerExecutable()

	if err := validations.CheckMinimumDockerVersion(ctx, docker); err != nil {
		return fmt.Errorf("failed to validate docker: %v", err)
	}

	if runtime.GOOS == "darwin" {
		if err = validations.CheckDockerDesktopVersion(ctx, docker); err != nil {
			return fmt.Errorf("failed to validate docker desktop: %v", err)
		}
	}

	validations.CheckDockerAllocatedMemory(ctx, docker)

	kubeconfigPath := kubeconfig.FromClusterName(clusterConfig.Name)
	if validations.FileExistsAndIsNotEmpty(kubeconfigPath) {
		return fmt.Errorf(
			"old cluster config file exists under %s, please use a different clusterName to proceed",
			clusterConfig.Name,
		)
	}

	clusterSpec, err := newClusterSpec(cc.clusterOptions)
	if err != nil {
		return err
	}

	cliConfig := buildCliConfig(clusterSpec)
	dirs, err := cc.directoriesToMount(clusterSpec, cliConfig, cc.installPackages)
	if err != nil {
		return err
	}

	clusterManagerOpts, err := buildClusterManagerOpts(cc.timeoutOptions)
	if err != nil {
		return fmt.Errorf("failed to build cluster manager opts: %v", err)
	}

	deps, err := dependencies.ForSpec(ctx, clusterSpec).WithExecutableMountDirs(dirs...).
		WithBootstrapper().
		WithCliConfig(cliConfig).
		WithClusterManager(clusterSpec.Cluster, clusterManagerOpts...).
		WithProvider(cc.fileName, clusterSpec.Cluster, cc.skipIpCheck, cc.hardwareCSVPath, cc.forceClean, cc.tinkerbellBootstrapIP).
		WithGitOpsFlux(clusterSpec.Cluster, clusterSpec.FluxConfig, cliConfig).
		WithWriter().
		WithEksdInstaller().
		WithPackageInstaller(clusterSpec, cc.installPackages).
		Build(ctx)
	if err != nil {
		return err
	}
	defer close(ctx, deps)

	if !features.IsActive(features.CloudStackProvider()) && deps.Provider.Name() == constants.CloudStackProviderName {
		return fmt.Errorf("provider cloudstack is not supported in this release")
	}

	if !features.IsActive(features.SnowProvider()) && deps.Provider.Name() == constants.SnowProviderName {
		return fmt.Errorf("provider snow is not supported in this release")
	}

	if !features.IsActive(features.NutanixProvider()) && deps.Provider.Name() == constants.NutanixProviderName {
		return fmt.Errorf("provider nutanix is not supported in this release")
	}

	createCluster := workflows.NewCreate(
		deps.Bootstrapper,
		deps.Provider,
		deps.ClusterManager,
		deps.GitOpsFlux,
		deps.Writer,
		deps.EksdInstaller,
		deps.PackageInstaller,
	)

	var cluster *types.Cluster
	if clusterSpec.ManagementCluster == nil {
		cluster = &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Cluster.Name),
		}
	} else {
		cluster = &types.Cluster{
			Name:           clusterSpec.ManagementCluster.Name,
			KubeconfigFile: clusterSpec.ManagementCluster.KubeconfigFile,
		}
	}

	validationOpts := &validations.Opts{
		Kubectl: deps.Kubectl,
		Spec:    clusterSpec,
		WorkloadCluster: &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Cluster.Name),
		},
		ManagementCluster: cluster,
		Provider:          deps.Provider,
		CliConfig:         cliConfig,
	}
	createValidations := createvalidations.New(validationOpts)

	err = createCluster.Run(ctx, clusterSpec, createValidations, cc.forceClean)

	cleanup(deps, &err)
	return err
}
