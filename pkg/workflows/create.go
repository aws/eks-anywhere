package workflows

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

type Create struct {
	bootstrapper     interfaces.Bootstrapper
	provider         providers.Provider
	clusterManager   interfaces.ClusterManager
	gitOpsManager    interfaces.GitOpsManager
	writer           filewriter.FileWriter
	eksdInstaller    interfaces.EksdInstaller
	packageInstaller interfaces.PackageInstaller
}

func NewCreate(bootstrapper interfaces.Bootstrapper, provider providers.Provider,
	clusterManager interfaces.ClusterManager, gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter, eksdInstaller interfaces.EksdInstaller,
	packageInstaller interfaces.PackageInstaller,
) *Create {
	return &Create{
		bootstrapper:     bootstrapper,
		provider:         provider,
		clusterManager:   clusterManager,
		gitOpsManager:    gitOpsManager,
		writer:           writer,
		eksdInstaller:    eksdInstaller,
		packageInstaller: packageInstaller,
	}
}

func (c *Create) Run(ctx context.Context, clusterSpec *cluster.Spec, validator interfaces.Validator, forceCleanup bool) error {
	if forceCleanup {
		if err := c.bootstrapper.DeleteBootstrapCluster(ctx, &types.Cluster{
			Name: clusterSpec.Cluster.Name,
		}, constants.Create, forceCleanup); err != nil {
			return err
		}
	}
	commandContext := &task.CommandContext{
		Bootstrapper:     c.bootstrapper,
		Provider:         c.provider,
		ClusterManager:   c.clusterManager,
		GitOpsManager:    c.gitOpsManager,
		ClusterSpec:      clusterSpec,
		Writer:           c.writer,
		Validations:      validator,
		EksdInstaller:    c.eksdInstaller,
		PackageInstaller: c.packageInstaller,
	}

	if clusterSpec.ManagementCluster != nil {
		commandContext.BootstrapCluster = clusterSpec.ManagementCluster
	}

	err := task.NewTaskRunner(&SetAndValidateTask{}, c.writer).RunTask(ctx, commandContext)

	return err
}

// task related entities

type CreateBootStrapClusterTask struct{}

type SetAndValidateTask struct{}

type CreateWorkloadClusterTask struct{}

type InstallResourcesOnManagementTask struct{}

type InstallEksaComponentsTask struct{}

type InstallGitOpsManagerTask struct{}

type MoveClusterManagementTask struct{}

type WriteClusterConfigTask struct{}

type DeleteBootstrapClusterTask struct {
	*CollectDiagnosticsTask
}

type InstallCuratedPackagesTask struct{}

// CreateBootStrapClusterTask implementation

