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

// UpgradeManagement prepares management cluster to upgrade.
type UpgradeManagement struct {
	provider           providers.Provider
	clusterManager     interfaces.ClusterManager
	gitOpsManager      interfaces.GitOpsManager
	writer             filewriter.FileWriter
	capiManager        interfaces.CAPIManager
	eksdInstaller      interfaces.EksdInstaller
	eksdUpgrader       interfaces.EksdUpgrader
	upgradeChangeDiff  *types.ChangeDiff
	managementUpgrader interfaces.ManagementUpgrader
}

// NewUpgradeManagement builds a new NewUpgradeManagement.
func NewUpgradeManagement(provider providers.Provider,
	capiManager interfaces.CAPIManager,
	clusterManager interfaces.ClusterManager,
	gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter,
	eksdUpgrader interfaces.EksdUpgrader,
	eksdInstaller interfaces.EksdInstaller,
	managementUpgrader interfaces.ManagementUpgrader,
) *UpgradeManagement {
	upgradeChangeDiff := types.NewChangeDiff()
	return &UpgradeManagement{
		provider:           provider,
		clusterManager:     clusterManager,
		gitOpsManager:      gitOpsManager,
		writer:             writer,
		capiManager:        capiManager,
		eksdUpgrader:       eksdUpgrader,
		eksdInstaller:      eksdInstaller,
		upgradeChangeDiff:  upgradeChangeDiff,
		managementUpgrader: managementUpgrader,
	}
}

// Run runs the upgrade management cluster workflow.
func (c *UpgradeManagement) Run(ctx context.Context, clusterSpec *cluster.Spec, managementCluster *types.Cluster, validator interfaces.Validator) error {
	commandContext := &task.CommandContext{
		Provider:           c.provider,
		ClusterManager:     c.clusterManager,
		GitOpsManager:      c.gitOpsManager,
		ManagementCluster:  managementCluster,
		ClusterSpec:        clusterSpec,
		Validations:        validator,
		Writer:             c.writer,
		CAPIManager:        c.capiManager,
		EksdInstaller:      c.eksdInstaller,
		EksdUpgrader:       c.eksdUpgrader,
		UpgradeChangeDiff:  c.upgradeChangeDiff,
		ManagementUpgrader: c.managementUpgrader,
	}
	if features.IsActive(features.CheckpointEnabled()) {
		return task.NewTaskRunner(&setupAndValidateManagementTasks{}, c.writer, task.WithCheckpointFile()).RunTask(ctx, commandContext)
	}

	return task.NewTaskRunner(&setupAndValidateManagementTasks{}, c.writer).RunTask(ctx, commandContext)
}

type setupAndValidateManagementTasks struct{}

type updateManagementSecrets struct{}

type ensureManagementEtcdCAPIComponentsExistTask struct{}

type upgradeManagementCoreComponents struct {
	UpgradeChangeDiff *types.ChangeDiff
}

type upgradeManagementNeeded struct{}

type pauseManagementGitOpsReconcile struct{}

type upgradeManagementClusterTask struct{}

type managementClusterBeforeUpgrade struct{}

// reconcileManagementClusterDefinitions updates all the places that have a cluster definition to follow the cluster config provided to this workflow:
// the eks-a objects in the management cluster and the cluster config in the git repo if GitOps is enabled. It also resumes the eks-a controller
// manager and GitOps reconciliations.
type reconcileManagementClusterDefinitions struct {
	eksaSpecDiff bool
}

type writeManagementClusterConfigTask struct {
	*CollectDiagnosticsTask
}

func (s *setupAndValidateManagementTasks) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
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

	return &updateManagementSecrets{}
}

func (s *setupAndValidateManagementTasks) providerValidation(ctx context.Context, commandContext *task.CommandContext) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: fmt.Sprintf("%s provider validation", commandContext.Provider.Name()),
				Err:  commandContext.Provider.SetupAndValidateUpgradeCluster(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.CurrentClusterSpec),
			}
		},
	}
}

func (s *setupAndValidateManagementTasks) Name() string {
	return "setup-and-validate"
}

func (s *setupAndValidateManagementTasks) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
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
	return &updateManagementSecrets{}, nil
}

func (s *setupAndValidateManagementTasks) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *updateManagementSecrets) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	err := commandContext.Provider.UpdateSecrets(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}
	return &ensureManagementEtcdCAPIComponentsExistTask{}
}

