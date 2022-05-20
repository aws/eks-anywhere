package workflows

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

type Create struct {
	bootstrapper   interfaces.Bootstrapper
	provider       providers.Provider
	clusterManager interfaces.ClusterManager
	addonManager   interfaces.AddonManager
	writer         filewriter.FileWriter
	eksdInstaller  interfaces.EksdInstaller
}

func NewCreate(bootstrapper interfaces.Bootstrapper, provider providers.Provider,
	clusterManager interfaces.ClusterManager, addonManager interfaces.AddonManager, writer filewriter.FileWriter, eksdInstaller interfaces.EksdInstaller,
) *Create {
	return &Create{
		bootstrapper:   bootstrapper,
		provider:       provider,
		clusterManager: clusterManager,
		addonManager:   addonManager,
		writer:         writer,
		eksdInstaller:  eksdInstaller,
	}
}

func (c *Create) Run(ctx context.Context, clusterSpec *cluster.Spec, validator interfaces.Validator, forceCleanup bool, packagesLocation string) error {
	if forceCleanup {
		if err := c.bootstrapper.DeleteBootstrapCluster(ctx, &types.Cluster{
			Name: clusterSpec.Cluster.Name,
		}, false); err != nil {
			return err
		}
	}
	commandContext := &task.CommandContext{
		Bootstrapper:   c.bootstrapper,
		Provider:       c.provider,
		ClusterManager: c.clusterManager,
		AddonManager:   c.addonManager,
		ClusterSpec:    clusterSpec,
		Writer:         c.writer,
		Validations:    validator,
		EksdInstaller:  c.eksdInstaller,
	}

	if clusterSpec.ManagementCluster != nil {
		commandContext.BootstrapCluster = clusterSpec.ManagementCluster
	}

	err := task.NewTaskRunner(&SetAndValidateTask{}).RunTask(ctx, commandContext)
	if err != nil {
		return err
	}

	if packagesLocation != "" {
		curatedpackages.PrintLicense()
		err = installCuratedPackages(ctx, clusterSpec, packagesLocation)
	}
	return err
}

// task related entities

type CreateBootStrapClusterTask struct{}

type SetAndValidateTask struct{}

type CreateWorkloadClusterTask struct{}

type InstallResourcesOnManagementTask struct{}

type InstallEksaComponentsTask struct{}

type InstallAddonManagerTask struct{}

type MoveClusterManagementTask struct{}

type WriteClusterConfigTask struct{}

type DeleteBootstrapClusterTask struct {
	*CollectDiagnosticsTask
}

// CreateBootStrapClusterTask implementation

func (s *CreateBootStrapClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.BootstrapCluster != nil {
		return &CreateWorkloadClusterTask{}
	}
	logger.Info("Creating new bootstrap cluster")

	bootstrapOptions, err := commandContext.Provider.BootstrapClusterOpts()
	if err != nil {
		commandContext.SetError(err)
		return nil
	}

	bootstrapCluster, err := commandContext.Bootstrapper.CreateBootstrapCluster(ctx, commandContext.ClusterSpec, bootstrapOptions...)
	if err != nil {
		commandContext.SetError(err)
		return nil
	}
	commandContext.BootstrapCluster = bootstrapCluster

	logger.Info("Provider specific pre-capi-install-setup on bootstrap cluster")
	if err = commandContext.Provider.PreCAPIInstallOnBootstrap(ctx, bootstrapCluster, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	logger.Info("Installing cluster-api providers on bootstrap cluster")
	if err = commandContext.ClusterManager.InstallCAPI(ctx, commandContext.ClusterSpec, bootstrapCluster, commandContext.Provider); err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	if commandContext.ClusterSpec.AWSIamConfig != nil {
		logger.Info("Creating aws-iam-authenticator certificate and key pair secret on bootstrap cluster")
		if err = commandContext.ClusterManager.CreateAwsIamAuthCaSecret(ctx, bootstrapCluster); err != nil {
			commandContext.SetError(err)
			return &CollectMgmtClusterDiagnosticsTask{}
		}
	}

	logger.Info("Provider specific post-setup")
	if err = commandContext.Provider.PostBootstrapSetup(ctx, commandContext.ClusterSpec.Cluster, bootstrapCluster); err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	return &CreateWorkloadClusterTask{}
}

func (s *CreateBootStrapClusterTask) Name() string {
	return "bootstrap-cluster-init"
}

// SetAndValidateTask implementation

func (s *SetAndValidateTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Performing setup and validations")
	runner := validations.NewRunner()
	runner.Register(s.providerValidation(ctx, commandContext)...)
	runner.Register(commandContext.AddonManager.Validations(ctx, commandContext.ClusterSpec)...)
	runner.Register(s.validations(ctx, commandContext)...)

	err := runner.Run()
	if err != nil {
		commandContext.SetError(err)
		return nil
	}
	return &CreateBootStrapClusterTask{}
}

func (s *SetAndValidateTask) validations(ctx context.Context, commandContext *task.CommandContext) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: "create preflight validations pass",
				Err:  commandContext.Validations.PreflightValidations(ctx),
			}
		},
	}
}

