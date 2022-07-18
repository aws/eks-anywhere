package workflows

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

type Upgrade struct {
	bootstrapper      interfaces.Bootstrapper
	provider          providers.Provider
	clusterManager    interfaces.ClusterManager
	addonManager      interfaces.AddonManager
	writer            filewriter.FileWriter
	capiManager       interfaces.CAPIManager
	eksdInstaller     interfaces.EksdInstaller
	eksdUpgrader      interfaces.EksdUpgrader
	upgradeChangeDiff *types.ChangeDiff
}

func NewUpgrade(bootstrapper interfaces.Bootstrapper, provider providers.Provider,
	capiManager interfaces.CAPIManager,
	clusterManager interfaces.ClusterManager, addonManager interfaces.AddonManager, writer filewriter.FileWriter, eksdUpgrader interfaces.EksdUpgrader, eksdInstaller interfaces.EksdInstaller,
) *Upgrade {
	upgradeChangeDiff := types.NewChangeDiff()
	return &Upgrade{
		bootstrapper:      bootstrapper,
		provider:          provider,
		clusterManager:    clusterManager,
		addonManager:      addonManager,
		writer:            writer,
		capiManager:       capiManager,
		eksdUpgrader:      eksdUpgrader,
		eksdInstaller:     eksdInstaller,
		upgradeChangeDiff: upgradeChangeDiff,
	}
}

func (c *Upgrade) Run(ctx context.Context, clusterSpec *cluster.Spec, managementCluster *types.Cluster, workloadCluster *types.Cluster, validator interfaces.Validator, forceCleanup bool) error {
	if forceCleanup {
		if err := c.bootstrapper.DeleteBootstrapCluster(ctx, &types.Cluster{
			Name: clusterSpec.Cluster.Name,
		}, true); err != nil {
			return err
		}
	}

	commandContext := &task.CommandContext{
		Bootstrapper:      c.bootstrapper,
		Provider:          c.provider,
		ClusterManager:    c.clusterManager,
		AddonManager:      c.addonManager,
		ManagementCluster: managementCluster,
		WorkloadCluster:   workloadCluster,
		ClusterSpec:       clusterSpec,
		Validations:       validator,
		Writer:            c.writer,
		CAPIManager:       c.capiManager,
		EksdInstaller:     c.eksdInstaller,
		EksdUpgrader:      c.eksdUpgrader,
		UpgradeChangeDiff: c.upgradeChangeDiff,
	}
	if features.IsActive(features.CheckpointEnabled()) {
		return task.NewTaskRunner(&setupAndValidateTasks{}, c.writer, task.WithCheckpointFile()).RunTask(ctx, commandContext)
	}

	return task.NewTaskRunner(&setupAndValidateTasks{}, c.writer).RunTask(ctx, commandContext)
}

type setupAndValidateTasks struct{}

type updateSecrets struct{}

type ensureEtcdCAPIComponentsExistTask struct{}

type upgradeCoreComponents struct {
	UpgradeChangeDiff *types.ChangeDiff
}

type upgradeNeeded struct{}

type pauseEksaAndFluxReconcile struct{}

type createBootstrapClusterTask struct {
	bootstrapCluster *types.Cluster
}

type installCAPITask struct{}

type moveManagementToBootstrapTask struct{}

type moveManagementToWorkloadTask struct{}

type upgradeWorkloadClusterTask struct{}

type deleteBootstrapClusterTask struct {
	*CollectDiagnosticsTask
}

type updateClusterAndGitResources struct{}

type resumeFluxReconcile struct{}

type writeClusterConfigTask struct{}

func (s *setupAndValidateTasks) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Performing setup and validations")
	currentSpec, err := commandContext.ClusterManager.GetCurrentClusterSpec(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec.Cluster.Name)
	if err != nil {
		commandContext.SetError(err)
		return nil
	}
	commandContext.CurrentClusterSpec = currentSpec
	runner := validations.NewRunner()
	runner.Register(s.validations(ctx, commandContext)...)

	err = runner.Run()
	if err != nil {
		commandContext.SetError(err)
		return nil
	}

	return &updateSecrets{}
}

func (s *setupAndValidateTasks) validations(ctx context.Context, commandContext *task.CommandContext) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: fmt.Sprintf("%s provider validation", commandContext.Provider.Name()),
				Err:  commandContext.Provider.SetupAndValidateUpgradeCluster(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.CurrentClusterSpec),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: "upgrade preflight validations pass",
				Err:  commandContext.Validations.PreflightValidations(ctx),
			}
		},
	}
}

