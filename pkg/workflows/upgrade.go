package workflows

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/constants"
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
	gitOpsManager     interfaces.GitOpsManager
	writer            filewriter.FileWriter
	capiManager       interfaces.CAPIManager
	eksdInstaller     interfaces.EksdInstaller
	eksdUpgrader      interfaces.EksdUpgrader
	upgradeChangeDiff *types.ChangeDiff
}

func NewUpgrade(bootstrapper interfaces.Bootstrapper, provider providers.Provider,
	capiManager interfaces.CAPIManager,
	clusterManager interfaces.ClusterManager,
	gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter,
	eksdUpgrader interfaces.EksdUpgrader,
	eksdInstaller interfaces.EksdInstaller,
) *Upgrade {
	upgradeChangeDiff := types.NewChangeDiff()
	return &Upgrade{
		bootstrapper:      bootstrapper,
		provider:          provider,
		clusterManager:    clusterManager,
		gitOpsManager:     gitOpsManager,
		writer:            writer,
		capiManager:       capiManager,
		eksdUpgrader:      eksdUpgrader,
		eksdInstaller:     eksdInstaller,
		upgradeChangeDiff: upgradeChangeDiff,
	}
}

func (c *Upgrade) Run(ctx context.Context, clusterSpec *cluster.Spec, managementCluster *types.Cluster, workloadCluster *types.Cluster, validator interfaces.Validator, forceCleanup bool) error {
	commandContext := &task.CommandContext{
		Bootstrapper:      c.bootstrapper,
		Provider:          c.provider,
		ClusterManager:    c.clusterManager,
		GitOpsManager:     c.gitOpsManager,
		ManagementCluster: managementCluster,
		WorkloadCluster:   workloadCluster,
		ClusterSpec:       clusterSpec,
		Validations:       validator,
		Writer:            c.writer,
		CAPIManager:       c.capiManager,
		EksdInstaller:     c.eksdInstaller,
		EksdUpgrader:      c.eksdUpgrader,
		UpgradeChangeDiff: c.upgradeChangeDiff,
		ForceCleanup:      forceCleanup,
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

type pauseEksaReconcile struct{}

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

// reconcileClusterDefinitions updates all the places that have a cluster definition to follow the cluster config provided to this workflow:
// the eks-a objects in the management cluster and the cluster config in the git repo if GitOps is enabled. It also resumes the eks-a controller
// manager and GitOps reconciliations.
type reconcileClusterDefinitions struct {
	eksaSpecDiff bool
}

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
	runner.Register(s.providerValidation(ctx, commandContext)...)
	runner.Register(commandContext.Validations.PreflightValidations(ctx)...)

	err = runner.Run()
	if err != nil {
		commandContext.SetError(err)
		return nil
	}

	return &updateSecrets{}
}

func (s *setupAndValidateTasks) providerValidation(ctx context.Context, commandContext *task.CommandContext) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: fmt.Sprintf("%s provider validation", commandContext.Provider.Name()),
				Err:  commandContext.Provider.SetupAndValidateUpgradeCluster(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.CurrentClusterSpec),
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
	err := commandContext.Provider.UpdateSecrets(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec)
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
	return &pauseEksaReconcile{}
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
	return &pauseEksaReconcile{}, nil
}

func (s *upgradeCoreComponents) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Upgrading core components")

	err := commandContext.Provider.PreCoreComponentsUpgrade(
		ctx,
		commandContext.ManagementCluster,
		commandContext.ClusterSpec,
	)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

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

	if err = commandContext.GitOpsManager.Install(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	changeDiff, err = commandContext.GitOpsManager.Upgrade(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
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
		return &createBootstrapClusterTask{}
	}
	diff, err := commandContext.ClusterManager.EKSAClusterSpecChanged(ctx, commandContext.ManagementCluster, newSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	if !diff {
		logger.Info("No upgrades needed from cluster spec")
		return &reconcileClusterDefinitions{eksaSpecDiff: false}
	}

	return &createBootstrapClusterTask{}
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
	return &createBootstrapClusterTask{}, nil
}

func (s *pauseEksaReconcile) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Pausing EKS-A cluster controller reconcile")
	err := commandContext.ClusterManager.PauseEKSAControllerReconcile(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	logger.Info("Pausing GitOps cluster resources reconcile")
	err = commandContext.GitOpsManager.PauseClusterResourcesReconcile(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	return &upgradeCoreComponents{}
}

func (s *pauseEksaReconcile) Name() string {
	return "pause-controllers-reconcile"
}

func (s *pauseEksaReconcile) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *pauseEksaReconcile) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &upgradeCoreComponents{}, nil
}

