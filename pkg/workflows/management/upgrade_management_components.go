package management

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/workflows"
	"github.com/aws/eks-anywhere/pkg/workflows/interfaces"
)

// UpgradeManagementComponentsWorkflow is a schema for upgrade management components.
type UpgradeManagementComponentsWorkflow struct {
	clientFactory  interfaces.ClientFactory
	provider       providers.Provider
	clusterManager interfaces.ClusterManager
	gitOpsManager  interfaces.GitOpsManager
	writer         filewriter.FileWriter
	capiManager    interfaces.CAPIManager
	eksdInstaller  interfaces.EksdInstaller
	eksdUpgrader   interfaces.EksdUpgrader
}

// NewUpgradeManagementComponentsRunner builds a new UpgradeManagementCommponents construct.
func NewUpgradeManagementComponentsRunner(
	clientFactory interfaces.ClientFactory,
	provider providers.Provider,
	capiManager interfaces.CAPIManager,
	clusterManager interfaces.ClusterManager,
	gitOpsManager interfaces.GitOpsManager,
	writer filewriter.FileWriter,
	eksdUpgrader interfaces.EksdUpgrader,
	eksdInstaller interfaces.EksdInstaller,
) *UpgradeManagementComponentsWorkflow {
	return &UpgradeManagementComponentsWorkflow{
		clientFactory:  clientFactory,
		provider:       provider,
		clusterManager: clusterManager,
		gitOpsManager:  gitOpsManager,
		writer:         writer,
		capiManager:    capiManager,
		eksdUpgrader:   eksdUpgrader,
		eksdInstaller:  eksdInstaller,
	}
}

// UMCValidator is a struct that holds a cluster and a kubectl executable.
// It is used to perform preflight validations on the cluster.
type UMCValidator struct {
	cluster *types.Cluster
	kubectl *executables.Kubectl
}

// NewUMCValidator is a constructor function that creates a new instance of UMCValidator.
func NewUMCValidator(cluster *types.Cluster, kubectl *executables.Kubectl) *UMCValidator {
	return &UMCValidator{
		cluster: cluster,
		kubectl: kubectl,
	}
}

// PreflightValidations is a method of the UMCValidator struct.
// It performs preflight validations on the cluster and returns a slice of Validation objects.
func (u *UMCValidator) PreflightValidations(ctx context.Context) []validations.Validation {
	return []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "control plane ready",
				Remediation: fmt.Sprintf("ensure control plane nodes and pods for cluster %s are Ready", u.cluster.Name),
				Err:         u.kubectl.ValidateControlPlaneNodes(ctx, u.cluster, u.cluster.Name),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "cluster CRDs ready",
				Remediation: "",
				Err:         u.kubectl.ValidateClustersCRD(ctx, u.cluster),
			}
		},
	}
}

// Run Upgrade implements upgrade functionality for management cluster's upgrade operation.
func (umc *UpgradeManagementComponentsWorkflow) Run(ctx context.Context, clusterSpec *cluster.Spec, managementCluster *types.Cluster, validator interfaces.Validator) error {
	commandContext := &task.CommandContext{
		ClientFactory:     umc.clientFactory,
		Provider:          umc.provider,
		ClusterManager:    umc.clusterManager,
		ManagementCluster: managementCluster,
		ClusterSpec:       clusterSpec,
		Validations:       validator,
		Writer:            umc.writer,
		CAPIManager:       umc.capiManager,
		UpgradeChangeDiff: types.NewChangeDiff(),
		GitOpsManager:     umc.gitOpsManager,
		EksdUpgrader:      umc.eksdUpgrader,
		EksdInstaller:     umc.eksdInstaller,
	}

	return task.NewTaskRunner(&setupAndValidateMC{}, umc.writer).RunTask(ctx, commandContext)
}

type setupAndValidateMC struct{}

// Run setupAndValidate validates management cluster before upgrade process starts.
func (s *setupAndValidateMC) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Performing setup and validations")
	currentSpec, err := commandContext.ClusterManager.GetCurrentClusterSpec(ctx, commandContext.ManagementCluster, commandContext.ClusterSpec.Cluster.Name)
	if err != nil {
		commandContext.SetError(err)
		return nil
	}
	commandContext.CurrentClusterSpec = currentSpec
	runner := validations.NewRunner()
	runner.Register(
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name: fmt.Sprintf("%s provider setup and validation", commandContext.Provider.Name()),
				Err:  commandContext.Provider.SetupAndValidateUpgradeManagementComponents(ctx, commandContext.ClusterSpec),
			}
		},
	)
	runner.Register(commandContext.Validations.PreflightValidations(ctx)...)

	err = runner.Run()
	if err != nil {
		commandContext.SetError(err)
		return nil
	}

	return &upgradeCoreComponentsMC{
		UpgradeChangeDiff: &types.ChangeDiff{},
	}
}

func (s *setupAndValidateMC) Name() string {
	return "validate"
}

func (s *setupAndValidateMC) Restore(_ context.Context, _ *task.CommandContext, _ *task.CompletedTask) (task.Task, error) {
	return nil, nil
}

func (s *setupAndValidateMC) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

// This struct is similar to upgradeCoreComponents, but its returned value is different in Run() function.
type upgradeCoreComponentsMC struct {
	UpgradeChangeDiff *types.ChangeDiff
}

func (s *upgradeCoreComponentsMC) Name() string {
	return "upgrade-core-components-mc"
}

func (s *upgradeCoreComponentsMC) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: s.UpgradeChangeDiff,
	}
}

func (s *upgradeCoreComponentsMC) Restore(_ context.Context, commandContext *task.CommandContext, completedTask *task.CompletedTask) (task.Task, error) {
	s.UpgradeChangeDiff = &types.ChangeDiff{}
	if err := task.UnmarshalTaskCheckpoint(completedTask.Checkpoint, s.UpgradeChangeDiff); err != nil {
		return nil, err
	}
	commandContext.UpgradeChangeDiff = s.UpgradeChangeDiff
	return &installNewComponentsMC{}, nil
}

func (s *upgradeCoreComponentsMC) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if err := runUpgradeCoreComponents(ctx, commandContext); err != nil {
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}
	return &installNewComponentsMC{}
}

// This struct is similar to installNewComponents, but its returned value is different in Run() function.
type installNewComponentsMC struct{}

func (s *installNewComponentsMC) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if err := runInstallNewComponents(ctx, commandContext); err != nil {
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}

	if commandContext.OriginalError == nil {
		logger.MarkSuccess("Management components upgraded!")
	}

	return nil
}

func (s *installNewComponentsMC) Name() string {
	return "install-new-eksa-version-components-mc"
}

func (s *installNewComponentsMC) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *installNewComponentsMC) Restore(_ context.Context, _ *task.CommandContext, _ *task.CompletedTask) (task.Task, error) {
	return nil, nil
}