func (s *SetAndValidateTask) providerValidation(ctx context.Context, commandContext *task.CommandContext) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: fmt.Sprintf("%s Provider setup is valid", commandContext.Provider.Name()),
				Err:  commandContext.Provider.SetupAndValidateCreateCluster(ctx, commandContext.ClusterSpec),
			}
		},
	}
}

func (s *SetAndValidateTask) Name() string {
	return "setup-validate"
}

// CreateWorkloadClusterTask implementation

func (s *CreateWorkloadClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Creating new workload cluster")
	workloadCluster, err := commandContext.ClusterManager.CreateWorkloadCluster(ctx, commandContext.BootstrapCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	commandContext.WorkloadCluster = workloadCluster

	logger.Info("Installing networking on workload cluster")
	err = commandContext.ClusterManager.InstallNetworking(ctx, workloadCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	if commandContext.ClusterSpec.AWSIamConfig != nil {
		logger.Info("Installing aws-iam-authenticator on workload cluster")
		err = commandContext.ClusterManager.InstallAwsIamAuth(ctx, commandContext.BootstrapCluster, workloadCluster, commandContext.ClusterSpec)
		if err != nil {
			commandContext.SetError(err)
			return &CollectDiagnosticsTask{}
		}
	}

	err = commandContext.ClusterManager.InstallStorageClass(ctx, workloadCluster, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	if !commandContext.BootstrapCluster.ExistingManagement {
		logger.Info("Installing cluster-api providers on workload cluster")
		err = commandContext.ClusterManager.InstallCAPI(ctx, commandContext.ClusterSpec, commandContext.WorkloadCluster, commandContext.Provider)
		if err != nil {
			commandContext.SetError(err)
			return &CollectDiagnosticsTask{}
		}

		logger.Info("Installing EKS-A secrets on workload cluster")
		err := commandContext.Provider.UpdateSecrets(ctx, commandContext.WorkloadCluster)
		if err != nil {
			commandContext.SetError(err)
			return &CollectDiagnosticsTask{}
		}
	}

	logger.V(4).Info("Installing machine health checks on bootstrap cluster")
	err = commandContext.ClusterManager.InstallMachineHealthChecks(ctx, commandContext.BootstrapCluster, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	return &InstallResourcesOnManagementTask{}
}

func (s *CreateWorkloadClusterTask) Name() string {
	return "workload-cluster-init"
}

// InstallResourcesOnManagement implementation
func (s *InstallResourcesOnManagementTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.BootstrapCluster.ExistingManagement {
		return &MoveClusterManagementTask{}
	}
	logger.Info("Installing resources on management cluster")

	if err := commandContext.Provider.PostWorkloadInit(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	return &MoveClusterManagementTask{}
}

func (s *InstallResourcesOnManagementTask) Name() string {
	return "install-resources-on-management-cluster"
}

// MoveClusterManagementTask implementation

func (s *MoveClusterManagementTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.BootstrapCluster.ExistingManagement {
		return &InstallEksaComponentsTask{}
	}
	logger.Info("Moving cluster management from bootstrap to workload cluster")
	err := commandContext.ClusterManager.MoveCAPI(ctx, commandContext.BootstrapCluster, commandContext.WorkloadCluster, commandContext.WorkloadCluster.Name, commandContext.ClusterSpec, types.WithNodeRef())
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	return &InstallEksaComponentsTask{}
}

func (s *MoveClusterManagementTask) Name() string {
	return "capi-management-move"
}

// InstallEksaComponentsTask implementation

func (s *InstallEksaComponentsTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if !commandContext.BootstrapCluster.ExistingManagement {
		logger.Info("Installing EKS-A custom components (CRD and controller) on workload cluster")
		err := commandContext.ClusterManager.InstallCustomComponents(ctx, commandContext.ClusterSpec, commandContext.WorkloadCluster, commandContext.Provider)
		if err != nil {
			commandContext.SetError(err)
			return &CollectDiagnosticsTask{}
		}
		logger.Info("Installing EKS-D components on workload cluster")
		err = commandContext.EksdInstaller.InstallEksdCRDs(ctx, commandContext.ClusterSpec, commandContext.WorkloadCluster)
		if err != nil {
			commandContext.SetError(err)
			return &CollectDiagnosticsTask{}
		}
	}

	logger.Info("Creating EKS-A CRDs instances on workload cluster")
	datacenterConfig := commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec)
	machineConfigs := commandContext.Provider.MachineConfigs(commandContext.ClusterSpec)

	// this disables create-webhook validation during create
	commandContext.ClusterSpec.Cluster.PauseReconcile()
	datacenterConfig.PauseReconcile()

	targetCluster := commandContext.WorkloadCluster
	if commandContext.BootstrapCluster.ExistingManagement {
		targetCluster = commandContext.BootstrapCluster
	}
	err := commandContext.ClusterManager.CreateEKSAResources(ctx, targetCluster, commandContext.ClusterSpec, datacenterConfig, machineConfigs)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	err = commandContext.EksdInstaller.InstallEksdManifest(ctx, commandContext.ClusterSpec, targetCluster)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	err = commandContext.ClusterManager.ResumeEKSAControllerReconcile(ctx, targetCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	return &InstallAddonManagerTask{}
}

func (s *InstallEksaComponentsTask) Name() string {
	return "eksa-components-install"
}

// InstallAddonManagerTask implementation

func (s *InstallAddonManagerTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Installing AddonManager and GitOps Toolkit on workload cluster")

	err := commandContext.AddonManager.InstallGitOps(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec))
	if err != nil {
		logger.MarkFail("Error when installing GitOps toolkits on workload cluster; EKS-A will continue with cluster creation, but GitOps will not be enabled", "error", err)
		return &WriteClusterConfigTask{}
	}
	return &WriteClusterConfigTask{}
}

func (s *InstallAddonManagerTask) Name() string {
	return "addon-manager-install"
}

func (s *WriteClusterConfigTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Writing cluster config file")
	err := clustermarshaller.WriteClusterConfig(commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec), commandContext.Writer)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	return &DeleteBootstrapClusterTask{}
}

func (s *WriteClusterConfigTask) Name() string {
	return "write-cluster-config"
}

// DeleteBootstrapClusterTask implementation

func (s *DeleteBootstrapClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if !commandContext.BootstrapCluster.ExistingManagement {
		logger.Info("Deleting bootstrap cluster")
		err := commandContext.Bootstrapper.DeleteBootstrapCluster(ctx, commandContext.BootstrapCluster, false)
		if err != nil {
			commandContext.SetError(err)
		}
	}
	if commandContext.OriginalError == nil {
		logger.MarkSuccess("Cluster created!")
	}
	return nil
}

func (s *DeleteBootstrapClusterTask) Name() string {
	return "delete-kind-cluster"
}

func installCuratedPackages(ctx context.Context, spec *cluster.Spec, packagesLocation string) error {
	err := installPackagesController(ctx, spec)
	if err != nil {
		logger.MarkFail("Error when installing curated packages on workload cluster; please install through eksctl anywhere install packagecontroller command", "error", err)
		return nil
	}

	err = installPackages(ctx, spec.Cluster.Name, packagesLocation)
	if err != nil {
		logger.MarkFail("Error when installing curated packages on workload cluster; please install through eksctl anywhere create packages command", "error", err)
	}
	return nil
}

func installPackagesController(ctx context.Context, spec *cluster.Spec) error {
	logger.Info("Installing curated packages controller on workload cluster")
	kubeConfig := kubeconfig.FromClusterName(spec.Cluster.Name)
	deps, err := curatedpackages.NewDependenciesForPackages(ctx, kubeConfig)
	if err != nil {
		return err
	}
	chart := spec.VersionsBundle.VersionsBundle.PackageController.HelmChart
	pc := curatedpackages.NewPackageControllerClient(deps.Helm, deps.Kubectl, kubeConfig, chart.Image(), chart.Name, chart.Tag())
	err = pc.InstallController(ctx)
	if err != nil {
		return err
	}
	return nil
}

func installPackages(ctx context.Context, clusterName, packagesLocation string) error {
	kubeConfig := kubeconfig.FromClusterName(clusterName)
	deps, err := curatedpackages.NewDependenciesForPackages(ctx, kubeConfig, packagesLocation)
	if err != nil {
		return err
	}
	packageClient := curatedpackages.NewPackageClient(
		nil,
		deps.Kubectl,
	)
	err = packageClient.CreatePackages(ctx, packagesLocation, kubeConfig)
	if err != nil {
		return err
	}
	return nil
}