func (s *createBootstrapClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if commandContext.ForceCleanup {
		if err := commandContext.Bootstrapper.DeleteBootstrapCluster(ctx, &types.Cluster{
			Name: commandContext.ClusterSpec.Cluster.Name,
		}, constants.Upgrade, commandContext.ForceCleanup); err != nil {
			commandContext.SetError(err)
			return nil
		}
	}
	if commandContext.ClusterSpec.Cluster.IsManaged() {
		return &upgradeWorkloadClusterTask{}
	}
	logger.Info("Creating bootstrap cluster")
	bootstrapOptions, err := commandContext.Provider.BootstrapClusterOpts(commandContext.ClusterSpec)
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
	if commandContext.ClusterSpec.Cluster.IsManaged() {
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
	// Take best effort CAPI backup of workload cluster without filter.
	// If that errors, then take CAPI backup filtering on only workload cluster.
	logger.Info("Backing up workload cluster's management resources before moving to bootstrap cluster")
	err := commandContext.ClusterManager.BackupCAPI(ctx, commandContext.WorkloadCluster, commandContext.BackupClusterStateDir, "")
	if err != nil {
		err = commandContext.ClusterManager.BackupCAPIWaitForInfrastructure(ctx, commandContext.WorkloadCluster, commandContext.BackupClusterStateDir, commandContext.WorkloadCluster.Name)
		if err != nil {
			commandContext.SetError(err)
			return &CollectDiagnosticsTask{}
		}
	}

	logger.V(3).Info("Pausing workload clusters before moving management cluster resources to bootstrap cluster")
	err = commandContext.ClusterManager.PauseCAPIWorkloadClusters(ctx, commandContext.WorkloadCluster)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	logger.Info("Moving management cluster from workload to bootstrap cluster")
	err = commandContext.ClusterManager.MoveCAPI(ctx, commandContext.WorkloadCluster, commandContext.BootstrapCluster, commandContext.WorkloadCluster.Name, commandContext.ClusterSpec, types.WithNodeRef(), types.WithNodeHealthy())
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	logger.V(3).Info("Provider specific post management move")
	err = commandContext.Provider.PostMoveManagementToBootstrap(ctx, commandContext.BootstrapCluster)
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
	if commandContext.ClusterSpec.Cluster.IsManaged() {
		eksaManagementCluster = commandContext.ManagementCluster
	}

	logger.Info("Upgrading workload cluster")
	err := commandContext.ClusterManager.UpgradeCluster(ctx, commandContext.ManagementCluster, commandContext.WorkloadCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		// Take backup of bootstrap cluster capi components
		if commandContext.BootstrapCluster != nil {
			logger.Info("Backing up management components from bootstrap cluster")
			err := commandContext.ClusterManager.BackupCAPIWaitForInfrastructure(ctx, commandContext.BootstrapCluster, fmt.Sprintf("bootstrap-%s", commandContext.BackupClusterStateDir), commandContext.WorkloadCluster.Name)
			if err != nil {
				logger.Info("Bootstrap management component backup failed, use existing workload cluster backup", "error", err)
			}
		}
		return &CollectDiagnosticsTask{}
	}

	if commandContext.UpgradeChangeDiff.Changed() {
		if err = commandContext.ClusterManager.ApplyBundles(ctx, commandContext.ClusterSpec, eksaManagementCluster); err != nil {
			commandContext.SetError(err)
			return &CollectDiagnosticsTask{}
		}

		if err = commandContext.ClusterManager.ApplyReleases(ctx, commandContext.ClusterSpec, eksaManagementCluster); err != nil {
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
	if commandContext.ClusterSpec.Cluster.IsManaged() {
		return &reconcileClusterDefinitions{eksaSpecDiff: true}
	}
	logger.Info("Moving cluster management from bootstrap to workload cluster")
	err := commandContext.ClusterManager.MoveCAPI(ctx, commandContext.BootstrapCluster, commandContext.WorkloadCluster, commandContext.WorkloadCluster.Name, commandContext.ClusterSpec, types.WithNodeRef(), types.WithNodeHealthy())
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}
	commandContext.ManagementCluster = commandContext.WorkloadCluster

	logger.V(3).Info("Resuming all workload clusters after moving management cluster resources from bootstrap to management clusters")
	err = commandContext.ClusterManager.ResumeCAPIWorkloadClusters(ctx, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	return &reconcileClusterDefinitions{eksaSpecDiff: true}
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
	if commandContext.ClusterSpec.Cluster.IsSelfManaged() {
		commandContext.ManagementCluster = commandContext.WorkloadCluster
	}
	return &reconcileClusterDefinitions{eksaSpecDiff: true}, nil
}

func (s *reconcileClusterDefinitions) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Updating EKS-A cluster resource")
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

	logger.Info("Resuming EKS-A controller reconcile")
	err = commandContext.ClusterManager.ResumeEKSAControllerReconcile(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	logger.Info("Updating Git Repo with new EKS-A cluster spec")
	err = commandContext.GitOpsManager.UpdateGitEksaSpec(ctx, commandContext.ClusterSpec, datacenterConfig, machineConfigs)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	logger.Info("Forcing reconcile Git repo with latest commit")
	err = commandContext.GitOpsManager.ForceReconcileGitRepo(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectDiagnosticsTask{}
	}

	logger.Info("Resuming GitOps cluster resources kustomization")
	err = commandContext.GitOpsManager.ResumeClusterResourcesReconcile(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &writeClusterConfigTask{}
	}
	if !s.eksaSpecDiff {
		return nil
	}
	return &writeClusterConfigTask{}
}

func (s *reconcileClusterDefinitions) Name() string {
	return "resume-eksa-and-gitops-kustomization"
}

func (s *reconcileClusterDefinitions) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *reconcileClusterDefinitions) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
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
	if commandContext.ClusterSpec.Cluster.IsSelfManaged() {
		if err := commandContext.Bootstrapper.DeleteBootstrapCluster(ctx, commandContext.BootstrapCluster, constants.Upgrade, false); err != nil {
			commandContext.SetError(err)
		}

		if commandContext.OriginalError == nil {
			logger.MarkSuccess("Cluster upgraded!")
		}
		if err := commandContext.Provider.PostBootstrapDeleteForUpgrade(ctx, commandContext.ManagementCluster); err != nil {
			// Cluster has been successfully upgraded, bootstrap cluster successfully deleted
			// We don't necessarily need to return with an error here and abort
			logger.Info(fmt.Sprintf("%v", err))
		}

		capiObjectFile := filepath.Join(commandContext.BootstrapCluster.Name, commandContext.BackupClusterStateDir)
		if err := os.RemoveAll(capiObjectFile); err != nil {
			logger.Info(fmt.Sprintf("management cluster CAPI backup file not found: %v", err))
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
