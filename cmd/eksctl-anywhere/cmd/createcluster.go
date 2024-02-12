package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/aflag"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
	newManagement "github.com/aws/eks-anywhere/pkg/workflow/management"
	"github.com/aws/eks-anywhere/pkg/workflows"
	"github.com/aws/eks-anywhere/pkg/workflows/management"
	"github.com/aws/eks-anywhere/pkg/workflows/workload"
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
	providerOptions       *dependencies.ProviderOptions
}

var cc = &createClusterOptions{
	providerOptions: &dependencies.ProviderOptions{
		Tinkerbell: &dependencies.TinkerbellOptions{
			BMCOptions: &hardware.BMCOptions{
				RPC: &hardware.RPCOpts{},
			},
		},
	},
}

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
	aflag.String(aflag.TinkerbellBootstrapIP, &cc.tinkerbellBootstrapIP, createClusterCmd.Flags())
	createClusterCmd.Flags().BoolVar(&cc.forceClean, "force-cleanup", false, "Force deletion of previously created bootstrap cluster")
	hideForceCleanup(createClusterCmd.Flags())
	createClusterCmd.Flags().BoolVar(&cc.skipIpCheck, "skip-ip-check", false, "Skip check for whether cluster control plane ip is in use")
	createClusterCmd.Flags().StringVar(&cc.installPackages, "install-packages", "", "Location of curated packages configuration files to install to the cluster")
	createClusterCmd.Flags().StringArrayVar(&cc.skipValidations, "skip-validations", []string{}, fmt.Sprintf("Bypass create validations by name. Valid arguments you can pass are --skip-validations=%s", strings.Join(createvalidations.SkippableValidations[:], ",")))
	tinkerbellFlags(createClusterCmd.Flags(), cc.providerOptions.Tinkerbell.BMCOptions.RPC)

	aflag.MarkRequired(createClusterCmd.Flags(), aflag.ClusterConfig.Name)
}

func tinkerbellFlags(fs *pflag.FlagSet, r *hardware.RPCOpts) {
	aflag.String(aflag.TinkerbellBMCConsumerURL, &r.ConsumerURL, fs)
	aflag.MarkHidden(fs, aflag.TinkerbellBMCConsumerURL.Name)
	aflag.String(aflag.TinkerbellBMCHTTPContentType, &r.Request.HTTPContentType, fs)
	aflag.MarkHidden(fs, aflag.TinkerbellBMCHTTPContentType.Name)
	aflag.String(aflag.TinkerbellBMCHTTPMethod, &r.Request.HTTPMethod, fs)
	aflag.MarkHidden(fs, aflag.TinkerbellBMCHTTPMethod.Name)
	aflag.String(aflag.TinkerbellBMCTimestampHeader, &r.Request.TimestampHeader, fs)
	aflag.MarkHidden(fs, aflag.TinkerbellBMCTimestampHeader.Name)
	aflag.HTTPHeader(aflag.TinkerbellBMCStaticHeaders, aflag.NewHeader(&r.Request.StaticHeaders), fs)
	aflag.MarkHidden(fs, aflag.TinkerbellBMCStaticHeaders.Name)
	aflag.String(aflag.TinkerbellBMCSigHeaderName, &r.Signature.HeaderName, fs)
	aflag.MarkHidden(fs, aflag.TinkerbellBMCSigHeaderName.Name)
	aflag.Bool(aflag.TinkerbellBMCAppendAlgoToHeaderDisabled, &r.Signature.AppendAlgoToHeaderDisabled, fs)
	aflag.MarkHidden(fs, aflag.TinkerbellBMCAppendAlgoToHeaderDisabled.Name)
	aflag.StringSlice(aflag.TinkerbellBMCSigIncludedPayloadHeaders, &r.Signature.IncludedPayloadHeaders, fs)
	aflag.MarkHidden(fs, aflag.TinkerbellBMCSigIncludedPayloadHeaders.Name)
	aflag.Bool(aflag.TinkerbellBMCPrefixSigDisabled, &r.HMAC.PrefixSigDisabled, fs)
	aflag.MarkHidden(fs, aflag.TinkerbellBMCPrefixSigDisabled.Name)
	aflag.StringSlice(aflag.TinkerbellBMCHMACSecrets, &r.HMAC.Secrets, fs)
	aflag.MarkHidden(fs, aflag.TinkerbellBMCHMACSecrets.Name)
	aflag.String(aflag.TinkerbellBMCCustomPayload, &r.Experimental.CustomRequestPayload, fs)
	aflag.MarkHidden(fs, aflag.TinkerbellBMCCustomPayload.Name)
	aflag.String(aflag.TinkerbellBMCCustomPayloadDotLocation, &r.Experimental.DotPath, fs)
	aflag.MarkHidden(fs, aflag.TinkerbellBMCCustomPayloadDotLocation.Name)
}