func (s *updateManagementSecrets) Name() string {
	return "update-secrets"
}

func (s *updateManagementSecrets) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *updateManagementSecrets) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &ensureManagementEtcdCAPIComponentsExistTask{}, nil
}

func (s *ensureManagementEtcdCAPIComponentsExistTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Ensuring etcd CAPI providers exist on management cluster before upgrade")
	if err := commandContext.CAPIManager.EnsureEtcdProvidersInstallation(ctx, commandContext.ManagementCluster, commandContext.Provider, commandContext.CurrentClusterSpec); err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}
	return &pauseManagementGitOpsReconcile{}
}

func (s *ensureManagementEtcdCAPIComponentsExistTask) Name() string {
	return "ensure-etcd-capi-components-exist"
}

func (s *ensureManagementEtcdCAPIComponentsExistTask) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *ensureManagementEtcdCAPIComponentsExistTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &pauseManagementGitOpsReconcile{}, nil
}

func (s *pauseManagementGitOpsReconcile) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Pausing GitOps cluster resources reconcile")
	err := commandContext.GitOpsManager.PauseClusterResourcesReconcile(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	return &upgradeManagementCoreComponents{}
}

func (s *pauseManagementGitOpsReconcile) Name() string {
	return "pause-controllers-reconcile"
}

func (s *pauseManagementGitOpsReconcile) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *pauseManagementGitOpsReconcile) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &upgradeManagementCoreComponents{}, nil
}

func (s *upgradeManagementCoreComponents) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Upgrading core components")
	err := commandContext.Provider.PreCoreComponentsUpgrade(
		ctx,
		commandContext.ManagementCluster,
		commandContext.ClusterSpec,
	)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	changeDiff, err := commandContext.ClusterManager.UpgradeNetworking(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	changeDiff, err = commandContext.CAPIManager.Upgrade(ctx, commandContext.ManagementCluster, commandContext.Provider, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	if err = commandContext.GitOpsManager.Install(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	changeDiff, err = commandContext.GitOpsManager.Upgrade(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	changeDiff, err = commandContext.ClusterManager.Upgrade(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	changeDiff, err = commandContext.EksdUpgrader.Upgrade(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	s.UpgradeChangeDiff = commandContext.UpgradeChangeDiff

	if err = commandContext.ClusterManager.ApplyBundles(ctx, commandContext.ClusterSpec, commandContext.ManagementCluster); err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	if err = commandContext.ClusterManager.ApplyReleases(ctx, commandContext.ClusterSpec, commandContext.ManagementCluster); err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	return &upgradeManagementNeeded{}
}

func (s *upgradeManagementCoreComponents) Name() string {
	return "upgrade-core-components"
}

func (s *upgradeManagementCoreComponents) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: s.UpgradeChangeDiff,
	}
}

func (s *upgradeManagementCoreComponents) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	s.UpgradeChangeDiff = &types.ChangeDiff{}
	if err := task.UnmarshalTaskCheckpoint(completedTask.Checkpoint, s.UpgradeChangeDiff); err != nil {
		return nil, err
	}
	commandContext.UpgradeChangeDiff = s.UpgradeChangeDiff
	return &upgradeManagementNeeded{}, nil
}

func (s *upgradeManagementNeeded) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	newSpec := commandContext.ClusterSpec

	if upgradeManagementNeeded, err := commandContext.Provider.UpgradeNeeded(ctx, newSpec, commandContext.CurrentClusterSpec, commandContext.ManagementCluster); err != nil {
		commandContext.SetError(err)
		return nil
	} else if upgradeManagementNeeded {
		logger.V(3).Info("Provider needs a cluster upgrade")
		return &managementClusterBeforeUpgrade{}
	}
	diff, err := commandContext.ClusterManager.EKSAClusterSpecChanged(ctx, commandContext.ManagementCluster, newSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	if !diff {
		logger.Info("No upgrades needed from cluster spec")
		return &reconcileManagementClusterDefinitions{eksaSpecDiff: false}
	}
	return &managementClusterBeforeUpgrade{}
}

func (s *upgradeManagementNeeded) Name() string {
	return "upgrade-needed"
}

func (s *upgradeManagementNeeded) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *upgradeManagementNeeded) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &managementClusterBeforeUpgrade{}, nil
}