func (s *CreateBootStrapClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.BootstrapCluster != nil {
		if commandContext.ClusterSpec.AWSIamConfig != nil {
			logger.Info("Creating aws-iam-authenticator certificate and key pair secret on bootstrap cluster")
			if err := commandContext.ClusterManager.CreateAwsIamAuthCaSecret(ctx, commandContext.BootstrapCluster, commandContext.ClusterSpec.Cluster.Name); err != nil {
				commandContext.SetError(err)
				return &CollectMgmtClusterDiagnosticsTask{}
			}
		}

		return &CreateWorkloadClusterTask{}
	}
	logger.Info("Creating new bootstrap cluster")

	bootstrapOptions, err := commandContext.Provider.BootstrapClusterOpts(commandContext.ClusterSpec)
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
		if err = commandContext.ClusterManager.CreateAwsIamAuthCaSecret(ctx, bootstrapCluster, commandContext.ClusterSpec.Cluster.Name); err != nil {
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

func (s *CreateBootStrapClusterTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *CreateBootStrapClusterTask) Checkpoint() *task.CompletedTask {
	return nil
}

// SetAndValidateTask implementation

func (s *SetAndValidateTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Performing setup and validations")
	runner := validations.NewRunner()
	runner.Register(s.providerValidation(ctx, commandContext)...)
	runner.Register(commandContext.GitOpsManager.Validations(ctx, commandContext.ClusterSpec)...)
	runner.Register(commandContext.Validations.PreflightValidations(ctx)...)

	err := runner.Run()
	if err != nil {
		commandContext.SetError(err)
		return nil
	}
	return &CreateBootStrapClusterTask{}
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

func (s *SetAndValidateTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *SetAndValidateTask) Checkpoint() *task.CompletedTask {
	return nil
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

	logger.V(4).Info("Installing machine health checks on bootstrap cluster")
	err = commandContext.ClusterManager.InstallMachineHealthChecks(ctx, commandContext.ClusterSpec, commandContext.BootstrapCluster)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	if err = commandContext.ClusterManager.RunPostCreateWorkloadCluster(ctx, commandContext.BootstrapCluster, commandContext.WorkloadCluster, commandContext.ClusterSpec); err != nil {
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

	if commandContext.ClusterSpec.Cluster.IsSelfManaged() {
		logger.Info("Creating EKS-A namespace")
		err = commandContext.ClusterManager.CreateEKSANamespace(ctx, workloadCluster)
		if err != nil {
			commandContext.SetError(err)
			return &CollectDiagnosticsTask{}
		}

		logger.Info("Installing cluster-api providers on workload cluster")
		err = commandContext.ClusterManager.InstallCAPI(ctx, commandContext.ClusterSpec, commandContext.WorkloadCluster, commandContext.Provider)
		if err != nil {
			commandContext.SetError(err)
			return &CollectDiagnosticsTask{}
		}

		logger.Info("Installing EKS-A secrets on workload cluster")
		err := commandContext.Provider.UpdateSecrets(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec)
		if err != nil {
			commandContext.SetError(err)
			return &CollectDiagnosticsTask{}
		}
	}

	return &InstallResourcesOnManagementTask{}
}

func (s *CreateWorkloadClusterTask) Name() string {
	return "workload-cluster-init"
}

func (s *CreateWorkloadClusterTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *CreateWorkloadClusterTask) Checkpoint() *task.CompletedTask {
	return nil
}

// InstallResourcesOnManagement implementation.
func (s *InstallResourcesOnManagementTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.ClusterSpec.Cluster.IsManaged() {
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

func (s *InstallResourcesOnManagementTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *InstallResourcesOnManagementTask) Checkpoint() *task.CompletedTask {
	return nil
}

// MoveClusterManagementTask implementation

func (s *MoveClusterManagementTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.ClusterSpec.Cluster.IsManaged() {
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

func (s *MoveClusterManagementTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *MoveClusterManagementTask) Checkpoint() *task.CompletedTask {
	return nil
}

// InstallEksaComponentsTask implementation

func (s *InstallEksaComponentsTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.ClusterSpec.Cluster.IsSelfManaged() {
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

	targetCluster := commandContext.WorkloadCluster
	if commandContext.ClusterSpec.Cluster.IsManaged() {
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
	return &InstallGitOpsManagerTask{}
}

func (s *InstallEksaComponentsTask) Name() string {
	return "eksa-components-install"
}

func (s *InstallEksaComponentsTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *InstallEksaComponentsTask) Checkpoint() *task.CompletedTask {
	return nil
}

// InstallGitOpsManagerTask implementation

func (s *InstallGitOpsManagerTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Installing GitOps Toolkit on workload cluster")

	err := commandContext.GitOpsManager.InstallGitOps(ctx, commandContext.WorkloadCluster, commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec))
	if err != nil {
		logger.MarkFail("Error when installing GitOps toolkits on workload cluster; EKS-A will continue with cluster creation, but GitOps will not be enabled", "error", err)
		return &WriteClusterConfigTask{}
	}
	return &WriteClusterConfigTask{}
}

func (s *InstallGitOpsManagerTask) Name() string {
	return "gitops-manager-install"
}

func (s *InstallGitOpsManagerTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *InstallGitOpsManagerTask) Checkpoint() *task.CompletedTask {
	return nil
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

func (s *WriteClusterConfigTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *WriteClusterConfigTask) Checkpoint() *task.CompletedTask {
	return nil
}

// DeleteBootstrapClusterTask implementation

func (s *DeleteBootstrapClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.ClusterSpec.Cluster.IsSelfManaged() {
		logger.Info("Deleting bootstrap cluster")
		err := commandContext.Bootstrapper.DeleteBootstrapCluster(ctx, commandContext.BootstrapCluster, constants.Create, false)
		if err != nil {
			commandContext.SetError(err)
		}
	}
	if commandContext.OriginalError == nil {
		logger.MarkSuccess("Cluster created!")
	}
	return &InstallCuratedPackagesTask{}
}

func (s *DeleteBootstrapClusterTask) Name() string {
	return "delete-kind-cluster"
}

func (cp *InstallCuratedPackagesTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	commandContext.PackageInstaller.InstallCuratedPackages(ctx)
	return nil
}

func (cp *InstallCuratedPackagesTask) Name() string {
	return "install-curated-packages"
}

func (s *InstallCuratedPackagesTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *InstallCuratedPackagesTask) Checkpoint() *task.CompletedTask {
	return nil
}