func (cc *createClusterOptions) createCluster(cmd *cobra.Command, _ []string) error {
	if cc.forceClean {
		logger.MarkFail(forceCleanupDeprecationMessageForCreateDelete)
		return errors.New("please remove the --force-cleanup flag")
	}

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

	if clusterConfig.Spec.EtcdEncryption != nil {
		return errors.New("etcdEncryption is not supported during cluster creation")
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

	createCLIConfig, err := buildCreateCliConfig(cc)
	if err != nil {
		return err
	}

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

	factory := dependencies.ForSpec(clusterSpec).WithExecutableMountDirs(dirs...).
		WithBootstrapper().
		WithCliConfig(cliConfig).
		WithClusterManager(clusterSpec.Cluster, clusterManagerTimeoutOpts).
		WithProvider(cc.fileName, clusterSpec.Cluster, cc.skipIpCheck, cc.hardwareCSVPath, cc.forceClean, cc.tinkerbellBootstrapIP, skippedValidations, cc.providerOptions).
		WithGitOpsFlux(clusterSpec.Cluster, clusterSpec.FluxConfig, cliConfig).
		WithWriter().
		WithEksdInstaller().
		WithPackageInstaller(clusterSpec, cc.installPackages, cc.managementKubeconfig).
		WithValidatorClients().
		WithCreateClusterDefaulter(createCLIConfig).
		WithClusterApplier().
		WithKubeconfigWriter(clusterSpec.Cluster).
		WithClusterCreator(clusterSpec.Cluster)

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
		deps.UnAuthKubeClient,
		deps.Bootstrapper,
		deps.Provider,
		deps.ClusterManager,
		deps.GitOpsFlux,
		deps.Writer,
		deps.EksdInstaller,
		deps.PackageInstaller,
	)

	mgmt := getManagementCluster(clusterSpec)

	validationOpts := &validations.Opts{
		Kubectl: deps.UnAuthKubectlClient,
		Spec:    clusterSpec,
		WorkloadCluster: &types.Cluster{
			Name:           clusterSpec.Cluster.Name,
			KubeconfigFile: kubeconfig.FromClusterName(clusterSpec.Cluster.Name),
		},
		ManagementCluster:  mgmt,
		Provider:           deps.Provider,
		CliConfig:          cliConfig,
		SkippedValidations: skippedValidations,
		KubeClient:         deps.UnAuthKubeClient.KubeconfigClient(mgmt.KubeconfigFile),
	}
	createValidations := createvalidations.New(validationOpts)

	if features.UseNewWorkflows().IsActive() {
		deps, err = factory.
			WithCNIInstaller(clusterSpec, deps.Provider).
			Build(ctx)
		if err != nil {
			return err
		}

		wflw := &newManagement.CreateCluster{
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
		if registrar, ok := deps.Provider.(newManagement.CreateClusterHookRegistrar); ok {
			wflw.WithHookRegistrar(registrar)
		}

		err = wflw.Run(ctx)
	} else if clusterConfig.IsManaged() {
		createWorkloadCluster := workload.NewCreate(
			deps.Provider,
			deps.ClusterManager,
			deps.GitOpsFlux,
			deps.Writer,
			deps.EksdInstaller,
			deps.PackageInstaller,
			deps.ClusterCreator,
		)
		err = createWorkloadCluster.Run(ctx, clusterSpec, createValidations)

	} else if clusterSpec.Cluster.IsSelfManaged() && features.UseControllerViaCLIWorkflow().IsActive() {
		logger.Info("Using the new workflow using the controller for management cluster create")

		createMgmtCluster := management.NewCreate(
			deps.Bootstrapper,
			deps.Provider,
			deps.ClusterManager,
			deps.GitOpsFlux,
			deps.Writer,
			deps.EksdInstaller,
			deps.PackageInstaller,
			deps.ClusterCreator,
			deps.EksaInstaller,
		)

		err = createMgmtCluster.Run(ctx, clusterSpec, createValidations)
	} else {
		err = createCluster.Run(ctx, clusterSpec, createValidations, cc.forceClean)
	}

	cleanup(deps, &err)
	return err
}