func (s *managementClusterBeforeUpgrade) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Backing up workload cluster's management resources before moving the upgrade")
	err := commandContext.ClusterManager.BackupCAPIWaitForInfrastructure(ctx, commandContext.ManagementCluster, commandContext.ManagementClusterStateDir, commandContext.ManagementCluster.Name)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	logger.V(3).Info("Pausing workload clusters before upgrading management cluster")
	err = commandContext.ClusterManager.PauseCAPIWorkloadClusters(ctx, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	return &upgradeManagementClusterTask{}
}

func (s *managementClusterBeforeUpgrade) Name() string {
	return "management-cluster-before-upgrade"
}

func (s *managementClusterBeforeUpgrade) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *managementClusterBeforeUpgrade) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &upgradeManagementClusterTask{}, nil
}

func (s *upgradeManagementClusterTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Updating EKS-A cluster resource")
	datacenterConfig := commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec)
	machineConfigs := commandContext.Provider.MachineConfigs(commandContext.ClusterSpec)
	err := commandContext.ClusterManager.UpgradeEKSAResources(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, datacenterConfig, machineConfigs)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}
	err = commandContext.EksdInstaller.InstallEksdManifest(ctx, commandContext.ClusterSpec, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	logger.Info("Upgrading management cluster")
	err = commandContext.ManagementUpgrader.UpgradeManagementCluster(ctx, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	return &reconcileManagementClusterDefinitions{eksaSpecDiff: true}
}

func (s *upgradeManagementClusterTask) Name() string {
	return "upgrade-management-cluster"
}

func (s *upgradeManagementClusterTask) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *upgradeManagementClusterTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &reconcileManagementClusterDefinitions{eksaSpecDiff: true}, nil
}

func (s *reconcileManagementClusterDefinitions) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Updating EKS-A cluster resource")
	datacenterConfig := commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec)
	machineConfigs := commandContext.Provider.MachineConfigs(commandContext.ClusterSpec)

	logger.V(3).Info("Resuming all workload clusters management cluster components have been upgraded")
	err := commandContext.ClusterManager.ResumeCAPIWorkloadClusters(ctx, commandContext.ManagementCluster)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	logger.Info("Updating Git Repo with new EKS-A cluster spec")
	err = commandContext.GitOpsManager.UpdateGitEksaSpec(ctx, commandContext.ClusterSpec, datacenterConfig, machineConfigs)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	logger.Info("Forcing reconcile Git repo with latest commit")
	err = commandContext.GitOpsManager.ForceReconcileGitRepo(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return &CollectMgmtClusterDiagnosticsTask{}
	}

	logger.Info("Resuming GitOps cluster resources kustomization")
	err = commandContext.GitOpsManager.ResumeClusterResourcesReconcile(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec, commandContext.Provider)
	if err != nil {
		commandContext.SetError(err)
		return &writeManagementClusterConfigTask{}
	}
	if !s.eksaSpecDiff {
		return nil
	}
	return &writeManagementClusterConfigTask{}
}

func (s *reconcileManagementClusterDefinitions) Name() string {
	return "resume-gitops-kustomization"
}

func (s *reconcileManagementClusterDefinitions) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *reconcileManagementClusterDefinitions) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &writeManagementClusterConfigTask{}, nil
}

func (s *writeManagementClusterConfigTask) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Writing cluster config file")
	err := clustermarshaller.WriteClusterConfig(commandContext.ClusterSpec, commandContext.Provider.DatacenterConfig(commandContext.ClusterSpec), commandContext.Provider.MachineConfigs(commandContext.ClusterSpec), commandContext.Writer)
	if err != nil {
		commandContext.SetError(err)
	}
	if commandContext.OriginalError != nil {
		c := CollectMgmtClusterDiagnosticsTask{}
		c.Run(ctx, commandContext)
	}
	if commandContext.OriginalError == nil {
		logger.MarkSuccess("Cluster upgraded!")
	}
	return nil
}

func (s *writeManagementClusterConfigTask) Name() string {
	return "write-cluster-config"
}

func (s *writeManagementClusterConfigTask) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *writeManagementClusterConfigTask) Restore(ctx context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	return &deleteBootstrapClusterTask{}, nil
}