func (s *setupAndValidateTasks) Name() string {
	return "setup-and-validate"
}

func (s *setupAndValidateTasks) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	if err := commandContext.Provider.SetupAndValidateUpgradeCluster(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.CurrentClusterSpec); err != nil {
		commandContext.SetError(err)
		return nil, err
	}
	logger.Info(fmt.Sprintf("%s Provider setup is valid", commandContext.Provider.Name()))
	currentSpec, err := commandContext.ClusterManager.GetCurrentClusterSpec(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec.Cluster.Name)
	if err != nil {
		commandContext.SetError(err)
		return nil, err
	}
	commandContext.CurrentClusterSpec = currentSpec
	return &updateSecrets{}, nil
}

func (s *setupAndValidateTasks) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *updateSecrets) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	err := commandContext.Provider.UpdateSecrets(ctx, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	return &ensureEtcdCAPIComponentsExistTask{}
}

func (s *updateSecrets) Name() string {
	return "update-secrets"
}

func (s *updateSecrets) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *updateSecrets) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &ensureEtcdCAPIComponentsExistTask{}, nil
}

func (s *ensureEtcdCAPIComponentsExistTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Ensuring etcd CAPI providers exist on management cluster before upgrade")
	if err := commandContext.CAPIManager.EnsureEtcdProvidersInstallation(ctx, commandContext.ManagementCluster, commandContext.Provider, commandContext.CurrentClusterSpec); err != nil {
		commandContext.SetError(err)
		return nil
	}
	return &upgradeCoreComponents{}
}

func (s *ensureEtcdCAPIComponentsExistTask) Name() string {
	return "ensure-etcd-capi-components-exist"
}

func (s *ensureEtcdCAPIComponentsExistTask) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *ensureEtcdCAPIComponentsExistTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &upgradeCoreComponents{}, nil
}

func (s *upgradeCoreComponents) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Upgrading core components")

	changeDiff, err := commandContext.ClusterManager.UpgradeNetworking(ctx, commandContext.WorkloadCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	changeDiff, err = commandContext.CAPIManager.Upgrade(ctx, commandContext.ManagementCluster, commandContext.Provider, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	err = commandContext.AddonManager.UpdateLegacyFileStructure(ctx, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	changeDiff, err = commandContext.AddonManager.Upgrade(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	changeDiff, err = commandContext.ClusterManager.Upgrade(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	changeDiff, err = commandContext.EksdUpgrader.Upgrade(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)
	s.UpgradeChangeDiff = commandContext.UpgradeChangeDiff

	return &upgradeNeeded{}
}

func (s *upgradeCoreComponents) Name() string {
	return "upgrade-core-components"
}

func (s *upgradeCoreComponents) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: s.UpgradeChangeDiff,
	}
}

func (s *upgradeCoreComponents) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	s.UpgradeChangeDiff = &types.ChangeDiff{}
	if err := task.UnmarshalTaskCheckpoint(completedTask.Checkpoint, s.UpgradeChangeDiff); err != nil {
		return nil, err
	}
	commandContext.UpgradeChangeDiff = s.UpgradeChangeDiff
	return &upgradeNeeded{}, nil
}

func (s *upgradeNeeded) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	newSpec := commandContext.ClusterSpec

	if upgradeNeeded, err := commandContext.Provider.UpgradeNeeded(ctx, newSpec, commandContext.CurrentClusterSpec, commandContext.ManagementCluster); err != nil {
		commandContext.SetError(err)
		return nil
	} else if upgradeNeeded {
		logger.V(3).Info("Provider needs a cluster upgrade")
		return &pauseEksaAndFluxReconcile{}
	}
	diff, err := commandContext.ClusterManager.EKSAClusterSpecChanged(ctx, commandContext.ManagementCluster, newSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	if !diff {
		logger.Info("No upgrades needed from cluster spec")
		return nil
	}

	return &pauseEksaAndFluxReconcile{}
}

func (s *upgradeNeeded) Name() string {
	return "upgrade-needed"
}

func (s *upgradeNeeded) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *upgradeNeeded) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &pauseEksaAndFluxReconcile{}, nil
}

