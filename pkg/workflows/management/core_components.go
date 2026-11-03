package management

import (
	"context"
	"fmt"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/task"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflows"
)

type ensureEtcdCAPIComponentsExist struct{}

// Run ensureEtcdCAPIComponentsExist ensures ETCD CAPI providers on the management cluster.
func (s *ensureEtcdCAPIComponentsExist) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	logger.Info("Ensuring etcd CAPI providers exist on management cluster before upgrade")
	managementComponents := cluster.ManagementComponentsFromBundles(commandContext.CurrentClusterSpec.Bundles)
	if err := commandContext.CAPIManager.EnsureEtcdProvidersInstallation(ctx, commandContext.ManagementCluster, commandContext.Provider, managementComponents, commandContext.CurrentClusterSpec); err != nil {
		commandContext.SetError(err)
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}
	return &pauseGitOpsReconcile{}
}

func (s *ensureEtcdCAPIComponentsExist) Name() string {
	return "ensure-etcd-capi-components-exist"
}

func (s *ensureEtcdCAPIComponentsExist) Checkpoint() *task.CompletedTask {
	return &task.CompletedTask{
		Checkpoint: nil,
	}
}

func (s *ensureEtcdCAPIComponentsExist) Restore(_ context.Context, _ *task.CommandContext, _ *task.CompletedTask) (task.Task, error) {
	return &pauseGitOpsReconcile{}, nil
}

type upgradeCoreComponents struct {
	UpgradeChangeDiff *types.ChangeDiff
}

func runUpgradeCoreComponents(ctx context.Context, commandContext *task.CommandContext) error {
	logger.Info("Upgrading core components")

	newManagementComponents := cluster.ManagementComponentsFromBundles(commandContext.ClusterSpec.Bundles)

	err := commandContext.Provider.PreCoreComponentsUpgrade(
		ctx,
		commandContext.ManagementCluster,
		newManagementComponents,
		commandContext.ClusterSpec,
	)
	if err != nil {
		commandContext.SetError(err)
		return err
	}

	client, err := commandContext.ClientFactory.BuildClientFromKubeconfig(commandContext.ManagementCluster.KubeconfigFile)
	if err != nil {
		commandContext.SetError(err)
		return err
	}

	currentManagementComponents, err := cluster.GetManagementComponents(ctx, client, commandContext.CurrentClusterSpec.Cluster)
	if err != nil {
		commandContext.SetError(err)
		return err
	}

	changeDiff, err := commandContext.CAPIManager.Upgrade(ctx, commandContext.ManagementCluster, commandContext.Provider, currentManagementComponents, newManagementComponents, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return err
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	if err = commandContext.GitOpsManager.Install(ctx, commandContext.ManagementCluster, newManagementComponents, commandContext.CurrentClusterSpec, commandContext.ClusterSpec); err != nil {
		commandContext.SetError(err)
		return err
	}

	changeDiff, err = commandContext.GitOpsManager.Upgrade(ctx, commandContext.ManagementCluster, currentManagementComponents, newManagementComponents, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return err
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	changeDiff, err = commandContext.ClusterManager.Upgrade(ctx, commandContext.ManagementCluster, currentManagementComponents, newManagementComponents, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return err
	}
	commandContext.UpgradeChangeDiff.Append(changeDiff)

	err = commandContext.EksdUpgrader.Upgrade(ctx, commandContext.ManagementCluster, commandContext.CurrentClusterSpec, commandContext.ClusterSpec)
	if err != nil {
		commandContext.SetError(err)
		return err
	}

	logger.V(5).Info("Applying cluster management components version update")
	eksaCluster := &anywherev1.Cluster{}
	err = client.Get(ctx, commandContext.CurrentClusterSpec.Cluster.Name, commandContext.CurrentClusterSpec.Cluster.Namespace, eksaCluster)
	if err != nil {
		commandContext.SetError(err)
		return err
	}
	eksaCluster.SetManagementComponentsVersion(commandContext.ClusterSpec.EKSARelease.Spec.Version)
	eksaCluster.ResourceVersion = ""
	if err := retryAdmissionConnection(ctx, time.Second, time.Minute, func(ctx context.Context) error {
		return client.ApplyServerSide(ctx,
			constants.EKSACLIFieldManager,
			eksaCluster,
			kubernetes.ApplyServerSideOptions{ForceOwnership: true},
		)
	}); err != nil {
		commandContext.SetError(err)
		return err
	}

	commandContext.ClusterSpec.Cluster.SetManagementComponentsVersion(commandContext.ClusterSpec.EKSARelease.Spec.Version)

	return nil
}

// A Ready controller can precede webhook Service endpoint propagation. Retry
// the idempotent version update only for admission connection failures.
func retryAdmissionConnection(ctx context.Context, interval, timeout time.Duration, apply func(context.Context) error) error {
	var lastError error
	err := wait.PollUntilContextTimeout(ctx, interval, timeout, true, func(ctx context.Context) (bool, error) {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		lastError = apply(ctx)
		if lastError == nil {
			return true, nil
		}
		if transientAdmissionConnection(lastError) {
			return false, nil
		}
		return false, lastError
	})
	if err != nil && lastError != nil && ctx.Err() == nil && wait.Interrupted(err) {
		return fmt.Errorf("waiting for admission webhook connection: %w: %v", err, lastError)
	}
	return err
}

func transientAdmissionConnection(err error) bool {
	if apierrors.IsForbidden(err) || apierrors.IsInvalid(err) || apierrors.IsUnauthorized(err) {
		return false
	}
	message := err.Error()
	if strings.Contains(message, "x509:") || strings.Contains(message, "tls:") {
		return false
	}
	return strings.Contains(message, "failed calling webhook") &&
		(strings.Contains(message, "connect: connection refused") ||
			strings.Contains(message, "no endpoints available for service"))
}

// Run upgradeCoreComponents upgrades pre cluster upgrade components.
func (s *upgradeCoreComponents) Run(ctx context.Context, commandContext *task.CommandContext) task.Task {
	if err := runUpgradeCoreComponents(ctx, commandContext); err != nil {
		return &workflows.CollectMgmtClusterDiagnosticsTask{}
	}
	s.UpgradeChangeDiff = commandContext.UpgradeChangeDiff

	return &preClusterUpgrade{}
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
	return &preClusterUpgrade{}, nil
}
