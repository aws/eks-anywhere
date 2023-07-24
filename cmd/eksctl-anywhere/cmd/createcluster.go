package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
	"github.com/aws/eks-anywhere/pkg/workflow/management"
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
	skipValidations       []string
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
	createClusterCmd.Flags().StringArrayVar(&cc.skipValidations, "skip-validations", []string{}, fmt.Sprintf("Bypass create validations by name. Valid arguments you can pass are --skip-validations=%s", strings.Join(createvalidations.SkippableValidations[:], ",")))

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
		if err := checkTinkerbellFlags(cmd.Flags(), cc.hardwareCSVPath, Create); err != nil {
			return err
		}
	}

	docker := executables.BuildDockerExecutable()

	if err := validations.CheckMinimumDockerVersion(ctx, docker); err != nil {
		return fmt.Errorf("failed to validate docker: %v", err)
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

	if err := validations.ValidateAuthenticationForRegistryMirror(clusterSpec); err != nil {
		return err
	}

	cliConfig := buildCliConfig(clusterSpec)
	dirs, err := cc.directoriesToMount(clusterSpec, cliConfig, cc.installPackages)
	if err != nil {
		return err
	}

	createCLIConfig := buildCreateCliConfig(cc)

	clusterManagerTimeoutOpts, err := buildClusterManagerOpts(cc.timeoutOptions, clusterSpec.Cluster.Spec.DatacenterRef.Kind)
	if err != nil {
		return fmt.Errorf("failed to build cluster manager opts: %v", err)
	}

	var skippedValidations map[string]bool
	if len(cc.skipValidations) != 0 {
		skippedValidations, err = validations.ValidateSkippableValidation(cc.skipValidations, createvalidations.SkippableValidations)
		if err != nil {
			return err
		}
	}

	factory := dependencies.ForSpec(ctx, clusterSpec).WithExecutableMountDirs(dirs...).
		WithBootstrapper().
		WithCliConfig(cliConfig).
		WithClusterManager(clusterSpec.Cluster, clusterManagerTimeoutOpts).
		WithProvider(cc.fileName, clusterSpec.Cluster, cc.skipIpCheck, cc.hardwareCSVPath, cc.forceClean, cc.tinkerbellBootstrapIP, skippedValidations).
		WithGitOpsFlux(clusterSpec.Cluster, clusterSpec.FluxConfig, cliConfig).
		WithWriter().
		WithEksdInstaller().
		WithPackageInstaller(clusterSpec, cc.installPackages, cc.managementKubeconfig).
		WithValidatorClients().
		WithCreateClusterDefaulter(createCLIConfig)

	if cc.timeoutOptions.noTimeouts {
		factory.WithNoTimeouts()
	}

	deps, err := factory.Build(ctx)
	if err != nil {
		return err
	}
	defer close(ctx, deps)

	clusterSpec, err = deps.CreateClusterDefaulter.Run(ctx, clusterSpec)
	if err != nil {
		return err
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

	validationOpts := &validations.Opts{
		Kubectl: deps.UnAuthKubectlClient,
		Spec:    clusterSpec,
		WorkloadCluster: &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Cluster.Name),
		},
		ManagementCluster:  getManagementCluster(clusterSpec),
		Provider:           deps.Provider,
		CliConfig:          cliConfig,
		SkippedValidations: skippedValidations,
	}
	createValidations := createvalidations.New(validationOpts)

	if features.UseNewWorkflows().IsActive() {
		deps, err = factory.
			WithCNIInstaller(clusterSpec, deps.Provider).
			Build(ctx)
		if err != nil {
			return err
		}

		wflw := &management.CreateCluster{
			Spec:                          clusterSpec,
			Bootstrapper:                  deps.Bootstrapper,
			CreateBootstrapClusterOptions: deps.Provider,
			CNIInstaller:                  deps.CNIInstaller,
			Cluster:                       clustermanager.NewCreateClusterShim(clusterSpec, deps.ClusterManager, deps.Provider),
			FS:                            deps.Writer,
		}
		wflw.WithHookRegistrar(awsiamauth.NewHookRegistrar(deps.AwsIamAuth, clusterSpec))

		// Not all provider implementations want to bind hooks so we explicitly check if they
		// want to bind hooks before registering it.
		if registrar, ok := deps.Provider.(management.CreateClusterHookRegistrar); ok {
			wflw.WithHookRegistrar(registrar)
		}

		err = wflw.Run(ctx)
	} else {
		err = createCluster.Run(ctx, clusterSpec, createValidations, cc.forceClean)
	}

	cleanup(deps, &err)
	return err
}