func (s *pauseEksaAndFluxReconcile) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Pausing EKS-A cluster controller reconcile")
	err := commandContext.ClusterManager.PauseEKSAControllerReconcile(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	logger.Info("Pausing Flux kustomization")
	err = commandContext.AddonManager.PauseGitOpsKustomization(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	return &createBootstrapClusterTask{}
}

func (s *pauseEksaAndFluxReconcile) Name() string {
	return "pause-controllers-reconcile"
}

func (s *pauseEksaAndFluxReconcile) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *pauseEksaAndFluxReconcile) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &createBootstrapClusterTask{}, nil
}

func (s *createBootstrapClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.ManagementCluster != nil && commandContext.ManagementCluster.ExistingManagement {
		return &upgradeWorkloadClusterTask{}
	}
	logger.Info("Creating bootstrap cluster")
	bootstrapOptions, err := commandContext.Provider.BootstrapClusterOpts()
	if err != nil {
		commandContext.SetError(err)
		return nil
	}

	bootstrapCluster, err := commandContext.Bootstrapper.CreateBootstrapCluster(ctx, commandContext.ClusterSpec, bootstrapOptions...)
	commandContext.BootstrapCluster = bootstrapCluster
	if err != nil {
		commandContext.SetError(err)
		return &deleteBootstrapClusterTask{}
	}

	logger.Info("Provider specific pre-capi-install-setup on bootstrap cluster")
	if err = commandContext.Provider.PreCAPIInstallOnBootstrap(ctx, bootstrapCluster, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	logger.Info("Provider specific post-setup")
	if err = commandContext.Provider.PostBootstrapSetupUpgrade(ctx, commandContext.ClusterSpec.Cluster, bootstrapCluster); err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}
	s.bootstrapCluster = bootstrapCluster

	return &installCAPITask{}
}

func (s *createBootstrapClusterTask) Name() string {
	return "bootstrap-cluster-init"
}

func (s *createBootstrapClusterTask) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: s.bootstrapCluster,
	}
}

func (s *createBootstrapClusterTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	s.bootstrapCluster = &types.Cluster{}
	if err := task.UnmarshalTaskCheckpoint(completedTask.Checkpoint, s.bootstrapCluster); err != nil {
		return nil, err
	}
	commandContext.BootstrapCluster = s.bootstrapCluster
	if commandContext.ManagementCluster != nil && commandContext.ManagementCluster.ExistingManagement {
		return &upgradeWorkloadClusterTask{}, nil
	}
	return &installCAPITask{}, nil
}

func (s *installCAPITask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Installing cluster-api providers on bootstrap cluster")
	err := commandContext.ClusterManager.InstallCAPI(ctx, commandContext.ClusterSpec, commandContext.BootstrapCluster, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}
	return &moveManagementToBootstrapTask{}
}

func (s *installCAPITask) Name() string {
	return "install-capi"
}

func (s *installCAPITask) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *installCAPITask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &moveManagementToBootstrapTask{}, nil
}

func (s *moveManagementToBootstrapTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Moving cluster management from workload to bootstrap cluster")
	err := commandContext.ClusterManager.MoveCAPI(ctx, commandContext.WorkloadCluster, commandContext.BootstrapCluster, commandContext.WorkloadCluster.Name, commandContext.ClusterSpec, types.WithNodeRef(), types.WithNodeHealthy())
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	commandContext.ManagementCluster = commandContext.BootstrapCluster
	return &upgradeWorkloadClusterTask{}
}

func (s *moveManagementToBootstrapTask) Name() string {
	return "capi-management-move-to-bootstrap"
}

func (s *moveManagementToBootstrapTask) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *moveManagementToBootstrapTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	commandContext.ManagementCluster = commandContext.BootstrapCluster
	return &upgradeWorkloadClusterTask{}, nil
}

func (s *upgradeWorkloadClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	eksaManagementCluster := commandContext.WorkloadCluster
	if commandContext.ManagementCluster != nil && commandContext.ManagementCluster.ExistingManagement {
		eksaManagementCluster = commandContext.ManagementCluster
	}

	logger.Info("Upgrading workload cluster")
	err := commandContext.ClusterManager.UpgradeCluster(ctx, commandContext.ManagementCluster, commandContext.WorkloadCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	if commandContext.UpgradeChangeDiff.Changed() {
		if err = commandContext.ClusterManager.ApplyBundles(ctx, commandContext.ClusterSpec, eksaManagementCluster); err != nil {
			commandContext.SetError(err)
			return &CollectDiagnosticsTask{}
		}
	}

	return &moveManagementToWorkloadTask{}
}

func (s *upgradeWorkloadClusterTask) Name() string {
	return "upgrade-workload-cluster"
}

func (s *upgradeWorkloadClusterTask) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *upgradeWorkloadClusterTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &moveManagementToWorkloadTask{}, nil
}

func (s *moveManagementToWorkloadTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.ManagementCluster.ExistingManagement {
		return &updateClusterAndGitResources{}
	}
	logger.Info("Moving cluster management from bootstrap to workload cluster")
	err := commandContext.ClusterManager.MoveCAPI(ctx, commandContext.BootstrapCluster, commandContext.WorkloadCluster, commandContext.WorkloadCluster.Name, commandContext.ClusterSpec, types.WithNodeRef(), types.WithNodeHealthy())
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	commandContext.ManagementCluster = commandContext.WorkloadCluster
	return &updateClusterAndGitResources{}
}

func (s *moveManagementToWorkloadTask) Name() string {
	return "capi-management-move-to-workload"
}

func (s *moveManagementToWorkloadTask) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *moveManagementToWorkloadTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	if !commandContext.ManagementCluster.ExistingManagement {
		commandContext.ManagementCluster = commandContext.WorkloadCluster
	}
	return &updateClusterAndGitResources{}, nil
}

func (s *updateClusterAndGitResources) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Applying new EKS-A cluster resource; resuming reconcile")
	datacenterConfig := commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec)
	machineConfigs := commandContext.Provider.MachineConfigs(commandContext.ClusterSpec)
	err := commandContext.ClusterManager.CreateEKSAResources(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, datacenterConfig, machineConfigs)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	err = commandContext.EksdInstaller.InstallEksdManifest(ctx, commandContext.ClusterSpec, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	logger.Info("Resuming EKS-A controller reconciliation")
	err = commandContext.ClusterManager.ResumeEKSAControllerReconcile(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	logger.Info("Updating Git Repo with new EKS-A cluster spec")
	err = commandContext.AddonManager.UpdateGitEksaSpec(ctx, commandContext.ClusterSpec, datacenterConfig, machineConfigs)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	return &resumeFluxReconcile{}
}

func (s *updateClusterAndGitResources) Name() string {
	return "update-resources"
}

func (s *updateClusterAndGitResources) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *updateClusterAndGitResources) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &resumeFluxReconcile{}, nil
}

func (s *resumeFluxReconcile) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Forcing reconcile Git repo with latest commit")
	err := commandContext.AddonManager.ForceReconcileGitRepo(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	logger.Info("Resuming Flux kustomization")
	err = commandContext.AddonManager.ResumeGitOpsKustomization(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &writeClusterConfigTask{}
	}
	return &writeClusterConfigTask{}
}

func (s *resumeFluxReconcile) Name() string {
	return "resume-flux-kustomization"
}

func (s *resumeFluxReconcile) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *resumeFluxReconcile) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &writeClusterConfigTask{}, nil
}

func (s *writeClusterConfigTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Writing cluster config file")
	err := clustermarshaller.WriteClusterConfig(commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec), commandContext.Writer)
	if err != nil {
		commandContext.SetError(err)
	}
	return &deleteBootstrapClusterTask{}
}

func (s *writeClusterConfigTask) Name() string {
	return "write-cluster-config"
}

func (s *writeClusterConfigTask) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *writeClusterConfigTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &deleteBootstrapClusterTask{}, nil
}

func (s *deleteBootstrapClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.OriginalError != nil {
		c := CollectDiagnosticsTask{}
		c.Run(ctx, commandContext)
	}
	if commandContext.BootstrapCluster != nil && !commandContext.BootstrapCluster.ExistingManagement {
		if err := commandContext.Bootstrapper.DeleteBootstrapCluster(ctx, commandContext.BootstrapCluster, true); err != nil {
			commandContext.SetError(err)
		}
		if commandContext.OriginalError == nil {
			logger.MarkSuccess("Cluster upgraded!")
		}
		return nil
	}
	logger.Info("Bootstrap cluster information missing - skipping delete kind cluster")
	if commandContext.OriginalError == nil {
		logger.MarkSuccess("Cluster upgraded!")
	}
	return nil
}

func (s *deleteBootstrapClusterTask) Name() string {
	return "delete-kind-cluster"
}
